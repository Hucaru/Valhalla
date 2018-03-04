package chat

import (
	"strings"

	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/gopacket"
)

func HandleAllChat(conn interfaces.ClientConn, reader gopacket.Reader) (uint32, string, bool, gopacket.Packet) {
	mapID := charsPtr.GetOnlineCharacterHandle(conn).GetCurrentMap()
	charID := charsPtr.GetOnlineCharacterHandle(conn).GetCharID()

	msg := reader.ReadString(int(reader.ReadInt16()))
	ind := strings.Index(msg, "!")

	if ind == 0 && conn.IsAdmin() {
		return mapID, msg, true, []byte{}
	}

	return mapID, "", false, allChatPacket(charID, conn.IsAdmin(), msg)
}
