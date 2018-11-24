package handlers

import (
	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/npcdialogue"

	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/game/packet"
)

func handleNPCMovement(conn mnet.MConnChannel, reader maplepacket.Reader) {
	data := reader.GetRestAsBytes()
	id := reader.ReadInt32()

	conn.Write(packet.NPCMovement(data))
	conn.Write(packet.NPCSetController(id, true))
}

func handleNPCChat(conn mnet.MConnChannel, reader maplepacket.Reader) {
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

func handleNPCChatContinue(conn mnet.MConnChannel, reader maplepacket.Reader) {
	msgType := reader.ReadByte()

	stateChange := reader.ReadByte()
	npcdialogue.GetSession(conn).Continue(msgType, stateChange, reader)
}

func handleNPCShop(conn mnet.MConnChannel, reader maplepacket.Reader) {
	npcdialogue.GetSession(conn).Shop(reader)
}
