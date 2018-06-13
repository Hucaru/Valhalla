package channel

import (
	"sync"

	"github.com/Hucaru/Valhalla/nx"
)

var Maps = mapleMaps{maps: make(map[uint32]*mapleMap), mutex: &sync.RWMutex{}}

func GenerateMaps() {
	for mapID, stage := range nx.Maps {
		m := mapleMap{mutex: &sync.RWMutex{}}

		m.SetReturnMap(stage.ReturnMap)

		for _, portal := range stage.Portals {
			p := maplePortal{}

			p.SetName(portal.Name)
			p.SetX(portal.X)
			p.SetY(portal.Y)
			p.SetIsSpawn(portal.IsSpawn)
			p.SetToMap(portal.Tm)
			p.SetToPortal(portal.Tn)

			m.AddPortal(p)
		}

		Maps.AddMap(mapID, &m)
	}
}

func init() {
	// setup any timers e.g. boats
}

type mapleMaps struct {
	maps  map[uint32]*mapleMap
	mutex *sync.RWMutex
}

func (m *mapleMaps) AddMap(mapID uint32, newMap *mapleMap) {
	m.mutex.Lock()
	m.maps[mapID] = newMap
	m.mutex.Unlock()
}

func (m *mapleMaps) GetMap(mapID uint32) *mapleMap {
	m.mutex.RLock()
	result := m.maps[mapID]
	m.mutex.RUnlock()

	return result
}
