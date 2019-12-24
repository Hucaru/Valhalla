package mob

import (
	"fmt"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/server/movement"
	"github.com/Hucaru/Valhalla/server/pos"
)

// Controller of mob
type Controller interface {
	Send(mpacket.Packet)
}

// Data for mob
type Data struct {
	controller, summoner   Controller
	id                     int32
	spawnID                int32
	pos                    pos.Data
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

	lastAttackTime int64
	lastSkillTime  int64
	skillTimes     map[byte]int64

	DmgTaken map[mnet.Client]int32
}

// CreateFromData - creates a mob from nx data
func CreateFromData(spawnID int32, life nx.Life, m nx.Mob) Data {
	return Data{id: life.ID,
		spawnID:    spawnID,
		pos:        pos.New(life.X, life.Y),
		faceLeft:   life.FaceLeft,
		hp:         int16(m.HP),
		mp:         int16(m.HP),
		maxHP:      int16(m.MaxHP),
		maxMP:      int16(m.MaxMP),
		foothold:   life.Foothold,
		summonType: -2,
	}
}

// CreateFromID - creates a mob from an id and position data
func CreateFromID(spawnID, id int32, p pos.Data, conn mnet.Client) (Data, error) {
	m, err := nx.GetMob(id)

	if err != nil {
		return Data{}, fmt.Errorf("Unknown mob id: %v", id)
	}

	// If this isn't working with regards to position make the foothold equal to player? nearest to pos?
	mob := CreateFromData(spawnID, nx.Life{Foothold: 0, X: p.X(), Y: p.Y(), FaceLeft: true}, m)
	mob.summoner = conn
	return mob, nil
}

// Controller of mob
func (m Data) Controller() Controller {
	return m.controller
}

// SetController of mob
func (m *Data) SetController(controller Controller, follow bool) {
	m.controller = controller
	controller.Send(packetMobControl(*m, follow))
}

// RemoveController from mob
func (m *Data) RemoveController() {
	m.controller.Send(packetMobEndControl(*m))
}

// AcknowledgeController movement bytes
func (m *Data) AcknowledgeController(moveID int16, movData movement.Frag) {
	m.pos.SetX(movData.X())
	m.pos.SetY(movData.Y())
	m.foothold = movData.Foothold()
	m.stance = movData.Stance()

	allowedToUseSkill := false
	var skill, level byte = 0, 0
	m.controller.Send(packetMobControlAcknowledge(m.spawnID, moveID, allowedToUseSkill, m.mp, skill, level))
}

// SpawnID of mob
func (m Data) SpawnID() int32 {
	return m.spawnID
}

// SetSummonType of mob
func (m *Data) SetSummonType(v int8) {
	m.summonType = v
}

// PerformSkill - mob skill action
func (m *Data) PerformSkill(delay int16, skillLevel, skillID byte) {

}

// PerformAttack - mob attack action
func (m *Data) PerformAttack(attackID byte) {

}

type instance interface {
}

// HandleDamage on the mob
func (m Data) HandleDamage(dmg int32, conn mnet.Client, inst instance) {

}

// DisplayBytes to show mob
func (m Data) DisplayBytes() []byte {
	p := mpacket.NewPacket()

	p.WriteInt32(m.spawnID)
	p.WriteByte(0x00) // control status?
	p.WriteInt32(m.id)

	p.WriteInt32(0) // some kind of status?

	p.WriteInt16(m.pos.X())
	p.WriteInt16(m.pos.Y())

	var bitfield byte

	if m.summoner != nil {
		bitfield = 0x08
	} else {
		bitfield = 0x02
	}

	if m.faceLeft {
		bitfield |= 0x01
	} else {
		bitfield |= 0x04
	}

	if m.stance%2 == 1 {
		bitfield |= 0x01
	} else {
		bitfield |= 0
	}

	if m.flySpeed > 0 {
		bitfield |= 0x04
	}

	p.WriteByte(bitfield)    // 0x08 - a summon, 0x04 - flying, 0x02 - ???, 0x01 - faces left
	p.WriteInt16(m.foothold) // foothold to oscillate around
	p.WriteInt16(m.foothold) // spawn foothold
	p.WriteInt8(m.summonType)

	if m.summonType == -3 || m.summonType >= 0 {
		p.WriteInt32(m.summonOption) // some sort of summoning options, not sure what this is
	}

	p.WriteInt32(0) // encode mob status

	return p
}
