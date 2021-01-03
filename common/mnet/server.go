package mnet

import (
	"net"

	"github.com/Hucaru/Valhalla/common/mpacket"
)

type Server interface {
	MConn
}

type server struct {
	baseConn
}

func NewServer(conn net.Conn, eRecv chan *Event, queueSize int) *server {
	s := &server{}
	s.Conn = conn

	s.eSend = make(chan mpacket.Packet, queueSize)
	s.eRecv = eRecv

	s.reader = func() {
		serverReader(s, s.eRecv, 2)
	}

	s.interServer = true

	return s
}
