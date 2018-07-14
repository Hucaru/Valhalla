package nx

import (
	"encoding/binary"
	"math"
	"strconv"
	"strings"
)

type Item struct {
	Price   int32
	SlotMax int16
	Cash    bool

	AttackSpeed   int16
	Accuracy      int16
	Evasion       int16
	WeaponAttack  int16
	MagicAttack   int16
	MagicDefence  int16
	WeaponDefence int16

	Str int16
	Dex int16
	Int int16
	Luk int16

	ReqStr   int16
	ReqDex   int16
	ReqInt   int16
	ReqLuk   int16
	ReqJob   int16
	ReqLevel byte

	Upgrades  byte
	UnitPrice float64
}

var Items = make(map[int32]Item)

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

					Items[int32(itemID)] = getItem(itemIDNode)
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

			Items[int32(itemID)] = getItem(itemIDNode)
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

				Items[int32(itemID)] = getItem(itemIDNode)

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
					item.SlotMax = dataToInt16(property.Data)
				case "price":
					item.Price = dataToInt32(property.Data)
				case "attackSpeed":
					item.AttackSpeed = dataToInt16(property.Data)
				case "incAcc":
					item.Accuracy = dataToInt16(property.Data)
				case "incEVA":
					item.Evasion = dataToInt16(property.Data)
				case "incPAD":
					item.WeaponAttack = dataToInt16(property.Data)
				case "incMAD":
					item.MagicAttack = dataToInt16(property.Data)
				case "incMDD":
					item.MagicDefence = dataToInt16(property.Data)
				case "incPDD":
					item.WeaponDefence = dataToInt16(property.Data)
				case "incSTR":
					item.Str = dataToInt16(property.Data)
				case "incDEX":
					item.Dex = dataToInt16(property.Data)
				case "incINT":
					item.Int = dataToInt16(property.Data)
				case "incLUK":
					item.Luk = dataToInt16(property.Data)
				case "reqSTR":
					item.ReqStr = dataToInt16(property.Data)
				case "reqDEX":
					item.ReqDex = dataToInt16(property.Data)
				case "reqINT":
					item.ReqInt = dataToInt16(property.Data)
				case "reqLUK":
					item.ReqLuk = dataToInt16(property.Data)
				case "reqJob":
					item.ReqJob = dataToInt16(property.Data)
				case "reqLevel":
					item.ReqLevel = property.Data[0]
				case "tuc":
					item.Upgrades = property.Data[0]
				case "unitPrice":
					bits := binary.LittleEndian.Uint64([]byte(property.Data[:]))
					item.UnitPrice = math.Float64frombits(bits)
				case "icon":
				case "iconRaw":
				case "vslot":
				case "islot":
				default:
					//fmt.Println(strLookup[property.NameID])
				}
			}
		default:
		}
	}
	return item
}
