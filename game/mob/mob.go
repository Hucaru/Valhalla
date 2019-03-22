package mob

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
)

type SpawnInfo struct {
	Mob          Mob
	Count        int
	TimeCanSpawn int64
}

func CreateSpawnInfo(mob Mob) SpawnInfo {
	return SpawnInfo{
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

func Create(spawnID int32, life nx.Life, mob nx.Mob, summoner mnet.MConnChannel, mapID int32) Mob {
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
		m.Controller.Send(packetEndControl(*m))
	}

	m.Controller = newController
	newController.Send(packetControl(*m, false))
}

func (m *Mob) ShowTo(conn mnet.MConnChannel) {
	conn.Send(packetShow(*m))
}

func (m *Mob) RemoveFrom(conn mnet.MConnChannel, method byte) {
	conn.Send(packetRemove(*m, method)) // 0 keeps it there and is no longer attackable, 1 normal death, 2 disaapear instantly
}

func (m *Mob) RemoveController() {
	m.Controller.Send(packetEndControl(*m))
}

func (m *Mob) Acknowledge(moveID int16, allowedToUseSkill bool, skill byte, level byte) {
	m.Controller.Send(packetControlAcknowledge(m.SpawnID, moveID, allowedToUseSkill, int16(m.MP), skill, level))
}

// func (m *gameMob) FindNewControllerExcept(conn mnet.MConnChannel) {
// 	var newController mnet.MConnChannel

// 	for c, v := range Players {
// 		if v.char.MapID == m.mapID {
// 			if c == conn {
// 				continue
// 			} else {
// 				newController = c
// 			}
// 		}
// 	}

// 	if newController == nil {
// 		return
// 	}

// 	m.ChangeController(Players[newController])
// }
