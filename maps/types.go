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

type mapleMap struct {
	npcs         []npc
	mobs         []mob
	forcedReturn uint32
	returnMap    uint32
	mobRate      float64
	isTown       bool
	portals      []portal
}

func createMapleMap(npc []npc,
	mobs []mob,
	forcedReturn uint32,
	returnMap uint32,
	mobRate float64,
	isTown bool,
	portals []portal) *mapleMap {

	return &mapleMap{npc,
		mobs,
		forcedReturn,
		returnMap,
		mobRate,
		isTown,
		portals}
}

func (m *mapleMap) AddNpc() {

}

func (m *mapleMap) GetNpcs() {

}

func (m *mapleMap) AddMob() {

}

func (m *mapleMap) GetMobs() {

}

func (m *mapleMap) AddPortal() {

}

func (m *mapleMap) GetPortal() {

}
