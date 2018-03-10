package skills

import (
	"fmt"

	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/gopacket"
)

func HandleStandardSkill(conn interfaces.ClientConn, reader gopacket.Reader) (gopacket.Packet, uint32) {
	char := charsPtr.GetOnlineCharacterHandle(conn)

	tByte := reader.ReadByte()

	targets := tByte / 0x10
	hits := tByte % 0x10

	skillID := reader.ReadUint32()

	reader.ReadByte()

	display := reader.ReadByte()
	animation := reader.ReadByte()

	reader.ReadInt32()

	damages := make(map[uint32][]uint32)

	for i := byte(0); i < targets; i++ {
		objID := reader.ReadUint32()

		reader.ReadInt32() // ?
		reader.ReadInt16() // objx
		reader.ReadInt16() // objy
		reader.ReadInt32() // ?
		reader.ReadInt16() // objy

		var dmgs []uint32

		for j := byte(0); j < hits; j++ {
			dmgs = append(dmgs, reader.ReadUint32())
		}

		damages[objID] = dmgs
	}

	// playerX := reader.ReadInt16()
	// playerY := reader.ReadInt16()

	// char.SetY(playerY)

	return skillAnimationPacket(char.GetCharID(), skillID, tByte, targets, hits, display, animation, damages), char.GetCurrentMap()
}

func HandleRangedSkill(conn interfaces.ClientConn, reader gopacket.Reader) (gopacket.Packet, uint32) {
	fmt.Println("Ranged skill", reader)
	char := charsPtr.GetOnlineCharacterHandle(conn)

	return []byte{}, char.GetCurrentMap()
}

func HandleSpecialSkill(conn interfaces.ClientConn, reader gopacket.Reader) (gopacket.Packet, uint32) {
	fmt.Println("Special skill", reader)
	// skillID := reader.ReadUint32()
	// level := reader.ReadByte()

	// for now just create the show packet

	char := charsPtr.GetOnlineCharacterHandle(conn)

	return []byte{}, char.GetCurrentMap()
}
