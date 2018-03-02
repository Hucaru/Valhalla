package data

import (
	"sync"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/interfaces"
)

type onlineCharactersList map[interfaces.OcClientConn]*character.Character

var onlineChars = make(onlineCharactersList)

var onlineCharactersMutex = &sync.RWMutex{}

// GetOnlineCharsPtr -
func GetOnlineCharsPtr() onlineCharactersList {
	return onlineChars
}

func (oc onlineCharactersList) AddOnlineCharacter(conn interfaces.OcClientConn, char *character.Character) {
	onlineCharactersMutex.RLock()
	if _, exists := oc[conn]; exists {
		return
	}
	onlineCharactersMutex.RUnlock()

	onlineCharactersMutex.Lock()
	oc[conn] = char
	onlineCharactersMutex.Unlock()
}

func (oc onlineCharactersList) RemoveOnlineCharacter(conn interfaces.OcClientConn) {
	onlineCharactersMutex.RLock()
	if _, exists := oc[conn]; !exists {
		return
	}
	onlineCharactersMutex.RUnlock()

	onlineCharactersMutex.Lock()
	delete(oc, conn)
	onlineCharactersMutex.Unlock()
}

func (oc onlineCharactersList) GetOnlineCharacterHandle(conn interfaces.OcClientConn) *character.Character {
	onlineCharactersMutex.RLock()
	char := oc[conn]
	onlineCharactersMutex.RUnlock()

	return char
}

func (oc onlineCharactersList) GetConnectionHandle(name string) interfaces.OcClientConn {
	var handle interfaces.OcClientConn

	onlineCharactersMutex.RLock()
	for k, v := range oc {
		if v.GetName() == name {
			handle = k
			break
		}
	}
	onlineCharactersMutex.RUnlock()

	return handle
}
