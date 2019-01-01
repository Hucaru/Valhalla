package game

import (
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/game/packet"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

var Rooms = make(map[int32]*Room)

const (
	omokRoom     = 0x01
	memoryRoom   = 0x02
	tradeRoom    = 0x03
	personalShop = 0x04
)

type Room struct {
	RoomType byte
	players  []mnet.MConnChannel

	Name, Password string
	inProgress     bool
	board          [15][15]byte
	cards          []byte
	BoardType      byte
	leaveAfterGame [2]bool

	accepted int
	items    [2][9]Item
	mesos    [2]int32
}

func CreateMemoryRoom() *Room {
	return &Room{RoomType: memoryRoom}
}

func CreateOmokRoom() *Room {
	return &Room{RoomType: omokRoom}
}

func CreateTradeRoom() *Room {
	return &Room{RoomType: tradeRoom}
}

func (r *Room) Broadcast(p mpacket.Packet) {
	for _, v := range r.players {
		v.Send(p)
	}
}

func (r *Room) GetGameBox() {

}

func (r *Room) AddPlayer(conn mnet.MConnChannel) {
	maxPlayers := 2

	if r.RoomType == omokRoom || r.RoomType == memoryRoom {
		maxPlayers = 4
	}

	if len(r.players) == maxPlayers {
		conn.Send(packet.RoomFull())
	}

	r.players = append(r.players, conn)

	player := Players[conn]
	roomPos := byte(len(r.players)) - 1

	r.Broadcast(packet.RoomJoin(r.RoomType, roomPos, player.Char()))

	displayInfo := []def.Character{}

	for _, v := range r.players {
		displayInfo = append(displayInfo, Players[v].Char())
	}

	if len(displayInfo) > 0 {
		conn.Send(packet.RoomShowWindow(r.RoomType, r.BoardType, byte(maxPlayers), roomPos, r.Name, displayInfo))
		// update box on map
	}
}

func (r *Room) RemovePlayer(conn mnet.MConnChannel, msgCode byte) {

}
