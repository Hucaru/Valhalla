package mnet

import (
	"net"

	"github.com/Hucaru/Valhalla/mpacket"
)

const (
	MEClientConnected = iota
	MEClientDisconnect
	MEClientPacket
	MEServerConnected
	MEServerDisconnect
	MEServerPacket
)

type Event struct {
	Type     int
	Protocol uint32
	Packet   mpacket.Packet
	Conn     net.Conn
}
