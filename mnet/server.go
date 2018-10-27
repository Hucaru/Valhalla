package mnet

import (
	"net"

	"github.com/Hucaru/Valhalla/consts"
	"github.com/Hucaru/Valhalla/maplepacket"
)

type MConnServer interface {
	MConn
}

type server struct {
	baseConn
}

func NewServer(conn net.Conn, eRecv chan *Event, queueSize int) *server {
	s := &server{}
	s.Conn = conn

	s.eSend = make(chan maplepacket.Packet, queueSize)
	s.eRecv = eRecv

	s.reader = func() {
		serverReader(s, s.eRecv, consts.ClientHeaderSize)
	}

	return s
}
