package mob

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/server/movement"
	"github.com/Hucaru/Valhalla/server/pos"
)

// Controller of mob
type controller interface {
	Conn() mnet.Client
	Send(mpacket.Packet)
}

type instance interface {
	Send(mpacket.Packet) error
	RemoveMob(int32, byte) error
	NextID() int32
	SpawnReviveMob(Data, interface{})
	ShowMobBossHPBar(Data)
}

type sender interface {
	Send(mpacket.Packet) error
}

type player interface {
	MapID() int32
	GiveEXP(int32, bool, bool)
}

// Data for mob
type Data struct {
	controller, summoner   controller
	id                     int32
	spawnID                int32
	pos                    pos.Data
	faceLeft               bool
	hp, mp                 int32
	maxHP, maxMP           int32
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

	lastAttackTime int64
	lastSkillTime  int64
	skillTimes     map[byte]int64

	dmgTaken map[controller]int32

	dropsItems bool
	dropsMesos bool

	hpBgColour byte
	hpFgColour byte

	spawnInterval int64
	timeToSpawn   time.Time
}

// CreateFromData - creates a mob from nx data
func CreateFromData(spawnID int32, life nx.Life, m nx.Mob, dropsItems, dropsMesos bool) Data {
	return Data{id: life.ID,
		spawnID:       spawnID,
		pos:           pos.New(life.X, life.Y, life.Foothold),
		faceLeft:      life.FaceLeft,
		hp:            m.HP,
		mp:            m.MP,
		maxHP:         m.MaxHP,
		maxMP:         m.MaxMP,
		exp:           int32(m.Exp),
		revives:       m.Revives,
		summonType:    -2,
		boss:          m.Boss >= 0,
		hpBgColour:    byte(m.HPTagBGColor),
		hpFgColour:    byte(m.HPTagColor),
		spawnInterval: life.MobTime,
		dmgTaken:      make(map[controller]int32),
	}
}

// CreateFromID - creates a mob from an id and position data
func CreateFromID(spawnID, id int32, p pos.Data, controller controller, dropsItems, dropsMesos bool) (Data, error) {
	m, err := nx.GetMob(id)

	if err != nil {
		return Data{}, fmt.Errorf("Unknown mob id: %v", id)
	}

	// If this isn't working with regards to position make the foothold equal to player? nearest to pos?
	mob := CreateFromData(spawnID, nx.Life{ID: id, Foothold: p.Foothold(), X: p.X(), Y: p.Y(), FaceLeft: true}, m, dropsItems, dropsMesos)
	mob.summoner = controller

	return mob, nil
}

// Controller of mob
func (m Data) Controller() controller {
	return m.controller
}

// SetController of mob
func (m *Data) SetController(controller controller, follow bool) {
	if controller == nil {
		return
	}

	m.controller = controller
	controller.Send(packetMobControl(*m, follow))
}

// RemoveController from mob
func (m *Data) RemoveController() {
	if m.controller != nil {
		m.controller.Send(packetMobEndControl(*m))
		m.controller = nil
	}
}

// AcknowledgeController movement bytes
func (m *Data) AcknowledgeController(moveID int16, movData movement.Frag, allowedToUseSkill bool, skill, level byte) {
	m.pos.SetX(movData.X())
	m.pos.SetY(movData.Y())
	m.pos.SetFoothold(movData.Foothold())
	m.stance = movData.Stance()
	m.faceLeft = m.stance%2 == 1

	if m.controller == nil {
		return
	}

	m.controller.Send(packetMobControlAcknowledge(m.spawnID, moveID, allowedToUseSkill, int16(m.mp), skill, level))
}

// ID of mob
func (m Data) ID() int32 {
	return m.id
}

// SpawnID of mob
func (m Data) SpawnID() int32 {
	return m.spawnID
}

// SetSpawnID of mob
func (m *Data) SetSpawnID(v int32) {
	m.spawnID = v
}

// SetSummonType of mob
func (m *Data) SetSummonType(v int8) {
	m.summonType = v
}

// SummonType of mob
func (m Data) SummonType() int8 {
	return m.summonType
}

// SetSummonOption of mob
func (m *Data) SetSummonOption(v int32) {
	m.summonOption = v
}

// FaceLeft property
func (m Data) FaceLeft() bool {
	return m.faceLeft
}

// SetFaceLeft property
func (m *Data) SetFaceLeft(v bool) {
	m.faceLeft = v
}

// HP of mob
func (m Data) HP() int32 {
	return m.hp
}

// MaxHP of mob
func (m Data) MaxHP() int32 {
	return m.maxHP
}

// Exp of mob
func (m Data) Exp() int32 {
	return m.exp
}

// Revives this mob spawns
func (m Data) Revives() []int32 {
	return m.revives
}

// Pos of the mob
func (m Data) Pos() pos.Data {
	return m.pos
}

// Boss value of mob
func (m Data) Boss() bool {
	return m.boss
}

// HasHPBar that can be shown
func (m Data) HasHPBar() (bool, int32, int32, int32, byte, byte) {
	return (m.boss && m.hpBgColour > 0), m.id, m.hp, m.maxHP, m.hpFgColour, m.hpBgColour
}

// SpawnInterval between mob spawning
func (m Data) SpawnInterval() int64 {
	return m.spawnInterval
}

// TimeToSpawn for boss monsters
func (m Data) TimeToSpawn() time.Time {
	return m.timeToSpawn
}

// SetTimeToSpawn for the mob
func (m *Data) SetTimeToSpawn(t time.Time) {
	m.timeToSpawn = t
}

// PerformSkill - mob skill action
func (m *Data) PerformSkill(delay int16, skillLevel, skillID byte) {

}

// PerformAttack - mob attack action
func (m *Data) PerformAttack(attackID byte) {

}

// GiveDamage to mob
func (m *Data) GiveDamage(damager controller, dmg ...int32) {
	for _, v := range dmg {
		if v > m.hp {
			v = m.hp
		}

		m.hp -= v

		if damager == nil {
			return
		}

		if _, ok := m.dmgTaken[damager]; ok {
			m.dmgTaken[damager] += v
		} else {
			m.dmgTaken[damager] = v
		}
	}
}

// GetDamage done to mob
func (m Data) GetDamage() map[controller]int32 {
	return m.dmgTaken
}

// Kill the mob silently
func (m *Data) Kill(inst instance, plr player) {
	inst.RemoveMob(m.spawnID, 0x0)
	plr.GiveEXP(m.exp, true, false)
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

	p.WriteByte(bitfield)          // 0x08 - a summon, 0x04 - flying, 0x02 - ???, 0x01 - faces left
	p.WriteInt16(m.pos.Foothold()) // foothold to oscillate around
	p.WriteInt16(m.pos.Foothold()) // spawn foothold
	p.WriteInt8(m.summonType)

	if m.summonType == -3 || m.summonType >= 0 {
		p.WriteInt32(m.summonOption) // when -3 used to link mob to a death using spawnID
	}

	p.WriteInt32(0) // encode mob status
	return p
}

func (m Data) String() string {
	sid := strconv.Itoa(int(m.spawnID))
	mid := strconv.Itoa(int(m.id))

	hp := strconv.Itoa(int(m.hp))
	mhp := strconv.Itoa(int(m.maxHP))

	mp := strconv.Itoa(int(m.mp))
	mmp := strconv.Itoa(int(m.maxMP))

	return sid + "(" + mid + ") " + hp + "/" + mhp + " " + mp + "/" + mmp + " (" + m.pos.String() + ")"
}

// Update mob for status changes e.g. posion, hp/mp recover, finding a new controller after inactivity
func (m *Data) Update(t time.Time) {

}
