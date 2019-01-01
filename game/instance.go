package game

import (
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/game/packet"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

type Instance struct {
	mapID   int32
	npcs    []def.NPC
	mobs    []gameMob
	players []mnet.MConnChannel
}

func createInstanceFromMapData(mapData nx.Map, mapID int32) Instance {
	npcs := []def.NPC{}
	mobs := []gameMob{}

	for _, l := range mapData.Mobs {
		nxMob, err := nx.GetMob(l.ID)

		if err != nil {
			continue
		}

		mobs = append(mobs, gameMob{Mob: def.CreateMob(int32(len(mobs)+1), l, nxMob, nil), mapID: mapID})
	}

	for _, l := range mapData.NPCs {
		npcs = append(npcs, def.CreateNPC(int32(len(npcs)), l))
	}

	return Instance{mapID: mapID, npcs: npcs, mobs: mobs}
}

func (inst *Instance) send(p mpacket.Packet) {
	for _, v := range inst.players {
		v.Send(p)
	}
}

func (inst *Instance) sendExcept(p mpacket.Packet, exception mnet.MConnChannel) {
	for _, v := range inst.players {
		if v == exception {
			continue
		}

		v.Send(p)
	}
}

func (inst *Instance) addPlayer(conn mnet.MConnChannel) {
	for i, mob := range inst.mobs {
		if mob.HP > 0 {
			mob.SummonType = -1 // -2: fade in spawn animation, -1: no spawn animation
			conn.Send(packet.MobShow(mob.Mob))

			if mob.Controller == nil {
				inst.mobs[i].ChangeController(conn)
			}
		}
	}

	for _, npc := range inst.npcs {
		conn.Send(packet.NpcShow(npc))
		conn.Send(packet.NPCSetController(npc.SpawnID, true))
	}

	player := Players[conn]

	for _, other := range inst.players {
		otherPlayer := Players[other]
		player.Send(packet.MapPlayerEnter(otherPlayer.Char()))
		otherPlayer.Send(packet.MapPlayerEnter(player.Char()))
	}

	inst.players = append(inst.players, conn)
}

func (inst *Instance) removePlayer(conn mnet.MConnChannel) {
	ind := -1
	for i, v := range inst.players {
		if v == conn {
			ind = i
		}
	}

	if ind == -1 {
		return // This should not be possible
	}

	inst.players = append(inst.players[:ind], inst.players[ind+1:]...)

	for i, v := range inst.mobs {
		if v.Controller == conn {
			inst.mobs[i].ChangeController(inst.findController())
		}
	}

	player := Players[conn]

	for _, other := range inst.players {
		other.Send(packet.MapPlayerLeft(player.Char().ID))
	}
}

func (inst *Instance) findController() mnet.MConnChannel {
	for _, p := range inst.players {
		return p
	}

	return nil
}

func (inst *Instance) findControllerExcept(conn mnet.MConnChannel) mnet.MConnChannel {
	for _, p := range inst.players {
		if p == conn {
			continue
		}

		return p
	}

	return nil
}

func (inst *Instance) generateMobSpawnID() int32 {
	var l int32
	for _, v := range inst.mobs {
		if v.SpawnID > l {
			l = v.SpawnID
		}
	}

	l++

	if l == 0 {
		l++
	}

	return l
}

func (inst *Instance) handleDeadMobs() {
	y := inst.mobs[:0]

	for _, mob := range inst.mobs {
		if mob.HP < 1 {
			mob.Controller.Send(packet.MobEndControl(mob.Mob))

			for _, id := range mob.Revives {
				inst.SpawnMobNoRespawn(id, inst.generateMobSpawnID(), mob.X, mob.Y, mob.Foothold, -3, mob.SpawnID, mob.FacesLeft())
				y = append(y, inst.mobs[len(inst.mobs)-1])
			}

			inst.send(packet.MobRemove(mob.Mob, 1)) // 0 keeps it there and is no longer attackable, 1 normal death, 2 disaapear instantly
		} else {
			y = append(y, mob)
		}
	}

	inst.mobs = y
}

func (inst *Instance) SpawnMob(mobID, spawnID int32, x, y, foothold int16, summonType int8, summonOption int32, facesLeft bool) {

}

func (inst *Instance) SpawnMobNoRespawn(mobID, spawnID int32, x, y, foothold int16, summonType int8, summonOption int32, facesLeft bool) {
	m, err := nx.GetMob(mobID)

	if err != nil {
		return
	}

	mob := def.CreateMob(spawnID, nx.Life{}, m, nil)
	mob.ID = mobID

	mob.X = x
	mob.Y = y
	mob.Foothold = foothold

	mob.Respawns = false

	mob.SummonType = summonType
	mob.SummonOption = summonOption

	mob.FaceLeft = facesLeft

	inst.send(packet.MobShow(mob))

	if summonType != -4 {
		mob.SummonType = -1
		mob.SummonOption = 0
	}

	inst.mobs = append(inst.mobs, gameMob{Mob: mob, mapID: inst.mapID})

	inst.mobs[len(inst.mobs)-1].Controller = inst.findController()
}
