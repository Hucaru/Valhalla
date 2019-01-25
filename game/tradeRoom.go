package game

import (
	"github.com/Hucaru/Valhalla/game/packet"
	"github.com/Hucaru/Valhalla/mnet"
)

type TradeRoom struct {
	*baseRoom
	accepted int
	items    [2][9]Item
	mesos    [2]int32
}

func (rc *roomContainer) CreateTradeRoom() int32 {
	id := rc.getNewRoomID()

	r := &TradeRoom{}
	r.baseRoom = &baseRoom{ID: id, RoomType: RoomTypeTrade, maxPlayers: tradeMaxPlayers}

	Rooms[id] = r

	return id
}

func (r *TradeRoom) AddPlayer(conn mnet.MConnChannel) {
	_, _ = r.baseRoom.AddPlayer(conn)
}

func (r *TradeRoom) RemovePlayer(conn mnet.MConnChannel, msgCode byte) bool {
	if roomSlot := r.baseRoom.RemovePlayer(conn); roomSlot > -1 {
		if r.accepted > 0 {
			r.Broadcast(packet.RoomLeave(byte(roomSlot), 7))
		} else {
			r.Broadcast(packet.RoomLeave(byte(roomSlot), 2))
		}

		return true
	}

	return false
}

func (r *TradeRoom) Close() {

}
