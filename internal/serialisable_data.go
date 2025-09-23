package internal

import (
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
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
	CashShop      CashShop
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
type CashShop struct {
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

type Party struct {
	ID        int32
	ChannelID [constant.MaxPartySize]int32
	PlayerID  [constant.MaxPartySize]int32
	Name      [constant.MaxPartySize]string
	MapID     [constant.MaxPartySize]int32 // TODO: this can be removed as plr ptr is used
	Job       [constant.MaxPartySize]int32
	Level     [constant.MaxPartySize]int32
}

func (party Party) GeneratePacket() mpacket.Packet {
	p := mpacket.NewPacket()
	p.WriteInt32(party.ID)
	p.WriteByte(byte(len(party.PlayerID)))

	for i := 0; i < len(party.PlayerID); i++ {
		p.WriteInt32(party.ChannelID[i])
		p.WriteInt32(party.PlayerID[i])
		p.WriteString(party.Name[i])
		p.WriteInt32(party.MapID[i])
		p.WriteInt32(party.Job[i])
		p.WriteInt32(party.Level[i])
	}

	return p
}

func (party *Party) SerialisePacket(reader *mpacket.Reader) {
	party.ID = reader.ReadInt32()

	amount := reader.ReadByte()

	for i := byte(0); i < amount; i++ {
		party.ChannelID[i] = reader.ReadInt32()
		party.PlayerID[i] = reader.ReadInt32()
		party.Name[i] = reader.ReadString(reader.ReadInt16())
		party.MapID[i] = reader.ReadInt32()
		party.Job[i] = reader.ReadInt32()
		party.Level[i] = reader.ReadInt32()
	}
}

type KV struct {
	K byte
	V int32
}
