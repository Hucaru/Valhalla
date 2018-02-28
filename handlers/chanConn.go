package handlers

import (
	"github.com/Hucaru/Valhalla/connection"
)

type clientChanConn struct {
	connection.ClientConnection
	userID    uint32
	isLogedIn bool
	isAdmin   bool
	worldID   uint32
	chanID    byte
}

func NewChanConnection(conn connection.ClientConnection) *clientChanConn {
	loginConn := &clientChanConn{ClientConnection: conn, isAdmin: false}
	return loginConn
}

func (c *clientChanConn) Close() {
	if c.isLogedIn {

	}

	c.Conn.Close()
}

func (c *clientChanConn) String() string {
	return c.ClientConnection.String()
}

func (c *clientChanConn) SetUserID(val uint32) {
	c.userID = val
}

func (c *clientChanConn) GetUserID() uint32 {
	return c.userID
}

func (c *clientChanConn) SetAdmin(val bool) {
	c.isAdmin = val
}

func (c *clientChanConn) IsAdmin() bool {
	return c.isAdmin
}

func (c *clientChanConn) SetIsLogedIn(val bool) {
	c.isLogedIn = val
}

func (c *clientChanConn) GetIsLogedIn() bool {
	return c.isLogedIn
}

func (c *clientChanConn) SetWorldID(val uint32) {
	c.worldID = val
}

func (c *clientChanConn) GetWorldID() uint32 {
	return c.worldID
}

func (c *clientChanConn) SetChanID(val byte) {
	c.chanID = val
}

func (c *clientChanConn) GetChanID() byte {
	return c.chanID
}
