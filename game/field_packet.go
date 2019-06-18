package game

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

func packetMapPlayerEnter(char character) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCharacterEnterField)
	p.WriteInt32(char.id)    // player id
	p.WriteString(char.name) // char name

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
	p.WriteInt32(char.chairID) // 0 means no chair in use, stance needs to be changed to match

	p.WriteInt16(char.pos.X)
	p.WriteInt16(char.pos.Y)

	p.WriteByte(char.stance)
	p.WriteInt16(char.foothold)
	p.WriteInt32(0) // ?

	return p
}

func packetMapPlayerLeft(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCharacterLeaveField)
	p.WriteInt32(charID)

	return p
}

func packetMapChange(mapID int32, channelID int32, mapPos byte, hp int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelWarpToMap)
	p.WriteInt32(channelID)
	p.WriteByte(0) // character portal counter
	p.WriteByte(0) // Is connecting
	p.WriteInt32(mapID)
	p.WriteByte(mapPos)
	p.WriteInt16(hp)
	p.WriteByte(0) // flag for more reading

	return p
}

func packetMapShowGameBox(charID, roomID int32, roomType, boardType byte, name string, hasPassword, koreanText bool, ammount byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoomBox)
	p.WriteInt32(charID)
	p.WriteByte(roomType)
	p.WriteInt32(roomID)
	p.WriteString(name)
	p.WriteBool(hasPassword)
	p.WriteByte(boardType)
	// win loss record since room opened?
	p.WriteByte(ammount)
	p.WriteByte(2)
	p.WriteBool(koreanText)

	return p
}

func packetMapRemoveGameBox(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoomBox)
	p.WriteInt32(charID)
	p.WriteInt32(0)

	return p
}
