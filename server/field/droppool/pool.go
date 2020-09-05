package droppool

import (
	"fmt"
	"math"
	"time"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/item"
	"github.com/Hucaru/Valhalla/server/pos"
)

const (
	SpawnDisappears      = 0
	SpawnNormal          = 1
	SpawnShow            = 2
	SpawnFadeAtTopOfDrop = 3
)

type field interface {
	Send(mpacket.Packet) error
	CalculateFinalDropPos(pos.Data) pos.Data
}

type controller interface {
	Send(mpacket.Packet)
	Conn() mnet.Client
}

// Data structure for the pool
type Data struct {
	instance field
	poolID   int32
	drops    map[int32]drop // If this struct doesn't stay static change to a ptr
}

// CreateNewPool for drops
func CreateNewPool(inst field) Data {
	return Data{instance: inst, drops: make(map[int32]drop)}
}

func (pool *Data) nextID() (int32, error) {
	for i := 0; i < 100; i++ { // Try 99 times to generate an id if first time fails
		pool.poolID++

		if pool.poolID == math.MaxInt32-1 {
			pool.poolID = math.MaxInt32 / 2
		} else if pool.poolID == 0 {
			pool.poolID = 1
		}

		if _, ok := pool.drops[pool.poolID]; !ok {
			return pool.poolID, nil
		}
	}

	return 0, fmt.Errorf("No space to generate id in drop pool")
}

// CanClose the pool down
func (pool Data) CanClose() bool {
	return false
}

// PlayerShowDrops when entering instance
func (pool Data) PlayerShowDrops(plr controller) {
	for _, drop := range pool.drops {
		plr.Send(packetShowDrop(SpawnShow, drop))
	}
}

func (pool *Data) RemoveDrop(instant bool, id ...int32) {
	for _, id := range id {
		pool.instance.Send(packetRemoveDrop(instant, id))

		if _, ok := pool.drops[id]; ok {
			delete(pool.drops, id)
		}
	}
}

// PlayerAttemptPickup of item
func (pool *Data) PlayerAttemptPickup(dropID int32, position pos.Data) (bool, item.Data) {
	return false, item.Data{}
}

const itemDistance = 20 // Between 15 and 20?
const itemDisppearTimeout = time.Minute * 2
const itemLootableByAllTimeout = time.Minute * 1

// CreateDrop into field
func (pool *Data) CreateDrop(spawnType byte, dropType byte, mesos int32, dropFrom pos.Data, expire bool, ownerID, partyID int32, items ...item.Data) {
	iCount := len(items)
	var offset int16 = 0

	if mesos > 0 {
		iCount++
	}

	if iCount > 0 {
		offset = int16(itemDistance * (iCount / 2))
	}

	currentTime := time.Now()
	expireTime := currentTime.Add(itemDisppearTimeout).Unix()
	var timeoutTime int64 = 0

	if dropType == DropTimeoutNonOwner || dropType == DropTimeoutNonOwnerParty {
		timeoutTime = currentTime.Add(itemLootableByAllTimeout).Unix()
	}

	for i, item := range items {
		finalPos := pool.instance.CalculateFinalDropPos(dropFrom) // (dropFrom, xShift)

		finalPos.SetX(finalPos.X() - offset + int16(i*itemDistance)) // This calculation needs to be interpolated to be placed on correct position on ledge

		if poolID, err := pool.nextID(); err == nil {
			drop := drop{
				ID:      poolID,
				ownerID: ownerID,
				partyID: partyID,
				mesos:   0,
				item:    item,

				expireTime:  expireTime,
				timeoutTime: timeoutTime,
				neverExpire: false,

				originPos: dropFrom,
				finalPos:  finalPos,

				dropType: dropType,
			}

			pool.drops[drop.ID] = drop

			pool.instance.Send(packetShowDrop(spawnType, drop))
		}
	}

	if mesos > 0 {
		finalPos := pool.instance.CalculateFinalDropPos(dropFrom)

		if iCount > 1 {
			finalPos.SetX(finalPos.X() - offset + int16((iCount-1)*itemDistance))
		}

		if poolID, err := pool.nextID(); err == nil {
			drop := drop{
				ID:      poolID,
				ownerID: ownerID,
				partyID: partyID,
				mesos:   mesos,

				expireTime:  expireTime,
				timeoutTime: timeoutTime,
				neverExpire: false,

				originPos: dropFrom,
				finalPos:  finalPos,

				dropType: dropType,
			}

			pool.drops[drop.ID] = drop

			pool.instance.Send(packetShowDrop(spawnType, drop))
		}
	}
}

// Update logic for the pool e.g. drops disappear
func (pool *Data) Update(t time.Time) {
	id := make([]int32, 0, len(pool.drops))

	currentTime := time.Now().Unix()

	for _, v := range pool.drops {
		if v.expireTime <= currentTime {
			id = append(id, v.ID)
		}
	}

	if len(id) > 0 {
		pool.RemoveDrop(false, id...)
	}
}
