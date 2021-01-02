package mnet

import (
	"net"

	"github.com/Hucaru/Valhalla/common/mpacket"
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
	Type   int
	Packet mpacket.Packet
	Conn   net.Conn
}
