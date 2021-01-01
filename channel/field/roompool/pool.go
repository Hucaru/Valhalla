package roompool

import (
	"fmt"
	"math"

	"github.com/Hucaru/Valhalla/channel/field/roompool/room"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type field interface {
	Send(mpacket.Packet) error
}

type controller interface {
	ID() int32
	Send(p mpacket.Packet)
	// Need the following to match the player interface in room, could expose it if this becomes an issue
	Conn() mnet.Client
	Name() string
	DisplayBytes() []byte
	MiniGameWins() int32
	MiniGameDraw() int32
	MiniGameLoss() int32
	MiniGamePoints() int32
	SetMiniGameWins(int32)
	SetMiniGameDraw(int32)
	SetMiniGameLoss(int32)
	SetMiniGamePoints(int32)
}

// Data for the pool
type Data struct {
	instance field
	rooms    map[int32]room.Room
	poolID   int32
}

// CreateNewPool for rooms
func CreateNewPool(inst field) Data {
	return Data{instance: inst, rooms: make(map[int32]room.Room)}
}

func (pool *Data) nextID() (int32, error) {
	for i := 0; i < 100; i++ { // Try 99 times to generate an id if first time fails
		pool.poolID++

		if pool.poolID == math.MaxInt32-1 {
			pool.poolID = math.MaxInt32 / 2
		} else if pool.poolID == 0 {
			pool.poolID = 1
		}

		if _, ok := pool.rooms[pool.poolID]; !ok {
			return pool.poolID, nil
		}
	}

	return 0, fmt.Errorf("No space to generate id in drop pool")
}

// PlayerShowRooms in pool
func (pool *Data) PlayerShowRooms(plr controller) {
	for _, r := range pool.rooms {
		if game, valid := r.(room.Game); valid {
			plr.Send(packetMapShowGameBox(game.DisplayBytes()))
		}
	}
}

// AddRoom to the pool
func (pool *Data) AddRoom(r room.Room) error {
	id, err := pool.nextID()

	if err != nil {
		return err
	}

	r.SetID(id)

	pool.rooms[id] = r

	pool.UpdateGameBox(r)

	return nil
}

// RemoveRoom from the pool with the associated id
func (pool *Data) RemoveRoom(id int32) error {
	if _, ok := pool.rooms[id]; !ok {
		return fmt.Errorf("Could not delete room as id was not found")
	}

	if _, valid := pool.rooms[id].(room.Game); valid {
		pool.instance.Send(packetMapRemoveGameBox(pool.rooms[id].OwnerID()))
	}

	delete(pool.rooms, id)

	return nil
}

// GetRoom with associated id
func (pool Data) GetRoom(id int32) (room.Room, error) {
	if _, ok := pool.rooms[id]; !ok {
		return nil, fmt.Errorf("Could not retrieve room as id was not found")
	}

	return pool.rooms[id], nil
}

// GetPlayerRoom from the player's id
func (pool Data) GetPlayerRoom(id int32) (room.Room, error) {
	for _, r := range pool.rooms {
		if r.Present(id) {
			return r, nil
		}
	}

	return nil, fmt.Errorf("no room with id")
}

// UpdateGameBox above player head in map
func (pool Data) UpdateGameBox(r room.Room) {
	if game, valid := r.(room.Game); valid {
		pool.instance.Send(packetMapShowGameBox(game.DisplayBytes()))
	}
}

// RemovePlayer checks if a player is in a room and removes them and closes the room if they are the owner
func (pool *Data) RemovePlayer(plr controller) {
	r, err := pool.GetPlayerRoom(plr.ID())

	if err != nil {
		return
	}

	if game, valid := r.(room.Game); valid {
		game.KickPlayer(plr, 0x0)

		if r.Closed() {
			pool.RemoveRoom(r.ID())
		}
	}
}
