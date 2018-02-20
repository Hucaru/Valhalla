package maps

import (
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/channelServer/handlers/npc"
	"github.com/Hucaru/Valhalla/channelServer/handlers/playerConn"
	"github.com/Hucaru/Valhalla/common/nx"
	"github.com/Hucaru/gopacket"
	"golang.org/x/exp/rand"
)

var playerMapList = make(map[uint32][]*playerConn.Conn)
var playerMapListMutex = &sync.Mutex{}

func RegisterNewPlayer(conn *playerConn.Conn, mapID uint32) {
	playerMapListMutex.Lock()

	playerMapList[mapID] = append(playerMapList[mapID], conn)
	displayMapObjects(conn, mapID)

	playerMapListMutex.Unlock()
}

func PlayerLeftGame(conn *playerConn.Conn) {
	playerMapListMutex.Lock()
	// Remove from current map
	currentMap := conn.GetCharacter().GetCurrentMap()
	for i, v := range playerMapList[currentMap] {
		if v == conn {
			playerMapList[currentMap] = append(playerMapList[currentMap][:i], playerMapList[currentMap][i+1:]...)
			break
		}
	}

	alertMapPlayerLeft(conn)

	playerMapListMutex.Unlock()
}

func PlayerChangeMap(conn *playerConn.Conn, newMapID uint32) {
	playerMapListMutex.Lock()
	previousMapID := conn.GetCharacter().GetCurrentMap()
	// Remove from current map
	for i, v := range playerMapList[previousMapID] {
		if v == conn {
			playerMapList[previousMapID] = append(playerMapList[previousMapID][:i], playerMapList[previousMapID][i+1:]...)
			break
		}
	}

	alertMapPlayerLeft(conn)

	playerMapList[newMapID] = append(playerMapList[newMapID], conn)

	char := conn.GetCharacter()
	char.SetCurrentMap(newMapID)
	char.SetPreviousMap(previousMapID)

	displayMapObjects(conn, newMapID)

	playerMapListMutex.Unlock()
}

func alertMapPlayerLeft(conn *playerConn.Conn) {
	for _, v := range playerMapList[conn.GetCharacter().GetCurrentMap()] {
		if v != conn {
			v.Write(playerLeftField(conn.GetCharacter().GetCharID()))
		}
	}
}

func PlayerMove(conn *playerConn.Conn, p gopacket.Packet) {
	SendPacketToMap(conn.GetCharacter().GetCurrentMap(), playerMove(conn.GetCharacter().GetCharID(), p))
}

func displayMapObjects(conn *playerConn.Conn, mapID uint32) {
	// Spawn pet

	// For all connections except player
	for _, v := range playerMapList[mapID] {
		if v != conn {
			v.Write(playerEnterField(conn.GetCharacter())) // send new player enter map to existing player
			conn.Write(playerEnterField(v.GetCharacter())) // send existing player enter map to new player
			// show existing player pet?
		}
	}

	// show npcs
	life := nx.Maps[mapID].Life
	for i, v := range life {
		if v.Npc {
			conn.Write(npc.SpawnNPC(uint32(i), v))
		}
	}

	// send mob data

	// show droped items

	// show player shops

	// show omok games

	// show kites

	// if map undergoing weather effect send it
}

func SendPacketToMap(mapID uint32, p gopacket.Packet) {
	playerMapListMutex.Lock()

	for _, v := range playerMapList[mapID] {
		pCopy := gopacket.NewPacket() // WTF why is conn.Write mutating the byte array?!
		pCopy.Append(p)

		v.Write(pCopy)
	}
	playerMapListMutex.Unlock()
}

func GetRandomSpawnPortal(mapID uint32) nx.Portal {
	var portals []nx.Portal
	for _, v := range nx.Maps[mapID].Portals {
		if v.IsSpawn {
			portals = append(portals, v)
		}
	}

	rand.Seed(uint64(time.Now().Unix()))

	return portals[rand.Int()%len(portals)]
}

func GetSpawnPortal(mapID uint32, portalID byte) nx.Portal {
	for _, v := range nx.Maps[mapID].Portals {
		if v.IsSpawn && v.ID == portalID {
			return v
		}
	}

	return GetRandomSpawnPortal(mapID)
}
