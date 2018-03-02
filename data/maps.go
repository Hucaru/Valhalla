package data

import "sync"

type mMap map[uint32]*mapleMap

var mapleMaps = make(mMap)
var mapleMapsMutex = &sync.RWMutex{}

// GeMapsPtr -
func GeMapsPtr() mMap {
	return mapleMaps
}

// GenerateMapsObject -
func GenerateMapsObject() {

}

func (mM mMap) GetMap(mapID uint32) *mapleMap {
	mapleMapsMutex.RLock()
	result := mapleMaps[mapID]
	mapleMapsMutex.RUnlock()

	return result
}

type mapleMap struct {
	npcs         []mapleNpc
	mobs         []mapleMob
	forcedReturn uint32
	returnMap    uint32
	mobRate      float64
	isTown       bool
	portals      []maplePortal
	mutex        sync.RWMutex
}

func (m mapleMap) GetNps() []mapleNpc {
	m.mutex.RLock()
	result := m.npcs
	m.mutex.RUnlock()

	return result
}

func (m mapleMap) AddNpc(npc mapleNpc) {
	m.mutex.Lock()
	m.npcs = append(m.npcs, npc)
	m.mutex.Unlock()
}

func (m mapleMap) GetMobs() []mapleMob {
	m.mutex.RLock()
	result := m.mobs
	m.mutex.RUnlock()

	return result
}

func (m mapleMap) AddMob(mob mapleMob) {
	m.mutex.Lock()
	m.mobs = append(m.mobs, mob)
	m.mutex.Unlock()
}

func (m mapleMap) GetPortals() []maplePortal {
	m.mutex.RLock()
	result := m.portals
	m.mutex.RUnlock()

	return result
}

func (m mapleMap) AddPortal(portal maplePortal) {
	m.mutex.Lock()
	m.portals = append(m.portals, portal)
	m.mutex.Unlock()
}
