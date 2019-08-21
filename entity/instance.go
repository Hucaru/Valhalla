package entity

import (
	"fmt"
	"math/rand"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type instance struct {
	id      int
	fieldID int32
	npcs    []npc
	portals []portal
	conns   []mnet.Client
	players *Players
}

func (inst *instance) delete() error {
	return nil
}

func (inst *instance) addPlayer(player *Player) error {
	for _, npc := range inst.npcs {
		player.Send(PacketNpcShow(npc))
		player.Send(PacketNpcSetController(npc.spawnID, true))
	}

	for _, other := range inst.conns {
		otherPlayer, _ := inst.players.GetFromConn(other)
		other.Send(PacketMapPlayerEnter(player.char))
		player.conn.Send(PacketMapPlayerEnter(otherPlayer.char))
	}

	// show all monsters on field

	// show all the rooms

	// show portals e.g. mystic door

	inst.conns = append(inst.conns, player.conn)
	return nil
}

func (inst *instance) removePlayer(player *Player) error {
	index := -1

	for i, v := range inst.conns {
		if v == player.conn {
			index = i
			break
		} else {

		}
	}
	if index == -1 {
		return fmt.Errorf("player does not exist in instance")
	}

	inst.conns = append(inst.conns[:index], inst.conns[index+1:]...)

	// if in room, remove
	for _, v := range inst.conns {
		v.Send(PacketMapPlayerLeft(player.char.id))
	}

	return nil
}

func (inst instance) send(p mpacket.Packet) error {
	for _, v := range inst.conns {
		v.Send(p)
	}

	return nil
}

func (inst instance) sendExcept(p mpacket.Packet, exception mnet.Client) error {
	for _, v := range inst.conns {
		if v == exception {
			continue
		}

		v.Send(p)
	}

	return nil
}

func (inst instance) getRandomSpawnPortal() (portal, error) {
	portals := []portal{}

	for _, p := range inst.portals {
		if p.name == "sp" {
			portals = append(portals, p)
		}
	}

	if len(portals) == 0 {
		return portal{}, fmt.Errorf("No spawn portals in map")
	}

	return portals[rand.Intn(len(portals))], nil
}

func (inst instance) getPortalFromName(name string) (portal, error) {
	for _, p := range inst.portals {
		if p.name == name {
			return p, nil
		}
	}

	return portal{}, fmt.Errorf("No portal with that name")
}

func (inst instance) getPortalFromID(id byte) (portal, error) {
	for _, p := range inst.portals {
		if p.id == id {
			return p, nil
		}
	}

	return portal{}, fmt.Errorf("No portal with that name")
}
