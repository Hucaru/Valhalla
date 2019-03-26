package entity

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

var Maps = make(map[int32]*GameMap)

type GameMap struct {
	id           int32
	instances    []*MapInstance
	mapData      nx.Map
	workDispatch chan func()

	vrlimit, mbr, ombr             mapRectangle
	mobCapacityMin, mobCapacityMax int
}

func InitMaps(dispatcher chan func()) {
	for mapID, nxMap := range nx.GetMaps() {

		Maps[mapID] = &GameMap{
			id:           mapID,
			mapData:      nxMap,
			workDispatch: dispatcher,
		}

		Maps[mapID].calculateMapLimits()
		Maps[mapID].CreateNewInstance()
	}
}

func (gm *GameMap) CreateNewInstance() int {
	inst := createInstanceFromMapData(gm.mapData, gm.id, gm.workDispatch, gm.mobCapacityMin, gm.mobCapacityMax)
	gm.instances = append(gm.instances, inst)
	return len(gm.instances) - 1
}

func (gm *GameMap) DeleteInstance(instance int) error {
	if len(gm.instances) > 0 {
		if instance == 0 {
			return fmt.Errorf("Not allowed to delete instance zero")
		}

		if instance < len(gm.instances) {
			if len(gm.instances[instance].players) > 0 {
				return fmt.Errorf("Cannot delete instance whilst players are present in it")
			}

			gm.instances = append(gm.instances[:instance], gm.instances[instance+1:]...)
			return nil
		}
	}

	return fmt.Errorf("Unable to delete instance")
}

func (gm *GameMap) GetNumberOfInstances() int {
	return len(gm.instances)
}

func (gm *GameMap) AddPlayer(conn mnet.Client, instance int) error {
	if len(gm.instances) > 0 {
		if instance < len(gm.instances) {
			gm.instances[instance].addPlayer(conn)
			return nil
		}

		Players[conn].InstanceID = 0
		gm.instances[0].addPlayer(conn)
		return nil
	}

	return fmt.Errorf("Unable to add player to map as there are no instances")
}

func (gm *GameMap) RemovePlayer(conn mnet.Client) {
	gm.instances[Players[conn].InstanceID].removePlayer(conn)
}

func (gm *GameMap) GetRandomSpawnPortal() (nx.Portal, byte, error) {
	portals := []nx.Portal{}
	inds := []int{}

	nxMap, err := nx.GetMap(gm.id)

	if err != nil {
		return nx.Portal{}, 0, fmt.Errorf("Invalid map id")
	}

	for i, p := range nxMap.Portals {
		if p.Pn == "sp" {
			portals = append(portals, p)
			inds = append(inds, i)
		}
	}

	ind := rand.Intn(len(portals))
	return portals[ind], byte(inds[ind]), nil
}

func (gm *GameMap) GetPlayers(instance int) ([]mnet.Client, error) {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		return gm.instances[instance].players, nil
	}

	return []mnet.Client{}, fmt.Errorf("Unable to get players")
}

func (gm *GameMap) GetMobs(instance int) ([]Mob, error) {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		return gm.instances[instance].mobs, nil
	}

	return nil, fmt.Errorf("Unable to get mobs")
}

func (gm *GameMap) GetMobFromSpawnID(spawnID int32, instance int) (*Mob, error) {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		for i, v := range gm.instances[instance].mobs {
			if v.SpawnID == spawnID {
				return &gm.instances[instance].mobs[i], nil
			}
		}
	}

	return nil, fmt.Errorf("Unable to get mob")
}

func (gm *GameMap) SpawnMob(mobID int32, pos Pos, fh int16, facesLeft bool, instance int) {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		inst := gm.instances[instance]

		inst.SpawnMob(mobID, inst.generateMobSpawnID(), pos.X, pos.Y, fh, -2, 0, facesLeft)
	}
}

func (gm *GameMap) GetNpcFromSpawnID(spawnID int32, instance int) (*Npc, error) {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		for i, v := range gm.instances[instance].npcs {
			if v.SpawnID == spawnID {
				return &gm.instances[instance].npcs[i], nil
			}
		}
	}

	return nil, fmt.Errorf("Unable to get npc")
}

func (gm *GameMap) HandleDeadMobs(instance int) {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		gm.instances[instance].handleDeadMobs()
	}
}

func (gm *GameMap) FindControllerExcept(conn mnet.Client, instance int) mnet.Client {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		return gm.instances[instance].findControllerExcept(conn)
	}

	return nil
}

func (gm *GameMap) Send(p mpacket.Packet, instance int) {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		gm.instances[instance].send(p)
	}
}

func (gm *GameMap) SendExcept(p mpacket.Packet, exception mnet.Client, instance int) {
	if len(gm.instances) > 0 && instance < len(gm.instances) {
		gm.instances[instance].sendExcept(p, exception)
	}
}

type mapRectangle struct {
	left, top, right, bottom int
}

func (r mapRectangle) empty() bool {
	if (r.left | r.top | r.right | r.bottom) == 0 {
		return true
	}

	return false
}

func (r *mapRectangle) inflate(x, y int) {
	r.left -= x
	r.top += y
	r.right += x
	r.bottom -= y
}

func (r mapRectangle) width() int {
	return r.right - r.left
}

func (r mapRectangle) height() int {
	return r.top - r.bottom
}

func (r mapRectangle) contains(x, y int) bool {
	if r.left > x {
		return false
	}

	if r.top < y {
		return false
	}

	if r.right < x {
		return false
	}

	if r.bottom > y {
		return false
	}

	return true
}

func (gm *GameMap) calculateMapLimits() {
	left, top := math.MaxInt32, math.MaxInt32
	right, bottom := math.MinInt32, math.MinInt32

	for _, fh := range gm.mapData.Footholds { //edge adjustments
		if fh.X1 < left {
			left = fh.X1
		}

		if fh.Y1 < top {
			top = fh.Y1
		}

		if fh.X2 < left {
			left = fh.X2
		}

		if fh.Y2 < top {
			top = fh.Y2
		}

		if fh.X1 > right {
			right = fh.X1
		}

		if fh.Y1 > bottom {
			bottom = fh.Y1
		}

		if fh.X2 > right {
			right = fh.X2
		}

		if fh.Y2 > bottom {
			bottom = fh.Y2
		}
	}

	gm.vrlimit = mapRectangle{int(gm.mapData.VRLeft), int(gm.mapData.VRTop), int(gm.mapData.VRRight), int(gm.mapData.VRBottom)}

	if gm.vrlimit.empty() {
		gm.vrlimit.left, gm.vrlimit.top, gm.vrlimit.right, gm.vrlimit.bottom = left, top-300, right, bottom+75
	}

	left += 30
	top -= 300
	right -= 30
	bottom += 10

	if !gm.vrlimit.empty() {
		if gm.vrlimit.left+20 < left {
			left = gm.vrlimit.left + 20
		}

		if gm.vrlimit.top+65 < top {
			top = gm.vrlimit.top + 65
		}

		if gm.vrlimit.right-5 > right {
			right = gm.vrlimit.right - 5
		}

		if gm.vrlimit.bottom > bottom {
			bottom = gm.vrlimit.bottom
		}
	}

	gm.mbr.left, gm.mbr.top, gm.mbr.right, gm.mbr.bottom = left+10, top-375, right-10, bottom+60
	gm.mbr.inflate(10, 10)
	gm.ombr = gm.mbr
	gm.ombr.inflate(60, 60)

	mobX, mobY := 800, 600

	if gm.mbr.width() > 800 {
		mobX = gm.mbr.width()
	}

	if gm.mbr.height() > 800 {
		mobY = gm.mbr.height()
	}

	gm.mobCapacityMin = int((float64(mobX*mobY) * gm.mapData.MobRate) * 0.0000078125)

	if gm.mobCapacityMin < 1 {
		gm.mobCapacityMin = 1
	} else if gm.mobCapacityMin > 40 {
		gm.mobCapacityMin = 40
	}

	gm.mobCapacityMax = gm.mobCapacityMin * 2
}
