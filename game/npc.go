package game

import "github.com/Hucaru/Valhalla/nx"

type npc struct {
	spawnID int32
	nx.Life
}

func createNpc(spawnID int32, life nx.Life) npc {
	return npc{spawnID, life}
}
