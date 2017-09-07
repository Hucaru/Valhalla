package loginConn

import (
	"fmt"
	"net"

	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/packet"
	"github.com/Hucaru/Valhalla/loginServer/handlers/worlds"
)

type Connection struct {
	conn      *connection.ClientConnection
	userID    uint32
	isLogedIn bool
	hash      string
	WorldMngr chan worlds.Message
	worldID   uint32
	gender    byte
}

func NewConnection(conn net.Conn) *Connection {
	loginConn := &Connection{conn: connection.NewClientConnection(conn)}
	return loginConn
}

func (c *Connection) Write(p packet.Packet) error {
	return c.conn.Write(p)
}

func (c *Connection) Read(p packet.Packet) error {
	return c.conn.Read(p)
}

func (c *Connection) Close() {
	if c.isLogedIn {
		_, err := connection.Db.Query("UPDATE users set isLogedIn=0 WHERE userID=?", c.userID)

		if err != nil {
			fmt.Println("Error in auto log out of user on disconnect, userID:", c.userID)
		}
	}
	c.WorldMngr <- worlds.Message{Opcode: worlds.CLIENT_NOT_ACTIVE, Message: nil}
	c.conn.Close()
}

func (c *Connection) String() string {
	return c.conn.String()
}

func (c *Connection) SetUserID(val uint32) {
	c.userID = val
}

func (c *Connection) GetUserID() uint32 {
	return c.userID
}

func (c *Connection) SetSessionHash(val string) {
	c.hash = val
}

func (c *Connection) GetSessionHash() string {
	return c.hash
}

func (c *Connection) SetIsLogedIn(val bool) {
	c.isLogedIn = val
}

func (c *Connection) GetIsLogedIn() bool {
	return c.isLogedIn
}

func (c *Connection) SetWorldID(val uint32) {
	c.worldID = val
}

func (c *Connection) GetWorldID() uint32 {
	return c.worldID
}

func (c *Connection) SetGender(val byte) {
	c.gender = val
}

func (c *Connection) GetGender() byte {
	return c.gender
}
