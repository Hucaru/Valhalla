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

	return char.GetCurrentMap(), standardSkillPacket(char.GetCharID(), skillID, tByte, targets, hits, display, animation, damages), damages
}

func HandleRangedSkill(conn interfaces.ClientConn, reader gopacket.Reader) (uint32, gopacket.Packet) {
	fmt.Println("Ranged skill", reader)
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

	fmt.Println(skillID, display, animation, damages, targets, hits)

	return char.GetCurrentMap(), rangedSkillPacket(char.GetCharID(), skillID, tByte, targets, hits, display, animation, damages)
}

func HandleSpecialSkill(conn interfaces.ClientConn, reader gopacket.Reader) (uint32, gopacket.Packet) {
	fmt.Println("Special skill", reader)
	// skillID := reader.ReadUint32()
	// level := reader.ReadByte()

	// for now just create the show packet

	char := charsPtr.GetOnlineCharacterHandle(conn)

	return char.GetCurrentMap(), []byte{}
}
