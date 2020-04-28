package server

import (
	"log"
	"time"

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// WorldServer data
type WorldServer struct {
	info  world
	login mnet.Server
}

// RegisterWithLogin server
func (server *WorldServer) RegisterWithLogin(conn mnet.Server, message string, ribbon byte) {
	server.info.message = message
	server.info.ribbon = ribbon

	server.login = conn
	server.registerWithLogin()
}

func (server *WorldServer) registerWithLogin() {
	p := mpacket.CreateInternal(opcode.WorldNew)
	p.WriteString(server.info.name)
	server.login.Send(p)
}

// HandleServerPacket from servers
func (server *WorldServer) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.WorldRequestOk:
		server.handleRequestOk(conn, reader)
	case opcode.WorldRequestBad:
		server.handleRequestBad(conn, reader)
	case opcode.ChannelNew:
		server.handleNewChannel(conn, reader)
	default:
		log.Println("UNKNOWN SERVER PACKET:", reader)
	}
}

// ServerDisconnected handler
func (server *WorldServer) ServerDisconnected(conn mnet.Server) {
	for i, v := range server.info.channels {
		if v.conn == conn {
			server.info.channels[i].conn = nil
			server.info.channels[i].maxPop = 0
			server.info.channels[i].pop = 0
			server.info.channels[i].port = 0
			log.Println("Lost channel", i)
			server.sendChannelInfo()
			break
		}
	}

	server.login.Send(server.info.generateInfoPacket())
}

func (server *WorldServer) handleRequestOk(conn mnet.Server, reader mpacket.Reader) {
	server.info.name = reader.ReadString(reader.ReadInt16())
	log.Println("Registered as", server.info.name, "with login server at", conn)
	server.login.Send(server.info.generateInfoPacket())
}

func (server *WorldServer) handleRequestBad(conn mnet.Server, reader mpacket.Reader) {
	log.Println("Rejected by login server at", conn)
	timer := time.NewTimer(30 * time.Second)

	<-timer.C

	server.registerWithLogin()
}

func (server *WorldServer) handleNewChannel(conn mnet.Server, reader mpacket.Reader) {
	log.Println("New channel request")
	ip := reader.ReadBytes(4)
	port := reader.ReadInt16()
	maxPop := reader.ReadInt16()

	if len(server.info.channels) > 19 {
		p := mpacket.CreateInternal(opcode.ChannelBad)
		conn.Send(p)
		return
	}

	// check to see if we have lost any channels
	for i, v := range server.info.channels {
		if v.conn == nil {
			server.info.channels[i].conn = conn
			server.info.channels[i].ip = ip
			server.info.channels[i].port = port
			server.info.channels[i].maxPop = maxPop

			p := mpacket.CreateInternal(opcode.ChannelOk)
			p.WriteByte(byte(i))
			conn.Send(p)
			server.login.Send(server.info.generateInfoPacket())

			log.Println("Re-registered channel", i)
			server.sendChannelInfo()
			return
		}
	}

	newChannel := channel{conn: conn, ip: ip, port: port, maxPop: maxPop, pop: 0}
	server.info.channels = append(server.info.channels, newChannel)

	p := mpacket.CreateInternal(opcode.ChannelOk)
	p.WriteString(server.info.name)
	p.WriteByte(byte(len(server.info.channels) - 1))
	conn.Send(p)
	server.login.Send(server.info.generateInfoPacket())

	log.Println("Registered channel", len(server.info.channels)-1)
	server.sendChannelInfo()
}

func (server *WorldServer) sendChannelInfo() {
	p := mpacket.CreateInternal(opcode.ChannelConnectionInfo)
	p.WriteByte(byte(len(server.info.channels)))

	for _, v := range server.info.channels {
		p.WriteBytes(v.ip)
		p.WriteInt16(v.port)
	}

	for _, v := range server.info.channels {
		if v.conn == nil {
			continue
		}

		v.conn.Send(p)
	}
}
