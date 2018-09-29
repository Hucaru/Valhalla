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
	// Move these to be under mutex protection
	ID         int32
	MapID      int32 // Change to retrieve from players
	P1Turn     bool
	InProgress bool

	RoomType     byte
	maxPlayers   byte
	participants [4]*MapleCharacter

	sItems, rItems [9]inventory.Item
	sMesos, rMesos int32
	accept         int

	name, password string
	boardType      byte
	board          [15][15]byte
	leaveAfterGame [2]bool
	lastTurnP1     [2]int32
	lastTurnP2     [2]int32

	mutex *sync.RWMutex
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

func (r *Room) GetBoardType() byte {
	r.mutex.RLock()
	boardType := r.boardType
	r.mutex.RUnlock()

	return boardType
}

func (r *Room) GetParticipantFromSlot(slotId byte) *MapleCharacter {
	r.mutex.RLock()

	var char *MapleCharacter

	if slotId < r.maxPlayers && r.participants[slotId] != nil {
		char = r.participants[slotId]
	}

	r.mutex.RUnlock()

	return char
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

		char.SendPacket(packets.RoomShowWindow(r.RoomType, r.boardType, r.maxPlayers, byte(index), r.name, displayInfo))

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
			var ammount byte = 0x1

			if r.InProgress {
				ammount = 2
			}

			p = packets.RoomShowMapBox(r.participants[0].GetCharID(), r.ID, r.RoomType, r.boardType, r.name, hasPassword, r.InProgress, ammount)
			valid = true
		}
	}
	r.mutex.RUnlock()

	return p, valid
}

func (r *Room) removeParticipant(char *MapleCharacter) (int, byte) {
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

	return roomSlot, counter
}

func (r *Room) RemoveParticipant(char *MapleCharacter, msgCode byte) (bool, int32) {
	roomSlot, counter := r.removeParticipant(char)
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
				char.SendPacket(packets.RoomLeave(byte(roomSlot), msgCode))
				r.Broadcast(packets.RoomLeave(byte(roomSlot), msgCode))

				if msgCode == 5 {
					r.Broadcast(packets.RoomYellowChat(0, char.GetName()))
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

func (r *Room) GetP1Turn() bool {
	r.mutex.RLock()
	result := r.P1Turn
	r.mutex.RUnlock()

	return result
}

func (r *Room) GetSlotIDFromChar(char *MapleCharacter) byte {
	id := byte(0)
	r.mutex.RLock()
	for i, v := range r.participants {
		if v == char {
			id = byte(i)
		}
	}
	r.mutex.RUnlock()

	return id
}

func (r *Room) AddLeave(char *MapleCharacter) {
	id := r.GetSlotIDFromChar(char)

	r.mutex.Lock()
	r.leaveAfterGame[id] = true
	r.mutex.Unlock()
}

func (r *Room) ChangeTurn() {
	r.mutex.Lock()
	r.P1Turn = !r.P1Turn
	r.mutex.Unlock()
}

func (r *Room) UndoTurn(p1 bool) { // figure out packet for this
	if p1 {
		r.mutex.Lock()
		r.board[r.lastTurnP1[0]][r.lastTurnP1[1]] = 0
		r.mutex.Unlock()
	} else {
		r.mutex.Lock()
		r.board[r.lastTurnP2[0]][r.lastTurnP2[1]] = 0
		r.mutex.Unlock()
	}
}

func (r *Room) PlacePiece(x, y int32, piece byte) {
	if r.board[x][y] != 0 {
		if r.P1Turn {
			r.participants[0].SendPacket(packets.RoomOmokInvalidPlaceMsg())
		} else {
			r.participants[1].SendPacket(packets.RoomOmokInvalidPlaceMsg())
		}

		return
	}

	var slotId byte

	if r.GetP1Turn() == true {
		r.mutex.RLock()
		r.board[x][y] = piece
		r.lastTurnP1[0] = x
		r.lastTurnP1[1] = y
		r.mutex.RUnlock()
		slotId = 0
	} else {
		r.mutex.RLock()
		r.board[x][y] = piece
		r.lastTurnP2[0] = x
		r.lastTurnP2[1] = y
		r.mutex.RUnlock()
		slotId = 1
	}

	r.mutex.RLock()
	win := checkOmokWin(r.board, piece)
	r.mutex.RUnlock()

	draw := false

	if !win {
		r.mutex.RLock()
		draw = checkOmokDraw(r.board)
		r.mutex.RUnlock()
	}

	r.Broadcast(packets.RoomPlaceOmokPiece(x, y, piece))

	if win || draw {
		r.GameEnd(draw, slotId, false)
	}

	r.ChangeTurn()
}

func (r *Room) GameEnd(draw bool, slotID byte, forfeit bool) {
	// Update the map box for current map players
	p, _ := r.GetBox()

	r.mutex.RLock()
	r.InProgress = false
	Maps.GetMap(r.participants[0].GetCurrentMap()).SendPacket(p)
	r.mutex.RUnlock()

	// Update players records
	if forfeit {
		if slotID == 1 { // for forfeits slot id is inversed
			r.participants[0].SetOmokLosses(r.participants[0].GetOmokLosses() + 1)
			r.participants[1].SetOmokWins(r.participants[1].GetOmokWins() + 1)
		} else {
			r.participants[1].SetOmokLosses(r.participants[1].GetOmokLosses() + 1)
			r.participants[0].SetOmokWins(r.participants[0].GetOmokWins() + 1)
		}

	} else if draw {
		r.participants[0].SetOmokTies(r.participants[0].GetOmokTies() + 1)
		r.participants[1].SetOmokTies(r.participants[1].GetOmokTies() + 1)
	} else {
		r.participants[slotID].SetOmokWins(r.participants[slotID].GetOmokWins() + 1)

		if slotID == 1 {
			r.participants[0].SetOmokLosses(r.participants[0].GetOmokLosses() + 1)
		} else {
			r.participants[1].SetOmokLosses(r.participants[1].GetOmokLosses() + 1)
		}

	}

	chars := make([]character.Character, 0)

	r.mutex.RLock()
	for i := 0; i < 2; i++ {
		if r.participants[i] != nil {
			chars = append(chars, r.participants[i].Character)
		}
	}
	r.mutex.RUnlock()

	r.Broadcast(packets.RoomGameResult(draw, slotID, forfeit, chars))

	r.mutex.Lock()
	r.board = [15][15]byte{}
	r.mutex.Unlock()

	// Remove players registered to leave
	if r.leaveAfterGame[1] == true {
		r.RemoveParticipant(r.GetParticipantFromSlot(1), 0)
	}

	if r.leaveAfterGame[0] == true {
		char := r.participants[0]
		r.RemoveParticipant(char, 0)
		char.SendPacket(packets.RoomLeave(byte(0), 0))
	}
}

func checkOmokDraw(board [15][15]byte) bool {
	for i := 0; i < 15; i++ {
		for j := 0; j < 15; j++ {
			if board[i][j] > 0 {
				return false
			}
		}
	}

	return true
}

func checkOmokWin(board [15][15]byte, piece byte) bool {
	// Check horizontal
	for i := 0; i < 15; i++ {
		for j := 0; j < 11; j++ {
			if board[j][i] == piece &&
				board[j+1][i] == piece &&
				board[j+2][i] == piece &&
				board[j+3][i] == piece &&
				board[j+4][i] == piece {
				return true
			}
		}
	}

	// Check vertical
	for i := 0; i < 11; i++ {
		for j := 0; j < 15; j++ {
			if board[j][i] == piece &&
				board[j][i+1] == piece &&
				board[j][i+2] == piece &&
				board[j][i+3] == piece &&
				board[j][i+4] == piece {
				return true
			}
		}
	}

	// Check diagonal 1
	for i := 4; i < 15; i++ {
		for j := 0; j < 11; j++ {
			if board[j][i] == piece &&
				board[j+1][i-1] == piece &&
				board[j+2][i-2] == piece &&
				board[j+3][i-3] == piece &&
				board[j+4][i-4] == piece {
				return true
			}
		}
	}

	// Check diagonal 2
	for i := 0; i < 11; i++ {
		for j := 0; j < 11; j++ {
			if board[j][i] == piece &&
				board[j+1][i+1] == piece &&
				board[j+2][i+2] == piece &&
				board[j+3][i+3] == piece &&
				board[j+4][i+4] == piece {
				return true
			}
		}
	}

	return false
}

func (r *Room) UpdateCharDisplay() {

}

func (r *Room) AddItem() {

}

func (r *Room) RemoveItem() {

}

func (r *Room) AddMesos() {

}
