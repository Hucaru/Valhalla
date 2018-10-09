package mnet

import (
	"github.com/Hucaru/Valhalla/maplepacket"
)

const (
	MEClientConnected = iota
	MEClientDisconnect
	MEClientPacket
)

type Event struct {
	Type   int
	Packet maplepacket.Packet
	Conn   MConn
}
