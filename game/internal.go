package game

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type world struct {
	conn          mnet.Server
	icon          byte
	name, message string
	ribbon        byte
	channels      []channel
}

func (w *world) generateInfoPacket() mpacket.Packet {
	p := mpacket.CreateInternal(opcode.WorldInfo)
	p.WriteByte(w.icon)
	p.WriteString(w.name)
	p.WriteString(w.message)
	p.WriteByte(w.ribbon)
	p.WriteByte(byte(len(w.channels)))

	for _, v := range w.channels {
		p.WriteBytes(v.generatePacket())
	}

	return p
}

func (w *world) serialisePacket(reader mpacket.Reader) {
	w.icon = reader.ReadByte()
	w.name = reader.ReadString(reader.ReadInt16())
	w.message = reader.ReadString(reader.ReadInt16())
	w.ribbon = reader.ReadByte()

	nOfChannels := int(reader.ReadByte())
	w.channels = make([]channel, nOfChannels)

	for i := 0; i < nOfChannels; i++ {
		w.channels[i].serialisePacket(&reader)
	}
}

type channel struct {
	conn        mnet.Server
	ip          []byte
	port        int16
	maxPop, pop int16
}

func (c channel) generatePacket() mpacket.Packet {
	p := mpacket.NewPacket()
	p.WriteBytes(c.ip)
	p.WriteInt16(c.port)
	p.WriteInt16(c.maxPop)
	p.WriteInt16(c.pop)
	return p
}

func (c *channel) serialisePacket(reader *mpacket.Reader) {
	c.ip = reader.ReadBytes(4)
	c.port = reader.ReadInt16()
	c.maxPop = reader.ReadInt16()
	c.pop = reader.ReadInt16()
}
