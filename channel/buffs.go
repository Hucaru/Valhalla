package channel

import (
	"log"
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/constant/skill"
	"github.com/Hucaru/Valhalla/nx"
)

// BuffValueTypes is a bitmask for buff flags. Matches the C# enum layout.
const (
	// Byte 1
	BuffWeaponAttack  uint32 = 1 << 0
	BuffWeaponDefense uint32 = 1 << 1
	BuffMagicAttack   uint32 = 1 << 2
	BuffMagicDefense  uint32 = 1 << 3
	BuffAccuracy      uint32 = 1 << 4
	BuffAvoidability  uint32 = 1 << 5
	BuffHands         uint32 = 1 << 6
	BuffSpeed         uint32 = 1 << 8

	// Byte 2
	BuffJump       uint32 = 1 << 10
	BuffMagicGuard uint32 = 1 << 9
	BuffDarkSight  uint32 = 1 << 10
	BuffBooster    uint32 = 1 << 11

	BuffPowerGuard uint32 = 1 << 12
	BuffMaxHP      uint32 = 1 << 13
	BuffMaxMP      uint32 = 1 << 14
	BuffInvincible uint32 = 1 << 15

	// Byte 3
	BuffSoulArrow   uint32 = 1 << 16
	BuffStun        uint32 = 1 << 17
	BuffPoison      uint32 = 1 << 18
	BuffSeal        uint32 = 1 << 19
	BuffDarkness    uint32 = 1 << 20
	BuffComboAttack uint32 = 1 << 21
	BuffCharges     uint32 = 1 << 22
	BuffDragonBlood uint32 = 1 << 23

	// Byte 4
	BuffHolySymbol       uint32 = 1 << 24
	BuffMesoUP           uint32 = 1 << 25
	BuffShadowPartner    uint32 = 1 << 26
	BuffPickPocketMesoUP uint32 = 1 << 27

	BuffMesoGuard uint32 = 1 << 28
	BuffThaw      uint32 = 1 << 29
	BuffWeakness  uint32 = 1 << 30
	BuffCurse     uint32 = 1 << 31
)

// skillBuffValues stores per-skill aggregate buff mask.
var skillBuffValues map[int32]uint32

// AddSkillBuff registers one or more buff flags for a skill.
func AddSkillBuff(skillID int32, flags ...uint32) {
	if skillBuffValues == nil {
		skillBuffValues = make(map[int32]uint32)
	}
	var mask uint32
	for _, f := range flags {
		mask |= f
	}
	skillBuffValues[skillID] |= mask
}

// LoadBuffs seeds known skill -> buff mask mappings.
// Fill this from your job constants so buffs resolve correctly at runtime.
func LoadBuffs() {
	skillBuffValues = make(map[int32]uint32)

	// 1st Job
	AddSkillBuff(int32(skill.IronBody), BuffWeaponDefense)
	AddSkillBuff(int32(skill.MagicGuard), BuffMagicGuard)
	AddSkillBuff(int32(skill.MagicArmor), BuffWeaponDefense) // some sources also set BuffMagicDefense
	AddSkillBuff(int32(skill.Focus), BuffAvoidability)       // can add BuffAccuracy as well
	AddSkillBuff(int32(skill.DarkSight), BuffDarkSight)

	// 2nd Job - Warrior branches
	AddSkillBuff(int32(skill.SwordBooster), BuffBooster)
	AddSkillBuff(int32(skill.AxeBooster), BuffBooster)
	AddSkillBuff(int32(skill.PageSwordBooster), BuffBooster)
	AddSkillBuff(int32(skill.BwBooster), BuffBooster)
	AddSkillBuff(int32(skill.SpearBooster), BuffBooster)
	AddSkillBuff(int32(skill.PolearmBooster), BuffBooster)

	AddSkillBuff(int32(skill.Rage), BuffWeaponAttack)
	AddSkillBuff(int32(skill.PowerGuard), BuffPowerGuard)
	AddSkillBuff(int32(skill.PagePowerGuard), BuffPowerGuard)

	AddSkillBuff(int32(skill.IronWill), BuffWeaponDefense, BuffMagicDefense)
	AddSkillBuff(int32(skill.HyperBody), BuffMaxHP, BuffMaxMP)

	// 2nd Job - Magician branches
	AddSkillBuff(int32(skill.Meditation), BuffMagicAttack)
	AddSkillBuff(int32(skill.ILMeditation), BuffMagicAttack)
	AddSkillBuff(int32(skill.Invincible), BuffInvincible) // Cleric passive/active invincible

	// 2nd Job - Archer branches
	AddSkillBuff(int32(skill.BowBooster), BuffBooster)
	AddSkillBuff(int32(skill.CrossbowBooster), BuffBooster)
	AddSkillBuff(int32(skill.SoulArrow), BuffSoulArrow)
	AddSkillBuff(int32(skill.CBSoulArrow), BuffSoulArrow)

	// 2nd Job - Thief branches
	AddSkillBuff(int32(skill.ClawBooster), BuffBooster)
	AddSkillBuff(int32(skill.DaggerBooster), BuffBooster)
	AddSkillBuff(int32(skill.Haste), BuffSpeed, BuffJump)
	AddSkillBuff(int32(skill.BanditHaste), BuffSpeed, BuffJump)

	// 3rd Job - Warrior branches
	AddSkillBuff(int32(skill.ComboAttack), BuffComboAttack)
	AddSkillBuff(int32(skill.DragonBlood), BuffWeaponAttack, BuffDragonBlood)
	AddSkillBuff(int32(skill.DragonRoar), BuffStun)

	// 3rd Job - White Knight charges: magic amp + charge
	AddSkillBuff(int32(skill.BwFireCharge), BuffMagicAttack, BuffCharges)
	AddSkillBuff(int32(skill.BwIceCharge), BuffMagicAttack, BuffCharges)
	AddSkillBuff(int32(skill.BwLitCharge), BuffMagicAttack, BuffCharges)
	AddSkillBuff(int32(skill.SwordFireCharge), BuffMagicAttack, BuffCharges)
	AddSkillBuff(int32(skill.SwordIceCharge), BuffMagicAttack, BuffCharges)
	AddSkillBuff(int32(skill.SwordLitCharge), BuffMagicAttack, BuffCharges)

	// 3rd Job - Chief Bandit
	AddSkillBuff(int32(skill.MesoGuard), BuffMesoGuard)
	AddSkillBuff(int32(skill.Pickpocket), BuffPickPocketMesoUP)

	// 3rd Job - Hermit
	AddSkillBuff(int32(skill.MesoUp), BuffMesoUP)
	AddSkillBuff(int32(skill.ShadowPartner), BuffShadowPartner)

	// 3rd Job - Priest
	AddSkillBuff(int32(skill.HolySymbol), BuffHolySymbol)

	// GM skills
	AddSkillBuff(int32(skill.GMShadowPartner), BuffShadowPartner)
	AddSkillBuff(int32(skill.GMBless), BuffWeaponAttack, BuffWeaponDefense, BuffMagicAttack, BuffMagicDefense, BuffAccuracy, BuffAvoidability)
	AddSkillBuff(int32(skill.GMHaste), BuffSpeed, BuffJump)
	AddSkillBuff(int32(skill.GMHolySymbol), BuffHolySymbol)
	AddSkillBuff(int32(skill.Hide), BuffInvincible)
}

// Initialize the skill->buff map once on package load so it "just works".
func init() {
	LoadBuffs()
}

// CharacterBuffs is a Go adaptation of the provided C# CharacterBuffs class.
type CharacterBuffs struct {
	plr               *player
	comboCount        byte
	activeSkillLevels map[int32]byte // skillID -> level
}

// NewCharacterBuffs creates a new CharacterBuffs holder for a player.
func NewCharacterBuffs(p *player) *CharacterBuffs {
	return &CharacterBuffs{
		plr:               p,
		activeSkillLevels: make(map[int32]byte),
	}
}

func (cb *CharacterBuffs) HasGMHide() bool {
	if cb == nil {
		return false
	}
	_, ok := cb.activeSkillLevels[int32(skill.Hide)]
	return ok
}

// GetActiveSkillLevel returns the cached active level for a skill if set.
func (cb *CharacterBuffs) GetActiveSkillLevel(skillID int32) byte {
	if lvl, ok := cb.activeSkillLevels[skillID]; ok {
		return lvl
	}
	return 0
}

// AddBuff applies a buff for a skill by computing its duration from nx data (Skill.Time)
// and then saving/sending it.
func (cb *CharacterBuffs) AddBuff(skillID int32, level byte, sinc1, sinc2 int, delay int16) {
	if cb == nil || cb.plr == nil {
		return
	}

	// Resolve level if not specified (0xFF => use player's current skill level).
	if level == 0xFF {
		if s, ok := cb.plr.skills[skillID]; ok && s.Level > 0 {
			level = s.Level
		} else {
			// No level, nothing to add.
			return
		}
	}

	// Lookup skill data for duration. nx.GetPlayerSkill returns []SkillLevelData (index level-1).
	skillInfo, err := nx.GetPlayerSkill(skillID)
	if err != nil || int(level) < 1 || int(level) > len(skillInfo) {
		log.Printf("AddBuff: invalid skill or level for skillID=%d level=%d: %v", skillID, level, err)
		return
	}
	durationSec := skillInfo[level-1].Time

	expiresAtMs := int64(0)
	if durationSec > 0 {
		expiresAtMs = time.Now().Add(time.Duration(durationSec) * time.Second).UnixMilli()
	}
	cb.AddBuffFromCC(skillID, expiresAtMs, level, sinc1, sinc2, delay)
}

// AddBuffFromCC applies a buff coming from a cross-channel or persisted source where
// the expiration time is already known (ms since epoch).
func (cb *CharacterBuffs) AddBuffFromCC(skillID int32, expiresAtMs int64, level byte, sinc1, sinc2 int, delay int16) {
	if cb == nil || cb.plr == nil {
		return
	}
	if skillID == 0 || level == 0 {
		return
	}

	cb.check(skillID)

	mask := getBuffMask(skillID)
	values := cb.buildBuffValues(skillID, level, mask, expiresAtMs)

	// Nothing meaningful to apply
	if mask == 0 || len(values) == 0 {
		return
	}

	log.Printf("AddBuffFromCC: skillID=%d level=%d mask=%d expiresAtMs=%d", skillID, level, mask, expiresAtMs)
	cb.plr.send(packetPlayerGiveBuff(skillID))
	cb.saveBuff(cb.plr.id, skillID, expiresAtMs, mask, level, sinc1, sinc2)
	cb.activeSkillLevels[skillID] = level
}

// For now, it's a no-op, but you can extend this to check for conflicts.
func (cb *CharacterBuffs) check(skillID int32) {
	// Example: If skillID == 1101006 and buff 1001003 active, remove it (Rage vs Iron Body).
}

// saveBuff persists/updates a buff record for this character.
func (cb *CharacterBuffs) saveBuff(charID int32, buffID int32, expiresAtMs int64, flags uint32, level byte, sinc, sinc2 int) {
	const insert = `
INSERT INTO character_buffs (bid, cid, time, flags, level, sinc, sinc2)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE time=VALUES(time), flags=VALUES(flags), level=VALUES(level), sinc=VALUES(sinc), sinc2=VALUES(sinc2)
`
	_, err := common.DB.Exec(insert,
		buffID,
		charID,
		expiresAtMs,
		flags,
		int(level),
		sinc,
		sinc2,
	)
	if err != nil {
		log.Printf("saveBuff: failed for cid=%d bid=%d: %v", charID, buffID, err)
	}
}

func (cb *CharacterBuffs) LoadBuffs() {
	if cb == nil || cb.plr == nil {
		return
	}

	type buffRow struct {
		BID   int32
		Time  int64
		Level int
	}

	rows, err := common.DB.Query(`SELECT bid, time, level FROM character_buffs WHERE cid=?`, cb.plr.id)
	if err != nil {
		log.Printf("LoadBuffs: query error for cid=%d: %v", cb.plr.id, err)
		return
	}
	defer rows.Close()

	now := time.Now().UnixMilli()
	toReapply := make([]buffRow, 0, 8)

	for rows.Next() {
		var r buffRow
		if err := rows.Scan(&r.BID, &r.Time, &r.Level); err != nil {
			log.Printf("LoadBuffs: scan error for cid=%d: %v", cb.plr.id, err)
			continue
		}
		// Skip expired
		if r.Time > 0 && r.Time <= now {
			continue
		}
		// Skip empty/invalid entries
		if r.BID == 0 || r.Level <= 0 {
			continue
		}
		toReapply = append(toReapply, r)
	}

	if err := rows.Err(); err != nil {
		log.Printf("LoadBuffs: rows error for cid=%d: %v", cb.plr.id, err)
	}

	for _, b := range toReapply {
		cb.AddBuffFromCC(b.BID, b.Time, byte(b.Level), 0, 0, 0)
	}
}

// RemoveExpiredBuffs deletes expired buffs for a player. Optional helper.
func (cb *CharacterBuffs) RemoveExpiredBuffs() {
	if cb == nil || cb.plr == nil {
		return
	}
	_, err := common.DB.Exec(`DELETE FROM character_buffs WHERE cid=? AND time>0 AND time<=?`, cb.plr.id, time.Now().UnixMilli())
	if err != nil {
		log.Printf("RemoveExpiredBuffs: cleanup failed for cid=%d: %v", cb.plr.id, err)
	}
}

// ClearBuff removes a specific buff from player (packet and DB).
// Flags param is optional; if unknown, pass 0 and only DB will be affected.
func (cb *CharacterBuffs) ClearBuff(skillID int32, flags uint32) {
	if cb == nil || cb.plr == nil {
		return
	}
	if flags == 0 {
		flags = getBuffMask(skillID)
	}
	if flags != 0 && cb.plr.inst != nil {
		_ = cb.plr.inst.send(packetPlayerResetForeignBuff(cb.plr.id, flags))
	}
	delete(cb.activeSkillLevels, skillID)

	_, err := common.DB.Exec(`DELETE FROM character_buffs WHERE cid=? AND bid=?`, cb.plr.id, skillID)
	if err != nil {
		log.Printf("ClearBuff: delete failed for cid=%d bid=%d: %v", cb.plr.id, skillID, err)
	}
}

// valueType mirrors the reference getValue selector (SkillX, SkillY, SkillSpeed, ...).
type valueType byte

const (
	valX valueType = iota
	valY
	valSpeed
	valJump
	valWatk
	valWdef
	valMatk
	valMdef
	valAcc
	valAvo
	valProp
	valLv
)

// getSkillValue returns the short value for a skill at a given level for the requested field.
// This mirrors the reference implementation you provided.
func getSkillValue(skillID int32, level byte, sel valueType) int16 {
	skillLevels, err := nx.GetPlayerSkill(skillID)
	if err != nil || level == 0 || int(level) > len(skillLevels) {
		return 0
	}
	sl := skillLevels[level-1]

	switch sel {
	case valX:
		return int16(sl.X)
	case valY:
		return int16(sl.Y)
	case valSpeed:
		return int16(sl.Speed)
	case valJump:
		return int16(sl.Jump)
	case valWatk:
		return int16(sl.Pad)
	case valWdef:
		return int16(sl.Pdd)
	case valMatk:
		return int16(sl.Mad)
	case valMdef:
		return int16(sl.Mdd)
	case valAcc:
		return int16(sl.Acc)
	case valAvo:
		return int16(sl.Eva)
	case valProp:
		return int16(sl.Prop)
	case valLv:
		return int16(level)
	default:
		return 0
	}
}

// buildBuffValues builds the serialized per-effect values (short,int,int) for SendChannelTempStatChange.
func (cb *CharacterBuffs) buildBuffValues(skillID int32, level byte, mask uint32, expiresAtMs int64) []byte {
	values := make([]byte, 0, 32)

	// Compute remaining duration in milliseconds (fits int32)
	remain := int32(0)
	if skillLevels, err := nx.GetPlayerSkill(skillID); err == nil && level > 0 && int(level) <= len(skillLevels) {
		if expiresAtMs > 0 {
			if dur := expiresAtMs - time.Now().UnixMilli(); dur > 0 {
				remain = int32(dur)
			}
		} else if sec := skillLevels[level-1].Time; sec > 0 {
			remain = int32(sec * 1000)
		}
	}

	appendTriple := func(val int16) {
		// short value
		values = append(values, byte(val), byte(val>>8))
		// int32 source (skill id)
		id := skillID
		values = append(values, byte(id), byte(id>>8), byte(id>>16), byte(id>>24))
		// int32 duration ms
		d := remain
		values = append(values, byte(d), byte(d>>8), byte(d>>16), byte(d>>24))
	}

	// Minimal generic mapping (extend as needed for other buffs)
	type m struct {
		flag uint32
		sel  valueType
	}
	maps := []m{
		{BuffSpeed, valSpeed},
		{BuffJump, valJump},
		{BuffWeaponAttack, valWatk},
		{BuffWeaponDefense, valWdef},
		{BuffMagicAttack, valMatk},
		{BuffMagicDefense, valMdef},
		{BuffAccuracy, valAcc},
		{BuffAvoidability, valAvo},
	}

	for _, e := range maps {
		if mask&e.flag != 0 {
			appendTriple(getSkillValue(skillID, level, e.sel))
		}
	}

	return values
}

// getBuffMask returns the bitmask for the provided skill ID.
// Uses the loaded skillBuffValues map; falls back to zero if unknown.
func getBuffMask(skillID int32) uint32 {
	if skillBuffValues != nil {
		if m, ok := skillBuffValues[skillID]; ok {
			return m
		}
	}
	return 0
}
