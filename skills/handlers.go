package skills

import (
	"fmt"

	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func HandleStandardSkill(conn interfaces.ClientConn, reader maplepacket.Reader) (uint32, maplepacket.Packet, map[uint32][]uint32) {
	char := charsPtr.GetOnlineCharacterHandle(conn)

	tByte := reader.ReadByte()

	targets := tByte / 0x10
	hits := tByte % 0x10

	skillID := reader.ReadUint32()

	reader.ReadByte()

	display := reader.ReadByte()
	animation := reader.ReadByte()

	reader.ReadUint32()

	damages := make(map[uint32][]uint32)

	for i := byte(0); i < targets; i++ {
		objID := reader.ReadUint32()

		reader.ReadInt32() // ?
		reader.ReadInt32() // ?
		reader.ReadInt32() // ?
		reader.ReadInt16() // ?

		var dmgs []uint32

		for j := byte(0); j < hits; j++ {
			dmgs = append(dmgs, reader.ReadUint32())
		}

		damages[objID] = dmgs
	}

	// playerX := reader.ReadInt16()
	// playerY := reader.ReadInt16()

	return char.GetCurrentMap(), standardSkillPacket(char.GetCharID(), skillID, targets, hits, display, animation, damages), damages
}

func HandleRangedSkill(conn interfaces.ClientConn, reader maplepacket.Reader) (uint32, maplepacket.Packet, map[uint32][]uint32) {
	fmt.Println("Ranged skill", reader)
	char := charsPtr.GetOnlineCharacterHandle(conn)

	tByte := reader.ReadByte()

	targets := tByte / 0x10
	hits := tByte % 0x10

	skillID := reader.ReadUint32()

	reader.ReadByte()

	display := reader.ReadByte()
	animation := reader.ReadByte()

	reader.ReadUint32() // ?
	invPos := reader.ReadByte()
	reader.ReadBytes(4)
	fmt.Println("Ranged weapon inventory location:", invPos)

	damages := make(map[uint32][]uint32)

	for i := byte(0); i < targets; i++ {
		objID := reader.ReadUint32()

		reader.ReadBytes(14)

		var dmgs []uint32

		for j := byte(0); j < hits; j++ {
			dmgs = append(dmgs, reader.ReadUint32())
		}

		damages[objID] = dmgs
	}

	// playerX := reader.ReadInt16()
	// playerY := reader.ReadInt16()

	// hard coded ilbi for now
	return char.GetCurrentMap(), rangedSkillPacket(char.GetCharID(), skillID, 2070006, targets, hits, display, animation, damages), damages
}

func HandleMagicSkill(conn interfaces.ClientConn, reader maplepacket.Reader) (uint32, maplepacket.Packet, map[uint32][]uint32) {
	char := charsPtr.GetOnlineCharacterHandle(conn)

	tByte := reader.ReadByte()

	targets := tByte / 0x10
	hits := tByte % 0x10

	skillID := reader.ReadUint32()

	reader.ReadByte()

	display := reader.ReadByte()
	animation := reader.ReadByte()

	reader.ReadUint32() //?

	damages := make(map[uint32][]uint32)

	for i := byte(0); i < targets; i++ {
		objID := reader.ReadUint32()

		reader.ReadInt32() // ?
		reader.ReadInt32() // ?
		reader.ReadInt32() // ?
		reader.ReadInt16() // ?

		var dmgs []uint32

		for j := byte(0); j < hits; j++ {
			dmgs = append(dmgs, reader.ReadUint32())
		}

		damages[objID] = dmgs
	}

	// playerX := reader.ReadInt16()
	// playerY := reader.ReadInt16()

	return char.GetCurrentMap(), magicSkillPacket(char.GetCharID(), skillID, targets, hits, display, animation, damages), damages
}

func HandleSpecialSkill(conn interfaces.ClientConn, reader maplepacket.Reader) (uint32, maplepacket.Packet) {
	fmt.Println("Special skill", reader)
	skillID := reader.ReadUint32()
	level := reader.ReadByte()

	char := charsPtr.GetOnlineCharacterHandle(conn)

	// add all the various skills that fall under this category
	switch skillID {

	// GM SKILLS
	case 5001000: // gm haste normal
	case 5001001: //gm super dragon roar
	case 5001002: // gm teleport
	case 5101000: // // gm heal + dispel
	case 5101001: // // gm super haste
	case 5101002: // gm holy symbol
	case 5101003: // gm bless
	case 5101004: // gm hide
		conn.Write(gmHidePacket(true))
	case 5101005: // gm resurect
	default:
		fmt.Println("Unkown skill id:", skillID)
	}

	conn.Write(continuePacket())
	return char.GetCurrentMap(), skillAnimationPacket(char.GetCharID(), skillID, level)
}
