package channel

import (
	"log"
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/constant/skill"
	"github.com/Hucaru/Valhalla/nx"
)

// BuffValueTypes now represent bit positions, not bitmasks.
const (
	// Byte 1 (bits 0..7)
	BuffWeaponAttack  = 0
	BuffWeaponDefense = 1
	BuffMagicAttack   = 2
	BuffMagicDefense  = 3

	BuffAccuracy     = 4
	BuffAvoidability = 5
	BuffHands        = 6
	BuffSpeed        = 7

	// Byte 2 (bits 8..15)
	BuffJump       = 8
	BuffMagicGuard = 9
	BuffDarkSight  = 10
	BuffBooster    = 11
	BuffPowerGuard = 12
	BuffMaxHP      = 13
	BuffMaxMP      = 14
	BuffInvincible = 15

	// Byte 3 (bits 16..23)
	BuffSoulArrow   = 16
	BuffStun        = 17
	BuffPoison      = 18
	BuffSeal        = 19
	BuffDarkness    = 20
	BuffComboAttack = 21
	BuffCharges     = 22
	BuffDragonBlood = 23

	// Byte 4 (bits 24..31)
	BuffHolySymbol       = 24
	BuffMesoUP           = 25
	BuffShadowPartner    = 26
	BuffPickPocketMesoUP = 27
	BuffMesoGuard        = 28
	BuffThaw             = 29
	BuffWeakness         = 30
	BuffCurse            = 31
)

// skillBuffBits stores per-skill bit positions in the CFlag.
var skillBuffBits map[int32][]int

// AddSkillBuff registers one or more Flag bit positions for a skill.
func AddSkillBuff(skillID int32, bits ...int) {
	if skillBuffBits == nil {
		skillBuffBits = make(map[int32][]int)
	}
	skillBuffBits[skillID] = append(skillBuffBits[skillID], bits...)
}

// LoadBuffs seeds known skill -> buff bit mappings.
func LoadBuffs() {
	skillBuffBits = make(map[int32][]int)

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
	AddSkillBuff(int32(skill.Invincible), BuffInvincible)

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

	// 3rd Job - White Knight charges
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

func init() {
	LoadBuffs()
}

// CharacterBuffs is a Go adaptation of the provided C# CharacterBuffs class.
type CharacterBuffs struct {
	plr               *player
	comboCount        byte
	activeSkillLevels map[int32]byte // skillID -> level

	// timers per skill for expiry
	expireTimers map[int32]*time.Timer
}

func NewCharacterBuffs(p *player) *CharacterBuffs {
	return &CharacterBuffs{
		plr:               p,
		activeSkillLevels: make(map[int32]byte),
		expireTimers:      make(map[int32]*time.Timer),
	}
}

func (cb *CharacterBuffs) HasGMHide() bool {
	if cb == nil {
		return false
	}
	_, ok := cb.activeSkillLevels[int32(skill.Hide)]
	return ok
}

func (cb *CharacterBuffs) GetActiveSkillLevel(skillID int32) byte {
	if lvl, ok := cb.activeSkillLevels[skillID]; ok {
		return lvl
	}
	return 0
}

func (cb *CharacterBuffs) AddBuff(skillID int32, level byte, sinc1, sinc2 int, delay int16) {
	if cb == nil || cb.plr == nil {
		return
	}

	if level == 0xFF {
		if s, ok := cb.plr.skills[skillID]; ok && s.Level > 0 {
			level = s.Level
		} else {
			return
		}
	}

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
	log.Printf("AddBuff: skillID=%d level=%d durationSec= %d expiresAtMs=%d", skillID, level, durationSec, expiresAtMs)
	cb.AddBuffFromCC(skillID, expiresAtMs, level, sinc1, sinc2, delay)
}

func buildMaskBytes64(bits []int) []byte {
	m := make([]byte, 8)
	for _, b := range bits {
		if b < 0 || b >= 64 {
			continue
		}
		byteIdx := b / 8
		bitOff := uint(b % 8)
		m[byteIdx] |= (1 << bitOff)
	}
	return m
}

// Derive triples by scanning maskBytes in the exact wire order we use:
// for byte = 0..7, for bit = 0..7 (LSB-first). Append a triple for each set bit.
func (cb *CharacterBuffs) buildBuffTriplesWireOrder(skillID int32, level byte, maskBytes []byte, expiresAtMs int64) ([]byte, int16) {
	levels, err := nx.GetPlayerSkill(skillID)
	if err != nil || level == 0 || int(level) > len(levels) {
		return nil, 0
	}
	sl := levels[level-1]

	// Compute remaining duration (seconds, int16)
	var remainSec int16
	if expiresAtMs > 0 {
		now := time.Now().UnixMilli()
		if d := expiresAtMs - now; d > 0 {
			sec := (d + 500) / 1000
			if sec > 32767 {
				sec = 32767
			}
			remainSec = int16(sec)
		}
	} else if sl.Time > 0 {
		if sl.Time > 32767 {
			remainSec = 32767
		} else {
			remainSec = int16(sl.Time)
		}
	}

	// Only concrete fields; toggles/percent-like flags -> 1
	valueForBit := func(bitIndex int) int16 {
		switch bitIndex {
		case BuffSpeed:
			if sl.Speed != 0 {
				return int16(sl.Speed)
			}
			return 1
		case BuffJump:
			if sl.Jump != 0 {
				return int16(sl.Jump)
			}
			return 1
		case BuffWeaponAttack:
			if sl.Pad != 0 {
				return int16(sl.Pad)
			}
			return 1
		case BuffWeaponDefense:
			if sl.Pdd != 0 {
				return int16(sl.Pdd)
			}
			return 1
		case BuffMagicAttack:
			if sl.Mad != 0 {
				return int16(sl.Mad)
			}
			return 1
		case BuffMagicDefense:
			if sl.Mdd != 0 {
				return int16(sl.Mdd)
			}
			return 1
		case BuffAccuracy:
			if sl.Acc != 0 {
				return int16(sl.Acc)
			}
			return 1
		case BuffAvoidability:
			if sl.Eva != 0 {
				return int16(sl.Eva)
			}
			return 1

		// Flags that don’t have a concrete numeric field in this layout
		case BuffMagicGuard, BuffBooster, BuffPowerGuard, BuffMaxHP, BuffMaxMP,
			BuffHolySymbol, BuffMesoUP, BuffPickPocketMesoUP, BuffMesoGuard,
			BuffDarkSight, BuffSoulArrow, BuffInvincible, BuffShadowPartner,
			BuffThaw, BuffWeakness, BuffCurse, BuffComboAttack, BuffCharges:
			return 1

		case BuffDragonBlood:
			if sl.Pad != 0 {
				return int16(sl.Pad)
			}
			return 1

		default:
			return 1
		}
	}

	values := make([]byte, 0, 64)
	appendTriple := func(val int16) {
		// short value
		values = append(values, byte(val), byte(val>>8))
		// int32 reason/source skill
		id := skillID
		values = append(values, byte(id), byte(id>>8), byte(id>>16), byte(id>>24))
		// short time (seconds)
		t := remainSec
		values = append(values, byte(t), byte(t>>8))
	}

	hasAny := false
	// Scan bytes 0..7, bit 0..7 (LSB-first within each byte)
	for byteIdx := 0; byteIdx < len(maskBytes) && byteIdx < 8; byteIdx++ {
		b := maskBytes[byteIdx]
		if b == 0 {
			continue
		}
		for bit := 0; bit < 8; bit++ {
			if (b & (1 << uint(bit))) != 0 {
				globalBit := byteIdx*8 + bit
				appendTriple(valueForBit(globalBit))
				hasAny = true
			}
		}
	}
	if !hasAny {
		return nil, 0
	}
	return values, remainSec
}

func (cb *CharacterBuffs) AddBuffFromCC(skillID int32, expiresAtMs int64, level byte, sinc1, sinc2 int, delay int16) {
	if cb == nil || cb.plr == nil {
		return
	}
	if skillID == 0 || level == 0 {
		return
	}

	cb.check(skillID)

	// Get the configured flag bits for this skill
	bits, ok := skillBuffBits[skillID]
	if !ok || len(bits) == 0 {
		return
	}

	// 64-bit mask bytes (wire order)
	maskBytes := buildMaskBytes64(bits)

	// Build value triples in the same wire-scan order
	values, remainSec := cb.buildBuffTriplesWireOrder(skillID, level, maskBytes, expiresAtMs)
	if len(values) == 0 {
		return
	}

	// Optional trailing byte (Combo/Charges)
	extra := byte(0)
	for _, b := range bits {
		if b == BuffComboAttack {
			extra = cb.comboCount
			break
		}
		if b == BuffCharges {
			extra = 1
		}
	}

	// Send to self and others
	cb.plr.send(packetPlayerGiveBuff(maskBytes, values, delay, extra))
	if cb.plr.inst != nil {
		cb.plr.inst.send(packetPlayerGiveForeignBuff(cb.plr.id, maskBytes, values, extra))
	}

	// Track active
	cb.activeSkillLevels[skillID] = level

	// Expiry
	if remainSec > 0 {
		cb.scheduleExpiry(skillID, time.Duration(remainSec)*time.Second)
	} else if expiresAtMs > 0 {
		cb.scheduleExpiry(skillID, 0)
	}
}

// Concrete-only values builder across the 64-bit mask.
// Emits triples for each set bit; values: Pad/Pdd/Mad/Mdd/Acc/Eva/Speed/Jump; toggles -> 1.
func (cb *CharacterBuffs) buildBuffValuesForMaskConcrete(skillID int32, level byte, mask *Flag, expiresAtMs int64) ([]byte, int16) {
	levels, err := nx.GetPlayerSkill(skillID)
	if err != nil || level == 0 || int(level) > len(levels) {
		return nil, 0
	}
	sl := levels[level-1]

	// Compute remaining duration in seconds (short)
	var remainSec int16
	if expiresAtMs > 0 {
		now := time.Now().UnixMilli()
		if d := expiresAtMs - now; d > 0 {
			sec := (d + 500) / 1000
			if sec > 32767 {
				sec = 32767
			}
			remainSec = int16(sec)
		}
	} else {
		if sl.Time > 32767 {
			remainSec = 32767
		} else {
			remainSec = int16(sl.Time)
		}
	}

	getVal := func(bit int) int16 {
		switch bit {
		case BuffSpeed:
			if sl.Speed != 0 {
				return int16(sl.Speed)
			}
			return 1
		case BuffJump:
			if sl.Jump != 0 {
				return int16(sl.Jump)
			}
			return 1
		case BuffWeaponAttack:
			if sl.Pad != 0 {
				return int16(sl.Pad)
			}
			return 1
		case BuffWeaponDefense:
			if sl.Pdd != 0 {
				return int16(sl.Pdd)
			}
			return 1
		case BuffMagicAttack:
			if sl.Mad != 0 {
				return int16(sl.Mad)
			}
			return 1
		case BuffMagicDefense:
			if sl.Mdd != 0 {
				return int16(sl.Mdd)
			}
			return 1
		case BuffAccuracy:
			if sl.Acc != 0 {
				return int16(sl.Acc)
			}
			return 1
		case BuffAvoidability:
			if sl.Eva != 0 {
				return int16(sl.Eva)
			}
			return 1

		// Toggles/percent-types (no concrete field used in this layout)
		case BuffMagicGuard, BuffBooster, BuffPowerGuard, BuffMaxHP, BuffMaxMP,
			BuffHolySymbol, BuffMesoUP, BuffPickPocketMesoUP, BuffMesoGuard,
			BuffDarkSight, BuffSoulArrow, BuffInvincible, BuffShadowPartner,
			BuffThaw, BuffWeakness, BuffCurse, BuffComboAttack, BuffCharges:
			return 1

		case BuffDragonBlood:
			if sl.Pad != 0 {
				return int16(sl.Pad)
			}
			return 1

		default:
			return 1
		}
	}

	values := make([]byte, 0, 64)
	appendTriple := func(val int16) {
		values = append(values, byte(val), byte(val>>8))
		id := skillID
		values = append(values, byte(id), byte(id>>8), byte(id>>16), byte(id>>24))
		t := remainSec
		values = append(values, byte(t), byte(t>>8))
	}

	nBits := len(mask.Data()) * 32 // 64 when default flag length
	hasAny := false
	for bit := 0; bit < nBits; bit++ {
		if mask.GetBitNumber(bit) == 1 {
			appendTriple(getVal(bit))
			hasAny = true
		}
	}
	if !hasAny {
		return nil, 0
	}
	return values, remainSec
}

// Timers

func (cb *CharacterBuffs) scheduleExpiry(skillID int32, after time.Duration) {
	if cb.expireTimers == nil {
		cb.expireTimers = make(map[int32]*time.Timer)
	}
	// Cancel previous
	if t, ok := cb.expireTimers[skillID]; ok && t != nil {
		t.Stop()
		delete(cb.expireTimers, skillID)
	}
	if after <= 0 {
		go cb.expireBuffNow(skillID)
		return
	}
	cb.expireTimers[skillID] = time.AfterFunc(after, func() {
		cb.expireBuffNow(skillID)
	})
}

func (cb *CharacterBuffs) expireBuffNow(skillID int32) {
	if cb == nil || cb.plr == nil {
		return
	}
	// Clear timer handle
	if t, ok := cb.expireTimers[skillID]; ok && t != nil {
		delete(cb.expireTimers, skillID)
	}

	// Build cancel mask (64-bit)
	mask := buildBuffMask(skillID)
	if mask == nil || mask.IsZero() {
		delete(cb.activeSkillLevels, skillID)
		return
	}
	maskBytes := mask.ToByteArray(false)

	// Self cancel
	cb.plr.send(packetPlayerCancelBuff(maskBytes))
	// Others cancel
	if cb.plr.inst != nil {
		cb.plr.inst.send(packetPlayerCancelForeignBuff(cb.plr.id, maskBytes))
	}

	delete(cb.activeSkillLevels, skillID)
}

func (cb *CharacterBuffs) check(skillID int32) {
	// Implement conflicting buff cleanup if needed.
}

func (cb *CharacterBuffs) RemoveExpiredBuffs() {
	if cb == nil || cb.plr == nil {
		return
	}
	_, err := common.DB.Exec(`DELETE FROM character_buffs WHERE cid=? AND time>0 AND time<=?`, cb.plr.id, time.Now().UnixMilli())
	if err != nil {
		log.Printf("RemoveExpiredBuffs: cleanup failed for cid=%d: %v", cb.plr.id, err)
	}
}

// ClearBuff removes a specific buff from player and DB.
func (cb *CharacterBuffs) ClearBuff(skillID int32, _ uint32) {
	if cb == nil || cb.plr == nil {
		return
	}
	mask := buildBuffMask(skillID)
	if mask != nil && !mask.IsZero() && cb.plr.inst != nil {
		cb.plr.inst.send(packetPlayerCancelForeignBuff(cb.plr.id, mask.ToByteArray(false)))
	}
	delete(cb.activeSkillLevels, skillID)

	_, err := common.DB.Exec(`DELETE FROM character_buffs WHERE cid=? AND bid=?`, cb.plr.id, skillID)
	if err != nil {
		log.Printf("ClearBuff: delete failed for cid=%d bid=%d: %v", cb.plr.id, skillID, err)
	}
}

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

func getSkillValue(skillID int32, level byte, sel valueType) int16 {
	skillLevels, err := nx.GetPlayerSkill(skillID)
	if err != nil || level == 0 || int(level) > len(skillLevels) {
		return 0
	}
	// Use 0-based index for the selected level
	sl := skillLevels[level-1]

	switch sel {
	case valX:
		return int16(sl.X)
	case valY:
		return int16(sl.Y)
	case valSpeed:
		log.Printf("getSkillValue: speed=%d", sl.Speed)
		return int16(sl.Speed)
	case valJump:
		log.Printf("getSkillValue: jump=%d", sl.Jump)
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

func collectBuffEntries(skillID int32, level byte) (entries map[int]int16, durationSec int16) {
	entries = make(map[int]int16)

	levels, err := nx.GetPlayerSkill(skillID)
	if err != nil || level == 0 || int(level) > len(levels) {
		return entries, 0
	}
	sl := levels[level-1]

	// Duration: prefer NX Time at level
	if sl.Time > 0 {
		if sl.Time > 32767 {
			durationSec = 32767
		} else {
			durationSec = int16(sl.Time)
		}
	}

	// Numeric stats -> bits. Only set if non-zero.
	if sl.Speed != 0 {
		entries[BuffSpeed] = int16(sl.Speed)
	}
	if sl.Jump != 0 {
		entries[BuffJump] = int16(sl.Jump)
	}
	if sl.Pad != 0 {
		entries[BuffWeaponAttack] = int16(sl.Pad)
	}
	if sl.Pdd != 0 {
		entries[BuffWeaponDefense] = int16(sl.Pdd)
	}
	if sl.Mad != 0 {
		entries[BuffMagicAttack] = int16(sl.Mad)
	}
	if sl.Mdd != 0 {
		entries[BuffMagicDefense] = int16(sl.Mdd)
	}
	if sl.Acc != 0 {
		entries[BuffAccuracy] = int16(sl.Acc)
	}
	if sl.Eva != 0 {
		entries[BuffAvoidability] = int16(sl.Eva)
	}

	// Optional: Some skills encode “percent” effects in X/Y. Add the mapping you need:
	// Example candidates (uncomment as you confirm for your client build):
	// if sl.X != 0 && isMagicGuard(skillID) { entries[BuffMagicGuard] = int16(sl.X) }
	// if sl.X != 0 && isBooster(skillID)    { entries[BuffBooster]    = int16(sl.X) }
	// if isDarkSight(skillID)                { entries[BuffDarkSight]  = 1 }
	// if isSoulArrow(skillID)                { entries[BuffSoulArrow]  = 1 }

	return entries, durationSec
}

func buildBuffMaskFromNX(skillID int32, level byte) *Flag {
	entries, _ := collectBuffEntries(skillID, level)
	if len(entries) == 0 {
		return nil
	}
	mask := NewFlag()
	for bit := range entries {
		mask.SetBitNumber(bit, 1)
	}
	return mask
}

// buildBuffValues builds the value triples using NX-derived entries.
// It emits triples in canonical bit order and uses either expiresAtMs or NX Time.
func (cb *CharacterBuffs) buildBuffValues(skillID int32, level byte, mask *Flag, expiresAtMs int64) []byte {
	values := make([]byte, 0, 64)

	// Derive which bits have values from NX
	entries, nxDuration := collectBuffEntries(skillID, level)
	if len(entries) == 0 {
		return nil
	}

	// Compute remaining duration in seconds (short). Prefer expiresAtMs if present.
	remainSec := nxDuration
	if expiresAtMs > 0 {
		now := time.Now().UnixMilli()
		if dur := expiresAtMs - now; dur > 0 {
			sec := (dur + 500) / 1000
			if sec > 32767 {
				sec = 32767
			}
			remainSec = int16(sec)
		} else {
			remainSec = 0
		}
	}

	appendTriple := func(val int16) {
		// short value
		values = append(values, byte(val), byte(val>>8))
		// int32 reason/source skill
		id := skillID
		values = append(values, byte(id), byte(id>>8), byte(id>>16), byte(id>>24))
		// short time (seconds)
		t := remainSec
		values = append(values, byte(t), byte(t>>8))
	}

	// Emit triples in the same canonical order you use for mask bits.
	nBits := len(mask.Data()) * 32
	for bit := 0; bit < nBits; bit++ {
		if mask.GetBitNumber(bit) == 1 {
			if val, ok := entries[bit]; ok {
				appendTriple(val)
			} else {
				// If a bit was set but we didn't derive a numeric value,
				// either skip or append a placeholder (1). Start with skip.
			}
		}
	}

	return values
}

func (cb *CharacterBuffs) buildBuffValuesForMask(skillID int32, level byte, mask *Flag, expiresAtMs int64) ([]byte, int16) {
	// Pull NX data for the skill/level
	levels, err := nx.GetPlayerSkill(skillID)
	if err != nil || level == 0 || int(level) > len(levels) {
		return nil, 0
	}
	sl := levels[level-1]

	// Duration preference: remaining time from expiresAtMs, else NX time (seconds, short)
	var remainSec int16
	if expiresAtMs > 0 {
		now := time.Now().UnixMilli()
		durMs := expiresAtMs - now
		if durMs < 0 {
			remainSec = 0
		} else {
			sec := (durMs + 500) / 1000
			if sec > 32767 {
				sec = 32767
			}
			remainSec = int16(sec)
		}
	} else {
		if sl.Time > 32767 {
			remainSec = 32767
		} else {
			remainSec = int16(sl.Time)
		}
	}

	// Helper to choose a value for each bit.
	// For boolean-style buffs, return 1 so the server-side decode registers them.
	getValueForBit := func(bit int) int16 {
		switch bit {
		// Numeric direct stats
		case BuffSpeed:
			return int16(sl.Speed)
		case BuffJump:
			return int16(sl.Jump)
		case BuffWeaponAttack:
			return int16(sl.Pad)
		case BuffWeaponDefense:
			return int16(sl.Pdd)
		case BuffMagicAttack:
			return int16(sl.Mad)
		case BuffMagicDefense:
			return int16(sl.Mdd)
		case BuffAccuracy:
			return int16(sl.Acc)
		case BuffAvoidability:
			return int16(sl.Eva)

		// Percent/ratio stats typically in X (or Y in rare cases)
		case BuffMagicGuard:
			return int16(sl.X)
		case BuffBooster:
			// Booster booster-speed is commonly X; if 0 in your NX, consider using 1
			if sl.X != 0 {
				return int16(sl.X)
			}
			return 1
		case BuffPowerGuard:
			return int16(sl.X)
		case BuffMaxHP:
			return int16(sl.X)
		case BuffMaxMP:
			return int16(sl.X)
		case BuffHolySymbol:
			return int16(sl.X)
		case BuffMesoUP:
			return int16(sl.X)
		case BuffPickPocketMesoUP:
			return int16(sl.X)
		case BuffMesoGuard:
			return int16(sl.X)

		// Boolean-style toggles
		case BuffDarkSight,
			BuffSoulArrow,
			BuffInvincible,
			BuffShadowPartner,
			BuffThaw,
			BuffWeakness, // typically a debuff from server; keep here for completeness
			BuffCurse:
			return 1

		// Special cases
		case BuffComboAttack:
			// Value is often treated as 1; the active orb count is sent as the extra trailing byte.
			return 1
		case BuffCharges:
			// White Knight charges; often treated as 1. Active charge count/type is handled elsewhere/extra.
			return 1
		case BuffDragonBlood:
			// Dragon Blood grants PAD in Pad and also acts like a toggle; ensure a non-zero.
			if sl.Pad != 0 {
				return int16(sl.Pad)
			}
			return 1

		default:
			// Fallback: if NX has a reasonable X/Y numeric, prefer that; else 1 to ensure application.
			if sl.X != 0 {
				return int16(sl.X)
			}
			if sl.Y != 0 {
				return int16(sl.Y)
			}
			return 1
		}
	}

	values := make([]byte, 0, 64)
	appendTriple := func(val int16) {
		// short value
		values = append(values, byte(val), byte(val>>8))
		// int32 reason/source skill
		id := skillID
		values = append(values, byte(id), byte(id>>8), byte(id>>16), byte(id>>24))
		// short time (seconds)
		t := remainSec
		values = append(values, byte(t), byte(t>>8))
	}

	// Emit a triple for every set bit in canonical order
	nBits := len(mask.Data()) * 32
	hasAny := false
	for bit := 0; bit < nBits; bit++ {
		if mask.GetBitNumber(bit) == 1 {
			val := getValueForBit(bit)
			appendTriple(val)
			hasAny = true
		}
	}

	if !hasAny {
		return nil, 0
	}
	return values, remainSec
}

// buildBuffMask builds a Flag (CFlag) with all relevant bits for the given skill set.
func buildBuffMask(skillID int32) *Flag {
	bits, ok := skillBuffBits[skillID]
	if !ok || len(bits) == 0 {
		return nil
	}
	uMask := NewFlag()
	for _, bit := range bits {
		uMask.SetBitNumber(bit, 1)
	}
	return uMask
}
