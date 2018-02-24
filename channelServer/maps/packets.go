package maps

import (
	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

func playerEnterField(char *character.Character) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_CHARCTER_ENTER_FIELD)
	p.WriteUint32(char.GetCharID()) // player id
	p.WriteString(char.GetName())   // char name
	p.WriteUint32(0)                // map buffs?
	p.WriteUint32(0)                // map buffs?
	p.WriteUint32(0)                // map buffs?
	p.WriteUint32(0)                // map buffs?

	character.WriteDisplayCharacter(char, &p)

	p.WriteUint32(0)                // ?
	p.WriteUint32(0)                // ?
	p.WriteUint32(0)                // ?
	p.WriteUint32(char.GetCharID()) // 0 means no chair in use, stance needs to be changed to match

	p.WriteInt16(char.GetX())
	p.WriteInt16(char.GetY())

	p.WriteByte(char.GetState())
	p.WriteUint16(char.GetFh())
	p.WriteUint32(0) // ?

	return p
}

func changeMap(mapID uint32, channelID uint32, mapPos byte, hp uint16) gopacket.Packet {
	p := gopacket.NewPacket()
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

func playerLeftField(charID uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_CHARCTER_LEAVE_FIELD)
	p.WriteUint32(charID)

	return p
}

func playerMove(charID uint32, leftOverBytes gopacket.Packet) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_MOVEMENT)
	p.WriteUint32(charID)
	p.WriteBytes(leftOverBytes)

	return p
}
