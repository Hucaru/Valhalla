package game

import (
	"fmt"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type instance struct {
	id      int
	fieldID int32
	npcs    []npc
	players []mnet.Client
	server  *ChannelServer
}

func (inst *instance) addPlayer(conn mnet.Client) error {
	for _, npc := range inst.npcs {
		conn.Send(packetNpcShow(npc))
		conn.Send(packetNpcSetController(npc.spawnID, true))
	}

	for _, other := range inst.players {
		other.Send(packetMapPlayerEnter(inst.server.players[conn].char))
		conn.Send(packetMapPlayerEnter(inst.server.players[other].char))
	}

	inst.players = append(inst.players, conn)
	return nil
}

func (inst *instance) removePlayer(conn mnet.Client) error {
	index := -1

	for i, v := range inst.players {
		if v == conn {
			index = i
			break
		} else {

		}
	}
	if index == -1 {
		return fmt.Errorf("player does not exist in instance")
	}

	inst.players = append(inst.players[:index], inst.players[index+1:]...)

	for _, v := range inst.players {
		v.Send(packetMapPlayerLeft(inst.server.players[conn].char.id))
	}

	return nil
}

func (inst *instance) delete() error {
	return nil
}

func (inst instance) send(p mpacket.Packet) error {
	for _, v := range inst.players {
		v.Send(p)
	}

	return nil
}

func (inst instance) sendExcept(p mpacket.Packet, exception mnet.Client) error {
	for _, v := range inst.players {
		if v == exception {
			continue
		}

		v.Send(p)
	}

	return nil
}
