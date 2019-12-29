package room

import (
	"fmt"
	"math"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type player interface {
	ID() int32
	Conn() mnet.Client
	Send(p mpacket.Packet)
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

// Room base behaviours
type Room interface {
	ID() int32
	AddPlayer(player) bool
	Closed() bool
	Present(int32) bool
	ChatMsg(player, string)
	OwnerID() int32
}

type room struct {
	id      int32
	ownerID int32
	players []player
}

func (r room) ID() int32 {
	return r.id
}

// OwnerID of the room
func (r room) OwnerID() int32 {
	return r.ownerID
}

func (r *room) addPlayer(plr player) bool {
	if len(r.players) == 0 {
		r.ownerID = plr.ID()
	} else if len(r.players) == 2 {
		plr.Send(packetRoomFull())
		return false
	}

	for _, v := range r.players {
		if v == plr {
			return false
		}
	}

	r.players = append(r.players, plr)

	return true
}

func (r *room) removePlayer(plr player) bool {
	for i, v := range r.players {
		if v.Conn() == plr.Conn() {
			r.players = append(r.players[:i], r.players[i+1:]...) // preserve order for slot numbers
			return true
		}
	}

	return false
}

func (r room) send(p mpacket.Packet) {
	for _, v := range r.players {
		v.Send(p)
	}
}

func (r room) sendExcept(p mpacket.Packet, plr player) {
	for _, v := range r.players {
		if v.Conn() == plr.Conn() {
			continue
		}
		v.Send(p)
	}
}

func (r room) Closed() bool {
	if len(r.players) == 0 {
		return true
	}

	return false
}

func (r room) ChatMsg(plr player, msg string) {
	for i, v := range r.players {
		if v.Conn() == plr.Conn() {
			r.send(packetRoomChat(plr.Name(), msg, byte(i)))
		}
	}
}

// Present checks that a player with the id passed is in the room
func (r room) Present(id int32) bool {
	for _, v := range r.players {
		if v.ID() == id {
			return true
		}
	}

	return false
}

// Game base behaviours
type Game interface {
	Ready(player)
	Unready(player)
	Start()
	DisplayBytes() []byte
	KickPlayer(player, byte) bool
	Expel() bool
	ChangeTurn()
	RequestTie(player)
	RequestTieResult(bool, player)
	Forfeit(player)
	RequestExit(bool, player)
}

type game struct {
	room

	roomType   byte
	boardType  byte
	ownerStart bool
	p1Turn     bool
	inProgress bool
	name       string
	password   string
	exit       [2]bool
}

// AddPlayer to game
func (r *game) AddPlayer(plr player) bool {
	if !r.room.addPlayer(plr) {
		return false
	}

	plr.Send(packetRoomShowWindow(r.roomType, r.boardType, byte(maxPlayers), byte(len(r.players)-1), r.name, r.players))

	if len(r.players) > 1 {
		r.sendExcept(packetRoomJoin(r.roomType, byte(len(r.players)-1), r.players[len(r.players)-1]), plr)
	}

	return true
}

// KickPlayer from game
func (r *game) KickPlayer(plr player, reason byte) bool {
	for i, v := range r.players {
		if v.Conn() == plr.Conn() {
			if r.inProgress {
				r.gameEnd(false, true, plr)
			}

			if !r.room.removePlayer(plr) {
				return false
			}

			plr.Send(packetRoomLeave(byte(i), reason))

			if i == 0 { // owner is always at index 0
				for j := range r.players {
					fmt.Println(packetRoomLeave(byte(j+1), 0x0))
					r.send(packetRoomLeave(byte(j+1), 0x0))
				}
				r.players = []player{} // sets the room into a closed state
			} else {
				fmt.Println(packetRoomLeave(byte(i), reason))
				r.send(packetRoomLeave(byte(i), reason))
			}

			return true
		}
	}

	return false
}

func (r *game) Expel() bool {
	if len(r.players) > 1 {
		r.send(packetRoomYellowChat(0, r.players[1].Name()))
		r.KickPlayer(r.players[1], 0x5)

		return true
	}

	return false
}

// Ready button pressed
func (r *game) Ready(plr player) {
	for i, v := range r.players {
		if v.Conn() == plr.Conn() && i == 1 {
			r.send(packetRoomReady())
		}
	}
}

// Unready button pressed
func (r *game) Unready(plr player) {
	for i, v := range r.players {
		if v.Conn() == plr.Conn() && i == 1 {
			r.send(packetRoomUnready())
		}
	}
}

// ChangeTurn of player
func (r *game) ChangeTurn() {
	r.p1Turn = !r.p1Turn
	r.send(packetRoomGameSkip(r.p1Turn))
}

func (r *game) gameEnd(draw, forfeit bool, plr player) {
	r.inProgress = false

	var winningSlot byte = 0x00

	if !r.p1Turn {
		winningSlot = 0x01
	}

	if forfeit {
		if plr.Conn() == r.players[0].Conn() {
			winningSlot = 0x01
		} else {
			winningSlot = 0x00
		}
	}

	r.assignPoints(draw, winningSlot)
	r.assignWinLossDraw(draw, winningSlot)
	r.send(packetRoomGameResult(draw, winningSlot, forfeit, r.players))

	if r.exit[0] == true {
		for _, v := range r.players {
			r.KickPlayer(v, 0)
		}
	} else if r.exit[1] == true {
		r.KickPlayer(r.players[1], 0)
		r.exit[1] = false // no need to clear owner entry, if they leave room closes
	}
}

func (r *game) assignWinLossDraw(draw bool, winningSlot byte) {
	if draw {
		r.players[0].SetMiniGameDraw(r.players[0].MiniGameDraw() + 1)
		r.players[1].SetMiniGameDraw(r.players[1].MiniGameDraw() + 1)
	} else {
		r.players[winningSlot].SetMiniGameWins(r.players[winningSlot].MiniGameWins() + 1)

		if winningSlot == 0x00 {
			r.players[1].SetMiniGameLoss(r.players[1].MiniGameLoss() + 1)
		} else {
			r.players[0].SetMiniGameLoss(r.players[0].MiniGameLoss() + 1)
		}
	}
}

// TODO: Correct points/elo calculation
func (r *game) assignPoints(draw bool, winningSlot byte) {
	// This is not the correct calculation
	diff := math.Abs(float64(r.players[0].MiniGamePoints() - r.players[1].MiniGamePoints()))
	pointChange := 17 - int32(diff/27)

	if draw {

	} else {
		r.players[winningSlot].SetMiniGamePoints(r.players[winningSlot].MiniGamePoints() + pointChange)

		if winningSlot == 0x00 {
			r.players[1].SetMiniGamePoints(r.players[1].MiniGamePoints() - pointChange)
		} else {
			r.players[0].SetMiniGamePoints(r.players[0].MiniGamePoints() - pointChange)
		}
	}

}

// ReqestTie of game
func (r *game) RequestTie(plr player) {
	for _, v := range r.players {
		if v.Conn() != plr.Conn() {
			v.Send(packetRoomRequestTie())
			return
		}
	}
}

// RequestTieResult of the choice the other player mode
func (r *game) RequestTieResult(tie bool, plr player) {
	if tie == true {
		r.gameEnd(true, false, nil)
	} else {
		for _, v := range r.players {
			if v.Conn() != plr.Conn() {
				v.Send(packetRoomRejectTie())
				return
			}
		}
	}
}

// Forfeit the game
func (r *game) Forfeit(plr player) {
	for _, v := range r.players {
		if v.Conn() == plr.Conn() {
			r.gameEnd(false, true, plr)
			return
		}
	}
}

// RequestExit at end of game
func (r *game) RequestExit(exit bool, plr player) {
	for i, v := range r.players {
		if v.Conn() == plr.Conn() {
			r.exit[i] = exit
			return
		}
	}
}

// DisplayBytes to show room game box
func (r game) DisplayBytes() []byte {
	p := mpacket.NewPacket()

	p.WriteInt32(r.players[0].ID())
	p.WriteByte(r.roomType)
	p.WriteInt32(r.id)
	p.WriteString(r.name)
	p.WriteBool(len(r.password) > 0)
	p.WriteByte(r.boardType)
	p.WriteByte(byte(len(r.players))) // number that is seen in the box? Player count?
	p.WriteByte(2)                    // ?
	p.WriteBool(r.inProgress)         //Sets some korean text, does it mean game is ongoing?

	return p
}
