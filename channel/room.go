package channel

import (
	"sync"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/packets"

	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/inventory"
)

var ActiveRooms = rooms{mutex: &sync.RWMutex{}}

type rooms struct {
	active []Room
	mutex  *sync.RWMutex
}

func (t *rooms) getNextRoomID() int32 {
	nextID := int32(0)

	t.mutex.RLock()

	if len(t.active) > 0 {
		nextID = int32(len(t.active)) // if somehow we overflow and get back to zero from negative max and the first trade is still open then....
	} else {
		nextID = 0
	}

	t.mutex.RUnlock()

	return nextID + 1
}

func (t *rooms) Add(val Room) {
	val.ID = t.getNextRoomID()

	t.mutex.Lock()
	t.active = append(t.active, val)
	t.mutex.Unlock()
}

func (t *rooms) Remove(val int32) {
	index := -1
	t.mutex.RLock()
	for i, v := range t.active {
		if v.ID == val {
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
		for _, c := range v.participants {
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

func (t *rooms) OnID(id int32, action func(r *Room)) {
	t.mutex.RLock()
	for i, v := range t.active {
		if v.ID == id {
			action(&t.active[i])
			break
		}
	}
	t.mutex.RUnlock()
}

func (t *rooms) OnRoom(action func(r *Room)) {
	t.mutex.RLock()
	for i := range t.active {
		action(&t.active[i])
	}
	t.mutex.RUnlock()
}

type Room struct {
	ID           int32
	participants [2]*MapleCharacter
	sitems       [9]inventory.Item
	ritems       [9]inventory.Item
	accept       int
	Type         byte
	mutex        *sync.RWMutex
}

func CreateTradeRoom(char *MapleCharacter) {
	newRoom := Room{Type: 0x03, mutex: &sync.RWMutex{}}
	newRoom.AddParticipant(char)
	ActiveRooms.Add(newRoom)
}

func (r *Room) Broadcast(packet maplepacket.Packet) {
	r.mutex.RLock()
	for _, p := range r.participants {
		if p != nil {
			p.SendPacket(packet)
		}
	}
	r.mutex.RUnlock()
}

func (r *Room) AddParticipant(char *MapleCharacter) {
	index := -1
	r.mutex.RLock()
	for i, p := range r.participants {
		if p == nil {
			r.Broadcast(packets.RoomJoin(byte(i), char.Character))
			index = i
			break
		}
	}
	r.mutex.RUnlock()

	if index > -1 {
		r.mutex.Lock()
		r.participants[index] = char
		r.mutex.Unlock()

		displayInfo := []character.Character{}

		r.mutex.RLock()
		for _, p := range r.participants {
			if p != nil {
				displayInfo = append(displayInfo, p.Character)
			}
		}
		r.mutex.RUnlock()

		char.SendPacket(packets.RoomShowTradeWindow(byte(index), displayInfo))
	} else {
		char.SendPacket(packets.RoomFull())
	}

}

func (r *Room) RemoveParticipant(char *MapleCharacter) (bool, int32) {
	roomSlot := -1
	counter := 0

	r.mutex.Lock()
	for i, p := range r.participants {
		if p == char {
			p = nil
			roomSlot = i
			continue
		}
		counter++
	}
	r.mutex.Unlock()

	if roomSlot > -1 {
		r.mutex.RLock()
		if r.Type == 0x03 && counter == 1 {
			if r.accept > 0 {
				r.Broadcast(packets.RoomLeave(byte(roomSlot), 7))
			} else {
				r.Broadcast(packets.RoomLeave(byte(roomSlot), 2))
			}
		}
		r.mutex.RUnlock()

		return true, r.ID
	}

	return false, -1
}

func (r *Room) SendMessage(name, msg string) {
	r.mutex.RLock()
	for i, p := range r.participants {
		if p == nil || p.GetName() == name {
			r.Broadcast(packets.RoomChat(name, msg, byte(i)))
		}
	}
	r.mutex.RUnlock()
}

func (r *Room) Accept(char *MapleCharacter) (bool, int32) {
	r.mutex.Lock()
	r.accept++
	r.mutex.Unlock()

	success := false

	r.mutex.RLock()
	for _, p := range r.participants {
		if p != nil && p != char {
			p.SendPacket(packets.RoomShowAccept())
			if r.accept == 2 {
				// do trade
				// RoomLeave of 8 is for when trading unique items, use this for cash shop items as well
				// RoomLeave of 9 is for when on seperate maps, make sure to ignore if one side is a gm

				for i := range r.participants {
					r.Broadcast(packets.RoomLeave(byte(i), 6))
				}

				success = true
			}
			break
		}
	}
	r.mutex.RUnlock()

	return success, r.ID
}

func (r *Room) UpdateCharDisplay() {

}

func (r *Room) AddItem() {

}

func (r *Room) RemoveItem() {

}

func (r *Room) AddMesos() {

}
