package game

import (
	"fmt"
	"math/rand"

	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

var Maps = make(map[int32]*GameMap)

type GameMap struct {
	id        int32
	instances []Instance
	mapData   nx.Map
}

func InitMaps() {
	for mapID, nxMap := range nx.GetMaps() {
		inst := make([]Instance, 1)
		inst[0] = createInstanceFromMapData(nxMap, mapID)

		Maps[mapID] = &GameMap{
			id:        mapID,
			instances: inst,
			mapData:   nxMap,
		}
	}
}

func (gm *GameMap) CreateNewInstance() {
	inst := createInstanceFromMapData(gm.mapData, gm.id)
	gm.instances = append(gm.instances, inst)
}

func (gm *GameMap) AddPlayer(conn mnet.MConnChannel, instance int) error {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		gm.instances[instance].addPlayer(conn)
		return nil
	}

	return fmt.Errorf("Unable to add player to map as there are no instances")
}

func (gm *GameMap) RemovePlayer(conn mnet.MConnChannel) {
	gm.instances[Players[conn].InstanceID].removePlayer(conn)
}

func (gm *GameMap) GetRandomSpawnPortal() (nx.Portal, byte) {
	portals := []nx.Portal{}
	inds := []int{}

	nxMap, _ := nx.GetMap(gm.id)

	for i, p := range nxMap.Portals {
		if p.Pn == "sp" {
			portals = append(portals, p)
			inds = append(inds, i)
		}
	}

	ind := rand.Intn(len(portals))
	return portals[ind], byte(inds[ind])
}

func (gm *GameMap) GetPlayers(instance int) ([]mnet.MConnChannel, error) {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		return gm.instances[instance].players, nil
	}

	return []mnet.MConnChannel{}, fmt.Errorf("Unable to get players")
}

func (gm *GameMap) GetMobs(instance int) ([]gameMob, error) {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		return gm.instances[instance].mobs, nil
	}

	return []gameMob{}, fmt.Errorf("Unable to get mobs")
}

func (gm *GameMap) GetMobFromSpawnID(spawnID int32, instance int) (*gameMob, error) {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		for i, v := range gm.instances[instance].mobs {
			if v.SpawnID == spawnID {
				return &gm.instances[instance].mobs[i], nil
			}
		}
	}

	return &gameMob{}, fmt.Errorf("Unable to get mob")
}

func (gm *GameMap) GetNpcFromSpawnID(spawnID int32, instance int) (*def.NPC, error) {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		for i, v := range gm.instances[instance].npcs {
			if v.SpawnID == spawnID {
				return &gm.instances[instance].npcs[i], nil
			}
		}
	}

	return &def.NPC{}, fmt.Errorf("Unable to get npc")
}

func (gm *GameMap) HandleDeadMobs(instance int) {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		gm.instances[instance].handleDeadMobs()
	}
}

func (gm *GameMap) Send(p mpacket.Packet, instance int) {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		gm.instances[instance].send(p)
	}
}

func (gm *GameMap) SendExcept(p mpacket.Packet, exception mnet.MConnChannel, instance int) {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		gm.instances[instance].sendExcept(p, exception)
	}
}
