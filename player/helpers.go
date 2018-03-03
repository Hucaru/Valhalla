package player

import "github.com/Hucaru/Valhalla/interfaces"

var charsPtr interfaces.Characters

// RegisterCharactersObj -
func RegisterCharactersObj(chars interfaces.Characters) {
	charsPtr = chars
}
