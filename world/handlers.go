package world

import (
	"log"
	"math"
	"time"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
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
		fallthrough
	case opcode.ChannePlayerDisconnect:
		fallthrough
	case opcode.ChannelPlayerBuddyEvent:
		fallthrough
	case opcode.ChannelPlayerChatEvent:
		server.forwardPacketToChannels(conn, reader)
	case opcode.ChannelPlayerPartyEvent:
		server.handlePartyEvent(conn, reader)
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

		server.channelBroadcast(internal.PacketChannelPartyCreateApproved(partyID, playerID, channelID, mapID, job, level, name))
	case 1:
		log.Println("World server should not receive a party event message type: 1")
	case 2: // leave party
		if destroy := reader.ReadBool(); destroy {
			partyID := reader.ReadInt32()

			for _, v := range server.reusablePartyIDs {
				if v == partyID {
					return
				}
			}

			server.reusablePartyIDs = append(server.reusablePartyIDs, partyID)
		}

		fallthrough
	case 3: // accept invite
		fallthrough
	case 4: //expel
		fallthrough
	case 5: // update party info
		p := mpacket.NewPacket()
		p.WriteByte(0)
		p.WriteBytes(reader.GetBuffer())
		server.channelBroadcast(p)
	default:
		log.Println("Unkown party event type:", op)
	}
}
