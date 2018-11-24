package channel

import (
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/consts"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/game/packet"

	"github.com/Hucaru/Valhalla/nx"
)

var Mobs = mapleMobs{alive: make(map[int32][]*MapleMob), dead: make(map[int32][]*MapleMob), mutex: &sync.RWMutex{}}

func GenerateMobs() {
	for mapID, stage := range nx.Maps {

		for spawnID, life := range stage.Life {
			if life.IsMob {
				mob := &MapleMob{}

				mob.SetID(life.ID)
				mob.SetSpawnID(int32(spawnID + 1))
				mob.SetX(life.X)
				mob.SetSx(life.X)
				mob.SetY(life.Y)
				mob.SetSy(life.Y)
				mob.SetRx0(life.Rx0)
				mob.SetRx1(life.Rx1)
				mob.SetFoothold(life.Fh)
				mob.SetFace(life.F)

				mob.SetMobTime(life.MobTime)
				mob.SetRespawns(true)

				mob.SetStatus(consts.MOB_STATUS_EMPTY)

				mon := nx.Mob[life.ID]

				mob.SetBoss(mon.Boss)
				mob.SetEXP(mon.Exp)
				mob.SetMaxHp(mon.MaxHp)
				mob.SetHp(mon.MaxHp)
				mob.SetMaxMp(mon.MaxMp)
				mob.SetMp(mon.MaxMp)
				mob.SetLevel(mon.Level)

				Mobs.AddMob(mapID, mob)
			}
		}

		go Mobs.mobRespawner(mapID)
	}
}

type mapleMobs struct {
	alive map[int32][]*MapleMob
	dead  map[int32][]*MapleMob
	mutex *sync.RWMutex
}

func (m *mapleMobs) AddMob(mapID int32, newMob *MapleMob) {
	m.mutex.Lock()
	m.alive[mapID] = append(m.alive[mapID], newMob)
	m.mutex.Unlock()
}

func (m *mapleMobs) OnMob(mapID, mobID int32, action func(mob *MapleMob)) {
	m.mutex.RLock()
	for _, value := range m.alive[mapID] {
		if value.GetSpawnID() == mobID {
			action(value)
		}
	}
	m.mutex.RUnlock()
}

func (m *mapleMobs) OnMobs(mapID int32, action func(mob *MapleMob)) {
	m.mutex.RLock()
	for _, value := range m.alive[mapID] {
		action(value)
	}
	m.mutex.RUnlock()
}

func (m *mapleMobs) MobTakeDamage(mapID, mobID int32, damage []int32) int32 {
	var exp int32

	var index = -1

	m.mutex.Lock()
	for i, mob := range m.alive[mapID] {
		if mob.GetSpawnID() == mobID {

			for _, dmg := range damage {
				if dmg > mob.GetHp() {
					// mob death
					exp = mob.GetEXP()
					Maps.GetMap(mapID).SendPacket(packet.MobRemove(mob, 1))
					mob.SetDeathTime(time.Now().Unix())
					index = i
					break
				} else {
					mob.SetHp(mob.GetHp() - dmg)
				}
			}
		}
	}

	if index > -1 {
		m.alive[mapID][index].SetHp(m.alive[mapID][index].GetMaxHp())
		m.alive[mapID][index].SetMp(m.alive[mapID][index].GetMaxMp())
		m.alive[mapID][index].SetX(m.alive[mapID][index].GetSx())
		m.alive[mapID][index].SetY(m.alive[mapID][index].GetSy())

		m.dead[mapID] = append(m.dead[mapID], m.alive[mapID][index])

		copy(m.alive[mapID][index:], m.alive[mapID][index+1:])
		m.alive[mapID][len(m.alive[mapID])-1] = nil
		m.alive[mapID] = m.alive[mapID][:len(m.alive[mapID])-1]
	}
	m.mutex.Unlock()

	return exp
}

func (m *mapleMobs) SpawnMob(mapID int32, mob *MapleMob) {
	Maps.GetMap(mapID).OnPlayers(func(conn mnet.MConnChannel) bool {
		mob.SetController(conn, true)
		return true
	})

	m.AddMob(mapID, mob)
	Maps.GetMap(mapID).SendPacket(packet.MobShow(mob, true))
}

func (m *mapleMobs) mobRespawner(mapID int32) {
	for {
		// Need to find proper way of handling respawns
		time.Sleep(time.Second * 5)

		m.mutex.RLock()
		size := len(m.dead[mapID])
		m.mutex.RUnlock()

		if size > 0 {
			m.mutex.Lock()
			deadMobs := m.dead[mapID]
			m.dead = nil // manually free memory just to be sure
			m.dead = make(map[int32][]*MapleMob)
			m.mutex.Unlock()

			for _, mob := range deadMobs {
				m.SpawnMob(mapID, mob)
			}
		}
	}
}
