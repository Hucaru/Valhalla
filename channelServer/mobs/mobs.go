package mobs

import (
	"sync"

	"github.com/Hucaru/Valhalla/channelServer/playerConn"
	"github.com/Hucaru/Valhalla/common/nx"
)

type monster struct {
	lifeData   nx.Life
	mobData    nx.Monster
	controller *playerConn.Conn
}

var mobsMap = make(map[uint32][]*monster)
var mobsMapMutex = &sync.RWMutex{}

func PlayerEnterMap(conn *playerConn.Conn, mapID uint32) {
	mobsMapMutex.RLock()
	_, exists := mobsMap[mapID]
	mobsMapMutex.RUnlock()

	if exists {
		return
	}

	// First time someone has entered the map, load in all the monsters it should have
	newMonster := &monster{}
	newMonster.controller = nil

	for _, v := range nx.Maps[mapID].Life {
		if v.Mob {
			newMonster.mobData = nx.Mob[v.ID]
			newMonster.lifeData = v
		}
	}

	mobsMapMutex.Lock()
	mobsMap[mapID] = append(mobsMap[mapID], newMonster)
	mobsMapMutex.Unlock()
	//server.MobsPlayerEnterMap(conn, mapID)

	//charactersOnMap := server.MapGetAllPlayers(mapID)

	// maxPos := float64(200000)
	// deltaX := float64(mob.LifeData.X - v.GetCharacter().GetX())
	// deltaY := float64(mob.LifeData.Y - v.GetCharacter().GetY())

	// scalarDelta := math.Sqrt(deltaX*deltaX + deltaY*deltaY)

	// if scalarDelta < maxPos {
	// }
}
