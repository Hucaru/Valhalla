package connection

import (
	"log"
)

type ClientLoginConn struct {
	ClientConnection
	userID    uint32
	isLogedIn bool
	isAdmin   bool
	hash      string
	worldID   uint32
	chanID    byte
	gender    byte
}

// NewLoginConnection -
func NewLoginConnection(conn ClientConnection) *ClientLoginConn {
	loginConn := &ClientLoginConn{ClientConnection: conn, isAdmin: false}
	return loginConn
}

func (c *ClientLoginConn) Close() {
	if c.isLogedIn {
		records, err := Db.Query("UPDATE users set isLogedIn=0 WHERE userID=?", c.userID)

		defer records.Close()

		if err != nil {
			log.Println("Error in auto log out of user on disconnect, userID:", c.userID)
		}
	}

	c.Conn.Close()
}

func (c *ClientLoginConn) String() string {
	return c.ClientConnection.String()
}

func (c *ClientLoginConn) SetUserID(val uint32) {
	c.userID = val
}

func (c *ClientLoginConn) GetUserID() uint32 {
	return c.userID
}

func (c *ClientLoginConn) SetAdmin(val bool) {
	c.isAdmin = val
}

func (c *ClientLoginConn) IsAdmin() bool {
	return c.isAdmin
}

func (c *ClientLoginConn) SetSessionHash(val string) {
	c.hash = val
}

func (c *ClientLoginConn) GetSessionHash() string {
	return c.hash
}

func (c *ClientLoginConn) SetIsLogedIn(val bool) {
	c.isLogedIn = val
}

func (c *ClientLoginConn) GetIsLogedIn() bool {
	return c.isLogedIn
}

func (c *ClientLoginConn) SetWorldID(val uint32) {
	c.worldID = val
}

func (c *ClientLoginConn) GetWorldID() uint32 {
	return c.worldID
}

func (c *ClientLoginConn) SetChanID(val byte) {
	c.chanID = val
}

func (c *ClientLoginConn) GetChanID() byte {
	return c.chanID
}

func (c *ClientLoginConn) SetGender(val byte) {
	c.gender = val
}

func (c *ClientLoginConn) GetGender() byte {
	return c.gender
}
