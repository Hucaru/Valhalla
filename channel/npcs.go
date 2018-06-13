package channel

import (
	"sync"

	"github.com/Hucaru/Valhalla/nx"
)

var NPCs = mapleNPCs{maps: make(map[uint32][]*mapleNpc), mutex: &sync.RWMutex{}}

func GenerateNPCs() {
	for mapID, stage := range nx.Maps {

		for spawnID, life := range stage.Life {
			if !life.IsMob {
				npc := &mapleNpc{}

				npc.SetID(life.ID)
				npc.SetSpawnID(uint32(spawnID + 1))
				npc.SetX(life.X)
				npc.SetY(life.Y)
				npc.SetRx0(life.Rx0)
				npc.SetRx1(life.Rx1)
				npc.SetFoothold(life.Fh)
				npc.SetFace(life.F)

				NPCs.AddNpc(mapID, npc)
			}
		}
	}
}

type mapleNPCs struct {
	maps  map[uint32][]*mapleNpc
	mutex *sync.RWMutex
}

func (m *mapleNPCs) AddNpc(mapID uint32, newNpc *mapleNpc) {
	m.mutex.Lock()
	m.maps[mapID] = append(m.maps[mapID], newNpc)
	m.mutex.Unlock()
}

func (m *mapleNPCs) GetNpcs(mapID uint32) []*mapleNpc {
	m.mutex.RLock()
	result := m.maps[mapID]
	m.mutex.RUnlock()

	return result
}
