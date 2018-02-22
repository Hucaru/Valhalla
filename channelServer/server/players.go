package server

import (
	"sync"

	"github.com/Hucaru/gopacket"

	"github.com/Hucaru/Valhalla/channelServer/playerConn"
)

var playerList = make(map[string]*playerConn.Conn)
var playerListMutex = &sync.RWMutex{}

func AddPlayerToList(conn *playerConn.Conn) {
	playerListMutex.Lock()

	// Is there any point in checking if it already exists?
	playerList[conn.GetCharacter().GetName()] = conn

	playerListMutex.Unlock()
}

func RemovePlayerFromList(conn *playerConn.Conn) {
	playerListMutex.Lock()

	if _, exists := playerList[conn.GetCharacter().GetName()]; exists {
		delete(playerList, conn.GetCharacter().GetName())
	}

	playerListMutex.Unlock()
}

func SendPacketToPlayerName(name string, p gopacket.Packet) bool {
	success := false

	playerListMutex.RLock()

	if v, exists := playerList[name]; exists {
		v.Write(p)
		success = true
	}

	playerListMutex.RUnlock()

	return success
}
