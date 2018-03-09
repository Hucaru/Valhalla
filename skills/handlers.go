package skills

import (
	"fmt"

	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/gopacket"
)

func HandleStandardSkill(conn interfaces.ClientConn, reader gopacket.Reader) (gopacket.Packet, uint32) {
	fmt.Println("Standard Skill", reader)
	usageType := reader.ReadByte()
	skillID := reader.ReadUint32()

	// for now just create the show packet

	char := charsPtr.GetOnlineCharacterHandle(conn)

	return skillAnimationPacket(char.GetCharID(), skillID, usageType), char.GetCurrentMap()
}

func HandleRangedSkill(conn interfaces.ClientConn, reader gopacket.Reader) (gopacket.Packet, uint32) {
	fmt.Println("Ranged skill", reader)
	usageType := reader.ReadByte()
	skillID := reader.ReadUint32()

	// for now just create the show packet

	char := charsPtr.GetOnlineCharacterHandle(conn)

	return skillAnimationPacket(char.GetCharID(), skillID, usageType), char.GetCurrentMap()
}

func HandleSpecialSkill(conn interfaces.ClientConn, reader gopacket.Reader) (gopacket.Packet, uint32) {
	fmt.Println("Special skill", reader)
	skillID := reader.ReadUint32()
	level := reader.ReadByte()

	// for now just create the show packet

	char := charsPtr.GetOnlineCharacterHandle(conn)

	return skillAnimationPacket(char.GetCharID(), skillID, level), char.GetCurrentMap()
}
