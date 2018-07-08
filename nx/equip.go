package nx

import (
	"strconv"
	"strings"
)

type EquipItem struct {
	Cash     bool
	ReqDex   int16
	ReInt    int16
	ReqJob   int16
	ReqLuk   int16
	ReqLevel int16
	ReqStr   int16
}

var Equip = make(map[int32]EquipItem)

func validEquipID(ID int32) bool {
	if _, ok := Equip[ID]; ok {
		return true
	}

	return false
}

func IsCashItem(ID int32) bool {
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

					Equip[int32(equipID)] = getEquipItem(itemNode)
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
					equip.ReqDex = int16(dataToUint16(option.Data))
				case "reqINT":
					equip.ReInt = int16(dataToUint16(option.Data))
				case "reqJob":
					equip.ReqJob = int16(dataToUint16(option.Data))
				case "reqLUK":
					equip.ReqLuk = int16(dataToUint16(option.Data))
				case "reqLevel":
					equip.ReqLevel = int16(dataToUint16(option.Data))
				case "reqSTR":
					equip.ReqStr = int16(dataToUint16(option.Data))
				default:
				}

			}
		default:
		}
	}

	return equip
}
