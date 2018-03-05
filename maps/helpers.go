package maps

import (
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/gopacket"
)

var mapsPtr interfaces.Maps

var charsPtr interfaces.Characters

// RegisterCharactersObj -
func RegisterCharactersObj(chars interfaces.Characters) {
	charsPtr = chars
}

// RegisterMapsObj -
func RegisterMapsObj(mapList interfaces.Maps) {
	mapsPtr = mapList
}

// SendPacketToMap -
func SendPacketToMap(mapID uint32, p gopacket.Packet) {
	if len(p) > 0 {

		players := mapsPtr.GetMap(mapID).GetPlayers()

		for _, v := range players {
			v.Write(p)
		}
	}
}

func getRandomSpawnPortal(mapID uint32) (interfaces.Portal, byte) {
	var portals []interfaces.Portal
	for _, portal := range mapsPtr.GetMap(mapID).GetPortals() {
		if portal.GetIsSpawn() {
			portals = append(portals, portal)
		}
	}
	rand.Seed(time.Now().UnixNano())
	pos := rand.Intn(len(portals))
	return portals[pos], byte(pos)
}

// SpawnMob -
func SpawnMob(mapID uint32, mobID uint32) {

}
