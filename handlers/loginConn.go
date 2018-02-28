package handlers

import (
	"log"

	"github.com/Hucaru/Valhalla/connection"
)

type clientConn struct {
	connection.ClientConnection
	userID    uint32
	isLogedIn bool
	isAdmin   bool
	hash      string
	worldID   uint32
	chanID    byte
	gender    byte
}

func NewConnection(conn connection.ClientConnection) *clientConn {
	loginConn := &clientConn{ClientConnection: conn, isAdmin: false}
	return loginConn
}

func (c *clientConn) Close() {
	if c.isLogedIn {
		records, err := connection.Db.Query("UPDATE users set isLogedIn=0 WHERE userID=?", c.userID)

		defer records.Close()

		if err != nil {
			log.Println("Error in auto log out of user on disconnect, userID:", c.userID)
		}
	}

	c.Conn.Close()
}

func (c *clientConn) String() string {
	return c.ClientConnection.String()
}

func (c *clientConn) SetUserID(val uint32) {
	c.userID = val
}

func (c *clientConn) GetUserID() uint32 {
	return c.userID
}

func (c *clientConn) SetAdmin(val bool) {
	c.isAdmin = val
}

func (c *clientConn) IsAdmin() bool {
	return c.isAdmin
}

func (c *clientConn) SetSessionHash(val string) {
	c.hash = val
}

func (c *clientConn) GetSessionHash() string {
	return c.hash
}

func (c *clientConn) SetIsLogedIn(val bool) {
	c.isLogedIn = val
}

func (c *clientConn) GetIsLogedIn() bool {
	return c.isLogedIn
}

func (c *clientConn) SetWorldID(val uint32) {
	c.worldID = val
}

func (c *clientConn) GetChanID() byte {
	return c.chanID
}

func (c *clientConn) SetChanID(val byte) {
	c.chanID = val
}

func (c *clientConn) GetWorldID() uint32 {
	return c.worldID
}

func (c *clientConn) SetGender(val byte) {
	c.gender = val
}

func (c *clientConn) GetGender() byte {
	return c.gender
}
