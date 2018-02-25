package skills

import (
	"github.com/Hucaru/Valhalla/channelServer/playerConn"
	"github.com/Hucaru/Valhalla/channelServer/server"
	"github.com/Hucaru/gopacket"
)

func HandlePlayerSkillUsage(reader gopacket.Reader, conn *playerConn.Conn) {
	char := conn.GetCharacter()

	level := reader.ReadByte()
	skillID := reader.ReadUint32()

	// For now just make the map aware that the skill was used

	server.SendPacketToMap(char.GetCurrentMap(), playerSkillAnimation(char.GetCharID(), skillID, level), nil)
}

func HandlePlayerSpecialSkillUsage(reader gopacket.Reader, conn *playerConn.Conn) {
	char := conn.GetCharacter()

	skillID := reader.ReadUint32()
	level := reader.ReadByte()

	// For now just make the map aware that the skill was used

	server.SendPacketToMap(char.GetCurrentMap(), playerSkillAnimation(char.GetCharID(), skillID, level), nil)

}
