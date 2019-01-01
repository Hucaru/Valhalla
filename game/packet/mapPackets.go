package packet

import (
	opcodes "github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/mpacket"
)

func MapPlayerEnter(char def.Character) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelCharacterEnterField)
	p.WriteInt32(char.ID)    // player id
	p.WriteString(char.Name) // char name

	if true {
		p.WriteString("test")
		p.WriteInt16(0)
		p.WriteByte(0)
		p.WriteByte(0)
		p.WriteInt16(0)
		p.WriteInt32(0)
		p.WriteInt32(0)
	} else {
		p.WriteInt32(0)
		p.WriteInt32(0)
		p.WriteInt32(0)
		p.WriteInt32(0)
	}

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

func MapPlayerLeft(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelCharacterLeaveField)
	p.WriteInt32(charID)

	return p
}

func MapChange(mapID int32, channelID int32, mapPos byte, hp int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelWarpToMap)
	p.WriteInt32(channelID)
	p.WriteByte(0) // character portal counter
	p.WriteByte(0) // Is connecting
	p.WriteInt32(mapID)
	p.WriteByte(mapPos)
	p.WriteInt16(hp)
	p.WriteByte(0) // flag for more reading

	return p
}
