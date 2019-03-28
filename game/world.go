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
	default:
		log.Println("UNKNOWN SERVER PACKET:", reader)
	}
}

func (server *World) handleRequestOk(conn mnet.Server, reader mpacket.Reader) {
	server.info.Name = reader.ReadString(int(reader.ReadInt16()))
	log.Println("Registered as", server.info.Name, "with login server at", conn)
	server.login.Send(server.info.GenerateInfoPacket())
}

func (server *World) handleRequestBad(conn mnet.Server, reader mpacket.Reader) {
	log.Println("Rejected by login server at", conn)
	timer := time.NewTimer(30 * time.Second)

	<-timer.C

	p := mpacket.CreateInternal(opcode.WorldNew)
	p.WriteString(server.info.Name)
	server.login.Send(p)
}
