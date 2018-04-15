package inventory

import (
	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/interfaces"
)

var charsPtr interfaces.Characters

// RegisterCharactersObj -
func RegisterCharactersObj(chars interfaces.Characters) {
	charsPtr = chars
}

func isStackable(item character.Item) bool {
	if item.GetItemID()/1e6 != 5 && // pet item
		item.GetInvID() != 1 && // equip
		item.GetItemID()/1e4 != 207 && // star/arrow etc
		item.GetAmount() < constants.MAX_ITEM_STACK {

		return true
	}

	return false
}

func isEquipable(item character.Item) bool {
	if item.GetItemID()/1e6 != 5 || item.GetInvID() != 1 {
		return true
	}

	return false
}
