package server

import (
	"log"

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

func channelPopUpdate(id byte, pop int16) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelInfo)
	p.WriteByte(id)
	p.WriteByte(0) // 0 is population
	p.WriteInt16(pop)

	return p
}

func channelPlayerConnected(id int32, name string, channelID byte, channelChange bool) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannePlayerConnect)
	p.WriteInt32(id)
	p.WriteString(name)
	p.WriteByte(channelID)
	p.WriteBool(channelChange)

	return p
}

func channelPlayerDisconnect(id int32, name string) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannePlayerDisconnect)
	p.WriteInt32(id)
	p.WriteString(name)

	return p
}

func channelBuddyEvent(op byte, recepientID, fromID int32, fromName string, channelID byte) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerBuddyEvent)
	p.WriteByte(op)

	switch op {
	case 1: // add
		fallthrough
	case 2: // accept
		p.WriteInt32(recepientID)
		p.WriteInt32(fromID)
		p.WriteString(fromName)
		p.WriteByte(channelID)
	case 3: // delete / reject
		p.WriteInt32(recepientID)
		p.WriteInt32(fromID)
		p.WriteByte(channelID)
	default:
		log.Println("unkown internal buddy event type:", op)
	}

	return p
}

func channelBuddyChat(fromName string, buffer []byte) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerChatEvent)
	p.WriteByte(1) // buddy
	p.WriteString(fromName)
	p.WriteBytes(buffer)

	return p
}
