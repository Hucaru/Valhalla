package nx

import (
	"strconv"
	"strings"
)

type EquipItem struct {
	Cash     bool
	ReqDex   uint16
	ReInt    uint16
	ReqJob   uint16
	ReqLuk   uint16
	ReqLevel uint16
	ReqStr   uint16
}

var Equip = make(map[uint32]EquipItem)

func validEquipID(ID uint32) bool {
	if _, ok := Equip[ID]; ok {
		return true
	}

	return false
}

func IsCashItem(ID uint32) bool {
	if validEquipID(ID) {
		return Equip[ID].Cash
	}

	return false
}

func getEquipInfo() {

	path := "Character"

	result := searchNode("Character", func(cursor *node) {

		for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
			n := nodes[cursor.ChildID+i]

			switch strLookup[n.NameID] {
			case "Accessory":
				fallthrough
			case "Cap":
				fallthrough
			case "Coat":
				fallthrough
			case "Face":
				fallthrough
			case "Glove":
				fallthrough
			case "Hair":
				fallthrough
			case "Longcoat":
				fallthrough
			case "Pants":
				fallthrough
			case "PetEquip":
				fallthrough
			case "Ring":
				fallthrough
			case "Shield":
				fallthrough
			case "Shoes":
				fallthrough
			case "Weapon":
				for j := uint32(0); j < uint32(n.ChildCount); j++ {
					itemNode := nodes[n.ChildID+j]

					// Get Name
					nameSplit := strings.Split(strLookup[itemNode.NameID], "/")
					equipID, err := strconv.Atoi(strings.Split(nameSplit[len(nameSplit)-1], ".")[0])

					if err != nil {
						panic(err)
					}

					Equip[uint32(equipID)] = getEquipItem(itemNode)
				}
			default:
				//
			}
		}
	})

	if !result {
		panic("Bad search:" + path)
	}
}

func getEquipItem(item node) EquipItem {
	equip := EquipItem{}

	for i := uint32(0); i < uint32(item.ChildCount); i++ {
		property := nodes[item.ChildID+i]

		switch strLookup[property.NameID] {
		case "info":
			for j := uint32(0); j < uint32(property.ChildCount); j++ {
				option := nodes[property.ChildID+j]

				switch strLookup[option.NameID] {
				case "cash":
					equip.Cash = bool(dataToInt64(option.Data) == 1)
				case "reqDEX":
					equip.ReqDex = dataToUint16(option.Data)
				case "reqINT":
					equip.ReInt = dataToUint16(option.Data)
				case "reqJob":
					equip.ReqJob = dataToUint16(option.Data)
				case "reqLUK":
					equip.ReqLuk = dataToUint16(option.Data)
				case "reqLevel":
					equip.ReqLevel = dataToUint16(option.Data)
				case "reqSTR":
					equip.ReqStr = dataToUint16(option.Data)
				default:
				}

			}
		default:
		}
	}

	return equip
}
