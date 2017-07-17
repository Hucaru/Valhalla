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
	p             Player
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
	return fmt.Sprintf("[Address] %s", handle.conn.RemoteAddr())
}

// Close -
func (handle *ClientConnection) Close() {
	fmt.Println("is loged in?", handle.GetPlayer().GetIsLogedIn())
	if handle.GetPlayer().GetIsLogedIn() {
		_, err := Db.Query("UPDATE users set isLogedIn=0 WHERE userID=?", handle.GetPlayer().GetUserID())

		if err != nil {
			fmt.Println("Error in auto log out of user on disconnect, userID:", handle.GetPlayer().GetUserID())
		}
	}

	handle.conn.Close()
}

func (handle *ClientConnection) sendPacket(p packet.Packet) error {
	_, err := handle.conn.Write(p)
	return err
}

func (handle *ClientConnection) Write(p packet.Packet) error {
	crypt.Encrypt(p)

	header := packet.NewPacket()
	header = crypt.GenerateHeader(len(p), handle.ivSend, constants.MAPLE_VERSION)
	header.Append(p)

	handle.ivSend = crypt.GenerateNewIV(handle.ivSend)

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
		handle.ivRecv = crypt.GenerateNewIV(handle.ivRecv)
		crypt.Decrypt(p)
	}

	return err
}

func sendHandshake(client *ClientConnection) error {
	packet := packet.NewPacket()

	packet.WriteInt16(13)
	packet.WriteInt16(constants.MAPLE_VERSION)
	packet.WriteString("")
	packet.Append(client.ivRecv)
	packet.Append(client.ivSend)
	packet.WriteByte(8)

	err := client.sendPacket(packet)

	return err
}

// GetPlayer -
func (handle *ClientConnection) GetPlayer() *Player {
	return &handle.p
}
