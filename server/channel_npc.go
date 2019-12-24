package server

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (server *ChannelServer) npcMovement(conn mnet.Client, reader mpacket.Reader) {
	data := reader.GetRestAsBytes()
	id := reader.ReadInt32()

	player, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[player.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(player.InstanceID())

	if err != nil {
		return
	}

	npc := inst.GetNpc(id)
	npc.AcknowledgeController(player, inst, data)
}
