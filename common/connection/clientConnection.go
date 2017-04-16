package connection

import (
	"fmt"
	"net"

	"github.com/Hucaru/Valhalla/common/constants"
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
	client := &ClientConnection{conn: conn, readingHeader: true, ivSend: make([]byte, 4), ivRecv: make([]byte, 4)}

	err := sendHandshake(client)

	if err != nil {
		client.connectionOpen = false
	}

	client.connectionOpen = true

	return client
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
	crypt.Encrypt(p)
	header := packet.NewPacket(constants.CLIENT_HEADER_SIZE)
	header = crypt.GenerateHeader(len(p), handle.ivSend, constants.MAPLE_VERSION)
	handle.ivSend = crypt.GenerateNewIV(handle.ivSend)
	header.Append(p)

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
	} else {
		handle.readingHeader = true
		handle.ivRecv = crypt.GenerateNewIV(handle.ivRecv)
		crypt.Decrypt(p)
	}

	fmt.Println("Client -> Server::", p)

	return err
}

func sendHandshake(client *ClientConnection) error {
	packet := packet.NewPacket(0)

	client.ivRecv = []byte{1, 2, 3, 4}
	client.ivSend = []byte{1, 2, 3, 4}

	// rand.Read(client.ivRecv) - Causes bad header to be returned
	// rand.Read(client.ivSend) - Causes bad header to be returned

	packet.WriteShort(13)
	packet.WriteShort(constants.MAPLE_VERSION)
	packet.WriteString("")
	packet.Append(client.ivRecv)
	packet.Append(client.ivSend)
	packet.WriteByte(8)

	err := client.sendPacket(packet)

	return err
}
