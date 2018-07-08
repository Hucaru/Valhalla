package nx

import (
	"encoding/binary"
	"math"
	"strconv"
	"strings"

	"github.com/Hucaru/Valhalla/constants"
)

type Item struct {
	Price   int32
	SlotMax int16
	Cash    bool

	AttackSpeed   int32
	Accuracy      int32
	Evasion       int32
	WeaponAttack  int32
	MagicAttack   int32
	WeaponDefence int32

	ReqStr   int32
	ReqDex   int32
	ReqInt   int32
	ReqLuk   int32
	ReqJob   int32
	ReqLevel int32

	Upgrades  int32
	UnitPrice float64
}

var Items = make(map[int32]Item)

func IsRechargeAble(itemID int32) bool {
	return (math.Floor(float64(itemID/10000)) == 207) // Taken from cliet
}

func IsStackable(itemID int32, invID byte, ammount int16) bool {
	if itemID/1e6 != 5 && // pet item
		invID != 1 && // equip
		itemID/1e4 != 207 && // star/arrow etc
		ammount < constants.MAX_ITEM_STACK {

		return true
	}

	return false
}

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
					item.AttackSpeed = dataToInt32(property.Data)
				case "incAcc":
					item.Accuracy = dataToInt32(property.Data)
				case "incEVA":
					item.Evasion = dataToInt32(property.Data)
				case "incPAD":
					item.WeaponAttack = dataToInt32(property.Data)
				case "incMAD":
					item.MagicAttack = dataToInt32(property.Data)
				case "incPDD":
					item.WeaponDefence = dataToInt32(property.Data)
				case "reqSTR":
					item.ReqStr = dataToInt32(property.Data)
				case "reqDEX":
					item.ReqDex = dataToInt32(property.Data)
				case "reqINT":
					item.ReqInt = dataToInt32(property.Data)
				case "reqLUK":
					item.ReqLuk = dataToInt32(property.Data)
				case "reqJob":
					item.ReqJob = dataToInt32(property.Data)
				case "reqLevel":
					item.ReqLevel = dataToInt32(property.Data)
				case "tuc":
					item.Upgrades = dataToInt32(property.Data)
				case "unitPrice":
					bits := binary.LittleEndian.Uint64([]byte(property.Data[:]))
					item.UnitPrice = math.Float64frombits(bits)
				default:
					//fmt.Println(strLookup[property.NameID])
				}
			}
		default:
		}
	}
	return item
}
