package nx

import (
	"strconv"
	"strings"
)

type Monster struct {
	Boss         bool
	Accuracy     int16
	Exp          int32
	Level        byte
	MaxHp        int32
	Hp           int32
	MaxMp        int32
	Mp           int32
	FlySpeed     int32
	SummonEffect byte
}

var Mob = make(map[int32]Monster)

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

					Mob[int32(ID)] = getMob(options)
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
			monst.Accuracy = dataToInt16(options.Data)
		case "exp":
			monst.Exp = dataToInt32(options.Data)
		case "level":
			monst.Level = options.Data[0]
		case "maxHP":
			monst.MaxHp = dataToInt32(options.Data)
			monst.Hp = monst.MaxHp
		case "maxMP":
			monst.MaxMp = dataToInt32(options.Data)
			monst.Mp = monst.MaxMp
		case "flySpeed":
			monst.FlySpeed = dataToInt32(options.Data)
		case "summonType":
			monst.SummonEffect = options.Data[0]
		default:
		}
	}

	return monst
}
