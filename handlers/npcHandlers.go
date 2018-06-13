package handlers

import (
	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/npcChat"

	"github.com/Hucaru/Valhalla/interop"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/packets"
)

func handleNPCMovement(conn interop.ClientConn, reader maplepacket.Reader) {
	data := reader.GetRestAsBytes()
	id := reader.ReadUint32()

	conn.Write(packets.NPCMovement(data))
	conn.Write(packets.NPCSetController(id, true))
}

func handleNPCChat(conn interop.ClientConn, reader maplepacket.Reader) {
	npcSpawnID := reader.ReadUint32()

	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
		for _, npc := range channel.NPCs.GetNpcs(char.GetCurrentMap()) {
			if npc.GetSpawnID() == npcSpawnID {
				npcChat.NewSession(conn, npc.GetID(), char)
				npcChat.GetSession(conn).Run()
			}
		}
	})
}

func handleNPCChatContinue(conn interop.ClientConn, reader maplepacket.Reader) {
	msgType := reader.ReadByte()

	stateChange := reader.ReadByte()
	npcChat.GetSession(conn).Continue(msgType, stateChange, reader)
}

func handleNPCShop(conn interop.ClientConn, reader maplepacket.Reader) {
	npcChat.GetSession(conn).Shop(reader)
}
