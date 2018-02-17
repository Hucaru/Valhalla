package nx

import "fmt"

type EquipItem struct {
}

var Equip map[uint32]EquipItem

func getEquipInfo() {
	Equip = make(map[uint32]EquipItem)

	path := "Character"

	result := searchNode("Character", func(cursor *node) {

		for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
			n := nodes[cursor.ChildID+i]

			switch strLookup[n.NameID] {
			case "Accessory":
			case "Cap":
			case "Coat":
			case "Face":
			case "Glove":
			case "Hair":
			case "Longcoat":
			case "Pants":
			case "PetEquip":
			case "Ring":
			case "Shield":
			case "Shoes":
			case "Weapon":
			default:
				fmt.Println("Unkown Character type", strLookup[n.NameID])
			}
		}
	})

	if !result {
		panic("Bad search:" + path)
	}
}
