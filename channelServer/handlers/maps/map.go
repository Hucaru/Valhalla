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
var mutex = &sync.Mutex{}

func RegisterNewPlayer(conn *playerConn.Conn, mapID uint32) {
	mutex.Lock()

	playerMapList[mapID] = append(playerMapList[mapID], conn)
	DisplayMapObjects(conn, mapID)

	mutex.Unlock()
}

func PlayerLeftGame(conn *playerConn.Conn, mapID uint32) {

}

func PlayerChangeMap(conn *playerConn.Conn, newMapID uint32) {
	mutex.Lock()
	previousMapID := conn.GetCharacter().GetCurrentMap()
	// Remove from current map
	for i, v := range playerMapList[previousMapID] {
		if v == conn {
			playerMapList[previousMapID] = append(playerMapList[previousMapID][:i], playerMapList[previousMapID][i+1:]...)
			break
		}
	}

	playerMapList[newMapID] = append(playerMapList[newMapID], conn)

	char := conn.GetCharacter()
	char.SetCurrentMap(newMapID)
	char.SetPreviousMap(previousMapID)

	DisplayMapObjects(conn, newMapID)

	mutex.Unlock()
}

func PlayerMovement() {
	mutex.Lock()
	mutex.Unlock()
}

func PlayerUseSkill() {
	mutex.Lock()
	mutex.Unlock()
}

func DisplayMapObjects(conn *playerConn.Conn, mapID uint32) {
	// Spawn pet

	// For all connections except player
	for _, v := range playerMapList[mapID] {
		if v != conn {
			// send new player enter map to existing player
			// send existing player enter map to new player
			// show existing player pet
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
	mutex.Lock()
	for _, v := range playerMapList[mapID] {
		v.Write(p)
	}
	mutex.Unlock()
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
