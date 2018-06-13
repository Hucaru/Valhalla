package channel

import (
	"sync"

	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/nx"
)

var Mobs = mapleMobs{maps: make(map[uint32][]*MapleMob), mutex: &sync.RWMutex{}}

func GenerateMobs() {
	for mapID, stage := range nx.Maps {

		for spawnID, life := range stage.Life {
			if life.IsMob {
				mob := &MapleMob{}

				mob.SetID(life.ID)
				mob.SetSpawnID(uint32(spawnID + 1))
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

				mob.SetStatus(constants.MOB_STATUS_EMPTY)

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
	}
}

type mapleMobs struct {
	maps  map[uint32][]*MapleMob
	mutex *sync.RWMutex
}

func (m *mapleMobs) AddMob(mapID uint32, newMob *MapleMob) {
	m.mutex.Lock()
	m.maps[mapID] = append(m.maps[mapID], newMob)
	m.mutex.Unlock()
}

func (p *mapleMobs) OnMob(mapID, mobID uint32, action func(mob *MapleMob)) {
	p.mutex.RLock()
	for _, value := range p.maps[mapID] {
		if value.GetSpawnID() == mobID {
			action(value)
		}
	}
	p.mutex.RUnlock()
}

func (p *mapleMobs) OnMobs(mapID uint32, action func(mob *MapleMob)) {
	p.mutex.RLock()
	for _, value := range p.maps[mapID] {
		action(value)
	}
	p.mutex.RUnlock()
}
