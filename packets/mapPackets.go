package packets

import (
	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func MapPlayerEnter(char character.Character) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_CHARCTER_ENTER_FIELD)
	p.WriteUint32(char.GetCharID()) // player id
	p.WriteString(char.GetName())   // char name
	p.WriteUint32(0)                // map buffs?
	p.WriteUint32(0)                // map buffs?
	p.WriteUint32(0)                // map buffs?
	p.WriteUint32(0)                // map buffs?

	character.WriteDisplayCharacter(char, &p)

	p.WriteUint32(0)                 // ?
	p.WriteUint32(0)                 // ?
	p.WriteUint32(0)                 // ?
	p.WriteUint32(char.GetChairID()) // 0 means no chair in use, stance needs to be changed to match

	p.WriteInt16(char.GetX())
	p.WriteInt16(char.GetY())

	p.WriteByte(char.GetState())
	p.WriteInt16(char.GetFoothold())
	p.WriteUint32(0) // ?

	return p
}

func MapPlayerLeft(charID uint32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_CHARCTER_LEAVE_FIELD)
	p.WriteUint32(charID)

	return p
}

func MapChange(mapID uint32, channelID uint32, mapPos byte, hp uint16) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_WARP_TO_MAP)
	p.WriteUint32(channelID)
	p.WriteByte(0) // character portal counter
	p.WriteByte(0) // Is connecting
	p.WriteUint32(mapID)
	p.WriteByte(mapPos)
	p.WriteUint16(hp)
	p.WriteByte(0) // flag for more reading

	return p
}
