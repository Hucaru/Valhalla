package types

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
)

type Mob struct {
	SpawnID    int32
	Summoner   mnet.MConnChannel
	Controller mnet.MConnChannel
	Stance     byte
	nx.Life
	nx.Monster
}

func CreateMob(spawnID int32, life nx.Life, mob nx.Monster, summoner mnet.MConnChannel) Mob {
	return Mob{SpawnID: spawnID,
		Summoner: summoner,
		Stance:   0,
		Life:     life,
		Monster:  mob}
}
