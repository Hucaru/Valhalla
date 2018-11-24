package packets

import (
	"github.com/Hucaru/Valhalla/consts/opcodes"
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func MapPlayerEnter(char def.Character) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelCharacterEnterField)
	p.WriteInt32(char.ID)    // player id
	p.WriteString(char.Name) // char name
	p.WriteInt32(0)          // map buffs?
	p.WriteInt32(0)          // map buffs?
	p.WriteInt32(0)          // map buffs?
	p.WriteInt32(0)          // map buffs?

	p.WriteBytes(writeDisplayCharacter(char))

	p.WriteInt32(0)            // ?
	p.WriteInt32(0)            // ?
	p.WriteInt32(0)            // ?
	p.WriteInt32(char.ChairID) // 0 means no chair in use, stance needs to be changed to match

	p.WriteInt16(char.Pos.X)
	p.WriteInt16(char.Pos.Y)

	p.WriteByte(char.Stance)
	p.WriteInt16(char.Foothold)
	p.WriteInt32(0) // ?

	return p
}

func MapPlayerLeft(charID int32) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelCharacterLeaveField)
	p.WriteInt32(charID)

	return p
}

func MapChange(mapID int32, channelID int32, mapPos byte, hp int16) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelWarpToMap)
	p.WriteInt32(channelID)
	p.WriteByte(0) // character portal counter
	p.WriteByte(0) // Is connecting
	p.WriteInt32(mapID)
	p.WriteByte(mapPos)
	p.WriteInt16(hp)
	p.WriteByte(0) // flag for more reading

	return p
}
