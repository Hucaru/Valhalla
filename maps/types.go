package maps

import (
	"github.com/Hucaru/gopacket"
)

type mob interface {
	Show() gopacket.Packet
	Spawn() gopacket.Packet
	UpdatePosition(int16, int16, int16)
	RemoveController()
	AddController()
	IsController() bool
}

type npc interface {
}

type portal interface {
}
