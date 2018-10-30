package nx

import (
	"strconv"
	"strings"
	"sync"
)

type Monster struct {
	Boss       bool
	Accuracy   int16
	Exp        int32
	Level      byte
	MaxHP      int32
	HP         int32
	MaxMP      int32
	MP         int32
	FlySpeed   int32
	SummonType byte
}

var Mob = make(map[int32]Monster)

func getMobInfo(wg *sync.WaitGroup) {
	defer wg.Done()

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
			monst.MaxHP = dataToInt32(options.Data)
			monst.HP = monst.MaxHP
		case "maxMP":
			monst.MaxMP = dataToInt32(options.Data)
			monst.MP = monst.MaxMP
		case "flySpeed":
			monst.FlySpeed = dataToInt32(options.Data)
		case "summonType":
			monst.SummonType = options.Data[0]
		default:
		}
	}

	return monst
}
