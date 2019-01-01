package channel

import (
	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/game/npcchat"
	"github.com/Hucaru/Valhalla/game/packet"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func npcMovement(conn mnet.MConnChannel, reader mpacket.Reader) {
	data := reader.GetRestAsBytes()
	id := reader.ReadInt32()

	conn.Send(packet.NPCMovement(data))
	conn.Send(packet.NPCSetController(id, true))
}

func npcChatStart(conn mnet.MConnChannel, reader mpacket.Reader) {
	npcSpawnID := reader.ReadInt32()

	player, ok := game.Players[conn]

	if !ok {
		return
	}

	m := game.Maps[player.Char().MapID]

	if m != nil {
		npc, err := m.GetNpcFromSpawnID(npcSpawnID, player.InstanceID)

		if err != nil {
			return
		}

		npcchat.NewSession(conn, npc.ID)
	} else {
		script :=
			`if state == 1 {
				return SendOk("NPC ID does not exist either on this map or in the game.")
			}`
		npcchat.NewSessionWithOverride(conn, script, 9010000)
	}

	npcchat.Run(conn)
}

func npcChatContinue(conn mnet.MConnChannel, reader mpacket.Reader) {
	msgType := reader.ReadByte()
	stateChange := reader.ReadByte()

	npcchat.Continue(conn, msgType, stateChange, reader)
}

func npcShop(conn mnet.MConnChannel, reader mpacket.Reader) {

}

func npcStorage(conn mnet.MConnChannel, reader mpacket.Reader) {

}
