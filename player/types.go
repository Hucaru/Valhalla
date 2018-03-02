package player

import (
	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/interfaces"
)

type characters interface {
	AddOnlineCharacter(interfaces.OcClientConn, *character.Character)
	RemoveOnlineCharacter(interfaces.OcClientConn)
	GetOnlineCharacterHandle(interfaces.OcClientConn) *character.Character
	GetConnectionHandle(string) interfaces.OcClientConn
}

var dataPtr characters

func RegisterCharactersObj(chars characters) {
	dataPtr = chars
}
