package channel

import (
	"fmt"
	"sync"

	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/inventory"
)

var ActiveRooms = rooms{mutex: &sync.RWMutex{}}

type rooms struct {
	active []Room
	mutex  *sync.RWMutex
}

func (t *rooms) GetNextRoomID() int32 {
	previousID := int32(-1)

	t.mutex.RLock()
	for _, v := range t.active {
		fmt.Println(v)
		if v.ID != (previousID + 1) {
			break
		}
	}
	t.mutex.RUnlock()

	return previousID + 1
}

func (t *rooms) Add(val Room) {
	t.mutex.Lock()
	t.active = append(t.active, val)
	t.mutex.Unlock()
}

func (t *rooms) Remove(val Room) {
	index := -1
	t.mutex.RLock()
	for i, v := range t.active {
		if v.Participants == val.Participants {
			index = i
			break
		}
	}
	t.mutex.RUnlock()

	if index > -1 {
		t.mutex.Lock()
		t.active = append(t.active[:index], t.active[index+1:]...)
		t.mutex.Unlock()
	}

}

func (t *rooms) OnConn(conn *connection.Channel, action func(r *Room)) {
	didWork := false

	t.mutex.RLock()
	for i, v := range t.active {
		for _, c := range v.Participants {
			if c == nil {
				continue
			}

			if conn == c.GetConn() {
				action(&t.active[i])
				didWork = true
			}
		}

		if didWork {
			break
		}
	}
	t.mutex.RUnlock()
}

type Room struct {
	ID           int32
	Participants [2]*MapleCharacter
	Iitems       []inventory.Item
	Ritems       []inventory.Item
}
