package room

import (
	"fmt"
	"math/rand"
	"time"
)

const roomTypeMemory = 0x02

// Memory behaviours
type Memory interface {
	SelectCard(byte, byte, player) bool
}

type memory struct {
	game

	cards         []byte
	firstCardPick byte
	matches       [2]int
}

// NewMemory a new memory
func NewMemory(id int32, name, password string, boardType byte) Room {
	g := game{name: name, password: password, boardType: boardType, roomType: roomTypeMemory, ownerStart: false}
	return &memory{game: g}
}

// SelectCard on the board
func (r *memory) SelectCard(turn, cardID byte, plr player) bool {
	if int(cardID) >= len(r.cards) {
		return false
	}
	if turn == 1 {
		r.firstCardPick = cardID
		r.sendExcept(packetRoomSelectCard(turn, cardID, cardID, turn), plr)
	} else if r.cards[r.firstCardPick] == r.cards[cardID] {

		var points byte = 2

		if r.p1Turn {
			r.matches[0]++
		} else {
			r.matches[1]++
			points = 3
		}

		r.send(packetRoomSelectCard(turn, cardID, r.firstCardPick, points))
		r.send(packetRoomYellowChat(0x09, plr.Name()))

		win, draw := r.checkCardWin()

		if win || draw {
			r.gameEnd(draw, false, nil)

			if r.Closed() { // If owner exit as part of game leave
				return false
			}

			return true
		}

	} else if r.p1Turn {
		r.send(packetRoomSelectCard(turn, cardID, r.firstCardPick, 0))
		r.p1Turn = !r.p1Turn
	} else {
		r.send(packetRoomSelectCard(turn, cardID, r.firstCardPick, 1))
		r.p1Turn = !r.p1Turn
	}

	return false
}

func (r *memory) checkCardWin() (bool, bool) {
	win, draw := false, false
	totalMatches := r.matches[0] + r.matches[1]

	switch r.boardType {
	case 0:
		if totalMatches == 6 {
			if r.matches[0] == r.matches[1] {
				draw = true
			} else { // current player must have won
				win = true
			}
		}
	case 1:
		if totalMatches == 10 {
			if r.matches[0] == r.matches[1] {
				draw = true
			} else {
				win = true
			}
		}
	case 2:
		if totalMatches == 15 {
			if r.matches[0] == r.matches[1] {
				draw = true
			} else {
				win = true
			}
		}
	}

	return win, draw
}

// Start memory game
func (r *memory) Start() {
	if len(r.players) < 2 {
		return
	}

	r.inProgress = true
	r.ownerStart = !r.ownerStart
	r.p1Turn = r.ownerStart
	r.matches[0], r.matches[1] = 0, 0
	r.shuffleCards()
	r.send(packetRoomMemoryStart(r.ownerStart, int32(r.boardType), r.cards))
}

func (r *memory) shuffleCards() {
	loopCounter := byte(0)
	switch r.boardType {
	case 0:
		loopCounter = 6
	case 1:
		loopCounter = 10
	case 2:
		loopCounter = 15
	default:
		fmt.Println("Cannot shuffle unkown card type")
	}

	r.cards = make([]byte, 0)

	for i := byte(0); i < loopCounter; i++ {
		r.cards = append(r.cards, i, i)
	}

	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	random.Shuffle(len(r.cards), func(i, j int) {
		r.cards[i], r.cards[j] = r.cards[j], r.cards[i]
	})
}
