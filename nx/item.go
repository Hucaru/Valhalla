package nx

import (
	"strconv"
	"strings"
)

type Item struct {
	Price   uint32
	SlotMax uint16
	Cash    bool
}

var Items = make(map[uint32]Item)

func getItemInfo() {
	base := "Item/"
	commonPaths := []string{"Cash", "Consume", "Etc", "Install"}

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
				default:
					//fmt.Println(strLookup[property.NameID])
				}
			}
		default:
		}
	}
	return item
}
