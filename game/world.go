package game

import (
	"log"
	"net"

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type World struct {
	Icon          byte
	Name, Message string
	Ribbon        byte
	Channels      []channel
	login         mnet.Server
}

// RegisterWithLogin server
func (server *World) RegisterWithLogin(conn mnet.Server) {
	server.login = conn
	server.login.Send(server.generateInfoPacket())
}

// HandleChannelPacket from channel
func (server *World) HandleChannelPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	default:
		log.Println("UNKNOWN CHANNEL PACKET:", reader)
	}
}

// HandleLoginPacket from login
func (server *World) HandleLoginPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	default:
		log.Println("UNKNOWN LOGIN PACKET:", reader)
	}
}

func (w *World) AddChannel(ch channel) bool {
	if len(w.Channels) > 19 {
		return false
	}

	w.Channels = append(w.Channels, ch)
	return true
}

func (w *World) generateInfoPacket() mpacket.Packet {
	p := mpacket.CreateInterServer(opcode.WorldInfo)
	p.WriteByte(w.Icon)
	p.WriteString(w.Name)
	p.WriteString(w.Message)
	p.WriteByte(w.Ribbon)
	p.WriteByte(byte(len(w.Channels)))

	for _, v := range w.Channels {
		p.WriteBytes(v.generatePacket())
	}

	return p
}

type channel struct {
	ip          net.IP
	port        int16
	MaxPop, Pop int16
}

func (c channel) generatePacket() mpacket.Packet {
	p := mpacket.NewPacket()
	p.WriteBytes(c.ip)
	p.WriteInt16(c.port)
	p.WriteInt16(c.MaxPop)
	p.WriteInt16(c.Pop)
	return p
}
