package game

import (
	"log"
	"time"

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/game/entity"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type World struct {
	info  entity.World
	login mnet.Server
}

// RegisterWithLogin server
func (server *World) RegisterWithLogin(conn mnet.Server, message string, ribbon byte) {
	server.info.Message = message
	server.info.Ribbon = ribbon

	server.login = conn
	server.registerWithLogin()
}

func (server *World) registerWithLogin() {
	p := mpacket.CreateInternal(opcode.WorldNew)
	p.WriteString(server.info.Name)
	server.login.Send(p)
}

// HandleServerPacket from servers
func (server *World) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.WorldRequestOk:
		server.handleRequestOk(conn, reader)
	case opcode.WorldRequestBad:
		server.handleRequestBad(conn, reader)
	case opcode.ChannelNew:
		server.handleNewChannel(conn, reader)
	case opcode.ChannelConnectionInfo:
		server.handleGetChannelConnectionInfo(conn, reader)
	default:
		log.Println("UNKNOWN SERVER PACKET:", reader)
	}
}

func (server *World) ServerDisconnected(conn mnet.Server) {
	for i, v := range server.info.Channels {
		if v.Conn == conn {
			server.info.Channels[i].Conn = nil
			server.info.Channels[i].MaxPop = 0
			server.info.Channels[i].Pop = 0
			log.Println("Lost channel", i)
			break
		}
	}

	server.login.Send(server.info.GenerateInfoPacket())
}

func (server *World) handleRequestOk(conn mnet.Server, reader mpacket.Reader) {
	server.info.Name = reader.ReadString(reader.ReadInt16())
	log.Println("Registered as", server.info.Name, "with login server at", conn)
	server.login.Send(server.info.GenerateInfoPacket())
}

func (server *World) handleRequestBad(conn mnet.Server, reader mpacket.Reader) {
	log.Println("Rejected by login server at", conn)
	timer := time.NewTimer(30 * time.Second)

	<-timer.C

	server.registerWithLogin()
}

func (server *World) handleNewChannel(conn mnet.Server, reader mpacket.Reader) {
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
			return
		}
	}

	newChannel := entity.Channel{Conn: conn, IP: ip, Port: port, MaxPop: maxPop, Pop: 0}
	server.info.Channels = append(server.info.Channels, newChannel)

	p := mpacket.CreateInternal(opcode.ChannelOk)
	p.WriteByte(byte(len(server.info.Channels) - 1))
	conn.Send(p)
	server.login.Send(server.info.GenerateInfoPacket())

	log.Println("Registered channel", len(server.info.Channels)-1)
}

func (server *World) handleGetChannelConnectionInfo(conn mnet.Server, reader mpacket.Reader) {
	id := reader.ReadByte()

	found := false

	for i, v := range server.info.Channels {
		if i == int(id) && v.Conn != nil {
			found = true
			break
		}
	}

	p := mpacket.CreateInternal(opcode.ChannelConnectionInfo)
	p.WriteBool(found)
	p.WriteByte(id)

	if found {
		p.WriteBytes(server.info.Channels[id].IP)
		p.WriteInt16(server.info.Channels[id].Port)
	}

	conn.Send(p)
}
