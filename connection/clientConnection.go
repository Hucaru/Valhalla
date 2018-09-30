package connection

import (
	"crypto/rand"
	"fmt"
	"net"

	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/crypt"
	"github.com/Hucaru/Valhalla/maplepacket"
)

// Client -
type Client struct {
	net.Conn
	readingHeader bool
	ivRecv        []byte
	ivSend        []byte
}

// NewClient -
func NewClient(conn net.Conn) Client {
	client := Client{Conn: conn, readingHeader: true, ivSend: make([]byte, 4), ivRecv: make([]byte, 4)}

	rand.Read(client.ivSend[:])
	rand.Read(client.ivRecv[:])

	err := sendHandshake(client)

	if err != nil {
		client.Close()
	}

	return client
}

// String -
func (handle Client) String() string {
	return fmt.Sprintf("[Address] %s", handle.Conn.RemoteAddr())
}

// Close -
func (handle *Client) Close() error {
	return handle.Conn.Close()
}

func (handle *Client) sendPacket(p maplepacket.Packet) error {
	_, err := handle.Conn.Write(p)
	return err
}

func (handle *Client) Write(p maplepacket.Packet) error {
	crypt.GenerateHeader(p[:4], len(p[4:]), handle.ivSend, constants.MapleVersion)
	crypt.Encrypt(p[4:])

	handle.ivSend = crypt.GenerateNewIV(handle.ivSend)

	_, err := handle.Conn.Write(p)

	return err
}

func (handle *Client) Read(p maplepacket.Packet) error {
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

func (handle *Client) GetClientIPPort() net.Addr {
	return handle.Conn.RemoteAddr()
}

func sendHandshake(client Client) error {
	packet := maplepacket.NewPacket()

	packet.WriteInt16(13)
	packet.WriteInt16(constants.MapleVersion)
	packet.WriteString("")
	packet.Append(client.ivRecv)
	packet.Append(client.ivSend)
	packet.WriteByte(8)

	err := client.sendPacket(packet)

	return err
}
