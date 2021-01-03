package world

import (
	"log"
	"math"
	"time"

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
	default:
		log.Println("UNKNOWN SERVER PACKET:", reader)
	}
}

func (server *Server) handleRequestOk(conn mnet.Server, reader mpacket.Reader) {
	server.info.Name = reader.ReadString(reader.ReadInt16())
	log.Println("Registered as", server.info.Name, "with login server at", conn)
	server.login.Send(server.info.GenerateInfoPacket())
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

	if len(server.info.Channels) > 19 {
		p := mpacket.CreateInternal(opcode.ChannelBad)
		conn.Send(p)
		return
	}

	// check to see if we have lost any channels
	for i, v := range server.info.Channels {
		if v.Conn == nil {
			server.info.Channels[i].Conn = conn
			server.info.Channels[i].IP = ip
			server.info.Channels[i].Port = port
			server.info.Channels[i].MaxPop = maxPop

			p := mpacket.CreateInternal(opcode.ChannelOk)
			p.WriteByte(byte(i))
			conn.Send(p)
			server.login.Send(server.info.GenerateInfoPacket())

			log.Println("Re-registered channel", i)
			server.sendChannelInfo()
			return
		}
	}

	// TODO highest value party id and set the to current party id if it is larger

	newChannel := internal.Channel{Conn: conn, IP: ip, Port: port, MaxPop: maxPop, Pop: 0}
	server.info.Channels = append(server.info.Channels, newChannel)

	p := mpacket.CreateInternal(opcode.ChannelOk)
	p.WriteString(server.info.Name)
	p.WriteByte(byte(len(server.info.Channels) - 1))
	conn.Send(p)
	server.login.Send(server.info.GenerateInfoPacket())

	log.Println("Registered channel", len(server.info.Channels)-1)
	server.sendChannelInfo()
}

func (server Server) sendChannelInfo() {
	p := mpacket.CreateInternal(opcode.ChannelConnectionInfo)
	p.WriteByte(byte(len(server.info.Channels)))

	for _, v := range server.info.Channels {
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
		server.info.Channels[id].Pop = reader.ReadInt16()
	default:
		log.Println("Unkown channel update type", op)
	}
	server.login.Send(server.info.GenerateInfoPacket())
}

func (server *Server) handlePlayerConnect(conn mnet.Server, reader mpacket.Reader) {
	server.forwardPacketToChannels(conn, reader)

	playerID := reader.ReadInt32()
	_ = reader.ReadString(reader.ReadInt16()) // name
	channelID := reader.ReadByte()            // this needs to be -1 to show player in cash shop
	_ = reader.ReadBool()                     // channelChange
	mapID := reader.ReadInt32()

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

	// search db for player is in guild, if in guld send a guild connected message
	// channel server will need to display guild to players
}

func (server *Server) handlePlayerDisconnect(conn mnet.Server, reader mpacket.Reader) {
	server.forwardPacketToChannels(conn, reader)

	playerID := reader.ReadInt32()

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

	for _, guild := range server.guilds {
		end := false
		for i, v := range guild.PlayerID {
			if v == playerID {
				guild.Online[i] = false
				// send guild connection message
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
	case 0: // new party request
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
			server.info.Channels[channelID].Conn.Send(internal.PacketChannelPartyCreateApproved(playerID, false, server.parties[partyID]))
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

			server.channelBroadcast(internal.PacketChannelPartyCreateApproved(playerID, true, server.parties[partyID]))
		}

	case 2: // leave party
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
	case 3: // accept invite
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
	case 4: // update party info
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
		log.Println("Unkown party event type:", op)
	}
}

func (server *Server) handleGuildEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 0: // new guild
	default:
		log.Println("Unkown guild event type:", op)
	}
}
