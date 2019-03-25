package game

import (
	"log"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type World struct {
	// list of channel server
	// login server
}

// HandleChannelPacket from channel
func (server *World) HandleChannelPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	default:
		log.Println("Unkown packet:", reader)
	}
}
