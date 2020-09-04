package room

const roomTypeTrade = 0x03

// Trade behaviours
type Trade interface {
	RemovePlayer(player)
	SendInvite(player)
	Reject(byte, string)
}

// Trade window
type trade struct {
	room
}

// NewTrade a trade
func NewTrade(id int32) Trade {
	r := room{id: id, roomType: roomTypeTrade}
	return &trade{room: r}
}

// AddPlayer to game
func (r *trade) AddPlayer(plr player) bool {
	if !r.room.addPlayer(plr) {
		return false
	}

	plr.Send(packetRoomShowWindow(r.roomType, 0x00, byte(maxPlayers), byte(len(r.players)-1), "", r.players))

	if len(r.players) > 1 {
		r.sendExcept(packetRoomJoin(r.roomType, byte(len(r.players)-1), r.players[len(r.players)-1]), plr)
	}

	return true
}

// RemovePlayer from trade
func (r *trade) RemovePlayer(plr player) {
	// Note: since anyone leaving the room causes it to close we don't need to remove players
	for i, v := range r.players {
		if v.Conn() != plr.Conn() {
			v.Send(packetRoomLeave(byte(i), 0x02))
		}
	}
}

// SendInvite to player
func (r trade) SendInvite(plr player) {
	plr.Send(packetRoomInvite(roomTypeTrade, r.players[0].Name(), r.id))
}

// Reject the invite
func (r trade) Reject(code byte, name string) {
	r.send(packetRoomInviteResult(code, name))
}

// InsertItem into trading window
func (r *trade) InsertItem() {

}

// AddMesos to trade window
func (r *trade) AddMesos(amount int32, plr player) {

}

// SwapItems completeing the trade
func (r *trade) SwapItems() bool {
	return true
}
