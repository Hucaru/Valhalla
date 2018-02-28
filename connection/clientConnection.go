package connection

import (
	"crypto/rand"
	"fmt"
	"net"

	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/crypt"
	"github.com/Hucaru/gopacket"
)

// ClientConnection -
type ClientConnection struct {
	net.Conn
	readingHeader bool
	ivRecv        []byte
	ivSend        []byte
}

// NewClientConnection -
func NewClientConnection(conn net.Conn) ClientConnection {
	client := ClientConnection{Conn: conn, readingHeader: true, ivSend: make([]byte, 4), ivRecv: make([]byte, 4)}

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
	return fmt.Sprintf("[Address] %s", handle.Conn.RemoteAddr())
}

// Close -
func (handle *ClientConnection) Close() error {
	return handle.Conn.Close()
}

func (handle *ClientConnection) sendPacket(p gopacket.Packet) error {
	_, err := handle.Conn.Write(p)
	return err
}

func (handle *ClientConnection) Write(p gopacket.Packet) error {
	encryptedPacket := append([]byte{}, p...)

	crypt.Encrypt(encryptedPacket)

	header := gopacket.NewPacket()
	header = crypt.GenerateHeader(len(encryptedPacket), handle.ivSend, constants.MAPLE_VERSION)
	header.Append(encryptedPacket)

	handle.ivSend = crypt.GenerateNewIV(handle.ivSend)

	_, err := handle.Conn.Write(header)

	return err
}

func (handle *ClientConnection) Read(p gopacket.Packet) error {
	_, err := handle.Conn.Read(p)

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

	return err
}

func (handle *ClientConnection) GetClientIPPort() net.Addr {
	return handle.Conn.RemoteAddr()
}

func sendHandshake(client ClientConnection) error {
	packet := gopacket.NewPacket()

	packet.WriteInt16(13)
	packet.WriteInt16(constants.MAPLE_VERSION)
	packet.WriteString("")
	packet.Append(client.ivRecv)
	packet.Append(client.ivSend)
	packet.WriteByte(8)

	err := client.sendPacket(packet)

	return err
}
