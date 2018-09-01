package packets

import (
	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func MapPlayerEnter(char character.Character) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SendChannelCharacterEnterField)
	p.WriteInt32(char.GetCharID()) // player id
	p.WriteString(char.GetName())  // char name
	p.WriteInt32(0)                // map buffs?
	p.WriteInt32(0)                // map buffs?
	p.WriteInt32(0)                // map buffs?
	p.WriteInt32(0)                // map buffs?

	p.WriteBytes(writeDisplayCharacter(char))

	p.WriteInt32(0)                 // ?
	p.WriteInt32(0)                 // ?
	p.WriteInt32(0)                 // ?
	p.WriteInt32(char.GetChairID()) // 0 means no chair in use, stance needs to be changed to match

	p.WriteInt16(char.GetX())
	p.WriteInt16(char.GetY())

	p.WriteByte(char.GetState())
	p.WriteInt16(char.GetFoothold())
	p.WriteInt32(0) // ?

	return p
}

func MapPlayerLeft(charID int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SendChannelCharacterLeaveField)
	p.WriteInt32(charID)

	return p
}

func MapChange(mapID int32, channelID int32, mapPos byte, hp int16) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SendChannelWarpToMap)
	p.WriteInt32(channelID)
	p.WriteByte(0) // character portal counter
	p.WriteByte(0) // Is connecting
	p.WriteInt32(mapID)
	p.WriteByte(mapPos)
	p.WriteInt16(hp)
	p.WriteByte(0) // flag for more reading

	return p
}
