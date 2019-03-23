package game

import (
	"github.com/Hucaru/Valhalla/game/def"
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
	pos, ok := r.baseRoom.AddPlayer(conn)

	if !ok {
		return
	}
	displayInfo := []def.Character{}

	for _, v := range r.players {
		displayInfo = append(displayInfo, Players[v].Char())
	}

	conn.Send(PacketRoomShowWindow(byte(RoomTypeTrade), 0, 2, pos, "", displayInfo))
}

func (r *TradeRoom) RemovePlayer(conn mnet.MConnChannel, msgCode byte) bool {
	if roomSlot := r.baseRoom.RemovePlayer(conn); roomSlot > -1 {
		if r.accepted > 0 {
			r.Broadcast(PacketRoomLeave(byte(roomSlot), 7))
		} else {
			r.Broadcast(PacketRoomLeave(byte(roomSlot), 2))
		}

		for _, v := range r.players {
			player, err := Players.GetFromConn(v)

			if err != nil {
				continue
			}

			player.RoomID = 0
		}

		return true
	}

	return false
}
