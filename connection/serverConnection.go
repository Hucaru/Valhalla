package connection

import (
	"fmt"
	"net"

	"github.com/Hucaru/Valhalla/maplepacket"
)

// TODO: Add a crypt decrypt step to all packets sent/recv

type Message struct {
	Reader     maplepacket.Reader
	ReturnChan chan maplepacket.Packet
}

func NewMessage(p maplepacket.Packet, ch chan maplepacket.Packet) Message {
	if ch == nil {
		ch = make(chan maplepacket.Packet)
	}

	return Message{Reader: maplepacket.NewReader(&p), ReturnChan: ch}
}

type ServerConnection struct {
	conn net.Conn
}

func NewServerConnection(conn net.Conn) *ServerConnection {
	server := &ServerConnection{conn: conn}
	return server
}

func (handle ServerConnection) String() string {
	return fmt.Sprintf("[Server Address] %s", handle.conn.RemoteAddr())
}

func (handle *ServerConnection) Close() {
	handle.conn.Close()
}

func (handle *ServerConnection) Write(p maplepacket.Packet) error {
	header := maplepacket.NewPacket()
	header.WriteInt32(int32(len(p)))
	header.Append(p)

	_, err := handle.conn.Write(header)
	return err
}

func (handle *ServerConnection) Read(p maplepacket.Packet) error {
	_, err := handle.conn.Read(p)
	return err
}
