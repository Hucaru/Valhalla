package server

import (
	"sync"

	"github.com/Hucaru/Valhalla/channelServer/playerConn"
	"github.com/Hucaru/gopacket"
)

var playerMapList = make(map[uint32][]*playerConn.Conn)
var playerMapListMutex = &sync.RWMutex{}

func SendPacketToMap(mapID uint32, p gopacket.Packet) {
	playerMapListMutex.RLock()

	for i := range playerMapList[mapID] {
		playerMapList[mapID][i].Write(p)
	}

	playerMapListMutex.RUnlock()
}

func AddPlayerToMap(conn *playerConn.Conn, mapID uint32) {
	playerMapListMutex.Lock()

	playerMapList[mapID] = append(playerMapList[mapID], conn)

	playerMapListMutex.Unlock()
}

func RemovePlayerFromMap(conn *playerConn.Conn, mapID uint32) {
	playerMapListMutex.Lock()

	for i, v := range playerMapList[mapID] {
		if v == conn {
			playerMapList[mapID] = append(playerMapList[mapID][:i], playerMapList[mapID][i+1:]...)
			break
		}
	}

	playerMapListMutex.Unlock()
}

func PerformMapReadWork(work func(maps map[uint32][]*playerConn.Conn)) {
	playerMapListMutex.RLock()

	work(playerMapList)

	playerMapListMutex.RUnlock()
}

func MapGetAllPlayers(mapID uint32) []*playerConn.Conn {
	playerMapListMutex.RLock()
	val, exists := playerMapList[mapID]
	playerMapListMutex.RUnlock()

	if exists {
		return []*playerConn.Conn{}
	}

	return val
}
