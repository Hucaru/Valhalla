package skills

import (
	"fmt"

	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/gopacket"
)

func HandleStandardSkill(conn interfaces.ClientConn, reader gopacket.Reader) (uint32, gopacket.Packet, map[uint32][]uint32) {
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

func HandleRangedSkill(conn interfaces.ClientConn, reader gopacket.Reader) (uint32, gopacket.Packet, map[uint32][]uint32) {
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
	invPos := reader.ReadUint16()
	reader.ReadBytes(3)
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

func HandleMagicSkill(conn interfaces.ClientConn, reader gopacket.Reader) (uint32, gopacket.Packet, map[uint32][]uint32) {
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

func HandleSpecialSkill(conn interfaces.ClientConn, reader gopacket.Reader) (uint32, gopacket.Packet) {
	fmt.Println("Special skill", reader)
	skillID := reader.ReadUint32()
	level := reader.ReadByte()

	char := charsPtr.GetOnlineCharacterHandle(conn)

	return char.GetCurrentMap(), skillAnimationPacket(char.GetCharID(), skillID, level)
}
