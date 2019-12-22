package room

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type player interface {
	Conn() mnet.Client
	Send(p mpacket.Packet)
	Name() string
	DisplayBytes() []byte
	MiniGameWins() int32
	MiniGameDraw() int32
	MiniGameLoss() int32
}

// Room interface to the base room struct
type Room interface {
	ID() int32
	Close()
	AddPlayer(p player) bool
	RemovePlayer(conn mnet.Client) bool
	Send(p mpacket.Packet)
	Closed() bool
}

type room struct {
	id      int32
	players []player
}

func (r room) ID() int32 {
	return r.id
}

func (r *room) AddPlayer(p player) bool {
	for _, v := range r.players {
		if v == p {
			return false
		}
	}

	r.players = append(r.players, p)

	return true
}

func (r *room) RemovePlayer(conn mnet.Client) bool {
	for i, v := range r.players {
		if v.Conn() == conn {
			r.players[i] = r.players[len(r.players)-1]
			r.players = r.players[:len(r.players)-1]
			return true
		}
	}

	return false
}

func (r room) Send(p mpacket.Packet) {
	for _, v := range r.players {
		v.Send(p)
	}
}

func (r room) Closed() bool {
	if len(r.players) == 0 {
		return true
	}

	return false
}
