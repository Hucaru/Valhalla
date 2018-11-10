package game

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/packets"
	"github.com/Hucaru/Valhalla/types"
)

var maps = make(map[int32]*GameMap)

type GameMap struct {
	npcs []types.NPC
	mobs []mob
}

func InitMaps() {
	for mapID, nxMap := range nx.Maps {
		npcs := []types.NPC{}
		mobs := []mob{}

		for _, l := range nxMap.Life {
			if l.IsMob {
				mobs = append(mobs, mob{Mob: types.CreateMob(int32(len(mobs)+1), l, nx.Mob[l.ID], nil), mapID: mapID})
			} else {
				npcs = append(npcs, types.CreateNPC(int32(len(npcs)), l))
			}
		}

		maps[mapID] = &GameMap{
			npcs: npcs,
			mobs: mobs,
		}
	}
}

func (gm *GameMap) removeController(conn mnet.MConnChannel) {
	for i, m := range gm.mobs {
		if m.Controller == conn {
			gm.mobs[i].Controller = nil
			conn.Send(packets.MobEndControl(m.Mob))
		}
	}

	for c, p := range players {
		if c != conn && p.char.CurrentMap == players[conn].char.CurrentMap {
			for i, m := range gm.mobs {
				gm.mobs[i].Controller = c
				c.Send(packets.MobControl(m.Mob))
			}
		}
	}
}

func (gm *GameMap) addController(conn mnet.MConnChannel) {
	for i, m := range gm.mobs {
		if m.Controller == nil {
			gm.mobs[i].Controller = conn
			conn.Send(packets.MobControl(m.Mob))
		}
	}
}

func (gm *GameMap) GetMobFromID(id int32) *mob {
	for i, v := range gm.mobs {
		if v.SpawnID == id {
			return &gm.mobs[i]
		}
	}

	return nil
}

func (gm *GameMap) HandleDeadMobs() {
	y := gm.mobs[:0]

	for _, mob := range gm.mobs {
		if mob.HP < 1 {
			mob.Controller.Send(packets.MobEndControl(mob.Mob))

			// if len(mob.Revive) > 0 {
			// 	for _, id := range mob.Revive {
			// 		SpawnWithoutRespawn(mob.mapID, id, int32(len(gm.mobs)+1), mob.X, mob.Y, mob.Foothold, -3, mob.SpawnID)
			// 	}
			// }

			SendToMap(mob.mapID, packets.MobRemove(mob.Mob, 1)) // 0 keeps it there and is no longer attackable, 1 normal death, 2 disaapear instantly
		} else {
			y = append(y, mob)
		}
	}

	gm.mobs = y
}

func (gm *GameMap) SpawnMob(mobID, spawnID int32, x, y, foothold int16, summonType int8, summonOption int32, facesLeft byte) {

}

func (gm *GameMap) SpawnMobNoRespawn(mobID, spawnID int32, x, y, foothold int16, summonType int8, summonOption int32, facesLeft byte) {

}

// func SpawnWithoutRespawn(mapID, mobID, spawnID int32, x, y, foothold int16, summonType int8, summonOption int32) {
// 	mob := types.CreateMob(spawnID, nx.Life{}, nx.Mob[mobID], nil)
// 	mob.ID = mobID
// 	mob.X = x
// 	mob.Y = y
// 	mob.Foothold = foothold

// 	mob.Respawns = false

// 	mob.SummonType = summonType
// 	mob.SummonOption = summonOption

// 	maps[mapID].mobs = append(maps[mapID].mobs, mob)

// 	SendToMap(mapID, packets.MobShow(mob))

// 	findController(mapID, &mob)

// 	if summonType != -4 {
// 		mob.SummonType = -1
// 		mob.SummonOption = 0
// 	}
// }

// func findController(mapID int32, mob *types.Mob) {
// 	for _, p := range players {
// 		if p.char.CurrentMap == mapID {
// 			mob.Controller = p

// 			p.Send(packets.MobControl(*mob))

// 			return
// 		}
// 	}
// }
