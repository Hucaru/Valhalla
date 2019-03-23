package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

type Instance struct {
	mapID                int32
	npcs                 []def.NPC
	mobs                 []Mob
	spawnableMobs        []*MobSI
	players              []mnet.MConnChannel
	workDispatch         chan func()
	previousMobSpawnTime int64
	mapData              nx.Map
}

func createInstanceFromMapData(mapData nx.Map, mapID int32, dispatcher chan func()) *Instance {
	npcs := []def.NPC{}
	mobs := []Mob{}
	spawnableMobs := []*MobSI{}

	for _, l := range mapData.Mobs {
		nxMob, err := nx.GetMob(l.ID)

		if err != nil {
			continue
		}

		newMob := CreateMob(int32(len(mobs)+1), l, nxMob, nil, mapID)
		mobSpawn := CreateMobSpawnInfo(newMob)
		mobSpawn.Count++

		mobs = append(mobs, mobSpawn.Mob)
		spawnableMobs = append(spawnableMobs, &mobSpawn)
	}

	for _, l := range mapData.NPCs {
		npcs = append(npcs, def.CreateNPC(int32(len(npcs)), l))
	}

	inst := &Instance{mapID: mapID,
		npcs:          npcs,
		mobs:          mobs,
		spawnableMobs: spawnableMobs,
		workDispatch:  dispatcher,
		mapData:       mapData}

	// Periodic map work
	go func(inst *Instance) {
		timer := time.NewTicker(1000 * time.Millisecond)
		quit := make(chan bool)

		for {
			select {
			case <-timer.C:
				inst.workDispatch <- func() {
					if inst == nil {
						quit <- true
					} else {
						inst.periodicWork()
					}
				}
			case <-quit:
				return
			}
		}

	}(inst)

	return inst
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
			mob.ShowTo(conn)

			if mob.Controller == nil {
				inst.mobs[i].ChangeController(conn)
			}
		}
	}

	for _, npc := range inst.npcs {
		conn.Send(PacketNpcShow(npc))
		conn.Send(PacketNpcSetController(npc.SpawnID, true))
	}

	player := Players[conn]

	for _, other := range inst.players {
		otherPlayer := Players[other]
		otherPlayer.Send(PacketMapPlayerEnter(player.Char()))

		player.Send(PacketMapPlayerEnter(otherPlayer.Char()))

		if otherPlayer.RoomID > 0 {
			r := Rooms[otherPlayer.RoomID]

			switch r.(type) {
			case *OmokRoom:
				omokRoom := r.(*OmokRoom)

				if omokRoom.IsOwner(other) {
					player.Send(PacketMapShowGameBox(otherPlayer.Char().ID, omokRoom.ID, byte(omokRoom.RoomType), omokRoom.BoardType, omokRoom.Name, bool(len(omokRoom.Password) > 0), omokRoom.InProgress, 0x01))
				}
			case *MemoryRoom:
				memoryRoom := r.(*MemoryRoom)
				if memoryRoom.IsOwner(other) {
					player.Send(PacketMapShowGameBox(otherPlayer.Char().ID, memoryRoom.ID, byte(memoryRoom.RoomType), memoryRoom.BoardType, memoryRoom.Name, bool(len(memoryRoom.Password) > 0), memoryRoom.InProgress, 0x01))
				}
			}
		}
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
		other.Send(PacketMapPlayerLeft(player.Char().ID))
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

func (inst *Instance) SpawnMob(mobID, spawnID int32, x, y, foothold int16, summonType int8, summonOption int32, facesLeft bool) {
	m, err := nx.GetMob(mobID)

	if err != nil {
		return
	}

	mob := CreateMob(spawnID, nx.Life{}, m, nil, inst.mapID)
	mob.ID = mobID

	mob.X = x
	mob.Y = y
	mob.Foothold = foothold

	mob.SummonType = summonType
	mob.SummonOption = summonOption

	mob.FaceLeft = facesLeft

	for _, v := range inst.players {
		mob.ShowTo(v)
	}

	if summonType != -4 {
		mob.SummonType = -1
		mob.SummonOption = 0
	}

	inst.mobs = append(inst.mobs, mob)

	inst.mobs[len(inst.mobs)-1].ChangeController(inst.findController())
}

func (inst *Instance) handleDeadMobs() {
	y := inst.mobs[:0]

	for _, mob := range inst.mobs {
		if mob.HP < 1 {
			mob.RemoveController()

			for _, id := range mob.Revives {
				inst.SpawnMob(id, inst.generateMobSpawnID(), mob.X, mob.Y, mob.Foothold, -3, mob.SpawnID, mob.FacesLeft())
				y = append(y, inst.mobs[len(inst.mobs)-1])
			}

			if mob.Exp > 0 {
				for player, _ := range mob.DmgTaken {
					p, err := Players.GetFromConn(player)

					if err != nil {
						continue
					}

					// perform exp calculation

					p.GiveEXP(int32(mob.Exp), true, false)
				}
			}

			for _, v := range inst.players {
				mob.RemoveFrom(v, 1)
			}

			for _, spm := range inst.spawnableMobs {
				if spm.Mob.ID == mob.ID && spm.Count > 0 {
					spm.Count--
					spm.TimeCanSpawn = time.Now().UnixNano()/int64(time.Millisecond) + spm.Mob.MobTime
				}
			}
		} else {
			y = append(y, mob)
		}
	}

	inst.mobs = y
}

func (inst *Instance) capacity() int {
	// if no mob capacity limit flag present return current mob count

	if len(inst.players) > (inst.mapData.MobCapacityMin / 2) {
		if len(inst.players) < (inst.mapData.MobCapacityMin * 2) {
			return inst.mapData.MobCapacityMin + (inst.mapData.MobCapacityMax-inst.mapData.MobCapacityMin)*(2*len(inst.players)-inst.mapData.MobCapacityMin)/(3*inst.mapData.MobCapacityMin)
		}

		return inst.mapData.MobCapacityMax
	}

	return inst.mapData.MobCapacityMin
}

func (inst *Instance) handleMobRespawns(currentTime int64) {
	if currentTime-inst.previousMobSpawnTime < 7000 {
		return
	}

	inst.previousMobSpawnTime = currentTime

	capacity := inst.capacity()

	if capacity < 0 {
		return
	}

	amountCanSpawn := capacity - len(inst.mobs)

	if amountCanSpawn < 1 {
		return
	}

	mobsToSpawn := []*MobSI{}

	for _, spm := range inst.spawnableMobs {
		addInfront := true
		regenInterval := spm.Mob.MobTime

		if regenInterval == 0 { // Standard mobs
			anyMobSpawned := len(inst.mobs) != 0

			if anyMobSpawned {
				rect := nx.Rectangle{int(spm.Mob.X - 100), int(spm.Mob.Y - 100), int(spm.Mob.X + 100), int(spm.Mob.Y + 100)}
				for _, currentMob := range inst.mobs {
					if !rect.Contains(int(currentMob.X), int(currentMob.Y)) {
						continue
					}
				}
			} else {
				addInfront = false
			}
		} else if regenInterval < 0 { // ?
			fmt.Println("Hit less than zero regen interval for", spm)
			// if not reset continue
		} else { // Timer mobs
			if spm.Count != 0 {
				continue
			}

			if currentTime-spm.TimeCanSpawn < 0 {
				continue
			}
		}

		if addInfront {
			mobsToSpawn = append([]*MobSI{spm}, mobsToSpawn...)
		} else {
			mobsToSpawn = append(mobsToSpawn, spm)
		}
	}

	for len(mobsToSpawn) > 0 && amountCanSpawn > 0 {
		spm := mobsToSpawn[0]

		if spm.Mob.MobTime == 0 {
			ind := rand.Intn(len(mobsToSpawn))
			spm = mobsToSpawn[ind]
			mobsToSpawn = append(mobsToSpawn[:ind], mobsToSpawn[ind+1:]...)
		}

		inst.SpawnMob(spm.Mob.ID, inst.generateMobSpawnID(), spm.Mob.X, spm.Mob.Y, spm.Mob.Foothold, -2, 0, spm.Mob.FaceLeft)
		amountCanSpawn--
		spm.Count++

		mobsToSpawn = mobsToSpawn[1:]
	}
}

func (inst *Instance) periodicWork() {
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	// Update drops
	// Update mist
	// update portals

	if len(inst.players) > 0 {
		inst.handleMobRespawns(currentTime)
		// check vac hack

		// for each character
		// tick for map dmg e.g. drowning
		// if pet present perform duties
	}
}
