package game

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
)

type MobSI struct {
	Mob          Mob
	Count        int
	TimeCanSpawn int64
}

func CreateMobSpawnInfo(mob Mob) MobSI {
	return MobSI{
		Mob:          mob,
		Count:        0,
		TimeCanSpawn: 0}
}

type Mob struct {
	SpawnID          int32
	Summoner         mnet.MConnChannel
	Controller       mnet.MConnChannel
	Stance           byte
	SkillTimes       map[byte]int64
	CanUseSkill      bool
	LastSkillUseTime int64
	SkillID          byte
	SkillLevel       byte
	StatBuff         int32
	LastAttackTime   int64

	nx.Life
	nx.Mob

	mapID    int32
	DmgTaken map[mnet.MConnChannel]int32
}

func CreateMob(spawnID int32, life nx.Life, mob nx.Mob, summoner mnet.MConnChannel, mapID int32) Mob {
	return Mob{SpawnID: spawnID,
		Summoner:   summoner,
		Stance:     0,
		SkillTimes: make(map[byte]int64),
		Life:       life,
		Mob:        mob,
		mapID:      mapID,
		DmgTaken:   make(map[mnet.MConnChannel]int32)}
}

func (m Mob) FacesLeft() bool {
	return m.Stance%2 != 0
}

func (m *Mob) GiveDamage(conn mnet.MConnChannel, damages []int32) {
	if m.HP > 0 && m.Controller != conn {
		m.ChangeController(conn)
	}

	for _, dmg := range damages {
		if dmg > m.HP {
			m.HP = 0
			m.DmgTaken[conn] += m.HP
		} else {
			m.HP -= dmg
			m.DmgTaken[conn] += dmg
		}
	}
}

func (m *Mob) ChangeController(newController mnet.MConnChannel) {
	if newController == nil {
		m.Controller = nil
		return
	}

	if m.Controller == newController {
		return
	}

	if m.Controller != nil {
		m.Controller.Send(PacketMobEndControl(*m))
	}

	m.Controller = newController
	newController.Send(PacketMobControl(*m, false))
}

func (m *Mob) RemoveController() {
	m.Controller.Send(PacketMobEndControl(*m))
	m.Controller = nil
}

func (m *Mob) Acknowledge(moveID int16, allowedToUseSkill bool, skill byte, level byte) {
	m.Controller.Send(PacketMobControlAcknowledge(m.SpawnID, moveID, allowedToUseSkill, int16(m.MP), skill, level))
}

func (m *Mob) ResetAggro() {
	m.Controller.Send(PacketMobEndControl(*m))
	m.Controller.Send(PacketMobControl(*m, false))
}
