package maps

import (
	"sync"

	"github.com/Hucaru/Valhalla/channelServer/handlers/mobs"
	"github.com/Hucaru/Valhalla/channelServer/handlers/npc"
	"github.com/Hucaru/Valhalla/channelServer/handlers/playerConn"
	"github.com/Hucaru/Valhalla/common/nx"
	"github.com/Hucaru/gopacket"
)

var playerMapList = make(map[uint32][]*playerConn.Conn)
var playerMapListMutex = &sync.RWMutex{}

func RegisterNewPlayer(conn *playerConn.Conn, mapID uint32) {
	addPlayerToMap(conn, mapID)
	displayMapObjects(conn, mapID)
}

func PlayerLeftGame(conn *playerConn.Conn) {
	mapID := conn.GetCharacter().GetCurrentMap()
	removePlayerFromMap(conn, mapID)
	alertMapPlayerLeft(conn, mapID)
}

func PlayerChangeMap(conn *playerConn.Conn, newMapID uint32) {
	previousMapID := conn.GetCharacter().GetCurrentMap()

	char := conn.GetCharacter()
	char.SetCurrentMap(newMapID)
	char.SetPreviousMap(previousMapID)

	alertMapPlayerLeft(conn, previousMapID)

	removePlayerFromMap(conn, previousMapID)
	addPlayerToMap(conn, newMapID)

	displayMapObjects(conn, newMapID)
}

func addPlayerToMap(conn *playerConn.Conn, mapID uint32) {
	playerMapListMutex.Lock()

	playerMapList[mapID] = append(playerMapList[mapID], conn)

	playerMapListMutex.Unlock()
}

func removePlayerFromMap(conn *playerConn.Conn, mapID uint32) {
	playerMapListMutex.Lock()

	for i, v := range playerMapList[mapID] {
		if v == conn {
			playerMapList[mapID] = append(playerMapList[mapID][:i], playerMapList[mapID][i+1:]...)
			break
		}
	}
	playerMapListMutex.Unlock()
}

func alertMapPlayerLeft(conn *playerConn.Conn, mapID uint32) {
	playerMapListMutex.RLock()

	for _, v := range playerMapList[mapID] {
		if v != conn {
			v.Write(playerLeftField(conn.GetCharacter().GetCharID()))
		}
	}

	playerMapListMutex.RUnlock()
}

func PlayerMove(conn *playerConn.Conn, p gopacket.Packet) {
	SendPacketToMap(conn.GetCharacter().GetCurrentMap(), playerMove(conn.GetCharacter().GetCharID(), p))
}

func displayMapObjects(conn *playerConn.Conn, mapID uint32) {
	// Spawn pet

	// For all connections except player
	playerMapListMutex.RLock()

	for _, v := range playerMapList[mapID] {
		if v != conn {
			v.Write(playerEnterField(conn.GetCharacter())) // send new player enter map to existing player
			conn.Write(playerEnterField(v.GetCharacter())) // send existing player enter map to new player
			// show existing player pet?
		}
	}

	playerMapListMutex.RUnlock()

	// show npcs
	life := nx.Maps[mapID].Life
	for i, v := range life {
		if v.Npc {
			conn.Write(npc.SpawnNPC(uint32(i), v))
		}
	}

	// send mob data
	conn.Write(mobs.SpawnMob())   // test
	conn.Write(mobs.ControlMob()) // test

	// show droped items

	// show player shops

	// show omok games

	// show kites

	// if map undergoing weather effect send it
}

func SendPacketToMap(mapID uint32, p gopacket.Packet) {
	playerMapListMutex.RLock()

	for i := range playerMapList[mapID] {
		playerMapList[mapID][i].Write(p)
	}
	playerMapListMutex.RUnlock()
}
