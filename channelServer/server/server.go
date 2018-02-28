package server

import (
	"log"

	"github.com/Hucaru/Valhalla/common/nx"
)

// Init - Initialises the server state once data has been parsed
func Init() {

	// Instatiate all maps
	for k := range nx.Maps {
		maps[k] = createMap(k)
	}

	log.Print("Finished initialisation")
}
