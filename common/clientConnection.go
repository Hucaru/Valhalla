package common

import (
	"fmt"
	"net"
)

type Connection interface {
	Write(p Packet) error
	Read(p Packet) error
}

type ClientConnection struct {
	conn net.Conn
}

func NewClientConnection(conn net.Conn) *ClientConnection {
	return &ClientConnection{conn: conn}
}

func (handle *ClientConnection) Write(p Packet) error {
	// Do crypto - Chinese Shanda

	// Add length to front of packet - int16
	p.AddSize()

	fmt.Println("Sent::" + p.String())
	_, err := handle.conn.Write(p)

	return err
}

func (handle *ClientConnection) Read(p Packet) error {
	// Do crypto - Chinese Shanda

	_, err := handle.conn.Read(p)

	fmt.Println("Recv::" + p.String())

	return err
}
