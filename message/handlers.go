package message

import (
	"strings"

	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func HandleAllChat(conn interfaces.ClientConn, reader maplepacket.Reader) (uint32, string, bool, maplepacket.Packet) {
	mapID := charsPtr.GetOnlineCharacterHandle(conn).GetCurrentMap()
	charID := charsPtr.GetOnlineCharacterHandle(conn).GetCharID()

	msg := reader.ReadString(int(reader.ReadInt16()))
	ind := strings.Index(msg, "!")

	if ind == 0 && conn.IsAdmin() {
		return mapID, msg, true, []byte{}
	}

	return mapID, "", false, allChatPacket(charID, conn.IsAdmin(), msg)
}
