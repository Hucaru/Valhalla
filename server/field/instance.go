package field

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/field/droppool"
	"github.com/Hucaru/Valhalla/server/field/lifepool"
	"github.com/Hucaru/Valhalla/server/field/roompool"
	"github.com/Hucaru/Valhalla/server/pos"
)

type player interface {
	Conn() mnet.Client
	ID() int32
	InstanceID() int
	SetInstance(interface{})
	Name() string
	Pos() pos.Data
	DisplayBytes() []byte
	ChairID() int32
	Stance() byte
	Send(mpacket.Packet)
	MiniGameWins() int32
	MiniGameDraw() int32
	MiniGameLoss() int32
	MiniGamePoints() int32
	SetMiniGameWins(int32)
	SetMiniGameDraw(int32)
	SetMiniGameLoss(int32)
	SetMiniGamePoints(int32)
}

type socket interface {
	Send(mpacket.Packet)
}

type players interface {
	GetFromConn(mnet.Client) (player, error)
}

// Instance data for a field
type Instance struct {
	id          int
	fieldID     int32
	returnMapID int32
	timeLimit   int64

	lifePool lifepool.Data
	dropPool droppool.Data
	roomPool roompool.Data

	portals []Portal
	players []player

	// rooms []room.Room

	idCounter int32
	town      bool

	dispatch chan func()

	fieldTimer *time.Ticker
	runUpdate  bool
}

// ID of the instance within the field
func (inst Instance) ID() int {
	return inst.id
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

// LifePool pointer for instance
func (inst *Instance) LifePool() *lifepool.Data {
	return &inst.lifePool
}

// DropPool pointer for instance
func (inst *Instance) DropPool() *droppool.Data {
	return &inst.dropPool
}

// RoomPool pointer for instance
func (inst *Instance) RoomPool() *roompool.Data {
	return &inst.roomPool
}

// FindController in instance, need to return interface for casting
func (inst Instance) FindController() interface{} {
	for _, v := range inst.players {
		return v
	}

	return nil
}

// AddPlayer to the instance
func (inst *Instance) AddPlayer(plr player) error {
	plr.SetInstance(inst)

	for _, other := range inst.players {
		other.Send(packetMapPlayerEnter(plr))
		plr.Send(packetMapPlayerEnter(other))
	}

	inst.lifePool.AddPlayer(plr)
	inst.dropPool.PlayerShowDrops(plr)
	inst.roomPool.PlayerShowRooms(plr)

	// Play map animations e.g. ship arriving to dock

	inst.players = append(inst.players, plr)

	// For now pools run on all maps forever after first player enters.
	// If this hits perf too much then a set of params for each pool
	// will need to be determined to allow it to stop updating e.g.
	// drop pool, no drops and no players
	// life pool, max number of mobs spawned and no dot attacks in field
	if !inst.runUpdate {
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

	for _, v := range inst.players {
		v.Send(packetMapPlayerLeft(plr.ID()))
		plr.Send(packetMapPlayerLeft(v.ID()))
	}

	inst.lifePool.RemovePlayer(plr)
	inst.roomPool.RemovePlayer(plr)

	return nil
}

// GetPlayerFromID from the instance
func (inst Instance) GetPlayerFromID(id int32) (player, error) {
	for i, v := range inst.players {
		if v.ID() == id {
			return inst.players[i], nil
		}
	}

	return nil, fmt.Errorf("Player not in instance")
}

// MovePlayer for other players
func (inst Instance) MovePlayer(id int32, moveBytes []byte, plr player) {
	inst.SendExcept(packetPlayerMove(id, moveBytes), plr.Conn())
}

// NextID - gets the next available id to be used by the instance
func (inst *Instance) NextID() int32 {
	inst.idCounter++
	return inst.idCounter
}

// Send packet to instance
func (inst Instance) Send(p mpacket.Packet) error {
	for _, v := range inst.players {
		v.Send(p)
	}

	return nil
}

// SendExcept - sends packet to instance except a particular player
func (inst Instance) SendExcept(p mpacket.Packet, exception mnet.Client) error {
	for _, v := range inst.players {
		if v.Conn() == exception {
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

func (inst *Instance) startFieldTimer() {
	inst.runUpdate = true
	inst.fieldTimer = time.NewTicker(time.Millisecond * 1000) // Is this correct time?

	go func() {
		for t := range inst.fieldTimer.C {
			inst.dispatch <- func() { inst.fieldUpdate(t) }
		}
	}()
}

func (inst *Instance) stopFieldTimer() {
	inst.runUpdate = false
	inst.fieldTimer.Stop()
}

// Responsible for hadnling the removing of mystic doors, disappearence of loot, ships coming and going
func (inst *Instance) fieldUpdate(t time.Time) {
	inst.lifePool.Update(t)
	inst.dropPool.Update(t)

	if inst.lifePool.CanClose() && inst.dropPool.CanClose() {
		inst.stopFieldTimer()
	}
}

// CalculateFinalDropPos from a starting position
func (inst *Instance) CalculateFinalDropPos(from pos.Data) pos.Data {
	return from
}
