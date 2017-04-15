package connection

import (
	"fmt"
	"net"

	"github.com/Hucaru/Valhalla/common/crypt"
	"github.com/Hucaru/Valhalla/common/packet"
)

// ClientConnection -
type ClientConnection struct {
	conn           net.Conn
	connectionOpen bool
	readingHeader  bool
	ivRecv         []byte
	ivSend         []byte
}

// NewClientConnection -
func NewClientConnection(conn net.Conn) *ClientConnection {
	client := &ClientConnection{conn: conn, readingHeader: true}

	err := sendHandshake(client)

	if err != nil {
		client.connectionOpen = false
	}

	client.connectionOpen = true

	return &ClientConnection{conn: conn}
}

// String -
func (handle ClientConnection) String() string {
	return fmt.Sprintf("[Address]::%s", handle.conn.RemoteAddr())
}

// IsOpen -
func (handle *ClientConnection) IsOpen() bool {
	return handle.connectionOpen
}

// Close -
func (handle *ClientConnection) Close() {
	handle.conn.Close()
}

func (handle *ClientConnection) sendPacket(p packet.Packet) error {
	_, err := handle.conn.Write(p)
	return err
}

func (handle *ClientConnection) Write(p packet.Packet) error {
	// Do crypto

	fmt.Println("Server -> Client::", p)
	_, err := handle.conn.Write(p)

	return err
}

func (handle *ClientConnection) Read(p packet.Packet) error {

	_, err := handle.conn.Read(p)

	if err != nil {
		return err
	}
	if handle.readingHeader == true {
		handle.readingHeader = false
		crypt.Decrypt(p)
	} else {
		handle.readingHeader = true
	}

	fmt.Println("Client -> Server::", p)

	return err
}

func sendHandshake(client *ClientConnection) error {
	packet := packet.NewPacket(0)

	client.ivRecv = []byte{0, 0, 0, 1} // Change to random init
	client.ivSend = []byte{0, 0, 0, 2} // Change to random init

	packet.WriteShort(13)
	packet.WriteShort(28)
	packet.WriteString("")
	packet.Append(client.ivRecv)
	packet.Append(client.ivSend)
	packet.WriteByte(8)

	err := client.sendPacket(packet)

	return err
}
