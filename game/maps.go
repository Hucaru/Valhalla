package game

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/types"
)

var maps = make(map[int32]*gameMap)

type gameMap struct {
	npcs       []types.NPC
	mobs       []types.Mob
	controller mnet.MConnChannel
}

func InitMaps() {
	for mapID, nxMap := range nx.Maps {
		npcs := []types.NPC{}
		mobs := []types.Mob{}

		for _, l := range nxMap.Life {
			if l.IsMob {
				mobs = append(mobs, types.CreateMob(int32(len(mobs)), l, nx.Mob[l.ID], nil))
			} else {
				npcs = append(npcs, types.CreateNPC(int32(len(npcs)), l))
			}
		}

		maps[mapID] = &gameMap{
			npcs: npcs,
			mobs: mobs,
		}
	}
}
