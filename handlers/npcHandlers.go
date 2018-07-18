package handlers

import (
	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/npcdialogue"

	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/packets"
)

func handleNPCMovement(conn *connection.Channel, reader maplepacket.Reader) {
	data := reader.GetRestAsBytes()
	id := reader.ReadInt32()

	conn.Write(packets.NPCMovement(data))
	conn.Write(packets.NPCSetController(id, true))
}

func handleNPCChat(conn *connection.Channel, reader maplepacket.Reader) {
	npcSpawnID := reader.ReadInt32()

	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
		for _, npc := range channel.NPCs.GetNpcs(char.GetCurrentMap()) {
			if npc.GetSpawnID() == npcSpawnID {
				npcdialogue.NewSession(conn, npc.GetID(), char)
				npcdialogue.GetSession(conn).Run()
			}
		}
	})
}

func handleNPCChatContinue(conn *connection.Channel, reader maplepacket.Reader) {
	msgType := reader.ReadByte()

	stateChange := reader.ReadByte()
	npcdialogue.GetSession(conn).Continue(msgType, stateChange, reader)
}

func handleNPCShop(conn *connection.Channel, reader maplepacket.Reader) {
	npcdialogue.GetSession(conn).Shop(reader)
}
