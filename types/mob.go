package types

import (
	"github.com/Hucaru/Valhalla/nx"
)

type Mob struct {
	SpawnID int32
	nx.Life
	nx.Monster
}

func CreateMob(spawnID int32, life nx.Life, mob nx.Monster) Mob {
	return Mob{spawnID, life, mob}
}
