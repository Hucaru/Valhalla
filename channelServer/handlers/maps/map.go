package maps

import (
	"time"

	"github.com/Hucaru/Valhalla/channelServer/handlers/playerConn"
	"github.com/Hucaru/Valhalla/common/nx"
	"golang.org/x/exp/rand"
)

var playerLocations = make(map[uint32][]*playerConn.Conn)

func RegisterNewPlayer(conn *playerConn.Conn) {

}

func PlayerChangeMap() {

}

func PlayerMovement() {

}

func PlayerUserSkill() {

}

func PlayerSaysToAll() {

}

func GetRandomSpawnPortal(mapID uint32) byte {
	var portals []byte
	for _, v := range nx.Maps[mapID].Portals {
		if v.IsSpawn {
			portals = append(portals, v.ID)
		}
	}

	rand.Seed(uint64(time.Now().Unix()))

	return portals[rand.Int()%len(portals)]
}
