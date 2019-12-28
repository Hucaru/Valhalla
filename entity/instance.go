package entity

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/field/room"
)

type instance struct {
	id             int
	fieldID        int32
	npcs           []npc
	portals        []Portal
	conns          []mnet.Client
	players        *Players
	mobs           []mob
	rooms          []room.Room
	fieldTimer     *time.Ticker
	fieldTimerTime int64
	roomCounter    int32 // Counting on when this overflows rooms won't still be around to clash

	dispatch chan func()
}

type gameRoom interface {
	DisplayInfo()
}

func (inst *instance) delete() error {
	return nil
}

func (inst instance) String() string {
	var info string

	info += "players(" + strconv.Itoa(len(inst.conns)) + "): "

	for _, v := range inst.conns {
		player, _ := inst.players.GetFromConn(v)
		info += " " + player.char.name + "(" + player.Char().Pos().String() + ")"
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
	for i, m := range inst.mobs {
		player.Send(PacketMobShow(m))
		if m.controller == nil {
			inst.mobs[i].SetController(player.conn, false)
		}
	}

	// show all the rooms
	for _, v := range inst.rooms {
		if r, valid := v.(gameRoom); valid {
			r.DisplayInfo()
			//inst.Send()
		}
	}

	// show portals e.g. mystic door

	// Play map animations e.g. ship arriving to dock

	inst.conns = append(inst.conns, player.conn)

	if len(inst.conns) == 1 {
		inst.startFieldTimer()
	}

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

	// if in room, remove, if room is closed update map

	for _, v := range inst.conns {
		v.Send(PacketMapPlayerLeft(player.char.id))
		otherPlayer, err := inst.players.GetFromConn(v)

		if err != nil {
			continue
		}

		player.Send(PacketMapPlayerLeft(otherPlayer.char.id))
	}

	var newController mnet.Client

	for i, v := range inst.npcs {
		if v.controller == player.conn {
			player.Send(PacketNpcSetController(v.spawnID, false))
			inst.npcs[i].controller = newController

			if newController != nil {
				newController.Send(PacketNpcSetController(v.spawnID, true))
			}
		}
	}

	for i, m := range inst.mobs {
		if m.controller == player.conn {
			if newController == nil {
				inst.mobs[i].controller = nil
			} else {
				inst.mobs[i].SetController(newController, false)
			}
			player.Send(PacketMobEndControl(m))
		}
	}

	if len(inst.conns) == 0 {
		inst.stopFieldTimer()
	}

	return nil
}

func (inst *instance) NextRoomID() int32 {
	inst.roomCounter++
	return inst.roomCounter
}

func (inst *instance) AddRoom(r room.Room) {
	inst.rooms = append(inst.rooms, r)

	if room, valid := r.(gameRoom); valid {
		log.Println("Send game room display to map")
		room.DisplayInfo()
		//inst.Send()
	}
}

func (inst *instance) RemoveRoom() error {
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

func (inst instance) GetRandomSpawnPortal() (Portal, error) {
	portals := []Portal{}

	for _, p := range inst.portals {
		if p.name == "sp" {
			portals = append(portals, p)
		}
	}

	if len(portals) == 0 {
		return Portal{}, fmt.Errorf("No spawn portals in map")
	}

	return portals[rand.Intn(len(portals))], nil
}

func (inst instance) CalculateNearestSpawnPortal(pos pos) (Portal, error) {
	var portal Portal
	found := true
	err := fmt.Errorf("Portal not found")

	for _, p := range inst.portals {
		if p.name == "sp" && found {
			portal = p
			found = false
			err = nil
		} else if p.name == "sp" {
			delta1 := portal.pos.calcDistanceSquare(pos)
			delta2 := p.pos.calcDistanceSquare(pos)

			if delta2 < delta1 {
				portal = p
			}
		}
	}

	return portal, err
}

func (inst instance) GetPortalFromName(name string) (Portal, error) {
	for _, p := range inst.portals {
		if p.name == name {
			return p, nil
		}
	}

	return Portal{}, fmt.Errorf("No portal with that name")
}

func (inst instance) GetPortalFromID(id byte) (Portal, error) {
	for _, p := range inst.portals {
		if p.id == id {
			return p, nil
		}
	}

	return Portal{}, fmt.Errorf("No portal with that name")
}

func (inst *instance) GetNpc(id int32) *npc {
	if id < 0 || int(id) > len(inst.npcs) {
		return nil
	}

	return &inst.npcs[id]
}

func (inst *instance) GetMob(id int32) *mob {
	for i, v := range inst.mobs {
		if v.spawnID == id {
			return &inst.mobs[i]
		}
	}

	return nil
}

func (inst *instance) startFieldTimer() {
	inst.fieldTimer = time.NewTicker(time.Second * time.Duration(5)) // Change to correct time
	go func() {
		for t := range inst.fieldTimer.C {
			inst.dispatch <- func() { inst.fieldUpdate(t) }
		}
	}()
}

func (inst *instance) stopFieldTimer() {
	inst.fieldTimer.Stop()
}

func (inst *instance) fieldUpdate(t time.Time) {
}
