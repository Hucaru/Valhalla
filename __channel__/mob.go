package channel

import (
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/consts"
	"github.com/Hucaru/Valhalla/consts/skills"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/packets"
)

type MapleMob struct {
	mapleNpc
	exp, status                        int32
	hp, maxHp, mp, maxMp, flySpeed     int32
	boss, respawns                     bool
	level, nextSkillID, nextSkillLevel byte
	sx, sy                             int16
	mobTime, deathTime, respawnTime    int64
	controller, summoner               mnet.MConnChannel

	lastSkillUseTime int64
	nSpawns          int16
	skillTimes       map[byte]int64
}

func (m *MapleMob) GetEXP() int32              { return m.exp }
func (m *MapleMob) SetEXP(exp int32)           { m.exp = exp }
func (m *MapleMob) GetStatus() int32           { return m.status }
func (m *MapleMob) SetStatus(status int32)     { m.status = status }
func (m *MapleMob) GetHp() int32               { return m.hp }
func (m *MapleMob) SetHp(hp int32)             { m.hp = hp }
func (m *MapleMob) GetMaxHp() int32            { return m.maxHp }
func (m *MapleMob) SetMaxHp(maxHp int32)       { m.maxHp = maxHp }
func (m *MapleMob) GetMp() int32               { return m.mp }
func (m *MapleMob) SetMp(mp int32)             { m.mp = mp }
func (m *MapleMob) GetMaxMp() int32            { return m.maxMp }
func (m *MapleMob) SetMaxMp(maxMp int32)       { m.maxMp = maxMp }
func (m *MapleMob) GetFlySpeed() int32         { return m.flySpeed }
func (m *MapleMob) SetFlySpeed(flySpeed int32) { m.flySpeed = flySpeed }
func (m *MapleMob) GetBoss() bool              { return m.boss }
func (m *MapleMob) SetBoss(boss bool)          { m.boss = boss }

func (m *MapleMob) GetLevel() byte      { return m.level }
func (m *MapleMob) SetLevel(level byte) { m.level = level }

func (m *MapleMob) GetNextSkillID() byte                  { return m.nextSkillID }
func (m *MapleMob) SetNextSkillID(nextSkillID byte)       { m.nextSkillID = nextSkillID }
func (m *MapleMob) GetNextSkillLevel() byte               { return m.nextSkillLevel }
func (m *MapleMob) SetNextSkillLevel(nextSkillLevel byte) { m.nextSkillLevel = nextSkillLevel }

func (m *MapleMob) GetSx() int16                     { return m.sx }
func (m *MapleMob) SetSx(sx int16)                   { m.sx = sx }
func (m *MapleMob) GetSy() int16                     { return m.sy }
func (m *MapleMob) SetSy(sy int16)                   { m.sy = sy }
func (m *MapleMob) GetRespawns() bool                { return m.respawns }
func (m *MapleMob) SetRespawns(respawns bool)        { m.respawns = respawns }
func (m *MapleMob) GetMobTime() int64                { return m.mobTime }
func (m *MapleMob) SetMobTime(mobTime int64)         { m.mobTime = mobTime }
func (m *MapleMob) GetDeathTime() int64              { return m.deathTime }
func (m *MapleMob) SetDeathTime(mobTime int64)       { m.deathTime = mobTime }
func (m *MapleMob) GetRespawnTime() int64            { return m.respawnTime }
func (m *MapleMob) SetRespawnTime(respawnTime int64) { m.respawnTime = respawnTime }

func (m *MapleMob) SetSummoner(summoner mnet.MConnChannel) { m.summoner = summoner }
func (m *MapleMob) GetSummoner() mnet.MConnChannel         { return m.summoner }

func (m *MapleMob) GetController() mnet.MConnChannel { return m.controller }

func (m *MapleMob) SetController(controller mnet.MConnChannel, isSpawn bool) {
	m.controller = controller
	m.controller.Send(packets.MobControl(m, isSpawn))
}

func (m *MapleMob) RemoveController() {
	m.controller.Send(packets.MobEndControl(m))
	m.controller = nil
}

func (m *MapleMob) Spawn(conn mnet.MConnChannel) {
	conn.Send(packets.MobShow(m, true))
}

func (m *MapleMob) Show(conn mnet.MConnChannel) {
	conn.Send(packets.MobShow(m, false))
}

func (m *MapleMob) Hide(conn mnet.MConnChannel) {
	conn.Send(packets.MobRemove(m, 0))
}

func (m *MapleMob) CanCastSkills() bool {
	return !(m.HasStatus(consts.MOB_STATUS_FREEZE) || m.HasStatus(consts.MOB_STATUS_STUN) || m.HasStatus(consts.MOB_STATUS_SHADOW_WEB))
}

func (m *MapleMob) HasStatus(status int32) bool {
	return m.status&status > 0
}

func (m *MapleMob) HasImmunity() bool {
	var mask int32 = consts.MOB_STATUS_WEAPON_IMMUNITY | consts.MOB_STATUS_MAGIC_IMMUNITY | consts.MOB_STATUS_WEAPON_DAMAGE_REFLECT | consts.MOB_STATUS_MAGIC_DAMAGE_REFLECT
	return (m.status & mask) != 0
}

func (m *MapleMob) ChooseRandomSkill() {
	if !m.CanCastSkills() || m.nextSkillID != 0 {
		return
	}

	if m.lastSkillUseTime != 0 && (time.Now().Unix()-m.lastSkillUseTime) < 3 {
		return
	}

	availableSkills := nx.GetMobSkills(m.GetID())

	if len(availableSkills) == 0 {
		return
	}

	skillsToChooseFrom := make([]nx.MobSkill, 0)

	for _, skill := range availableSkills {
		var stop bool

		switch skill.SkillID {
		case skills.Mob.WeaponAttackUp:
			fallthrough
		case skills.Mob.WeaponAttackUpAoe:
			stop = m.HasStatus(consts.MOB_STATUS_WATK)
		case skills.Mob.MagicAttackUp:
			fallthrough
		case skills.Mob.MagicAttackUpAoe:
			stop = m.HasStatus(consts.MOB_STATUS_MAGIC_ATTACK_UP)
		case skills.Mob.WeaponDefenceUp:
			fallthrough
		case skills.Mob.WeaponDefenceUpAoe:
			stop = m.HasStatus(consts.MOB_STATUS_WEAPON_DEFENSE_UP)
		case skills.Mob.MagicDefenceUp:
		case skills.Mob.MagicDefenceUpAoe:
			stop = m.HasStatus(consts.MOB_STATUS_MAGIC_DEFENSE_UP)
		case skills.Mob.WeaponImmunity:
		case skills.Mob.MagicImmunity:
		case skills.Mob.WeaponDamageReflect:
		case skills.Mob.MagicDamageReflect:
			stop = m.HasImmunity()
		case skills.Mob.McSpeedUp:
			stop = m.HasStatus(consts.MOB_STATUS_SPEED)
		case skills.Mob.Summon:
			stop = m.nSpawns > 3 // get summon max count from the skillid
		default:
		}

		if stop {
			continue
		}

		for k, v := range m.skillTimes {
			if k == skill.SkillID {
				stop = time.Now().Unix() < int64(int32(v)+skill.Interval)
			}
		}

		if !stop {
			hpPercentage := m.GetHp() * 100 / m.GetMaxHp()
			stop = hpPercentage < skill.HP // is this correct?
		}

		if !stop {
			skillsToChooseFrom = append(skillsToChooseFrom, skill)
		}
	}

	if len(skillsToChooseFrom) != 0 {
		rand.Seed(time.Now().Unix())
		skill := skillsToChooseFrom[rand.Intn(len(skillsToChooseFrom))]

		m.nextSkillID = skill.SkillID
		m.nextSkillLevel = skill.Level
	}
}

func (m *MapleMob) UseSkill() {
	if m.nextSkillID == 0 {
		return
	}

	m.lastSkillUseTime = time.Now().Unix()

	if m.skillTimes == nil {
		m.skillTimes = make(map[byte]int64)
	}

	m.skillTimes[m.nextSkillID] = m.lastSkillUseTime

	m.nextSkillID = 0
	m.nextSkillLevel = 0
}
