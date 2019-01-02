package game

import (
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/game/packet"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type roomContainer map[int32]*Room

var Rooms = make(roomContainer)

const (
	OmokRoom     = 0x01
	MemoryRoom   = 0x02
	TradeRoom    = 0x03
	PersonalShop = 0x04
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

var roomCounter = int32(0)

func (rc *roomContainer) getNewRoomID() int32 {
	roomCounter++

	if roomCounter == 0 {
		roomCounter = 1
	}

	return roomCounter
}

func (rc *roomContainer) CreateMemoryRoom(name, password string, boardType byte) int32 {
	r := &Room{RoomType: MemoryRoom, Name: name, Password: password, BoardType: boardType}
	id := rc.getNewRoomID()
	Rooms[id] = r
	return id
}

func (rc *roomContainer) CreateOmokRoom(name, password string, boardType byte) int32 {
	r := &Room{RoomType: OmokRoom, Name: name, Password: password, BoardType: boardType}
	id := rc.getNewRoomID()
	Rooms[id] = r
	return id
}

func (rc *roomContainer) CreateTradeRoom() int32 {
	r := &Room{RoomType: TradeRoom}
	id := rc.getNewRoomID()
	Rooms[id] = r
	return id
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

	if r.RoomType == OmokRoom || r.RoomType == MemoryRoom {
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
