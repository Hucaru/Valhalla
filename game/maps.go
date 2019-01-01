package game

import (
	"fmt"
	"math/rand"

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

func (gm *GameMap) AddPlayer(conn mnet.MConnChannel) error {
	if len(gm.instances) > 0 {
		gm.instances[0].addPlayer(conn)
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

func (gm *GameMap) Send(p mpacket.Packet) { // Assumes instance 0
	if len(gm.instances) > 0 {
		gm.instances[0].send(p)
	}
}

func (gm *GameMap) SendExcept(p mpacket.Packet, exception mnet.MConnChannel) { // Assumes instance 0
	if len(gm.instances) > 0 {
		gm.instances[0].sendExcept(p, exception)
	}
}
