package handlers

import (
	"log"

	"github.com/Hucaru/Valhalla/connection"
)

type clientLoginConn struct {
	connection.ClientConnection
	userID    uint32
	isLogedIn bool
	isAdmin   bool
	hash      string
	worldID   uint32
	chanID    byte
	gender    byte
}

// NewLoginConnection -
func NewLoginConnection(conn connection.ClientConnection) *clientLoginConn {
	loginConn := &clientLoginConn{ClientConnection: conn, isAdmin: false}
	return loginConn
}

func (c *clientLoginConn) Close() {
	if c.isLogedIn {
		records, err := connection.Db.Query("UPDATE users set isLogedIn=0 WHERE userID=?", c.userID)

		defer records.Close()

		if err != nil {
			log.Println("Error in auto log out of user on disconnect, userID:", c.userID)
		}
	}

	c.Conn.Close()
}

func (c *clientLoginConn) String() string {
	return c.ClientConnection.String()
}

func (c *clientLoginConn) SetUserID(val uint32) {
	c.userID = val
}

func (c *clientLoginConn) GetUserID() uint32 {
	return c.userID
}

func (c *clientLoginConn) SetAdmin(val bool) {
	c.isAdmin = val
}

func (c *clientLoginConn) IsAdmin() bool {
	return c.isAdmin
}

func (c *clientLoginConn) SetSessionHash(val string) {
	c.hash = val
}

func (c *clientLoginConn) GetSessionHash() string {
	return c.hash
}

func (c *clientLoginConn) SetIsLogedIn(val bool) {
	c.isLogedIn = val
}

func (c *clientLoginConn) GetIsLogedIn() bool {
	return c.isLogedIn
}

func (c *clientLoginConn) SetWorldID(val uint32) {
	c.worldID = val
}

func (c *clientLoginConn) GetWorldID() uint32 {
	return c.worldID
}

func (c *clientLoginConn) SetChanID(val byte) {
	c.chanID = val
}

func (c *clientLoginConn) GetChanID() byte {
	return c.chanID
}

func (c *clientLoginConn) SetGender(val byte) {
	c.gender = val
}

func (c *clientLoginConn) GetGender() byte {
	return c.gender
}
