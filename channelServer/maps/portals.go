package maps

import (
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/common/nx"
	"golang.org/x/exp/rand"
)

var closedPortalist = make(map[uint32][]string)
var closedPortalListMutex = &sync.RWMutex{}

func IsPortalOpen(mapID uint32, name string) bool {
	closedPortalListMutex.RLock()

	if val, ok := closedPortalist[mapID]; ok {
		for _, v := range val {
			if v == name {
				return false
			}
		}
	}

	closedPortalListMutex.RUnlock()

	return true
}

func ClosePortal(mapID uint32, name string) {
	inList := !IsPortalOpen(mapID, name)

	if !inList {
		closedPortalListMutex.Lock()

		closedPortalist[mapID] = append(closedPortalist[mapID], name)

		closedPortalListMutex.Unlock()
	}
}

func IsValidPortal(mapID uint32, name string) bool {
	for _, v := range nx.Maps[mapID].Portals {
		if v.Name == name {
			return true
		}
	}
	return false
}

func GetRandomSpawnPortal(mapID uint32) nx.Portal {
	var portals []nx.Portal
	for i := range nx.Maps[mapID].Portals {
		if nx.Maps[mapID].Portals[i].IsSpawn {
			portals = append(portals, nx.Maps[mapID].Portals[i])
		}
	}

	rand.Seed(uint64(time.Now().Unix()))

	return portals[rand.Int()%len(portals)]
}

func GetPortalByID(mapID uint32, portalID byte) nx.Portal {
	for i := range nx.Maps[mapID].Portals {
		if nx.Maps[mapID].Portals[i].ID == portalID {
			return nx.Maps[mapID].Portals[i]
		}
	}

	return GetRandomSpawnPortal(mapID)
}

func GetPortalByName(mapID uint32, name string) nx.Portal {
	for i := range nx.Maps[mapID].Portals {
		if nx.Maps[mapID].Portals[i].Name == name {
			return nx.Maps[mapID].Portals[i]
		}
	}

	return GetRandomSpawnPortal(mapID)
}

func GetSpawnPortal(mapID uint32, portalID byte) nx.Portal {
	for i := range nx.Maps[mapID].Portals {
		if nx.Maps[mapID].Portals[i].IsSpawn && nx.Maps[mapID].Portals[i].ID == portalID {
			return nx.Maps[mapID].Portals[i]
		}
	}

	return GetRandomSpawnPortal(mapID)
}
