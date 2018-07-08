package handlers

import (
	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/npcdialogue"

	"github.com/Hucaru/Valhalla/interop"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/packets"
)

func handleNPCMovement(conn interop.ClientConn, reader maplepacket.Reader) {
	data := reader.GetRestAsBytes()
	id := reader.ReadInt32()

	conn.Write(packets.NPCMovement(data))
	conn.Write(packets.NPCSetController(id, true))
}

func handleNPCChat(conn interop.ClientConn, reader maplepacket.Reader) {
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

func handleNPCChatContinue(conn interop.ClientConn, reader maplepacket.Reader) {
	msgType := reader.ReadByte()

	stateChange := reader.ReadByte()
	npcdialogue.GetSession(conn).Continue(msgType, stateChange, reader)
}

func handleNPCShop(conn interop.ClientConn, reader maplepacket.Reader) {
	npcdialogue.GetSession(conn).Shop(reader)
}
