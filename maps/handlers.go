package maps

import (
	"log"

	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/gopacket"
)

func HandlePlayerEnterMap(conn interfaces.ClientConn, mapID uint32) {
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
			// Send control packet
		}
		conn.Write(showMobPacket(uint32(i), v, false))
	}
}

func HandlePlayerLeaveMap(conn interfaces.ClientConn, mapID uint32) {
	mapsPtr.GetMap(mapID).RemovePlayer(conn)
	SendPacketToMap(mapID, playerLeftMapPacket(conn.GetUserID()))
	// Remove player as controller
	// find new controller for mobs if players left in map
}

func HandlePlayerUserPortal(conn interfaces.ClientConn, reader gopacket.Reader) {
	log.Println("Change packet:", reader)
}
