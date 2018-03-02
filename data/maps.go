package data

import (
	"sync"

	"github.com/Hucaru/Valhalla/interfaces"
)

type mMap map[uint32]*mapleMap

var mapleMaps = make(mMap)
var mapleMapsMutex = &sync.RWMutex{}

// GetMapsPtr -
func GetMapsPtr() mMap {
	return mapleMaps
}

// GenerateMapsObject -
func GenerateMapsObject() {

}

func (mM mMap) GetMap(mapID uint32) interfaces.Map {
	mapleMapsMutex.RLock()
	result := mapleMaps[mapID]
	mapleMapsMutex.RUnlock()

	return result
}

type mapleMap struct {
	npcs         []interfaces.Npc
	mobs         []interfaces.Mob
	forcedReturn uint32
	returnMap    uint32
	mobRate      float64
	isTown       bool
	portals      []interfaces.Portal
	mutex        sync.RWMutex
}

func (m mapleMap) GetNps() []interfaces.Npc {
	m.mutex.RLock()
	result := m.npcs
	m.mutex.RUnlock()

	return result
}

func (m mapleMap) AddNpc(npc interfaces.Npc) {
	m.mutex.Lock()
	m.npcs = append(m.npcs, npc)
	m.mutex.Unlock()
}

func (m mapleMap) GetMobs() []interfaces.Mob {
	m.mutex.RLock()
	result := m.mobs
	m.mutex.RUnlock()

	return result
}

func (m mapleMap) AddMob(mob interfaces.Mob) {
	m.mutex.Lock()
	m.mobs = append(m.mobs, mob)
	m.mutex.Unlock()
}

func (m mapleMap) GetPortals() []interfaces.Portal {
	m.mutex.RLock()
	result := m.portals
	m.mutex.RUnlock()

	return result
}

func (m mapleMap) AddPortal(portal interfaces.Portal) {
	m.mutex.Lock()
	m.portals = append(m.portals, portal)
	m.mutex.Unlock()
}
