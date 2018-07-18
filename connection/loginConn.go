package connection

import (
	"log"
)

type Login struct {
	Client
	userID    int32
	isLogedIn bool
	isAdmin   bool
	hash      string
	worldID   int32
	chanID    byte
	gender    byte
}

// NewLoginConnection -
func NewLogin(conn Client) *Login {
	loginConn := &Login{Client: conn, isAdmin: false}
	return loginConn
}

func (c *Login) Close() {
	if c.isLogedIn {
		records, err := Db.Query("UPDATE users set isLogedIn=0 WHERE userID=?", c.userID)

		defer records.Close()

		if err != nil {
			log.Println("Error in auto log out of user on disconnect, userID:", c.userID)
		}
	}

	c.Conn.Close()
}

func (c *Login) String() string {
	return c.Client.String()
}

func (c *Login) SetUserID(val int32) {
	c.userID = val
}

func (c *Login) GetUserID() int32 {
	return c.userID
}

func (c *Login) SetAdmin(val bool) {
	c.isAdmin = val
}

func (c *Login) IsAdmin() bool {
	return c.isAdmin
}

func (c *Login) SetSessionHash(val string) {
	c.hash = val
}

func (c *Login) GetSessionHash() string {
	return c.hash
}

func (c *Login) SetIsLogedIn(val bool) {
	c.isLogedIn = val
}

func (c *Login) GetIsLogedIn() bool {
	return c.isLogedIn
}

func (c *Login) SetWorldID(val int32) {
	c.worldID = val
}

func (c *Login) GetWorldID() int32 {
	return c.worldID
}

func (c *Login) SetChanID(val byte) {
	c.chanID = val
}

func (c *Login) GetChanID() byte {
	return c.chanID
}

func (c *Login) SetGender(val byte) {
	c.gender = val
}

func (c *Login) GetGender() byte {
	return c.gender
}
