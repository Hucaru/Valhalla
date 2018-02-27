package skills

import (
	"fmt"

	"github.com/Hucaru/Valhalla/channelServer/playerConn"
	"github.com/Hucaru/Valhalla/channelServer/server"
	"github.com/Hucaru/gopacket"
)

func HandlePlayerSkillUsage(reader gopacket.Reader, conn *playerConn.Conn) {
	char := conn.GetCharacter()

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

		// validate object is where map thinks it is, within 500ms, keep previous position
		reader.ReadInt32() // ?
		reader.ReadInt16() // objx
		reader.ReadInt16() // objy
		reader.ReadInt32() // ?
		reader.ReadInt16() // objy

		for j := byte(0); j < hits; j++ {
			dmg := reader.ReadUint32()
			fmt.Println(objID, dmg)
			damages[objID] = append(damages[objID], dmg)
		}
	}

	playerX := reader.ReadInt16()
	playerY := reader.ReadInt16()

	char.SetX(playerX)
	char.SetY(playerY)

	fmt.Println(reader)

	server.SendPacketToMap(char.GetCurrentMap(), playerSkillAnimation(char.GetCharID(), skillID, tByte, targets, hits, display, animation, damages), conn)
}

func HandlePlayerSpecialSkillUsage(reader gopacket.Reader, conn *playerConn.Conn) {
	// char := conn.GetCharacter()

	// skillID := reader.ReadUint32()
	// level := reader.ReadByte()

	// For now just make the map aware that the skill was used

	// server.SendPacketToMap(char.GetCurrentMap(), playerSkillAnimation(char.GetCharID(), skillID, level), nil)

}
