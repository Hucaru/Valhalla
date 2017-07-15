package connection

import (
	"crypto/rand"
	"fmt"
	"net"

	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/crypt"
	"github.com/Hucaru/Valhalla/common/packet"
)

// ClientConnection -
type ClientConnection struct {
	conn          net.Conn
	readingHeader bool
	ivRecv        []byte
	ivSend        []byte
}

// NewClientConnection -
func NewClientConnection(conn net.Conn) *ClientConnection {
	client := &ClientConnection{conn: conn, readingHeader: true, ivSend: make([]byte, 4), ivRecv: make([]byte, 4)}

	rand.Read(client.ivSend[:])
	rand.Read(client.ivRecv[:])

	err := sendHandshake(client)

	if err != nil {
		client.Close()
	}

	return client
}

// String -
func (handle ClientConnection) String() string {
	return fmt.Sprintf("[Address]::%s", handle.conn.RemoteAddr())
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
	fmt.Println("Server -> Client::", p)

	crypt.Encrypt(p)

	header := packet.NewPacket()
	header = crypt.GenerateHeader(len(p), handle.ivSend, constants.MAPLE_VERSION)
	// handle.ivSend = crypt.GenerateNewIV(handle.ivSend) // Required if AES is in client
	header.Append(p)

	_, err := handle.conn.Write(header)

	return err
}

func (handle *ClientConnection) Read(p packet.Packet) error {

	_, err := handle.conn.Read(p)

	if err != nil {
		return err
	}
	if handle.readingHeader == true {
		handle.readingHeader = false
	} else {
		handle.readingHeader = true
		// handle.ivRecv = crypt.GenerateNewIV(handle.ivRecv) // Required if AES is in client
		crypt.Decrypt(p)
	}

	fmt.Println("Client -> Server::", p)

	return err
}

func sendHandshake(client *ClientConnection) error {
	packet := packet.NewPacket()

	packet.WriteShort(13)
	packet.WriteShort(constants.MAPLE_VERSION)
	packet.WriteString("")
	packet.Append(client.ivRecv)
	packet.Append(client.ivSend)
	packet.WriteByte(8)

	err := client.sendPacket(packet)

	return err
}
