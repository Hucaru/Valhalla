package mnet

import (
	"github.com/Hucaru/Valhalla/common/dataController"
	"github.com/Hucaru/Valhalla/common/db/model"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mnet/crypt"
	"github.com/Hucaru/Valhalla/mpacket"
	"net"
	"runtime"
	"sync"
)

type client struct {
	logedIn    bool
	accountID  int32
	gender     byte
	worldID    byte
	channelID  byte
	regionID   int64
	adminLevel int
	player     model.Player
}

type BaseConn struct {
	baseConn
}

type Client struct {
	BaseConn
	client
	MoveQueue        *dataController.MoveSyncLKQueue
	LastMoveSyncTime int64
}

func (c *Client) MetaMovoSync() {
	for {
		if c.closed {
			return
		}

		for {
			b := c.MoveQueue.Dequeue()
			if b == nil {
				break
			}

			if b.Time > c.LastMoveSyncTime {
				c.LastMoveSyncTime = b.Time
				b.Fn()
			}
		}
		runtime.Gosched()
	}
}

func NewClient(conn net.Conn, eRecv chan *Event, queueSize int, keySend, keyRecv [4]byte, latency, jitter int) *Client {
	c := &Client{}
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

func NewClientMeta(conn net.Conn, queueSize int, latency, jitter int) *Client {
	c := &Client{}
	c.Conn = conn
	//c.sendChannelLock = sync.RWMutex{}
	//c.sendChannelQueue = *dataController.NewLKQueue()

	c.sendChannelQueue = dataController.NewLKQueue()
	c.sendChannel = make(chan mpacket.Packet)
	c.sendChannelWaitGroup = sync.WaitGroup{}

	c.MoveQueue = dataController.NewMoveSyncLKQueue()
	c.LastMoveSyncTime = 0

	//c.eSend = make(chan mpacket.Packet, 4096*4)

	c.interServer = false
	c.latency = latency
	c.jitter = jitter
	//c.pSend = make(chan func(), queueSize*10) // Used only when simulating latency
	//
	//if latency > 0 {
	//	go func(pSend chan func(), conn net.Conn) {
	//		for {
	//			select {
	//			case p, ok := <-pSend:
	//				if !ok {
	//					return
	//				}
	//
	//				p()
	//			default:
	//			}
	//		}
	//	}(c.pSend, conn)
	//}
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

func (c *client) GetPlayer() model.Player {
	return c.player
}

func (c *client) GetPlayer_P() *model.Player {
	return &c.player
}

func (c *client) SetPlayer(player model.Player) {
	c.player = player
}

//func (c *client) EventLoop(f <-chan func()) {
//
//}

func (c *Client) String() string {
	return c.baseConn.String()
}
