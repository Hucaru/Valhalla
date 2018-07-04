package nx

import (
	"strconv"
	"strings"
)

type Item struct {
	Price   uint32
	SlotMax uint16
	Cash    bool

	AttackSpeed  uint32
	Accuracy     uint32
	Evasion      uint32
	WeaponAttack uint32

	ReqStr   uint32
	ReqDex   uint32
	ReqInt   uint32
	ReqLuk   uint32
	ReqJob   uint32
	ReqLevel uint32

	Upgrades uint32
}

var Items = make(map[uint32]Item)

func getItemInfo() {
	base := "Item/"
	commonPaths := []string{"Cash", "Consume", "Etc", "Install", "Special"}

	for _, path := range commonPaths {
		result := searchNode(base+path, func(cursor *node) {
			for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
				n := nodes[cursor.ChildID+i]

				for j := uint32(0); j < uint32(n.ChildCount); j++ {
					itemIDNode := nodes[n.ChildID+j]

					itemID, err := strconv.Atoi(strLookup[itemIDNode.NameID])

					if err != nil {
						panic(err)
					}

					Items[uint32(itemID)] = getItem(itemIDNode)
				}
			}
		})

		if !result {
			panic("Bad Search")
		}
	}

	result := searchNode(base+"Pet", func(cursor *node) {
		for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
			itemIDNode := nodes[cursor.ChildID+i]

			itemID, err := strconv.Atoi(strings.Split(strLookup[itemIDNode.NameID], ".")[0])

			if err != nil {
				panic(err)
			}

			Items[uint32(itemID)] = getItem(itemIDNode)
		}
	})

	if !result {
		panic("Bad Search")
	}

	base = "Character/"
	commonPaths = []string{"Accessory", "Cap", "Cape", "Coat", "Face",
		"Glove", "Hair", "Longcoat", "Pants", "PetEquip", "Ring", "Shield", "Shoes", "Weapon"}

	for _, path := range commonPaths {
		result := searchNode(base+path, func(cursor *node) {
			for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
				itemIDNode := nodes[cursor.ChildID+i]

				itemID, err := strconv.Atoi(strings.Split(strLookup[itemIDNode.NameID], ".")[0])

				if err != nil {
					panic(err)
				}

				Items[uint32(itemID)] = getItem(itemIDNode)

			}
		})

		if !result {
			panic("Bad Search")
		}
	}
}

func getItem(node node) Item {
	item := Item{}
	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		options := nodes[node.ChildID+i]

		switch strLookup[options.NameID] {
		case "info":
			for l := uint32(0); l < uint32(options.ChildCount); l++ {
				property := nodes[options.ChildID+l]

				switch strLookup[property.NameID] {
				case "cash":
					item.Cash = bool(dataToInt64(property.Data) == 1)
				case "slotMax":
					item.SlotMax = dataToUint16(property.Data)
				case "price":
					item.Price = dataToUint32(property.Data)
				case "attackSpeed":
					item.AttackSpeed = dataToUint32(property.Data)
				case "incAcc":
					item.Accuracy = dataToUint32(property.Data)
				case "incEVA":
					item.Evasion = dataToUint32(property.Data)
				case "incPAD":
					item.WeaponAttack = dataToUint32(property.Data)
				case "reqSTR":
					item.ReqStr = dataToUint32(property.Data)
				case "reqDEX":
					item.ReqDex = dataToUint32(property.Data)
				case "reqINT":
					item.ReqInt = dataToUint32(property.Data)
				case "reqLUK":
					item.ReqLuk = dataToUint32(property.Data)
				case "reqJob":
					item.ReqJob = dataToUint32(property.Data)
				case "reqLevel":
					item.ReqLevel = dataToUint32(property.Data)
				case "tuc":
					item.Upgrades = dataToUint32(property.Data)
				case "unitPrice":
					//item.unitPrice = dataToUint32(property.Data)
				default:
					//fmt.Println(strLookup[property.NameID])
				}
			}
		default:
		}
	}
	return item
}
