package mnet

import (
	"net"

	"github.com/Hucaru/Valhalla/consts"
	"github.com/Hucaru/Valhalla/database"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/mnet/crypt"
)

type MConnChannel interface {
	MConn

	GetLogedIn() bool
	SetLogedIn(bool)
	GetAccountID() int32
	SetAccountID(int32)
	GetAdminLevel() int
	SetAdminLevel(int)
}

type channel struct {
	baseConn

	logedIn    bool
	accountID  int32
	adminLevel int
}

func NewChannel(conn net.Conn, eRecv chan *Event, queueSize int, keySend, keyRecv [4]byte) *channel {
	c := &channel{}
	c.Conn = conn

	c.eSend = make(chan mpacket.Packet, queueSize)
	c.eRecv = eRecv
	c.endSend = make(chan bool, 1)

	c.cryptSend = crypt.New(keySend, consts.MapleVersion)
	c.cryptRecv = crypt.New(keyRecv, consts.MapleVersion)

	c.reader = func() {
		clientReader(c, c.eRecv, consts.MapleVersion, consts.ClientHeaderSize, c.cryptRecv)
	}

	return c
}

func (c *channel) Cleanup() {
	c.baseConn.Cleanup()

	if c.logedIn {
		records, err := database.Handle.Query("UPDATE accounts SET isInChannel=? WHERE accountID=?", -1, c.accountID)

		defer records.Close()

		if err != nil {
			panic(err)
		}
	}
}

func (c *channel) GetLogedIn() bool {
	return c.logedIn
}

func (c *channel) SetLogedIn(logedIn bool) {
	c.logedIn = logedIn
}

func (c *channel) GetAccountID() int32 {
	return c.accountID
}

func (c *channel) SetAccountID(accountID int32) {
	c.accountID = accountID
}

func (c *channel) GetAdminLevel() int {
	return c.adminLevel
}

func (c *channel) SetAdminLevel(level int) {
	c.adminLevel = level
}
