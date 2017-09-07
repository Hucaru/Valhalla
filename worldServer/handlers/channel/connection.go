package channel

import (
	"net"

	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

type Connection struct {
	conn      *connection.ServerConnection
	channelID byte
}

func newConnection(conn net.Conn) *Connection {
	serverConn := &Connection{conn: connection.NewServerConnection(conn)}
	return serverConn
}

func (c *Connection) Write(p gopacket.Packet) error {
	return c.conn.Write(p)
}

func (c *Connection) Read(p gopacket.Packet) error {
	return c.conn.Read(p)
}

func (c *Connection) String() string {
	return c.conn.String()
}

func (c *Connection) SetchannelID(val byte) {
	c.channelID = val
}

func (c *Connection) Close() {
	p := gopacket.NewPacket()
	p.WriteByte(constants.CHANNEL_DROPPED)
	p.WriteByte(c.channelID)
	ChannelServer <- connection.NewMessage(p, nil)
	c.conn.Close()
}
