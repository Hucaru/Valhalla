package data

import (
	"sync"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/interfaces"
)

type charactersList map[interfaces.ClientConn]*character.Character

var onlineChars = make(charactersList)

var onlineCharactersMutex = &sync.RWMutex{}

// GetCharsPtr -
func GetCharsPtr() charactersList {
	return onlineChars
}

func (oc charactersList) AddOnlineCharacter(conn interfaces.ClientConn, char *character.Character) {
	onlineCharactersMutex.RLock()
	if _, exists := oc[conn]; exists {
		return
	}
	onlineCharactersMutex.RUnlock()

	onlineCharactersMutex.Lock()
	oc[conn] = char
	onlineCharactersMutex.Unlock()
}

func (oc charactersList) RemoveOnlineCharacter(conn interfaces.ClientConn) {
	onlineCharactersMutex.RLock()
	if _, exists := oc[conn]; !exists {
		return
	}
	onlineCharactersMutex.RUnlock()

	onlineCharactersMutex.Lock()
	delete(oc, conn)
	onlineCharactersMutex.Unlock()
}

func (oc charactersList) GetOnlineCharacterHandle(conn interfaces.ClientConn) *character.Character {
	onlineCharactersMutex.RLock()
	char := oc[conn]
	onlineCharactersMutex.RUnlock()

	return char
}

func (oc charactersList) GetConnHandleFromName(name string) interfaces.ClientConn {
	var handle interfaces.ClientConn

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

func (oc charactersList) GetCharFromID(id uint32) *character.Character {
	var handle *character.Character

	onlineCharactersMutex.RLock()
	for _, v := range oc {
		if v.GetCharID() == id {
			handle = v
			break
		}
	}
	onlineCharactersMutex.RUnlock()

	return handle
}

func (oc charactersList) GetChars() map[interfaces.ClientConn]*character.Character {
	onlineCharactersMutex.RLock()
	result := oc
	onlineCharactersMutex.RUnlock()

	return result
}
