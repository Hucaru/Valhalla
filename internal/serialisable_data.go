package internal

import (
	"github.com/Hucaru/Valhalla/common/mnet"
	"github.com/Hucaru/Valhalla/common/mpacket"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
)

type World struct {
	Conn          mnet.Server
	Icon          byte
	Name, Message string
	Ribbon        byte
	Channels      []Channel
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

type Party struct {
	ID        int32
	ChannelID [constant.MaxPartySize]int32
	PlayerID  [constant.MaxPartySize]int32
	Name      [constant.MaxPartySize]string
	MapID     [constant.MaxPartySize]int32
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

type Guild struct {
	ID       int32
	Capacity byte
	Name     string
	Notice   string

	Master   string
	JrMaster string
	Member1  string
	Member2  string
	Member3  string

	LogoBg, Logo             int16
	LogoBgColour, LogoColour byte

	PlayerID [constant.MaxGuildSize]int32
	Names    [constant.MaxGuildSize]string
	Jobs     [constant.MaxGuildSize]int32
	Levels   [constant.MaxGuildSize]int32
	Online   [constant.MaxGuildSize]bool
	Ranks    [constant.MaxGuildSize]int32
}

func (guild *Guild) GeneratePacket() mpacket.Packet {
	p := mpacket.NewPacket()
	p.WriteInt32(guild.ID)
	p.WriteByte(guild.Capacity)
	p.WriteString(guild.Name)
	p.WriteString(guild.Notice)
	p.WriteString(guild.Master)
	p.WriteString(guild.JrMaster)
	p.WriteString(guild.Member1)
	p.WriteString(guild.Member2)
	p.WriteString(guild.Member3)
	p.WriteInt16(guild.LogoBg)
	p.WriteByte(guild.LogoBgColour)
	p.WriteInt16(guild.Logo)
	p.WriteByte(guild.LogoColour)

	validIndexes := make([]int32, 0, constant.MaxGuildSize)
	for i, v := range guild.PlayerID {
		if v != 0 {
			validIndexes = append(validIndexes, int32(i))
		}
	}

	p.WriteByte(byte(len(validIndexes)))

	for _, i := range validIndexes {
		p.WriteInt32(guild.PlayerID[i])
		p.WriteString(guild.Names[i])
		p.WriteInt32(guild.Jobs[i])
		p.WriteInt32(guild.Levels[i])
		p.WriteBool(guild.Online[i])
		p.WriteInt32(guild.Ranks[i])
	}

	return p
}

func (guild *Guild) SerialisePacket(reader *mpacket.Reader) {
	guild.ID = reader.ReadInt32()
	guild.Capacity = reader.ReadByte()
	guild.Name = reader.ReadString(reader.ReadInt16())
	guild.Notice = reader.ReadString(reader.ReadInt16())

	guild.Master = reader.ReadString(reader.ReadInt16())
	guild.JrMaster = reader.ReadString(reader.ReadInt16())
	guild.Member1 = reader.ReadString(reader.ReadInt16())
	guild.Member2 = reader.ReadString(reader.ReadInt16())
	guild.Member3 = reader.ReadString(reader.ReadInt16())

	guild.LogoBg = reader.ReadInt16()
	guild.LogoBgColour = reader.ReadByte()
	guild.Logo = reader.ReadInt16()
	guild.LogoColour = reader.ReadByte()

	amount := reader.ReadByte()

	for i := byte(0); i < amount; i++ {
		guild.PlayerID[i] = reader.ReadInt32()
		guild.Names[i] = reader.ReadString(reader.ReadInt16())
		guild.Jobs[i] = reader.ReadInt32()
		guild.Levels[i] = reader.ReadInt32()
		guild.Online[i] = reader.ReadBool()
		guild.Ranks[i] = reader.ReadInt32()
	}
}
