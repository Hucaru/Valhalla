package game

import (
	"log"
	"time"

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type World struct {
	info  world
	login mnet.Server
}

// RegisterWithLogin server
func (server *World) RegisterWithLogin(conn mnet.Server, message string, ribbon byte) {
	server.info.Message = message
	server.info.Ribbon = ribbon

	server.login = conn
	server.login.Send(mpacket.CreateInternal(opcode.WorldNew))
}

// HandleServerPacket from servers
func (server *World) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.WorldRequestOk:
		server.info.Name = reader.ReadString(int(reader.ReadInt16()))
		log.Println("Registered as", server.info.Name, "with login server at", conn)
		server.login.Send(server.info.generateInfoPacket())
	case opcode.WorldRequestBad:
		log.Println("Rejected by login server at", conn)
		timer := time.NewTimer(30 * time.Second)
		<-timer.C
		server.login.Send(mpacket.CreateInternal(opcode.WorldNew))
	default:
		log.Println("UNKNOWN SERVER PACKET:", reader)
	}
}
