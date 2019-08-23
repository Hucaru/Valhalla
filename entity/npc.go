package entity

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
)

type npc struct {
	controller mnet.Client
	id         int32
	spawnID    int32
	pos        pos
	foothold   int16
	faceLeft   bool
	rx0, rx1   int16
}

func createNpcFromID(spawnID int32, npcID int32) {
}

func createNpcFromData(spawnID int32, life nx.Life) npc {
	return npc{id: life.ID,
		spawnID:  spawnID,
		pos:      pos{x: life.X, y: life.Y},
		foothold: life.Foothold,
		faceLeft: life.FaceLeft,
		rx0:      life.Rx0,
		rx1:      life.Rx1}
}

func (n npc) Controller() mnet.Client {
	return n.controller
}
