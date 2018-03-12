package data

import (
	"sync"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/interfaces"
)

type charactersList map[interfaces.OcClientConn]*character.Character

var onlineChars = make(charactersList)

var onlineCharactersMutex = &sync.RWMutex{}

// GetCharsPtr -
func GetCharsPtr() charactersList {
	return onlineChars
}

func (oc charactersList) AddOnlineCharacter(conn interfaces.OcClientConn, char *character.Character) {
	onlineCharactersMutex.RLock()
	if _, exists := oc[conn]; exists {
		return
	}
	onlineCharactersMutex.RUnlock()

	onlineCharactersMutex.Lock()
	oc[conn] = char
	onlineCharactersMutex.Unlock()
}

func (oc charactersList) RemoveOnlineCharacter(conn interfaces.OcClientConn) {
	onlineCharactersMutex.RLock()
	if _, exists := oc[conn]; !exists {
		return
	}
	onlineCharactersMutex.RUnlock()

	onlineCharactersMutex.Lock()
	delete(oc, conn)
	onlineCharactersMutex.Unlock()
}

func (oc charactersList) GetOnlineCharacterHandle(conn interfaces.OcClientConn) *character.Character {
	onlineCharactersMutex.RLock()
	char := oc[conn]
	onlineCharactersMutex.RUnlock()

	return char
}

func (oc charactersList) GetConnectionHandle(name string) interfaces.OcClientConn {
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
