package maps

import (
	"sync"
)

var mapleMaps = make(map[uint32]*mapleMap)
var mapleMapsMutex = &sync.RWMutex{}
