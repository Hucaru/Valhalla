package common

import "net"

type ClientConnection struct {
	conn net.Conn
}

func NewClientConnection(conn net.Conn) *ClientConnection {
	return &ClientConnection{conn: conn}
}
