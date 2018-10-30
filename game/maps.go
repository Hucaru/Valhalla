package game

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/packets"
	"github.com/Hucaru/Valhalla/types"
)

var maps = make(map[int32]*gameMap)

type gameMap struct {
	npcs     []types.NPC
	mobs     []types.Mob
	mobQueue []types.Mob
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

func (gm *gameMap) removeController(conn mnet.MConnChannel) {
	for i, m := range gm.mobs {
		if m.Controller == conn {
			gm.mobs[i].Controller = nil
			conn.Send(packets.MobEndControl(m))
		}
	}

	for c, p := range players {
		if c != conn && p.char.CurrentMap == players[conn].char.CurrentMap {
			for i, m := range gm.mobs {
				gm.mobs[i].Controller = c
				c.Send(packets.MobControl(m, false))
			}
		}
	}
}

func (gm *gameMap) addController(conn mnet.MConnChannel) {
	for i, m := range gm.mobs {
		if m.Controller == nil {
			gm.mobs[i].Controller = conn
			conn.Send(packets.MobControl(m, false))
		}
	}
}
