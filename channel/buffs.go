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
}

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

func (cb *CharacterBuffs) AddBuffFromCC(skillID int32, expiresAtMs int64, level byte, sinc1, sinc2 int, delay int16) {
	if cb == nil || cb.plr == nil {
		return
	}
	if skillID == 0 || level == 0 {
		return
	}

	cb.check(skillID)

	mask := buildBuffMaskFromNX(skillID, level)
	if mask == nil || mask.IsZero() {
		return
	}

	values := cb.buildBuffValues(skillID, level, mask, expiresAtMs)
	if len(values) == 0 {
		return
	}

	maskBytes := mask.ToByteArray(false)

	// SELF: mask + triples + int16 delay + optional extra Decode1
	cb.plr.send(packetPlayerGiveBuff(maskBytes, values, delay, 0))

	// OTHERS: charId + mask + triples + optional extra Decode1
	if cb.plr.inst != nil {
		cb.plr.inst.send(packetPlayerGiveForeignBuff(cb.plr.id, maskBytes, values, 0))
	}

	cb.activeSkillLevels[skillID] = level
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
