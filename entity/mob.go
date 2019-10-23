package entity

import (
	"fmt"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
)

type mob struct {
	controller, summoner   mnet.Client
	id                     int32
	spawnID                int32
	pos                    pos
	faceLeft               bool
	hp, mp                 int16
	maxHP, maxMP           int16
	hpRecovery, mpRecovery int32
	level                  int32
	exp                    int32
	maDamage               int32
	mdDamage               int32
	paDamage               int32
	pdDamage               int32
	summonType             int8 // -2: fade in spawn animation, -1: no spawn animation, 0: balrog summon effect?
	summonOption           int32
	boss                   bool
	undead                 bool
	elemAttr               int32
	invincible             bool
	speed                  int32
	eva                    int32
	acc                    int32
	link                   int32
	flySpeed               int32
	noRegen                int32
	skills                 map[byte]byte
	revives                []int32
	stance                 byte
	foothold               int16
}

func createMobFromData(spawnID int32, life nx.Life, m nx.Mob) mob {
	mob := mob{id: life.ID,
		spawnID:    spawnID,
		pos:        pos{x: life.X, y: life.Y},
		faceLeft:   life.FaceLeft,
		hp:         int16(m.HP),
		mp:         int16(m.HP),
		maxHP:      int16(m.MaxHP),
		maxMP:      int16(m.MaxMP),
		foothold:   life.Foothold,
		summonType: -2,
	}

	return mob
}

func createMobFromID(spawnID, id int32, p pos) (mob, error) {
	m, err := nx.GetMob(id)

	if err != nil {
		return mob{}, fmt.Errorf("Unknown mob id: %v", id)
	}

	// If this isn't working with regards to position make the foothold equal to player? nearest to pos?
	return createMobFromData(spawnID, nx.Life{Foothold: 0, X: p.x, Y: p.y, FaceLeft: true}, m), nil
}

func (m mob) Controller() mnet.Client {
	return m.controller
}

func (m *mob) SetController(conn mnet.Client, follow bool) {
	m.controller = conn
	conn.Send(PacketMobControl(*m, follow))
}

func (m *mob) AcknowledgeController(moveID int16, movData movementFrag) {
	m.pos.x = movData.x
	m.pos.y = movData.y
	m.foothold = movData.foothold
	m.stance = movData.stance

	var allowedToUseSkill bool = false
	var skill, level byte = 0, 0
	m.controller.Send(PacketMobControlAcknowledge(m.spawnID, moveID, allowedToUseSkill, m.mp, skill, level))
}

func (m mob) handleDeath(inst *instance) {

}
