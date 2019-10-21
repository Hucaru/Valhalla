package server

import (
	"github.com/Hucaru/Valhalla/entity"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (server *ChannelServer) npcMovement(conn mnet.Client, reader mpacket.Reader) {
	data := reader.GetRestAsBytes()
	id := reader.ReadInt32()

	player, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[player.Char().MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(player.InstanceID())

	if err != nil {
		return
	}

	npc := inst.GetNpc(id)

	if npc.Controller() != conn {
		conn.Send(entity.PacketNpcSetController(id, false))
	}

	inst.Send(entity.PacketNpcMovement(data))
}
