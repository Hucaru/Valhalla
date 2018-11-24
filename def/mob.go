package def

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
)

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
	Respawns         bool
	nx.Life
	nx.Monster
}

func CreateMob(spawnID int32, life nx.Life, mob nx.Monster, summoner mnet.MConnChannel) Mob {
	return Mob{SpawnID: spawnID,
		Summoner:   summoner,
		Stance:     0,
		SkillTimes: make(map[byte]int64),
		Respawns:   true,
		Life:       life,
		Monster:    mob}
}
