package maps

import (
	"github.com/Hucaru/Valhalla/interfaces"
)

func HandlePlayerEnterMap(conn interfaces.ClientConn, mapID uint32) {
	m := mapsPtr.GetMap(mapID)

	for _, v := range m.GetPlayers() {
		v.Write(playerEnterMapPacket(charsPtr.GetOnlineCharacterHandle(conn)))
		conn.Write(playerEnterMapPacket(charsPtr.GetOnlineCharacterHandle(v)))
	}

	m.AddPlayer(conn)

	// Send npcs
	for i, v := range m.GetNps() {
		conn.Write(showNpcPacket(uint32(i), v))
	}

	// Send mobs
	for i, v := range m.GetMobs() {

		if v.GetHp() == 0 {
			conn.Write(showMobPacket(uint32(i), v, true))
			v.SetHp(v.GetMaxHp())
		} else {
			conn.Write(showMobPacket(uint32(i), v, false))
		}
	}
}

func HandlePlayerLeaveMap(conn interfaces.ClientConn, mapID uint32) {
	mapsPtr.GetMap(mapID).RemovePlayer(conn)
	SendPacketToMap(mapID, playerLeftMapPacket(conn.GetUserID()))
}
