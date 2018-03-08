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

// PlayerEnterMap -
func PlayerEnterMap(conn interfaces.ClientConn, mapID uint32) {
	m := mapsPtr.GetMap(mapID)

	for _, v := range m.GetPlayers() {
		v.Write(playerEnterMapPacket(charsPtr.GetOnlineCharacterHandle(conn)))
		conn.Write(playerEnterMapPacket(charsPtr.GetOnlineCharacterHandle(v)))
	}

	m.AddPlayer(conn)

	// Send npcs
	for i, v := range m.GetNpcs() {
		conn.Write(showNpcPacket(uint32(i), v))
	}

	// Send mobs
	for i, v := range m.GetMobs() {
		if v.GetController() == nil {
			v.SetController(conn)
			conn.Write(controlMobPacket(uint32(i), v, false))
		}

		conn.Write(showMobPacket(uint32(i), v, false))
	}
}

// PlayerLeaveMap -
func PlayerLeaveMap(conn interfaces.ClientConn, mapID uint32) {
	m := mapsPtr.GetMap(mapID)

	m.RemovePlayer(conn)

	// Remove player as mob controller
	for i, v := range m.GetMobs() {
		if v.GetController() == conn {
			v.SetController(nil)
		}

		conn.Write(endMobControlPacket(uint32(i)))
	}

	if len(m.GetPlayers()) > 0 {
		newController := m.GetPlayers()[0]
		for i, v := range m.GetMobs() {
			newController.Write(controlMobPacket(uint32(i), v, false))
		}
	}

	SendPacketToMap(mapID, playerLeftMapPacket(charsPtr.GetOnlineCharacterHandle(conn).GetCharID()))
}

func GetRandomSpawnPortal(mapID uint32) (interfaces.Portal, byte) {
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

func RegisterNewPlayerCallback(conn interfaces.ClientConn) {
	conn.AddCloseCallback(func() {
		PlayerLeaveMap(conn, charsPtr.GetOnlineCharacterHandle(conn).GetCurrentMap())
		charsPtr.RemoveOnlineCharacter(conn)
	})
}
