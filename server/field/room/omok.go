package room

import (
	"github.com/Hucaru/Valhalla/mnet"
)

const roomTypeOmok = 0x01

// Omok interface to the omok struct
type Omok interface {
}

type omok struct {
	room

	name     string
	password string

	boardType    byte
	board        [15][15]byte
	previousTurn [2][2]int32
}

const maxPlayers = 2

// NewOmok returns an interface of Omok
func NewOmok(id int32, name, password string, boardType byte) Omok {
	return &omok{name: name, password: password, boardType: boardType}
}

func (r *omok) AddPlayer(p player) bool {
	if !r.room.AddPlayer(p) {
		return false
	}

	p.Send(packetRoomShowWindow(roomTypeOmok, r.boardType, byte(maxPlayers), byte(len(r.players)-1), r.name, r.players))

	return true
}

func (r *omok) RemovePlayer(conn mnet.Client) bool {
	if !r.room.RemovePlayer(conn) {
		return false
	}

	return true
}
