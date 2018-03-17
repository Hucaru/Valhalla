package command

import "github.com/Hucaru/Valhalla/interfaces"

var charsPtr interfaces.Characters
var mapsPtr interfaces.Maps

// RegisterCharactersObj -
func RegisterCharactersObj(chars interfaces.Characters) {
	charsPtr = chars
}

// RegisterMapsObj -
func RegisterMapsObj(mapList interfaces.Maps) {
	mapsPtr = mapList
}
