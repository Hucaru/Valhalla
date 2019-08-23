package entity

import (
	"fmt"
	"math/rand"
	"strconv"

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

func (inst instance) String() string {
	var info string

	info += "players(" + strconv.Itoa(len(inst.conns)) + "): "

	for _, v := range inst.conns {
		player, _ := inst.players.GetFromConn(v)
		info += " " + player.char.name + "(" + player.Pos().String() + ")"
	}

	return info
}

func (inst instance) PlayerCount() int {
	return len(inst.conns)
}

func (inst *instance) AddPlayer(player *Player) error {
	for i, npc := range inst.npcs {
		player.Send(PacketNpcShow(npc))

		if npc.controller == nil {
			inst.npcs[i].controller = player.conn
			player.Send(PacketNpcSetController(npc.spawnID, true))
		}
	}

	for _, other := range inst.conns {
		otherPlayer, err := inst.players.GetFromConn(other)

		if err != nil {
			continue
		}

		other.Send(PacketMapPlayerEnter(player.char))
		player.conn.Send(PacketMapPlayerEnter(otherPlayer.char))
	}

	// show all monsters on field

	// show all the rooms

	// show portals e.g. mystic door

	inst.conns = append(inst.conns, player.conn)
	return nil
}

func (inst *instance) RemovePlayer(player *Player) error {
	index := -1

	for i, v := range inst.conns {
		if v == player.conn {
			index = i
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("player does not exist in instance")
	}

	inst.conns = append(inst.conns[:index], inst.conns[index+1:]...)

	// if in room, remove

	for _, v := range inst.conns {
		v.Send(PacketMapPlayerLeft(player.char.id))
		otherPlayer, err := inst.players.GetFromConn(v)

		if err != nil {
			continue
		}

		player.Send(PacketMapPlayerLeft(otherPlayer.char.id))
	}

	for i, v := range inst.npcs {
		if v.controller == player.conn {
			player.Send(PacketNpcSetController(v.spawnID, false))

			if len(inst.conns) > 0 {
				inst.conns[0].Send(PacketNpcSetController(v.spawnID, true))
				inst.npcs[i].controller = inst.conns[0]
			}
		}
	}

	return nil
}

func (inst instance) Send(p mpacket.Packet) error {
	for _, v := range inst.conns {
		v.Send(p)
	}

	return nil
}

func (inst instance) SendExcept(p mpacket.Packet, exception mnet.Client) error {
	for _, v := range inst.conns {
		if v == exception {
			continue
		}

		v.Send(p)
	}

	return nil
}

func (inst instance) GetRandomSpawnPortal() (portal, error) {
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

func (inst instance) GetPortalFromName(name string) (portal, error) {
	for _, p := range inst.portals {
		if p.name == name {
			return p, nil
		}
	}

	return portal{}, fmt.Errorf("No portal with that name")
}

func (inst instance) GetPortalFromID(id byte) (portal, error) {
	for _, p := range inst.portals {
		if p.id == id {
			return p, nil
		}
	}

	return portal{}, fmt.Errorf("No portal with that name")
}

func (inst *instance) GetNpc(id int32) *npc {
	if id < 0 || int(id) > len(inst.npcs) {
		return nil
	}

	return &inst.npcs[id]
}
