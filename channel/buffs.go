package channel

import (
	"log"
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/nx"
)

// BuffValueTypes is a bitmask for buff flags.
const (
	BuffWeaponAttack     uint32 = 1 << 0
	BuffWeaponDefense    uint32 = 1 << 1
	BuffMagicAttack      uint32 = 1 << 2
	BuffMagicDefense     uint32 = 1 << 3
	BuffAccuracy         uint32 = 1 << 4
	BuffAvoidability     uint32 = 1 << 5
	BuffSpeed            uint32 = 1 << 6
	BuffJump             uint32 = 1 << 7
	BuffMagicGuard       uint32 = 1 << 8
	BuffDarkSight        uint32 = 1 << 9
	BuffBooster          uint32 = 1 << 10
	BuffPowerGuard       uint32 = 1 << 11
	BuffMaxHP            uint32 = 1 << 12
	BuffMaxMP            uint32 = 1 << 13
	BuffInvincible       uint32 = 1 << 14
	BuffSoulArrow        uint32 = 1 << 15
	BuffComboAttack      uint32 = 1 << 16
	BuffCharges          uint32 = 1 << 17
	BuffDragonBlood      uint32 = 1 << 18
	BuffMesoUP           uint32 = 1 << 19
	BuffShadowPartner    uint32 = 1 << 20
	BuffPickPocketMesoUP uint32 = 1 << 21
	BuffMesoGuard        uint32 = 1 << 22
)

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
	return false
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
	if durationSec <= 0 {
		// Some buffs can be 0 => treat as immediate/visual if needed. Here we bail.
		// Adjust if you want to allow "infinite" buffs.
	}

	expiresAtMs := time.Now().Add(time.Duration(durationSec) * time.Second).UnixMilli()
	cb.AddBuffFromCC(skillID, expiresAtMs, level, sinc1, sinc2, delay)
}

// AddBuffFromCC applies a buff coming from a cross-channel or persisted source where
// the expiration time is already known.
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

	// Log as hex (not string) to avoid mojibake
	log.Printf("AddBuffFromCC: skillID=%d mask=%08x values_len=%d values_hex=% x",
		skillID, mask, len(values), values)

	// Only send if player is in an instance
	if cb.plr.inst != nil {
		err := cb.plr.inst.send(packetPlayerSetTempStats(mask, values, delay))
		if err != nil {
			log.Printf("AddBuffFromCC: failed to send packet for cid=%d bid=%d: %v", cb.plr.id, skillID, err)
			return
		}
	}

	cb.saveBuff(cb.plr.id, skillID, expiresAtMs, mask, level, sinc1, sinc2)
	cb.activeSkillLevels[skillID] = level
}

// For now, it's a no-op, but you can extend this to check for conflicts.
func (cb *CharacterBuffs) check(skillID int32) {
	// Example: If skillID == 1101006 and buff 1001003 active, remove it.
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
	if flags != 0 {
		err := cb.plr.inst.send(packetPlayerResetForeignBuff(cb.plr.id, flags))
		if err != nil {
			return
		}
	}
	delete(cb.activeSkillLevels, skillID)

	_, err := common.DB.Exec(`DELETE FROM character_buffs WHERE cid=? AND bid=?`, cb.plr.id, skillID)
	if err != nil {
		log.Printf("ClearBuff: delete failed for cid=%d bid=%d: %v", cb.plr.id, skillID, err)
	}
}

// buildBuffValues builds the serialized per-effect values (value:int16, id:int32, duration:int32)
// in the order of the active bits in mask (LSB to MSB within our 32-bit mask).
// It uses nx data to pick the appropriate value per effect.
func (cb *CharacterBuffs) buildBuffValues(skillID int32, level byte, mask uint32, expiresAtMs int64) []byte {
	values := make([]byte, 0, 3*4+2) // rough capacity
	// Pull nx skill level data once
	skillInfo, err := nx.GetPlayerSkill(skillID)
	if err != nil || int(level) < 1 || int(level) > len(skillInfo) {
		return values
	}
	sl := skillInfo[level-1]

	nowMs := time.Now().UnixMilli()
	remain := int32(0)
	if expiresAtMs > 0 {
		if dur := expiresAtMs - nowMs; dur > 0 {
			remain = int32(dur)
		}
	} else if sl.Time > 0 {
		remain = int32(sl.Time * 1000)
	}

	// Helper to append one effect
	appendEffect := func(val int16) {
		// value (int16)
		values = append(values, byte(val), byte(val>>8))
		// source id (int32, positive for skill; negative when you implement item buffs)
		id := skillID
		values = append(values, byte(id), byte(id>>8), byte(id>>16), byte(id>>24))
		// duration ms (int32)
		d := remain
		values = append(values, byte(d), byte(d>>8), byte(d>>16), byte(d>>24))
	}

	// Iterate flags in stable order and push values for each set bit.
	type entry struct {
		flag uint32
		get  func() int16
	}
	entries := []entry{
		{BuffWeaponAttack, func() int16 { return int16(sl.X) }},
		{BuffWeaponDefense, func() int16 { return int16(sl.X) }},
		{BuffMagicAttack, func() int16 { return int16(sl.X) }},
		{BuffMagicDefense, func() int16 { return int16(sl.X) }},
		{BuffAccuracy, func() int16 { return int16(sl.X) }},
		{BuffAvoidability, func() int16 { return int16(sl.X) }},
		{BuffSpeed, func() int16 { return int16(sl.X) }},
		{BuffJump, func() int16 { return int16(sl.Y) }}, // common Haste pattern: X=speed, Y=jump
		{BuffMagicGuard, func() int16 { return int16(sl.X) }},
		{BuffDarkSight, func() int16 { return 1 }},
		{BuffBooster, func() int16 { return int16(sl.X) }},    // attack speed stages
		{BuffPowerGuard, func() int16 { return int16(sl.X) }}, // percent reflect
		{BuffMaxHP, func() int16 { return int16(sl.X) }},      // percent increase
		{BuffMaxMP, func() int16 { return int16(sl.X) }},      // percent increase
		{BuffInvincible, func() int16 { return 1 }},
		{BuffSoulArrow, func() int16 { return 1 }},
		{BuffComboAttack, func() int16 { return int16(sl.X) }},
		{BuffCharges, func() int16 { return int16(sl.X) }},
		{BuffDragonBlood, func() int16 { return int16(sl.X) }},
		{BuffMesoUP, func() int16 { return int16(sl.X) }},
		{BuffShadowPartner, func() int16 { return 1 }},
		{BuffPickPocketMesoUP, func() int16 { return int16(sl.X) }},
		{BuffMesoGuard, func() int16 { return int16(sl.X) }},
	}

	for _, e := range entries {
		if mask&e.flag != 0 {
			appendEffect(e.get())
		}
	}

	return values
}

// getBuffMask maps a skill ID to a buff mask. This is a crucial place to encode
// which flags a given skill applies. You should extend this with your skill set.
// For now, a minimal example mapping is provided; fill in as needed.
func getBuffMask(skillID int32) uint32 {
	switch skillID {
	case 4101004, 8001001:
		return BuffSpeed | BuffJump
	case 1101004, 1201004, 1301004, 3101002, 3201002, 4101003, 4201002:
		return BuffBooster
	case 3101004, 3201004:
		return BuffSoulArrow
	case 1201007:
		return BuffPowerGuard
	case 1101007:
		return BuffMaxHP | BuffMaxMP
	case 1101006:
		return BuffWeaponAttack
	default:
		return 0
	}
}
