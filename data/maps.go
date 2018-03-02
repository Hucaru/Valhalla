package data

import "sync"

var mapleMaps = make(map[uint32]*mapleMap)
var mapleMapsMutex = &sync.RWMutex{}

type mapleMap struct {
	npcs         []mapleNpc
	mobs         []mapleMob
	forcedReturn uint32
	returnMap    uint32
	mobRate      float64
	isTown       bool
	portals      []maplePortal
}
