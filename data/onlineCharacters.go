package data

import (
	"sync"

	"github.com/Hucaru/Valhalla/character"
)

type clientConn interface {
	GetUserID() uint32
}

type onlineCharactersList map[clientConn]*character.Character

var onlineChars = make(onlineCharactersList)

var onlineCharactersMutex = &sync.RWMutex{}

func AddOnlineCharacter(conn clientConn, char *character.Character) {
	onlineCharactersMutex.RLock()
	if _, exists := onlineChars[conn]; exists {
		return
	}
	onlineCharactersMutex.RUnlock()

	onlineCharactersMutex.Lock()
	onlineChars[conn] = char
	onlineCharactersMutex.Unlock()
}

func RemoveOnlineCharacter(conn clientConn) {
	onlineCharactersMutex.RLock()
	if _, exists := onlineChars[conn]; !exists {
		return
	}
	onlineCharactersMutex.RUnlock()

	onlineCharactersMutex.Lock()
	delete(onlineChars, conn)
	onlineCharactersMutex.Unlock()
}

func GetOnlineCharacterHandle(conn clientConn) *character.Character {
	onlineCharactersMutex.RLock()
	char := onlineChars[conn]
	onlineCharactersMutex.RUnlock()

	return char
}

func GetConnectionHandle(name string) clientConn {
	var handle clientConn

	onlineCharactersMutex.RLock()
	for k, v := range onlineChars {
		if v.GetName() == name {
			handle = k
			break
		}
	}
	onlineCharactersMutex.RUnlock()

	return handle
}
