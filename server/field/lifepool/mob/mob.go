package mob

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	skills "github.com/Hucaru/Valhalla/constant/skill"

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
	Skills                 map[byte]byte
	revives                []int32
	stance                 byte
	poison                 bool

	lastAttackTime int64
	lastSkillTime  int64
	skillTimes     map[byte]int64

	skillID    byte
	skillLevel byte
	statBuff   int32

	dmgTaken map[controller]int32

	dropsItems bool
	dropsMesos bool

	hpBgColour byte
	hpFgColour byte

	spawnInterval int64
	timeToSpawn   time.Time

	lastStatusUpdate int64
	lastHeal         int64
	lastTimeAttacked int64
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
		Skills:        nx.GetMobSkills(life.ID),
		skillTimes:    make(map[byte]int64),
		poison:        false,
		lastHeal:      time.Now().Unix(),
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

// SetHP of mob
func (m *Data) SetHP(hp int32) {
	m.hp = hp
}

// MP of mob
func (m Data) MP() int32 {
	return m.mp
}

// MaxMP of mob
func (m Data) MaxMP() int32 {
	return m.maxMP
}

// SetMP of mob
func (m *Data) SetMP(mp int32) {
	m.mp = mp
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

// StatBuff value of mob
func (m Data) StatBuff() int32 {
	return m.statBuff
}

// LastSkillTime value of mob
func (m Data) LastSkillTime() int64 {
	return m.lastSkillTime
}

// SetLastAttackTime of mob
func (m *Data) SetLastAttackTime(newTime int64) {
	m.lastAttackTime = newTime
}

// SetLastSkillTime of mob
func (m *Data) SetLastSkillTime(newTime int64) {
	m.lastSkillTime = newTime
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
func (mob *Data) PerformSkill(delay int16, skillLevel, skillID byte) {
	currentTime := time.Now().Unix()
	mob.lastSkillTime = currentTime
	mob.skillTimes[skillID] = currentTime

	if skillID != mob.skillID || (mob.statBuff&skills.MobStat.SealSkill > 0) {
		skillID = 0
		return
	}

	levels, err := nx.GetMobSkill(skillID)

	if err != nil {
		mob.skillID = 0
		return
	}

	var skillData nx.MobSkill
	for i, v := range levels {
		if i == int(skillLevel) {
			skillData = v
		}
	}

	mob.mp = mob.mp - skillData.MpCon
	if mob.mp < 0 {
		mob.mp = 0
	}

	// Handle all the different skills!
	switch skillID {
	case skills.Mob.WeaponAttackUpAoe:
	case skills.Mob.MagicAttackUp:
	case skills.Mob.MagicAttackUpAoe:
	case skills.Mob.WeaponDefenceUp:
	case skills.Mob.WeaponDefenceUpAoe:
	case skills.Mob.MagicDefenceUp:
	case skills.Mob.MagicDefenceUpAoe:
	case skills.Mob.HealAoe:
	case skills.Mob.Seal:
	case skills.Mob.Darkness:
	case skills.Mob.Weakness:
	case skills.Mob.Stun:
	case skills.Mob.Curse:
	case skills.Mob.Poison:
	case skills.Mob.Slow:
	case skills.Mob.Dispel:
	case skills.Mob.Seduce:
	case skills.Mob.SendToTown:
	case skills.Mob.PoisonMist:
	case skills.Mob.CrazySkull:
	case skills.Mob.Zombify:
	case skills.Mob.WeaponImmunity:
	case skills.Mob.MagicImmunity:
	case skills.Mob.ArmorSkill:
	case skills.Mob.WeaponDamageReflect:
	case skills.Mob.MagicDamageReflect:
	case skills.Mob.AnyDamageReflect:
	case skills.Mob.McWeaponAttackUp:
	case skills.Mob.McMagicAttackUp:
	case skills.Mob.McWeaponDefenseUp:
	case skills.Mob.McMagicDefenseUp:
	case skills.Mob.McAccuracyUp:
	case skills.Mob.McAvoidUp:
	case skills.Mob.McSpeedUp:
	case skills.Mob.McSeal:
	case skills.Mob.Summon:
	}

}

// PerformAttack - mob attack action
func (m *Data) PerformAttack(attackID byte) {
	// do stuff
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
	m.lastTimeAttacked = time.Now().Unix() // Is there a better place to put this?
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
	checkTime := t.Unix()
	m.lastStatusUpdate = checkTime

	if m.hp <= 0 {
		return
	}

	if m.poison {
		// Handle poison
		m.hp = m.hp - 10 // Need to adjust to poison amount based on poison level
	}

	// Update mob status if one is applied
	if (checkTime - m.lastHeal) > 30 {
		// Heal the mob
		regenhp, regenmp := m.calculateHeal()

		m.HealMob(regenhp, regenmp)
		m.lastHeal = checkTime
	}

}

// GetNextSkill returns the value of function chooseNextSkill
// The function chooseNextSkill identifies a random skill for the mob to use
// Various checks include MP consumption, cooldown, etc
func (m *Data) GetNextSkill() (byte, byte) {
	return chooseNextSkill(m)
}

func chooseNextSkill(mob *Data) (byte, byte) {
	var skillID, skillLevel byte

	skillsToChooseFrom := []byte{}

	for id := range mob.Skills {

		levels, err := nx.GetMobSkill(id)

		if err != nil {
			continue
		}

		if int(skillLevel) >= len(levels) {
			continue
		}

		skillData := levels[skillLevel]

		// Skill MP check
		if mob.mp < skillData.MpCon {
			continue
		}

		// Skill cooldown check
		if val, ok := mob.skillTimes[id]; ok {
			if (val + skillData.Interval) > time.Now().Unix() {
				continue
			}
		}

		// Check summon limit
		// if skillData.Limit {

		// }

		// Determine if stats can be buffed
		if mob.statBuff > 0 {
			alreadySet := false

			switch id {
			case skills.Mob.WeaponAttackUp:
				fallthrough
			case skills.Mob.WeaponAttackUpAoe:
				alreadySet = mob.statBuff&skills.MobStat.PowerUp > 0

			case skills.Mob.MagicAttackUp:
				fallthrough
			case skills.Mob.MagicAttackUpAoe:
				alreadySet = mob.statBuff&skills.MobStat.MagicUp > 0

			case skills.Mob.WeaponDefenceUp:
				fallthrough
			case skills.Mob.WeaponDefenceUpAoe:
				alreadySet = mob.statBuff&skills.MobStat.PowerGuardUp > 0

			case skills.Mob.MagicDefenceUp:
				fallthrough
			case skills.Mob.MagicDefenceUpAoe:
				alreadySet = mob.statBuff&skills.MobStat.MagicGuardUp > 0

			case skills.Mob.WeaponImmunity:
				alreadySet = mob.statBuff&skills.MobStat.PhysicalImmune > 0

			case skills.Mob.MagicImmunity:
				alreadySet = mob.statBuff&skills.MobStat.MagicImmune > 0

			// case skills.Mob.WeaponDamageReflect:

			// case skills.Mob.MagicDamageReflect:

			case skills.Mob.McSpeedUp:
				alreadySet = mob.statBuff&skills.MobStat.Speed > 0

			default:
			}

			if alreadySet {
				continue
			}

		}

		skillsToChooseFrom = append(skillsToChooseFrom, id)
	}

	if len(skillsToChooseFrom) > 0 {
		nextID := skillsToChooseFrom[rand.Intn(len(skillsToChooseFrom))]

		skillID = nextID

		for id, level := range mob.Skills {
			if id == nextID {
				skillLevel = level
			}
		}
	}

	if skillLevel == 0 {
		skillID = 0
	}

	return skillID, skillLevel
}

// CanUseSkill returns new skill for mob to use
func (m Data) CanUseSkill(skillPossible bool) (byte, byte) {
	// 10 second default cooldown
	if !skillPossible || (m.statBuff&skills.MobStat.SealSkill > 0) || (time.Now().Unix()-m.lastSkillTime) < 10 {
		return 0, 0
	}
	skillID, skillLevel := m.GetNextSkill()
	return skillID, skillLevel

}

// HealMob heals the mobs HP or MP
func (m *Data) HealMob(hp, mp int32) {
	if hp > 0 && m.hp < m.maxHP {
		newHP := m.hp + hp
		if newHP < 0 || newHP > m.maxHP {
			newHP = m.maxHP
		}
		m.hp = newHP
	}

	if mp > 0 && m.mp < m.maxMP {
		newMP := m.mp + mp
		if newMP < 0 || newMP > m.maxMP {
			newMP = m.maxMP
		}
		mp = newMP
	}
}

func (m Data) calculateHeal() (hp int32, mp int32) {
	hp, mp = 0, 0

	// Calculate HP regen amount
	hp = m.maxHP / 100

	// Calculate MP regen amount
	mp = m.maxMP / 100

	// Always return MP because attack time does not matter for regen.
	// If someone is bossing the boss will always need MP available to attack
	// Because of this we should always return MP
	if m.lastTimeAttacked-time.Now().Unix() < 60 {
		return 0, mp
	}

	// We are healing 1% of hp/mp.
	return hp, mp
}
