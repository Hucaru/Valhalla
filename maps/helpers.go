package maps

import (
	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/gopacket"
)

var maps interfaces.Maps

// RegisterMaps -
func RegisterMapsObj(maps interfaces.Maps) {
	maps = maps
}

// SendPacketToMap -
func SendPacketToMap(mapID uint32, p gopacket.Packet) {

}

// SpawnMob -
func SpawnMob(mapID uint32, mobID uint32) {

}
