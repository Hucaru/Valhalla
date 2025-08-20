package world

import (
	"log"
	"math"
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/mnet"
	"github.com/Hucaru/Valhalla/common/mpacket"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/internal"
)

// HandleServerPacket from servers
func (server *Server) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.WorldRequestOk:
		server.handleRequestOk(conn, reader)
	case opcode.WorldRequestBad:
		server.handleRequestBad(conn, reader)
	case opcode.ChannelNew:
		server.handleNewChannel(conn, reader)
	case opcode.ChannelInfo:
		server.handleChannelUpdate(conn, reader)
	case opcode.ChannePlayerConnect:
		server.handlePlayerConnect(conn, reader)
	case opcode.ChannePlayerDisconnect:
		server.handlePlayerDisconnect(conn, reader)
	case opcode.ChannelPlayerBuddyEvent:
		fallthrough
	case opcode.ChannelPlayerChatEvent:
		server.forwardPacketToChannels(conn, reader)
	case opcode.ChannelPlayerPartyEvent:
		server.handlePartyEvent(conn, reader)
	case opcode.ChannelPlayerGuildEvent:
		server.handleGuildEvent(conn, reader)
	case opcode.ChangeRate:
		server.handleChangeRate(conn, reader)
	default:
		log.Println("UNKNOWN SERVER PACKET:", reader)
	}
}

func (server *Server) handleRequestOk(conn mnet.Server, reader mpacket.Reader) {
	server.Info.Name = reader.ReadString(reader.ReadInt16())
	log.Println("Registered as", server.Info.Name, "with login server at", conn)
	server.login.Send(server.Info.GenerateInfoPacket())
}

func (server *Server) handleRequestBad(conn mnet.Server, reader mpacket.Reader) {
	log.Println("Rejected by login server at", conn)
	timer := time.NewTimer(30 * time.Second)

	<-timer.C

	server.registerWithLogin()
}

func (server *Server) handleNewChannel(conn mnet.Server, reader mpacket.Reader) {
	log.Println("New channel request")
	ip := reader.ReadBytes(4)
	port := reader.ReadInt16()
	maxPop := reader.ReadInt16()

	if len(server.Info.Channels) > 19 {
		p := mpacket.CreateInternal(opcode.ChannelBad)
		conn.Send(p)
		return
	}

	pSend := func(id int) {
		p := mpacket.CreateInternal(opcode.ChannelOk)
		p.WriteString(server.Info.Name)
		p.WriteByte(byte(id))
		// Sending the registered channel the world's rates
		p.WriteFloat32(server.Info.Rates.Exp)
		p.WriteFloat32(server.Info.Rates.Drop)
		p.WriteFloat32(server.Info.Rates.Mesos)

		conn.Send(p)
		server.login.Send(server.Info.GenerateInfoPacket())
	}

	// check to see if we have lost any channels
	for i, v := range server.Info.Channels {
		if v.Conn == nil {
			server.Info.Channels[i].Conn = conn
			server.Info.Channels[i].IP = ip
			server.Info.Channels[i].Port = port
			server.Info.Channels[i].MaxPop = maxPop
			pSend(i)

			log.Println("Re-registered channel", i)
			server.sendChannelInfo()
			return
		}
	}

	// TODO highest value party id and set the to current party id if it is larger

	newChannel := internal.Channel{Conn: conn, IP: ip, Port: port, MaxPop: maxPop, Pop: 0}
	server.Info.Channels = append(server.Info.Channels, newChannel)

	pSend(len(server.Info.Channels) - 1)

	log.Println("Registered channel", len(server.Info.Channels)-1)
	server.sendChannelInfo()
}

func (server Server) sendChannelInfo() {
	p := mpacket.CreateInternal(opcode.ChannelConnectionInfo)
	p.WriteByte(byte(len(server.Info.Channels)))

	for _, v := range server.Info.Channels {
		p.WriteBytes(v.IP)
		p.WriteInt16(v.Port)
	}

	server.channelBroadcast(p)
}

func (server *Server) handleChannelUpdate(conn mnet.Server, reader mpacket.Reader) {
	id := reader.ReadByte()
	op := reader.ReadByte()
	switch op {
	case 0: //population
		server.Info.Channels[id].Pop = reader.ReadInt16()
	default:
		log.Println("Unkown channel update type", op)
	}
	server.login.Send(server.Info.GenerateInfoPacket())
}

func (server *Server) handlePlayerConnect(conn mnet.Server, reader mpacket.Reader) {
	server.forwardPacketToChannels(conn, reader)

	playerID := reader.ReadInt32()
	_ = reader.ReadString(reader.ReadInt16()) // name
	channelID := reader.ReadByte()            // this needs to be -1 to show player in cash shop
	_ = reader.ReadBool()                     // channelChange
	mapID := reader.ReadInt32()
	_ = reader.ReadInt32() // guildID

	for _, party := range server.parties {
		end := false
		for i, v := range party.PlayerID {
			if v == playerID {
				party.ChannelID[i] = int32(channelID)
				party.MapID[i] = mapID
				server.channelBroadcast(internal.PacketWorldPartyUpdate(party.ID, party.PlayerID[i], int32(i), true, party))
				end = true
				break
			}
		}

		if end {
			break
		}
	}
}

func (server *Server) handlePlayerDisconnect(conn mnet.Server, reader mpacket.Reader) {
	server.forwardPacketToChannels(conn, reader)

	playerID := reader.ReadInt32()
	_ = reader.ReadString(reader.ReadInt16()) // name
	_ = reader.ReadInt32()                    // guildID

	for _, party := range server.parties {
		end := false
		for i, v := range party.PlayerID {
			if v == playerID {
				party.ChannelID[i] = -2 // means offline

				server.channelBroadcast(internal.PacketWorldPartyUpdate(party.ID, party.PlayerID[i], int32(i), true, party))
				end = true
				break
			}
		}

		if end {
			break
		}
	}
}

func (server *Server) handlePartyEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case internal.OpPartyCreate:
		playerID := reader.ReadInt32()
		channelID := reader.ReadByte()
		mapID := reader.ReadInt32()
		job := reader.ReadInt32()
		level := reader.ReadInt32()
		name := reader.ReadString(reader.ReadInt16())

		var partyID int32
		if len(server.reusablePartyIDs) > 0 {
			partyID = server.reusablePartyIDs[0]
			server.reusablePartyIDs[0] = server.reusablePartyIDs[len(server.reusablePartyIDs)-1]
			server.reusablePartyIDs = server.reusablePartyIDs[:len(server.reusablePartyIDs)-1]
		} else {
			server.nextPartyID++

			if server.nextPartyID == math.MaxInt32 {
				server.nextPartyID = 1
			}

			partyID = server.nextPartyID
		}

		if partyID == 0 {
			server.Info.Channels[channelID].Conn.Send(internal.PacketWorldPartyCreateApproved(playerID, false, server.parties[partyID]))
		} else {
			server.parties[partyID] = &internal.Party{
				ID: partyID,
			}

			server.parties[partyID].ChannelID[0] = int32(channelID)
			server.parties[partyID].PlayerID[0] = playerID
			server.parties[partyID].Name[0] = name
			server.parties[partyID].MapID[0] = mapID
			server.parties[partyID].Job[0] = job
			server.parties[partyID].Level[0] = level

			server.channelBroadcast(internal.PacketWorldPartyCreateApproved(playerID, true, server.parties[partyID]))
		}

	case internal.OpPartyLeaveExpel:
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		kicked := reader.ReadBool()

		if party, ok := server.parties[partyID]; ok {
			for i, v := range party.PlayerID {
				if v == playerID {
					destory := i == 0

					party.ChannelID[i] = 0
					party.PlayerID[i] = 0
					party.Name[i] = ""
					party.MapID[i] = 0
					party.Job[i] = 0
					party.Level[i] = 0

					server.channelBroadcast(internal.PacketWorldPartyLeave(partyID, destory, kicked, int32(i), party))

					server.reusablePartyIDs = append(server.reusablePartyIDs, partyID)
					break
				}
			}
		}
	case internal.OpPartyAccept:
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		channelID := reader.ReadInt32()
		mapID := reader.ReadInt32()
		job := reader.ReadInt32()
		level := reader.ReadInt32()
		name := reader.ReadString(reader.ReadInt16())

		if party, ok := server.parties[partyID]; ok {
			for i, v := range party.PlayerID {
				if v == 0 { // empty slot
					party.PlayerID[i] = playerID
					party.ChannelID[i] = channelID
					party.MapID[i] = mapID
					party.Job[i] = job
					party.Level[i] = level
					party.Name[i] = name

					server.channelBroadcast(internal.PacketWorldPartyAccept(partyID, playerID, int32(i), party))

					break
				}
			}
		}
	case internal.OpPartyInfoUpdate:
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		job := reader.ReadInt32()
		level := reader.ReadInt32()
		mapID := reader.ReadInt32()
		name := reader.ReadString(reader.ReadInt16())

		if party, ok := server.parties[partyID]; ok {
			for i, v := range party.PlayerID {
				if v == playerID {
					party.Job[i] = job
					party.Level[i] = level
					party.Name[i] = name
					party.MapID[i] = mapID

					server.channelBroadcast(internal.PacketWorldPartyUpdate(partyID, playerID, int32(i), false, party))

					break
				}
			}
		}
	default:
		log.Println("Unknown party event type:", op)
	}
}

func (server *Server) handleChangeRate(conn mnet.Server, reader mpacket.Reader) {
	mode := reader.ReadByte()
	rate := reader.ReadFloat32()

	switch mode {
	case 1:
		server.Info.Rates.Exp = rate
	case 2:
		server.Info.Rates.Drop = rate
	case 3:
		server.Info.Rates.Mesos = rate
	}

	if server.Info.Rates.Exp != server.Info.DefaultRates.Exp ||
		server.Info.Rates.Drop != server.Info.DefaultRates.Drop ||
		server.Info.Rates.Mesos != server.Info.DefaultRates.Mesos { // Rates event
		server.Info.Ribbon = 1
		log.Println("GM triggered rates event")
	} else {
		server.Info.Ribbon = 0
	}
	server.login.Send(server.Info.GenerateInfoPacket())

	p := mpacket.CreateInternal(opcode.ChangeRate)
	p.Append(reader.GetBuffer()[1:])

	server.channelBroadcast(p)
}

func (server *Server) handleGuildEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case internal.OpGuildDisband:
		guildID := reader.ReadInt32()

		if _, err := common.DB.Exec("DELETE FROM guilds WHERE (id=?)", guildID); err != nil {
			log.Println(err)
		} else {
			server.forwardPacketToChannels(conn, reader)
		}

	case internal.OpGuildRankUpdate:
		fallthrough
	case internal.OpGuildAddPlayer:
		server.forwardPacketToChannels(conn, reader)
	case internal.OpGuildPointsUpdate:
		guildID := reader.ReadInt32()
		points := reader.ReadInt32()

		query := "UPDATE guilds set points=? WHERE id=?"

		if _, err := common.DB.Exec(query, points, guildID); err != nil {
			log.Println(err)
		} else {
			server.forwardPacketToChannels(conn, reader)
		}
	case internal.OpGuildTitlesChange:
		guildID := reader.ReadInt32()
		master := reader.ReadString(reader.ReadInt16())
		jrMaster := reader.ReadString(reader.ReadInt16())
		member1 := reader.ReadString(reader.ReadInt16())
		member2 := reader.ReadString(reader.ReadInt16())
		member3 := reader.ReadString(reader.ReadInt16())

		query := "UPDATE guilds set master=?, jrMaster=?, member1=?, member2=?, member3=? WHERE id=?"

		if _, err := common.DB.Exec(query, master, jrMaster, member1, member2, member3, guildID); err != nil {
			log.Println(err)
		} else {
			server.forwardPacketToChannels(conn, reader)
		}
	case internal.OpGuildRemovePlayer:
		reader.Skip(4)
		playerID := reader.ReadInt32()

		query := "UPDATE characters set guildID=?, guildRankID=? WHERE id=?"

		if _, err := common.DB.Exec(query, nil, 0, playerID); err != nil {
			log.Println(err)
		} else {
			server.forwardPacketToChannels(conn, reader)
		}
	case internal.OpGuildNoticeChange:
		guildID := reader.ReadInt32()
		notice := reader.ReadString(reader.ReadInt16())

		query := "UPDATE guilds SET notice=? WHERE id=?"

		if _, err := common.DB.Exec(query, notice, guildID); err != nil {
			log.Println(err)
		} else {
			server.forwardPacketToChannels(conn, reader)
		}
	case internal.OpGuildEmblemChange:
		guildID := reader.ReadInt32()
		logoBg := reader.ReadInt16()
		logo := reader.ReadInt16()
		logoBgColour := reader.ReadByte()
		logoColour := reader.ReadByte()

		query := "UPDATE guilds SET logoBg=?,logoBgColour=?,logo=?,logoColour=? WHERE id=?"

		if _, err := common.DB.Exec(query, logoBg, logoBgColour, logo, logoColour, guildID); err != nil {
			log.Println(err)
		} else {
			server.forwardPacketToChannels(conn, reader)
		}
	default:
		log.Println("Unkown guild event type:", op)
	}
}
