package playerConn

import (
	"log"
	"net"

	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/gopacket"
)

type Conn struct {
	conn      *connection.ClientConnection
	userID    uint32
	isLogedIn bool
	isAdmin   bool
	hash      string
	worldID   uint32
	channelID uint32
	character character.Character
}

func NewConnection(conn net.Conn) *Conn {
	channelConn := &Conn{conn: connection.NewClientConnection(conn), isAdmin: false}
	return channelConn
}

func (c *Conn) Write(p gopacket.Packet) error {
	return c.conn.Write(p)
}

func (c *Conn) Read(p gopacket.Packet) error {
	return c.conn.Read(p)
}

func (c *Conn) Close() {
	if c.isLogedIn {
		records, err := connection.Db.Query("UPDATE users set isLogedIn=0 WHERE userID=?", c.userID)

		defer records.Close()

		if err != nil {
			log.Println("Error in auto log out of user on disconnect, userID:", c.userID)
		}
	}

	c.conn.Close()
}

func (c *Conn) String() string {
	return c.conn.String()
}

func (c *Conn) SetUserID(val uint32) {
	c.userID = val
}

func (c *Conn) GetUserID() uint32 {
	return c.userID
}

func (c *Conn) SetAdmin(val bool) {
	c.isAdmin = val
}

func (c *Conn) IsAdmin() bool {
	return c.isAdmin
}

func (c *Conn) SetSessionHash(val string) {
	c.hash = val
}

func (c *Conn) GetSessionHash() string {
	return c.hash
}

func (c *Conn) SetIsLogedIn(val bool) {
	c.isLogedIn = val
}

func (c *Conn) GetIsLogedIn() bool {
	return c.isLogedIn
}

func (c *Conn) SetWorldID(val uint32) {
	c.worldID = val
}

func (c *Conn) GetWorldID() uint32 {
	return c.worldID
}

func (c *Conn) SetChanneldID(val uint32) {
	c.channelID = val
}

func (c *Conn) GetChannelID() uint32 {
	return c.channelID
}

func (c *Conn) GetClientIPPort() net.Addr {
	return c.conn.GetClientIPPort()
}

func (c *Conn) SetCharacter(char character.Character) {
	c.character = char
}

func (c *Conn) GetCharacter() *character.Character {
	return &c.character
}
