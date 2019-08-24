package entity

import (
	"fmt"

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

func PacketMapPlayerEnter(char Character) mpacket.Packet {
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

	p.WriteBytes(WriteDisplayCharacter(char))

	p.WriteInt32(0)            // ?
	p.WriteInt32(0)            // ?
	p.WriteInt32(0)            // ?
	p.WriteInt32(char.chairID) // 0 means no chair in use, stance needs to be changed to match

	p.WriteInt16(char.pos.x)
	p.WriteInt16(char.pos.y)
	p.WriteByte(char.stance)
	p.WriteInt16(char.foothold)
	p.WriteInt32(0) // ?

	return p
}

func PacketMapPlayerLeft(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCharacterLeaveField)
	p.WriteInt32(charID)

	return p
}

func PacketMapChange(mapID int32, channelID int32, mapPos byte, hp int16) mpacket.Packet {
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

func PacketMapShowGameBox(charID, roomID int32, roomType, boardType byte, name string, hasPassword, koreanText bool, ammount byte) mpacket.Packet {
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

func PacketMapRemoveGameBox(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoomBox)
	p.WriteInt32(charID)
	p.WriteInt32(0)

	return p
}

func PacketMapSpawnMysticDoor(spawnID int32, pos pos, instant bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSpawnDoor)
	p.WriteBool(instant)
	p.WriteInt32(spawnID)
	p.WriteInt16(pos.x)
	p.WriteInt16(pos.y)

	return p
}

func PacketMapPortal(srcMap, dstmap int32, pos pos) mpacket.Packet {
	p := mpacket.CreateWithOpcode(0x2d)
	p.WriteByte(26)
	p.WriteByte(0) // ?
	p.WriteInt32(srcMap)
	p.WriteInt32(dstmap)
	p.WriteInt16(pos.x)
	p.WriteInt16(pos.y)

	return p
}

func PacketMapSpawnTownPortal(dstMap, srcMap int32, destPos pos) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelTownPortal)
	p.WriteInt32(dstMap)
	p.WriteInt32(srcMap)
	p.WriteInt16(destPos.x)
	p.WriteInt16(destPos.y)
	fmt.Println(p)
	return p
}

func PacketMapRemoveMysticDoor(spawnID int32, fade bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRemoveDoor)
	p.WriteBool(fade)
	p.WriteInt32(spawnID)

	return p
}
