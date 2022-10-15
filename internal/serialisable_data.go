package internal

import (
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type Rates struct {
	Exp   float32
	Drop  float32
	Mesos float32
}

type World struct {
	Conn          mnet.Server
	Icon          byte
	Name, Message string
	Ribbon        byte
	Channels      []Channel
	Rates         Rates
	DefaultRates  Rates
}

func (w *World) GenerateInfoPacket() mpacket.Packet {
	p := mpacket.CreateInternal(opcode.WorldInfo)
	p.WriteByte(w.Icon)
	p.WriteString(w.Name)
	p.WriteString(w.Message)
	p.WriteByte(w.Ribbon)
	p.WriteByte(byte(len(w.Channels)))

	for _, v := range w.Channels {
		p.WriteBytes(v.GeneratePacket())
	}

	return p
}

func (w *World) SerialisePacket(reader mpacket.Reader) {
	w.Icon = reader.ReadByte()
	w.Name = reader.ReadString(reader.ReadInt16())
	w.Message = reader.ReadString(reader.ReadInt16())
	w.Ribbon = reader.ReadByte()

	nOfChannels := int(reader.ReadByte())
	w.Channels = make([]Channel, nOfChannels)

	for i := 0; i < nOfChannels; i++ {
		w.Channels[i].SerialisePacket(&reader)
	}
}

type Channel struct {
	Conn        mnet.Server
	IP          []byte
	Port        int16
	MaxPop, Pop int16
}

func (c Channel) GeneratePacket() mpacket.Packet {
	p := mpacket.NewPacket()
	p.WriteBytes(c.IP)
	p.WriteInt16(c.Port)
	p.WriteInt16(c.MaxPop)
	p.WriteInt16(c.Pop)
	return p
}

func (c *Channel) SerialisePacket(reader *mpacket.Reader) {
	c.IP = reader.ReadBytes(4)
	c.Port = reader.ReadInt16()
	c.MaxPop = reader.ReadInt16()
	c.Pop = reader.ReadInt16()
}
