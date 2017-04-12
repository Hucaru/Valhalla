package connection

import (
	"fmt"
	"net"

	"github.com/Hucaru/Valhalla/common/packet"
)

type ClientConnection struct {
	conn           net.Conn
	connectionOpen bool
}

func NewClientConnection(conn net.Conn) *ClientConnection {
	client := &ClientConnection{conn: conn}

	err := sendHandshake(client)

	if err != nil {
		client.connectionOpen = false
	}

	client.connectionOpen = true

	return &ClientConnection{conn: conn}
}

func (handle *ClientConnection) IsOpen() bool {
	return handle.connectionOpen
}

func (handle *ClientConnection) Close() {
	handle.conn.Close()
}

func (handle *ClientConnection) Write(p packet.Packet) error {
	// Do crypto - Chinese Shanda

	// Add length to front of packet.Packet - int16
	p.AddSize()

	fmt.Println("Client::Sent::" + p.String())
	_, err := handle.conn.Write(p)

	return err
}

func (handle *ClientConnection) Read(p packet.Packet) error {
	// Do crypto - Chinese Shanda

	_, err := handle.conn.Read(p)

	if err != nil {
		return err
	}

	fmt.Println("Client::Recv::" + p.String())

	return err
}

func sendHandshake(client *ClientConnection) error {
	packet := packet.NewPacket(0)

	packet.WriteShort(28)
	packet.WriteString("")
	packet.WriteInt(1)
	packet.WriteInt(2)
	packet.WriteByte(8)

	err := client.Write(packet)

	return err
}
