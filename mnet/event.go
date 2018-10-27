package mnet

import (
	"net"

	"github.com/Hucaru/Valhalla/maplepacket"
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
	Packet maplepacket.Packet
	Conn   net.Conn
}
