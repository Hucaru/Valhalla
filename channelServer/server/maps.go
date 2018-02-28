package server

import (
	"log"
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/Valhalla/common/nx"
	"github.com/Hucaru/gopacket"
	"golang.org/x/exp/rand"
)

type playerHandle interface {
	Write([]byte)
	GetCharacter() *character.Character
}

var maps = make(map[uint32]*mapleMap)
var mapsMutex = &sync.RWMutex{}

func PlayerEnterMap(player *playerHandle, mapID uint32) {
	mapsMutex.RLock()
	if _, exists := maps[mapID]; !exists {
		return
	}
	mapsMutex.RUnlock()

	mapsMutex.Lock()
	maps[mapID].addNewPlayer(player)
	mapsMutex.Unlock()

	// show mobs, npcs and other players
}

func PlayerLeaveMap(player *playerHandle, mapID uint32) {
	mapsMutex.RLock()
	if _, exists := maps[mapID]; !exists {
		return
	}
	mapsMutex.RUnlock()

	mapsMutex.Lock()
	maps[mapID].removePlayer(player)
	mapsMutex.Unlock()
}

func GetRandomSpawnPortal(mapID uint32) (pos byte, x int16, y int16) {
	mapsMutex.RLock()
	if _, exists := maps[mapID]; !exists {
		return
	}
	mapsMutex.RUnlock()

	mapsMutex.RLock()
	pos, x, y = maps[mapID].getRandomSpawnPortal()
	mapsMutex.RUnlock()

	return pos, x, y
}

type mapleMap struct {
	id        uint32
	mobs      []mapleMob
	npcs      []mapleNpc
	portals   []maplePortal
	players   []*playerHandle
	returnMap uint32
	isTown    bool
	mobRate   float64
	mutex     *sync.RWMutex
}

// CreateMap - returns a new map struct containing all the map information for a given map ID
func createMap(mapID uint32) *mapleMap {
	nxMap := nx.Maps[mapID]

	mobs := make([]mapleMob, 0)
	npcs := make([]mapleNpc, 0)
	portals := make([]maplePortal, 0)

	for i, v := range nxMap.Life {
		if v.Mob {
			m := nx.Mob[v.ID]

			mobs = append(mobs, mapleMob{
				v.ID,
				uint32(i),
				v.X,
				v.Y,
				v.Fh,
				v.X,
				v.Y,
				v.F,
				nil,
				m.Boss,
				m.Accuracy,
				m.Exp,
				m.Level,
				m.MaxHp,
				m.Hp,
				m.MaxMp,
				m.Mp})

		} else if v.Npc {

			npcs = append(npcs, mapleNpc{
				v.ID,
				uint32(i),
				v.X,
				v.Y,
				v.Fh,
				v.X,
				v.Y,
				v.F,
			})

		} else {
			log.Println("Unknown life type in map creation for map ID:", mapID)
		}
	}

	for _, v := range nxMap.Portals {
		portals = append(portals, maplePortal{
			v.ID,
			v.Tm,
			v.Tn,
			v.Pt,
			v.IsSpawn,
			v.X,
			v.Y,
			v.Name})
	}

	return &mapleMap{
		mapID,
		mobs,
		npcs,
		portals,
		make([]*playerHandle.Conn, 0),
		nxMap.ReturnMap,
		nxMap.IsTown,
		nxMap.MobRate,
		&sync.RWMutex{}}
}

func (this *mapleMap) isPlayerInMap(player *playerHandle) bool {
	present := false

	this.mutex.RLock()
	for i := range this.players {
		if this.players[i] == player {
			present = true
			break
		}
	}
	this.mutex.RUnlock()

	return present
}

func (this *mapleMap) isPlayerNameInMap(name string) bool {
	present := false

	this.mutex.RLock()
	for i := range this.players {
		if this.players[i].GetCharacter().GetName() == name {
			present = true
			break
		}
	}
	this.mutex.RUnlock()

	return present
}

func (this *mapleMap) addNewPlayer(player *playerHandle) {
	if this.isPlayerInMap(player) {
		return
	}

	this.mutex.Lock()
	this.players = append(this.players, player)
	this.mutex.Unlock()
}

func (this *mapleMap) removePlayer(player *playerHandle) {
	if !this.isPlayerInMap(player) {
		return
	}

	this.mutex.RLock()
	index := 0
	for i := range this.players {
		if this.players[i] == player {
			index = i
		}
	}
	this.mutex.RUnlock()

	this.mutex.Lock()
	// order not preserved slice ptr element remove
	this.players[index] = this.players[len(this.players)-1]
	this.players[len(this.players)-1] = nil
	this.players = this.players[:len(this.players)-1]
	this.mutex.Unlock()
}

func (this *mapleMap) sendPacketToPlayers(packet gopacket.Packet) {
	this.mutex.RLock()

	for i := range this.players {
		this.players[i].Write(packet)
	}

	this.mutex.RUnlock()
}

func (this *mapleMap) getRandomSpawnPortal() (pos byte, x int16, y int16) {
	rand.Seed(uint64(time.Now().Unix()))

	var portals []maplePortal

	this.mutex.RLock()
	for _, v := range this.portals {
		if v.isSpawn {
			portals = append(portals)
		}
	}
	this.mutex.RUnlock()

	portal := portals[rand.Int()%len(portals)]

	return portal.id, portal.x, portal.y
}

func (this *mapleMap) getPortalfromID(id byte) (pos byte, x int16, y int16) {
	var portal maplePortal

	this.mutex.RLock()
	for _, v := range this.portals {
		if v.id == id {
			portal = v
		}
	}
	this.mutex.RUnlock()

	return portal.id, portal.x, portal.y
}
