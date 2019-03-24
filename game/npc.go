package game

import (
	"github.com/Hucaru/Valhalla/nx"
)

type Npc struct {
	SpawnID int32
	nx.Life
}

func CreateNpc(spawnID int32, life nx.Life) Npc {
	return Npc{spawnID, life}
}
