package handlers

import (
	"log"
	"strings"

	"github.com/Hucaru/Valhalla/commands"
	"github.com/Hucaru/Valhalla/packets"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/interop"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func handleAllChat(conn interop.ClientConn, reader maplepacket.Reader) {
	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
		mapID := char.GetCurrentMap()

		msg := reader.ReadString(int(reader.ReadInt16()))

		if strings.Index(msg, "/") == 0 && conn.IsAdmin() {
			commands.HandleGmCommand(conn, msg)
		} else {
			channel.Maps.GetMap(mapID).SendPacket(packets.MessageAllChat(char.GetCharID(), conn.IsAdmin(), msg))
		}
	})
}

func handleSlashCommand(conn interop.ClientConn, reader maplepacket.Reader) {
	cmdType := reader.ReadByte()

	switch cmdType {
	case 5:
		length := reader.ReadInt16()
		name := reader.ReadString(int(length))

		found := false

		channel.Players.OnCharacterFromName(name, func(char *channel.MapleCharacter) {
			found = true
			conn.Write(packets.MessageFindResult(name, char.IsAdmin(), false, true, char.GetCurrentMap()))
		})

		if !found {
			// go ask world server if exist and on what channel
		}

	default:
		log.Println("Slash command not implemented:", cmdType)
	}
}
