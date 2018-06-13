package nx

import (
	"strconv"
	"strings"
)

type Monster struct {
	Boss         bool
	Accuracy     uint16
	Exp          uint32
	Level        byte
	MaxHp        uint32
	Hp           uint32
	MaxMp        uint32
	Mp           uint32
	FlySpeed     uint32
	SummonEffect byte
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
			monst.MaxHp = dataToUint32(options.Data)
			monst.Hp = monst.MaxHp
		case "maxMP":
			monst.MaxMp = dataToUint32(options.Data)
			monst.Mp = monst.MaxMp
		case "flySpeed":
			monst.FlySpeed = dataToUint32(options.Data)
		case "summonType":
			monst.SummonEffect = options.Data[0]
		default:
		}
	}

	return monst
}
