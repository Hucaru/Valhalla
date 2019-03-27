package game

import (
	"net"

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

type world struct {
	Icon          byte
	Name, Message string
	Ribbon        byte
	Channels      []channel
}

func (w *world) AddChannel(ch channel) bool {
	if len(w.Channels) > 19 {
		return false
	}

	w.Channels = append(w.Channels, ch)
	return true
}

func (w *world) generateInfoPacket() mpacket.Packet {
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
