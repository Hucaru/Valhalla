package def

import (
	"github.com/Hucaru/Valhalla/nx"
)

type NPC struct {
	SpawnID int32
	nx.Life
}

func CreateNPC(spawnID int32, life nx.Life) NPC {
	return NPC{spawnID, life}
}
