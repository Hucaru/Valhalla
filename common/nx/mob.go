package nx

import (
	"strconv"
	"strings"
)

type Monster struct {
	Boss     bool
	Accuracy uint16
	Exp      uint32
	Level    byte
	MaxHp    uint16
	MaxMp    uint16
}

var Mob = make(map[uint32]Monster)

func getMobInfo() {
	result := searchNode("Mob", func(cursor *node) {
		for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
			mob := nodes[cursor.ChildID+i]

			for j := uint32(0); j < uint32(mob.ChildCount); j++ {
				options := nodes[mob.ChildID+j]

				switch strLookup[options.NameID] {
				case "info":
					ID, err := strconv.Atoi(strings.Split(strLookup[mob.NameID], ".")[0])

					if err != nil {
						panic(err)
					}

					Mob[uint32(ID)] = getMob(options)
				default:
				}

			}

		}
	})

	if !result {
		panic("Bad Search")
	}
}

func getMob(options node) Monster {
	monst := Monster{}
	for i := uint32(0); i < uint32(options.ChildCount); i++ {
		options := nodes[options.ChildID+i]
		switch strLookup[options.NameID] {
		case "boss":
			monst.Boss = bool(options.Data[0] == 1)
		case "acc":
			monst.Accuracy = dataToUint16(options.Data)
		case "exp":
			monst.Exp = dataToUint32(options.Data)
		case "level":
			monst.Level = options.Data[0]
		case "maxHP":
			monst.MaxHp = dataToUint16(options.Data)
		case "maxMP":
			monst.MaxMp = dataToUint16(options.Data)
		default:
		}
	}

	return monst
}
