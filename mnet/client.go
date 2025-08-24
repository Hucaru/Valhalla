package mnet

import (
	"net"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mnet/crypt"
	"github.com/Hucaru/Valhalla/mpacket"
)

type Client interface {
	MConn

	GetLogedIn() bool
	SetLogedIn(bool)
	GetAccountID() int32
	SetAccountID(int32)
	GetGender() byte
	SetGender(byte)
	GetWorldID() byte
	SetWorldID(byte)
	GetChannelID() byte
	SetChannelID(byte)
	GetAdminLevel() int
	SetAdminLevel(int)
}

type client struct {
	baseConn

	logedIn    bool
	accountID  int32
	gender     byte
	worldID    byte
	channelID  byte
	adminLevel int
}

func NewClient(conn net.Conn, eRecv chan *Event, queueSize int, keySend, keyRecv [4]byte, latency, jitter int) *client {
	c := &client{}
	c.Conn = conn

	c.eSend = make(chan mpacket.Packet, queueSize)
	c.eRecv = eRecv

	c.cryptSend = crypt.New(keySend, constant.MapleVersion)
	c.cryptRecv = crypt.New(keyRecv, constant.MapleVersion)

	c.reader = func() {
		clientReader(c, c.eRecv, constant.MapleVersion, constant.ClientHeaderSize, c.cryptRecv)
	}

	c.interServer = false
	c.latency = latency
	c.jitter = jitter
	c.pSend = make(chan func(), queueSize*10) // Used only when simulating latency
	if latency > 0 {
		// Note: this routing doesn't close and eventually enough re-logs will start to eat more and more cpu, and should therefore be used when testing lag effects
		go func(pSend chan func(), conn net.Conn) {
			for {
				select {
				case p, ok := <-pSend:
					if !ok {
						return
					}

					p()
				default:
				}
			}
		}(c.pSend, conn)
	}

	return c
}

func (c *client) GetLogedIn() bool {
	return c.logedIn
}

func (c *client) SetLogedIn(logedIn bool) {
	c.logedIn = logedIn
}

func (c *client) GetAccountID() int32 {
	return c.accountID
}

func (c *client) SetAccountID(accountID int32) {
	c.accountID = accountID
}

func (c *client) GetGender() byte {
	return c.gender
}

func (c *client) SetGender(gender byte) {
	c.gender = gender
}

func (c *client) GetWorldID() byte {
	return c.worldID
}

func (c *client) SetWorldID(id byte) {
	c.worldID = id
}

func (c *client) GetChannelID() byte {
	return c.channelID
}

func (c *client) SetChannelID(id byte) {
	c.channelID = id
}

func (c *client) GetAdminLevel() int {
	return c.adminLevel
}

func (c *client) SetAdminLevel(level int) {
	c.adminLevel = level
}
