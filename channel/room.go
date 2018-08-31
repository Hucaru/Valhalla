package channel

import (
	"math"
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
	lastID int32
	mutex  *sync.RWMutex
}

func (t *rooms) getNextRoomID() int32 {
	nextID := int32(0)

	t.mutex.Lock()

	if t.lastID < math.MaxInt32 {
		t.lastID++
		nextID = t.lastID
	}

	t.mutex.Unlock()

	return nextID
}

func (t *rooms) Add(val Room) {
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
	ID             int32
	participants   [4]*MapleCharacter
	sItems, rItems [9]inventory.Item
	sMesos, rMesos int32
	name, password string
	boardType      byte
	accept         int
	RoomType       byte
	maxPlayers     byte
	MapID          int32
	mutex          *sync.RWMutex
	P1Turn         bool
	InProgress     bool
}

func CreateTradeRoom(char *MapleCharacter) {
	newRoom := Room{ID: ActiveRooms.getNextRoomID(), RoomType: 0x03, mutex: &sync.RWMutex{}, maxPlayers: 0x02}
	newRoom.AddParticipant(char)
	ActiveRooms.Add(newRoom)
}

func CreateMemoryGame(char *MapleCharacter, name, password string, boardType byte) {
	newRoom := Room{ID: ActiveRooms.getNextRoomID(), RoomType: 0x02, mutex: &sync.RWMutex{}, maxPlayers: 0x04, name: name, password: password, boardType: boardType, MapID: char.GetCurrentMap(), P1Turn: true}
	newRoom.AddParticipant(char)
	ActiveRooms.Add(newRoom)
}

func CreateOmokGame(char *MapleCharacter, name, password string, boardType byte) {
	newRoom := Room{ID: ActiveRooms.getNextRoomID(), RoomType: 0x01, mutex: &sync.RWMutex{}, maxPlayers: 0x04, name: name, password: password, boardType: boardType, MapID: char.GetCurrentMap(), P1Turn: true}
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

func (r *Room) GetPassword() string {
	r.mutex.RLock()
	password := r.password
	r.mutex.RUnlock()

	return password
}

func (r *Room) AddParticipant(char *MapleCharacter) {
	index := -1
	r.mutex.RLock()
	for i := 0; i < int(r.maxPlayers); i++ {
		if r.participants[i] == nil {
			r.Broadcast(packets.RoomJoin(r.RoomType, byte(i), char.Character))
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

		char.SendPacket(packets.RoomShowWindow(r.RoomType, r.maxPlayers, byte(index), r.name, displayInfo))

		if p, valid := r.GetBox(); valid {
			Maps.GetMap(r.MapID).SendPacket(p)
		}

		r.mutex.RUnlock()

	} else {
		char.SendPacket(packets.RoomFull())
	}

}

func (r *Room) GetBox() (maplepacket.Packet, bool) {
	p := []byte{}
	valid := false
	r.mutex.RLock()
	if r.RoomType != 0x03 {
		hasPassword := false

		if len(r.password) > 0 {
			hasPassword = true
		}

		if r.participants[0] != nil {
			p = packets.RoomShowMapBox(r.participants[0].GetCharID(), r.ID, r.RoomType, r.boardType, r.name, hasPassword, r.InProgress)
			valid = true
		}
	}
	r.mutex.RUnlock()

	return p, valid
}

func (r *Room) RemoveParticipant(char *MapleCharacter, expelled bool) (bool, int32) {
	roomSlot := -1
	counter := byte(0)

	r.mutex.Lock()
	for i := 0; i < int(r.maxPlayers); i++ {
		if r.participants[i] == char {
			r.participants[i] = nil
			roomSlot = i
		}

		if r.participants[i] == nil {
			counter++
		}
	}
	r.mutex.Unlock()

	closeRoom := false

	if roomSlot > -1 {
		r.mutex.RLock()
		if r.RoomType == 0x03 && (r.maxPlayers-counter) == 1 {
			if r.accept > 0 {
				r.Broadcast(packets.RoomLeave(byte(roomSlot), 7))
			} else {
				r.Broadcast(packets.RoomLeave(byte(roomSlot), 2))
			}
			closeRoom = true
		} else if r.RoomType == 0x01 || r.RoomType == 0x02 {
			if r.participants[0] == nil {
				// kick everyone
				Maps.GetMap(r.MapID).SendPacket(packets.RoomRemoveBox(char.GetCharID()))
				for i, c := range r.participants {
					if c != nil {
						c.SendPacket(packets.RoomLeave(byte(i), 0))
						closeRoom = true
					}
				}
			} else {
				// I think the numbers change on the box depending on how many inside?
				char.SendPacket(packets.RoomLeave(byte(roomSlot), 5))

				if expelled {
					r.Broadcast(packets.RoomYellowChat(0, char.GetName()))
				} else {
					r.Broadcast(packets.RoomLeave(byte(roomSlot), 5))
					// r.Broadcast(packets.RoomYellowChat(4, char.GetName())) // this is a yellow text of above
				}
			}
		}
		r.mutex.RUnlock()
	}

	return closeRoom, r.ID
}

func (r *Room) SendMessage(name, msg string) {
	r.mutex.RLock()
	for i, p := range r.participants {
		if p != nil || p.GetName() == name {
			r.Broadcast(packets.RoomChat(name, msg, byte(i)))
			break
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

func (r *Room) Expel(roomSlot int) {
	for i, p := range r.participants {
		if p != nil && i == roomSlot {
			r.RemoveParticipant(p, true)
			break
		}
	}
}

func (r *Room) PlacePiece(x, y int32) {
	// validate placement
	// if valid, place and check for win or other condition
	r.Broadcast(packets.RoomPlaceOmokPiece(x, y))
}

func (r *Room) UpdateCharDisplay() {

}

func (r *Room) AddItem() {

}

func (r *Room) RemoveItem() {

}

func (r *Room) AddMesos() {

}
