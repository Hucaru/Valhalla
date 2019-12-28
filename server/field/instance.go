package field

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/field/mob"
	"github.com/Hucaru/Valhalla/server/field/npc"
	"github.com/Hucaru/Valhalla/server/field/room"
	"github.com/Hucaru/Valhalla/server/item"
	"github.com/Hucaru/Valhalla/server/pos"
)

type player interface {
	Conn() mnet.Client
	ID() int32
	InstanceID() int
	SetInstanceID(int)
	Name() string
	Pos() pos.Data
	DisplayBytes() []byte
	ChairID() int32
	Stance() byte
	Foothold() int16
	Send(mpacket.Packet)
	MiniGameWins() int32
	MiniGameDraw() int32
	MiniGameLoss() int32
}

type players interface {
	GetFromConn(mnet.Client) (player, error)
}

type drop struct {
	item       item.Data
	pos        pos.Data
	expireTime int64
	ownerName  string
	partyID    int32
}

// Instance data for a field
type Instance struct {
	id             int
	fieldID        int32
	npcs           []npc.Data
	portals        []Portal
	players        []player
	mobs           []mob.Data
	drops          []drop
	rooms          []room.Room
	fieldTimer     *time.Ticker
	fieldTimerTime int64
	idCounter      int32

	dispatch chan func()
}

func (inst *Instance) delete() error {
	return nil
}

func (inst Instance) String() string {
	var info string

	info += "players(" + strconv.Itoa(len(inst.players)) + "): "

	for _, v := range inst.players {
		info += " " + v.Name() + "(" + v.Pos().String() + ")"
	}

	return info
}

// PlayerCount for the instance
func (inst Instance) PlayerCount() int {
	return len(inst.players)
}

// AddPlayer to the instance
func (inst *Instance) AddPlayer(plr player) error {
	for i, npc := range inst.npcs {
		plr.Send(packetNpcShow(npc))

		if npc.Controller() == nil {
			inst.npcs[i].SetController(plr)
		}
	}

	for _, other := range inst.players {
		other.Send(packetMapPlayerEnter(plr))
		plr.Send(packetMapPlayerEnter(other))
	}

	// show all monsters on field
	for i, m := range inst.mobs {
		plr.Send(packetMobShow(m))
		if m.Controller() == nil {
			inst.mobs[i].SetController(plr, false)
		}
	}

	// show all the rooms
	for _, v := range inst.rooms {
		if game, valid := v.(room.Game); valid {
			plr.Send(packetMapShowGameBox(game.DisplayBytes()))
		}
	}

	// show portals e.g. mystic door

	// Play map animations e.g. ship arriving to dock

	inst.players = append(inst.players, plr)

	if len(inst.players) == 1 {
		inst.startFieldTimer()
	}

	return nil
}

// RemovePlayer from instance
func (inst *Instance) RemovePlayer(plr player) error {
	index := -1

	for i, v := range inst.players {
		if v.Conn() == plr.Conn() {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("player does not exist in instance")
	}

	inst.players = append(inst.players[:index], inst.players[index+1:]...)

	// if in room, remove, if room is closed update map

	for _, v := range inst.players {
		v.Send(packetMapPlayerLeft(plr.ID()))
		plr.Send(packetMapPlayerLeft(v.ID()))
	}

	for i, v := range inst.npcs {
		if v.Controller().Conn() == plr.Conn() {
			inst.npcs[i].RemoveController()

			if len(inst.players) > 0 {
				inst.npcs[i].SetController(inst.players[0])
			}
		}
	}

	for i, v := range inst.mobs {
		if v.Controller().Conn() == plr.Conn() {
			inst.mobs[i].RemoveController()

			if len(inst.players) > 0 {
				inst.mobs[i].SetController(inst.players[0], false)
			}
		}
	}

	for _, v := range inst.rooms {
		if game, valid := v.(room.Game); valid {
			game.KickPlayer(plr, 0x0)

			if v.Closed() {
				inst.RemoveRoom(v)
			}
		}
	}

	if len(inst.players) == 0 {
		inst.stopFieldTimer()
	}

	return nil
}

// NextID - gets the next available id to be used by the instance
func (inst *Instance) NextID() int32 {
	inst.idCounter++
	return inst.idCounter
}

// AddRoom to the instance
func (inst *Instance) AddRoom(r room.Room) {
	inst.rooms = append(inst.rooms, r)

	if game, valid := r.(room.Game); valid {
		inst.Send(packetMapShowGameBox(game.DisplayBytes()))
	}
}

// UpdateGameBox above player head in map
func (inst *Instance) UpdateGameBox(r room.Room) {
	if game, valid := r.(room.Game); valid {
		inst.Send(packetMapShowGameBox(game.DisplayBytes()))
	}
}

// RemoveRoom from instance
func (inst *Instance) RemoveRoom(r room.Room) error {
	for i, v := range inst.rooms {
		if v.ID() == r.ID() {
			inst.rooms[i] = inst.rooms[len(inst.rooms)-1]
			inst.rooms = inst.rooms[:len(inst.rooms)-1]

			if _, valid := r.(room.Game); valid {
				inst.Send(packetMapRemoveGameBox(r.OwnerID()))
			}
			return nil
		}
	}

	return fmt.Errorf("Could not find room to delete")
}

// GetPlayerRoom the room the player belongs to
func (inst *Instance) GetPlayerRoom(id int32) (room.Room, error) {
	for _, v := range inst.rooms {
		if v.Present(id) {
			return v, nil
		}
	}

	return nil, fmt.Errorf("Player does not belong to a room")
}

// GetRoomID the room with the passed in id
func (inst *Instance) GetRoomID(id int32) (room.Room, error) {
	for _, v := range inst.rooms {
		if v.ID() == id {
			return v, nil
		}
	}
	return nil, fmt.Errorf("no room with id")
}

// Send packet to instance
func (inst Instance) Send(p mpacket.Packet) error {
	for _, v := range inst.players {
		v.Send(p)
	}

	return nil
}

// SendExcept - sends packet to instance except a particular player
func (inst Instance) SendExcept(p mpacket.Packet, exception player) error {
	for _, v := range inst.players {
		if v == exception {
			continue
		}

		v.Send(p)
	}

	return nil
}

// GetRandomSpawnPortal returns a spawn potal at random
func (inst Instance) GetRandomSpawnPortal() (Portal, error) {
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

// CalculateNearestSpawnPortalID from a given position
func (inst Instance) CalculateNearestSpawnPortalID(pos pos.Data) (byte, error) {
	var portal Portal
	found := true
	err := fmt.Errorf("Portal not found")

	for _, p := range inst.portals {
		if p.name == "sp" && found {
			portal = p
			found = false
			err = nil
		} else if p.name == "sp" {
			delta1 := portal.pos.CalcDistanceSquare(pos)
			delta2 := p.pos.CalcDistanceSquare(pos)

			if delta2 < delta1 {
				portal = p
			}
		}
	}

	return portal.id, err
}

// GetPortalFromName in the current instance
func (inst Instance) GetPortalFromName(name string) (Portal, error) {
	for _, p := range inst.portals {
		if p.name == name {
			return p, nil
		}
	}

	return Portal{}, fmt.Errorf("No portal with that name")
}

// GetPortalFromID in the current instance
func (inst Instance) GetPortalFromID(id byte) (Portal, error) {
	for _, p := range inst.portals {
		if p.id == id {
			return p, nil
		}
	}

	return Portal{}, fmt.Errorf("No portal with that name")
}

// GetNpc in current instance
func (inst *Instance) GetNpc(id int32) *npc.Data {
	for i, v := range inst.npcs {
		if v.SpawnID() == id {
			return &inst.npcs[i]
		}
	}

	return nil
}

// GetMob in current instance
func (inst *Instance) GetMob(id int32) *mob.Data {
	for i, v := range inst.mobs {
		if v.SpawnID() == id {
			return &inst.mobs[i]
		}
	}

	return nil
}

func (inst *Instance) startFieldTimer() {
	inst.fieldTimer = time.NewTicker(time.Second * time.Duration(5)) // Change to correct time
	go func() {
		for t := range inst.fieldTimer.C {
			inst.dispatch <- func() { inst.fieldUpdate(t) }
		}
	}()
}

func (inst *Instance) stopFieldTimer() {
	inst.fieldTimer.Stop()
}

func (inst *Instance) fieldUpdate(t time.Time) {
}
