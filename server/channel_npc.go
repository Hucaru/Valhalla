package server

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (server *ChannelServer) npcMovement(conn mnet.Client, reader mpacket.Reader) {
	data := reader.GetRestAsBytes()
	id := reader.ReadInt32()

	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[plr.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(plr.InstanceID())

	if err != nil {
		return
	}

	inst.LifePool().NpcAcknowledge(id, plr, data)
}
