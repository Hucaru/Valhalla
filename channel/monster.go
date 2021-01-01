package channel

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant/skill"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

type monster struct {
	controller, summoner   *player
	id                     int32
	spawnID                int32
	pos                    pos
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
	poison                 bool

	lastAttackTime int64
	lastSkillTime  int64
	skillTimes     map[byte]int64

	skillID    byte
	skillLevel byte
	statBuff   int32

	dmgTaken map[*player]int32

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

func createMonsterFromData(spawnID int32, life nx.Life, m nx.Mob, dropsItems, dropsMesos bool) monster {
	return monster{id: life.ID,
		spawnID:       spawnID,
		pos:           newPos(life.X, life.Y, life.Foothold),
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
		dmgTaken:      make(map[*player]int32),
		skills:        nx.GetMobSkills(life.ID),
		skillTimes:    make(map[byte]int64),
		poison:        false,
		lastHeal:      time.Now().Unix(),
	}
}

func createMonsterFromID(spawnID, id int32, p pos, controller *player, dropsItems, dropsMesos bool) (monster, error) {
	m, err := nx.GetMob(id)

	if err != nil {
		return monster{}, fmt.Errorf("Unknown mob id: %v", id)
	}

	// If this isn't working with regards to position make the foothold equal to player? nearest to pos?
	mob := createMonsterFromData(spawnID, nx.Life{ID: id, Foothold: p.foothold, X: p.x, Y: p.y, FaceLeft: true}, m, dropsItems, dropsMesos)

	mob.summoner = controller

	return mob, nil
}

func (m *monster) setController(controller *player, follow bool) {
	if controller == nil {
		return
	}

	m.controller = controller
	controller.send(packetMobControl(*m, follow))
}

func (m *monster) removeController() {
	if m.controller != nil {
		m.controller.send(packetMobEndControl(*m))
		m.controller = nil
	}
}

func (m *monster) acknowledgeController(moveID int16, movData movementFrag, allowedToUseSkill bool, skill, level byte) {
	m.pos.x = movData.x
	m.pos.y = movData.y
	m.pos.foothold = movData.foothold
	m.stance = movData.stance
	m.faceLeft = m.stance%2 == 1

	if m.controller == nil {
		return
	}

	m.controller.send(packetMobControlAcknowledge(m.spawnID, moveID, allowedToUseSkill, int16(m.mp), skill, level))
}

func (m monster) hasHPBar() (bool, int32, int32, int32, byte, byte) {
	return (m.boss && m.hpBgColour > 0), m.id, m.hp, m.maxHP, m.hpFgColour, m.hpBgColour
}

func (m *monster) performSkill(delay int16, skillLevel, skillID byte) {
	currentTime := time.Now().Unix()
	m.lastSkillTime = currentTime
	m.skillTimes[skillID] = currentTime

	if skillID != m.skillID || (m.statBuff&skill.MobStat.SealSkill > 0) {
		skillID = 0
		return
	}

	levels, err := nx.GetMobSkill(skillID)

	if err != nil {
		m.skillID = 0
		return
	}

	var skillData nx.MobSkill
	for i, v := range levels {
		if i == int(skillLevel) {
			skillData = v
		}
	}

	m.mp = m.mp - skillData.MpCon
	if m.mp < 0 {
		m.mp = 0
	}

	// Handle all the different skills!
	switch skillID {
	case skill.Mob.WeaponAttackUpAoe:
	case skill.Mob.MagicAttackUp:
	case skill.Mob.MagicAttackUpAoe:
	case skill.Mob.WeaponDefenceUp:
	case skill.Mob.WeaponDefenceUpAoe:
	case skill.Mob.MagicDefenceUp:
	case skill.Mob.MagicDefenceUpAoe:
	case skill.Mob.HealAoe:
	case skill.Mob.Seal:
	case skill.Mob.Darkness:
	case skill.Mob.Weakness:
	case skill.Mob.Stun:
	case skill.Mob.Curse:
	case skill.Mob.Poison:
	case skill.Mob.Slow:
	case skill.Mob.Dispel:
	case skill.Mob.Seduce:
	case skill.Mob.SendToTown:
	case skill.Mob.PoisonMist:
	case skill.Mob.CrazySkull:
	case skill.Mob.Zombify:
	case skill.Mob.WeaponImmunity:
	case skill.Mob.MagicImmunity:
	case skill.Mob.ArmorSkill:
	case skill.Mob.WeaponDamageReflect:
	case skill.Mob.MagicDamageReflect:
	case skill.Mob.AnyDamageReflect:
	case skill.Mob.McWeaponAttackUp:
	case skill.Mob.McMagicAttackUp:
	case skill.Mob.McWeaponDefenseUp:
	case skill.Mob.McMagicDefenseUp:
	case skill.Mob.McAccuracyUp:
	case skill.Mob.McAvoidUp:
	case skill.Mob.McSpeedUp:
	case skill.Mob.McSeal:
	case skill.Mob.Summon:
	}

}

func (m *monster) performAttack(attackID byte) {
	// do stuff
}

func (m *monster) giveDamage(damager *player, dmg ...int32) {
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

func (m *monster) kill(inst fieldInstance, plr *player) {
	inst.lifePool.removeMob(m.spawnID, 0x0)
	plr.giveEXP(m.exp, true, false)
}

func (m monster) displayBytes() []byte {
	p := mpacket.NewPacket()

	p.WriteInt32(m.spawnID)
	p.WriteByte(0x00) // control status?
	p.WriteInt32(m.id)

	p.WriteInt32(0) // some kind of status?

	p.WriteInt16(m.pos.x)
	p.WriteInt16(m.pos.y)

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

	p.WriteByte(bitfield)        // 0x08 - a summon, 0x04 - flying, 0x02 - ???, 0x01 - faces left
	p.WriteInt16(m.pos.foothold) // foothold to oscillate around
	p.WriteInt16(m.pos.foothold) // spawn foothold
	p.WriteInt8(m.summonType)

	if m.summonType == -3 || m.summonType >= 0 {
		p.WriteInt32(m.summonOption) // when -3 used to link mob to a death using spawnID
	}

	p.WriteInt32(0) // encode mob status
	return p
}

func (m monster) String() string {
	sid := strconv.Itoa(int(m.spawnID))
	mid := strconv.Itoa(int(m.id))

	hp := strconv.Itoa(int(m.hp))
	mhp := strconv.Itoa(int(m.maxHP))

	mp := strconv.Itoa(int(m.mp))
	mmp := strconv.Itoa(int(m.maxMP))

	return sid + "(" + mid + ") " + hp + "/" + mhp + " " + mp + "/" + mmp + " (" + m.pos.String() + ")"
}

func (m *monster) update(t time.Time) {
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

		m.healMob(regenhp, regenmp)
		m.lastHeal = checkTime
	}

}

// GetNextSkill returns the value of function chooseNextSkill
// The function chooseNextSkill identifies a random skill for the mob to use
// Various checks include MP consumption, cooldown, etc
func (m *monster) useChooseNextSkill() (byte, byte) {
	return chooseNextSkill(m)
}

func chooseNextSkill(mob *monster) (byte, byte) {
	var skillID, skillLevel byte

	skillsToChooseFrom := []byte{}

	for id := range mob.skills {

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
			case skill.Mob.WeaponAttackUp:
				fallthrough
			case skill.Mob.WeaponAttackUpAoe:
				alreadySet = mob.statBuff&skill.MobStat.PowerUp > 0

			case skill.Mob.MagicAttackUp:
				fallthrough
			case skill.Mob.MagicAttackUpAoe:
				alreadySet = mob.statBuff&skill.MobStat.MagicUp > 0

			case skill.Mob.WeaponDefenceUp:
				fallthrough
			case skill.Mob.WeaponDefenceUpAoe:
				alreadySet = mob.statBuff&skill.MobStat.PowerGuardUp > 0

			case skill.Mob.MagicDefenceUp:
				fallthrough
			case skill.Mob.MagicDefenceUpAoe:
				alreadySet = mob.statBuff&skill.MobStat.MagicGuardUp > 0

			case skill.Mob.WeaponImmunity:
				alreadySet = mob.statBuff&skill.MobStat.PhysicalImmune > 0

			case skill.Mob.MagicImmunity:
				alreadySet = mob.statBuff&skill.MobStat.MagicImmune > 0

			// case skill.Mob.WeaponDamageReflect:

			// case skill.Mob.MagicDamageReflect:

			case skill.Mob.McSpeedUp:
				alreadySet = mob.statBuff&skill.MobStat.Speed > 0

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

		for id, level := range mob.skills {
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

func (m monster) canUseSkill(skillPossible bool) (byte, byte) {
	// 10 second default cooldown
	if !skillPossible || (m.statBuff&skill.MobStat.SealSkill > 0) || (time.Now().Unix()-m.lastSkillTime) < 10 {
		return 0, 0
	}
	skillID, skillLevel := chooseNextSkill(&m)
	return skillID, skillLevel

}

func (m *monster) healMob(hp, mp int32) {
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

func (m monster) calculateHeal() (hp int32, mp int32) {
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

func packetMobControl(m monster, chase bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelControlMob)
	if chase {
		p.WriteByte(0x02) // 2 chase, 1 no chase, 0 no control
	} else {
		p.WriteByte(0x01)
	}

	p.Append(m.displayBytes())

	return p
}

func packetMobControlAcknowledge(mobID int32, moveID int16, allowedToUseSkill bool, mp int16, skill byte, level byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelControlMobAck)
	p.WriteInt32(mobID)
	p.WriteInt16(moveID)
	p.WriteBool(allowedToUseSkill)
	p.WriteInt16(mp) // check this shouldn't be int32 or uint16 as Zakum has 60,000 mp
	p.WriteByte(skill)
	p.WriteByte(level)

	return p
}

func packetMobEndControl(m monster) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelControlMob)
	p.WriteByte(0)
	p.WriteInt32(m.spawnID)

	return p
}

func packetMobShowHpChange(spawnID int32, dmg int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMobDamage)
	p.WriteInt32(spawnID)
	p.WriteByte(0)
	p.WriteInt32(dmg)

	return p
}
