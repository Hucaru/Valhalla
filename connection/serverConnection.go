package connection

import (
	"fmt"
	"net"

	"github.com/Hucaru/gopacket"
)

// TODO: Add a crypt decrypt step to all packets sent/recv

type Message struct {
	Reader     gopacket.Reader
	ReturnChan chan gopacket.Packet
}

func NewMessage(p gopacket.Packet, ch chan gopacket.Packet) Message {
	if ch == nil {
		ch = make(chan gopacket.Packet)
	}

	return Message{Reader: gopacket.NewReader(&p), ReturnChan: ch}
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

func (handle *ServerConnection) Write(p gopacket.Packet) error {
	header := gopacket.NewPacket()
	header.WriteInt32(int32(len(p)))
	header.Append(p)

	_, err := handle.conn.Write(header)
	return err
}

func (handle *ServerConnection) Read(p gopacket.Packet) error {
	_, err := handle.conn.Read(p)
	return err
}
