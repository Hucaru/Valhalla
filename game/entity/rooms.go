package entity

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type RoomType byte

type roomContainer map[int32]Room

var Rooms = make(roomContainer)

const (
	RoomTypeOmok         RoomType = 0x01
	RoomTypeMemory       RoomType = 0x02
	RoomTypeTrade        RoomType = 0x03
	RoomTypePersonalShop RoomType = 0x04
	omokMaxPlayers                = 2 // can these rooms also have two observers in a queue?
	memoryMaxPlayers              = 2 // can these rooms also have two observers in a queue?
	tradeMaxPlayers               = 2
)

type Room interface {
	Broadcast(p mpacket.Packet)
	SendMessage(name, msg string)
	AddPlayer(conn mnet.Client)
	RemovePlayer(conn mnet.Client, msgCode byte) bool
}

type baseRoom struct {
	ID         int32
	players    []mnet.Client
	RoomType   RoomType
	maxPlayers int
}

var roomCounter = int32(0)

func (rc *roomContainer) getNewRoomID() int32 {
	roomCounter++

	if roomCounter == 0 {
		roomCounter = 1
	}

	return roomCounter
}

func (r *baseRoom) Broadcast(p mpacket.Packet) {
	for _, v := range r.players {
		v.Send(p)
	}
}

func (r *baseRoom) SendMessage(name, msg string) {
	for roomSlot, v := range r.players {
		if Players[v].Char().Name == name {
			r.Broadcast(PacketRoomChat(name, msg, byte(roomSlot)))
			break
		}
	}
}

func (r *baseRoom) AddPlayer(conn mnet.Client) (byte, bool) {
	if len(r.players) == r.maxPlayers {
		conn.Send(PacketRoomFull())
		return 0, false
	}

	r.players = append(r.players, conn)

	player := Players[conn]
	roomPos := byte(len(r.players)) - 1

	r.Broadcast(PacketRoomJoin(byte(r.RoomType), roomPos, player.Char()))

	Players[conn].RoomID = r.ID

	return roomPos, true
}

func (r *baseRoom) RemovePlayer(conn mnet.Client) int {
	roomSlot := -1

	for i, v := range r.players {
		if v == conn {
			roomSlot = i
			break
		}
	}

	if roomSlot < 0 {
		return roomSlot
	}

	player := Players[conn]
	player.RoomID = 0

	return roomSlot
}
