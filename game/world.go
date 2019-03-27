package game

import (
	"log"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type World struct {
	info  world
	login mnet.Server
}

// RegisterWithLogin server
func (server *World) RegisterWithLogin(conn mnet.Server) {
	server.login = conn
	server.login.Send(server.info.generateInfoPacket())
}

// HandleChannelPacket from channel
func (server *World) HandleChannelPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	default:
		log.Println("UNKNOWN CHANNEL PACKET:", reader)
	}
}

// HandleLoginPacket from login
func (server *World) HandleLoginPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	default:
		log.Println("UNKNOWN LOGIN PACKET:", reader)
	}
}
