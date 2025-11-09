package channel

import (
	"log"
	"time"

	"github.com/Hucaru/Valhalla/constant"
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

	// Beginner Skills
	AddSkillBuff(int32(skill.NimbleFeet), BuffSpeed)
	AddSkillBuff(int32(skill.Recovery), BuffHands) // Recovery buff (HP regen handled by 5-second ticker)

	// 1st Job
	AddSkillBuff(int32(skill.IronBody), BuffWeaponDefense)
	AddSkillBuff(int32(skill.MagicGuard), BuffMagicGuard)
	AddSkillBuff(int32(skill.MagicArmor), BuffWeaponDefense) // some sources also set BuffMagicDefense
	AddSkillBuff(int32(skill.Focus), BuffAvoidability)       // can add BuffAccuracy as well
	AddSkillBuff(int32(skill.DarkSight), BuffDarkSight)
	AddSkillBuff(int32(skill.NimbleBody), BuffSpeed)

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
	AddSkillBuff(int32(skill.Bless), BuffWeaponAttack, BuffWeaponDefense, BuffMagicAttack, BuffMagicDefense, BuffAccuracy, BuffAvoidability)

	// 3rd Job - Magician
	AddSkillBuff(int32(skill.SpellBooster), BuffBooster)
	AddSkillBuff(int32(skill.ILSpellBooster), BuffBooster)

	// GM skills
	AddSkillBuff(int32(skill.GMShadowPartner), BuffShadowPartner)
	AddSkillBuff(int32(skill.GMBless), BuffWeaponAttack, BuffWeaponDefense, BuffMagicAttack, BuffMagicDefense, BuffAccuracy, BuffAvoidability)
	AddSkillBuff(int32(skill.GMHaste), BuffSpeed, BuffJump)
	AddSkillBuff(int32(skill.GMHolySymbol), BuffHolySymbol)
	AddSkillBuff(int32(skill.Hide), BuffInvincible)

	AddSkillBuff(int32(skill.SilverHawk), BuffComboAttack)
	AddSkillBuff(int32(skill.GoldenEagle), BuffComboAttack)
	AddSkillBuff(int32(skill.Puppet), BuffPickPocketMesoUP)
	AddSkillBuff(int32(skill.SniperPuppet), BuffPickPocketMesoUP)
	AddSkillBuff(int32(skill.SummonDragon), BuffComboAttack)
}

func init() {
	LoadBuffs()
}

type CharacterBuffs struct {
	plr               *Player
	comboCount        byte
	activeSkillLevels map[int32]byte // skillID -> level
	expireTimers      map[int32]*time.Timer
	itemMasks         map[int32][]byte // sourceID (-itemId) -> mask
	expireAt          map[int32]int64  // sourceID -> unix ms expiry
	recoveryTicker    *time.Ticker     // 5-second ticker for Recovery skill
}

func NewCharacterBuffs(p *Player) {
	p.buffs = &CharacterBuffs{
		plr:               p,
		activeSkillLevels: make(map[int32]byte),
		expireTimers:      make(map[int32]*time.Timer),
		itemMasks:         make(map[int32][]byte),
		expireAt:          make(map[int32]int64),
	}
}

func (cb *CharacterBuffs) HasGMHide() bool {
	_, ok := cb.activeSkillLevels[int32(skill.Hide)]
	return ok
}

func (cb *CharacterBuffs) hasMobDebuff(skillID byte) bool {
	if cb == nil {
		return false
	}
	for src := range cb.activeSkillLevels {
		base := src & 0xFFFF
		if base == int32(skillID) {
			return true
		}
	}
	return false
}

func (cb *CharacterBuffs) AddBuff(charId, skillID int32, level byte, foreign bool, delay int16) {
	if cb.plr == nil {
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

	cb.AddBuffFromCC(charId, skillID, expiresAtMs, level, foreign, delay)

	if !foreign {
		switch skill.Skill(skillID) {
		case skill.SilverHawk, skill.GoldenEagle, skill.SummonDragon:
			if cb.plr != nil && cb.plr.getSummon(skillID) == nil {
				spawn := cb.plr.pos
				if cb.plr.inst != nil {
					if snapped := cb.plr.inst.fhHist.getFinalPosition(newPos(spawn.x, spawn.y, 0)); snapped.foothold != 0 {
						spawn = snapped
					}
				}
				su := &summon{
					OwnerID:    cb.plr.ID,
					SkillID:    skillID,
					Level:      level,
					Pos:        spawn,
					Stance:     0,
					Foothold:   spawn.foothold,
					IsPuppet:   false,
					SummonType: 0,
				}
				cb.plr.addSummon(su)
			}
		}
	}
}

func buildMaskBytes64(bits []int) []byte {
	m := make([]byte, 8)
	for _, b := range bits {
		if b < 0 || b >= 64 {
			continue
		}
		byteIdx := b / 8
		shift := uint(b % 8) // LSB-first
		m[byteIdx] |= (1 << shift)
	}
	return m
}

// Emit triples by scanning maskBytes in the same wire order we Send:
// bytes 0..7, bits 0..7 (LSB-first).
func (cb *CharacterBuffs) buildBuffTriplesWireOrder(skillID int32, level byte, maskBytes []byte, remainSec int16) []byte {
	levels, err := nx.GetPlayerSkill(skillID)
	if err != nil || level == 0 || int(level) > len(levels) {
		return nil
	}
	sl := levels[level-1]

	val := func(bit int) int16 {
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
		// Toggles/percent-like flags -> 1 (or X if present)
		case BuffMagicGuard, BuffBooster, BuffPowerGuard, BuffMaxHP, BuffMaxMP,
			BuffHolySymbol, BuffMesoUP, BuffPickPocketMesoUP, BuffMesoGuard,
			BuffDarkSight, BuffSoulArrow, BuffInvincible, BuffShadowPartner,
			BuffThaw, BuffWeakness, BuffCurse, BuffComboAttack, BuffCharges, BuffHands:
			if sl.X != 0 {
				return int16(sl.X)
			}
			return 1
		case BuffDragonBlood:
			if sl.Pad != 0 {
				return int16(sl.Pad)
			}
			return 1
		case BuffStun:
			return int16(sl.X)
		default:
			return 1
		}
	}

	out := make([]byte, 0, 64)
	appendTriple := func(v int16) {
		out = append(out, byte(v), byte(v>>8))
		id := skillID
		out = append(out, byte(id), byte(id>>8), byte(id>>16), byte(id>>24))
		// short remaining time in seconds
		t := remainSec
		out = append(out, byte(t), byte(t>>8))
	}

	has := false
	for byteIdx := 0; byteIdx < 8 && byteIdx < len(maskBytes); byteIdx++ {
		b := maskBytes[byteIdx]
		if b == 0 {
			continue
		}
		for bit := 0; bit < 8; bit++ {
			if (b & (1 << uint(bit))) != 0 {
				globalBit := byteIdx*8 + bit // aligns with our Buff* constants
				tripVal := val(globalBit)
				appendTriple(tripVal)
				has = true
			}
		}
	}
	if !has {
		return nil
	}

	return out
}

func buildItemBuffTriplesWireOrder(meta nx.Item, maskBytes []byte, durationSec int16, sourceID int32) []byte {
	remain := durationSec
	if remain < 0 {
		remain = 0
	}

	valForBit := func(bit int) int16 {
		switch bit {
		case BuffAccuracy:
			return meta.ACC
		case BuffAvoidability:
			return meta.EVA
		case BuffSpeed:
			return meta.Speed
		case BuffJump:
			return meta.Jump
		case BuffMagicAttack:
			return meta.MAD
		case BuffMagicDefense:
			return meta.MDD
		case BuffWeaponAttack:
			return meta.PAD
		case BuffWeaponDefense:
			return meta.PDD
		default:
			return 1
		}
	}

	out := make([]byte, 0, 48)
	appendTriple := func(v int16) {
		// short value
		out = append(out, byte(v), byte(v>>8))
		// int32 sourceID (negative Item ID)
		id := sourceID
		out = append(out, byte(id), byte(id>>8), byte(id>>16), byte(id>>24))
		// short time (seconds)
		t := remain
		out = append(out, byte(t), byte(t>>8))
	}

	for byteIdx := 0; byteIdx < 8 && byteIdx < len(maskBytes); byteIdx++ {
		b := maskBytes[byteIdx]
		if b == 0 {
			continue
		}
		for bit := 0; bit < 8; bit++ {
			if (b & (1 << uint(bit))) != 0 {
				globalBit := byteIdx*8 + bit
				appendTriple(valForBit(globalBit))
			}
		}
	}
	return out
}

// durationSec is the client-visible remaining time in seconds. Source ID is encoded as -Item.ID.
func (cb *CharacterBuffs) AddItemBuff(it Item) {
	var durationSec int16 = 0
	meta, err := nx.GetItem(it.ID)
	if err != nil {
		return
	}

	// Handle debuff curing first
	cb.cureDebuffs(meta)

	bits := make([]int, 0, 6)
	if meta.ACC > 0 {
		bits = append(bits, BuffAccuracy)
	}
	if meta.EVA > 0 {
		bits = append(bits, BuffAvoidability)
	}
	if meta.Speed > 0 {
		bits = append(bits, BuffSpeed)
	}
	if meta.Jump > 0 {
		bits = append(bits, BuffJump)
	}
	if meta.MAD > 0 {
		bits = append(bits, BuffMagicAttack)
	}
	if meta.MDD > 0 {
		bits = append(bits, BuffMagicDefense)
	}
	if meta.PAD > 0 {
		bits = append(bits, BuffWeaponAttack)
	}
	if meta.PDD > 0 {
		bits = append(bits, BuffWeaponDefense)
	}
	if len(bits) == 0 {
		return
	}

	// NX Time is in milliseconds.
	if meta.Time > 0 {
		ms := int32(meta.Time)
		sec := int16((ms + 999) / 1000) // ceil(ms/1000)
		if sec > 0 {
			durationSec = sec
		}
	}

	// Build mask and per-stat triples in the same (LSB-first) wire order as skills.
	maskBytes := buildMaskBytes64(bits)
	sourceID := -it.ID
	values := buildItemBuffTriplesWireOrder(meta, maskBytes, durationSec, sourceID)

	// Send to self and others (items don't need extra combo/charges byte).
	const extra byte = 0
	const delay int16 = 0
	cb.plr.Send(packetPlayerGiveBuff(maskBytes, values, delay, extra))

	m := make([]byte, len(maskBytes))
	copy(m, maskBytes)
	cb.itemMasks[sourceID] = m

	// Track authoritative expiry, schedule using Dispatch
	if durationSec > 0 {
		exp := time.Now().Add(time.Duration(durationSec) * time.Second).UnixMilli()
		cb.expireAt[sourceID] = exp
		cb.scheduleExpiryLocked(sourceID, time.Duration(durationSec)*time.Second)
	}
}

// cureDebuffs removes debuffs based on the cure flags in the item
func (cb *CharacterBuffs) cureDebuffs(meta nx.Item) {
	if cb.plr == nil {
		return
	}

	// Check which debuffs this item can cure
	var debuffsToCure []int

	if meta.Poison > 0 {
		debuffsToCure = append(debuffsToCure, BuffPoison)
	}
	if meta.Weakness > 0 {
		debuffsToCure = append(debuffsToCure, BuffWeakness)
	}
	if meta.Curse > 0 {
		debuffsToCure = append(debuffsToCure, BuffCurse)
	}
	if meta.Darkness > 0 {
		debuffsToCure = append(debuffsToCure, BuffDarkness)
	}
	if meta.Seal > 0 {
		debuffsToCure = append(debuffsToCure, BuffSeal)
	}

	if len(debuffsToCure) == 0 {
		return
	}

	// Find and remove active debuffs that match
	for skillID := range cb.activeSkillLevels {
		// Check if this is a mob debuff (rValue format: skillID | (level << 16))
		baseSkillID := skillID & 0xFFFF
		if baseSkillID >= 100 && baseSkillID <= 200 {
			// This is a mob debuff, check if we should cure it
			bits, ok := skillBuffBits[skillID]
			if ok {
				for _, bit := range bits {
					for _, cureBit := range debuffsToCure {
						if bit == cureBit {
							cb.expireBuffNow(skillID)
							break
						}
					}
				}
			}
		}
	}
}

func (cb *CharacterBuffs) AddItemBuffFromCC(itemID int32, expiresAtMs int64) {
	meta, err := nx.GetItem(itemID)
	if err != nil {
		return
	}

	// Re-derive bits like AddItemBuff
	bits := make([]int, 0, 8)
	if meta.ACC > 0 {
		bits = append(bits, BuffAccuracy)
	}
	if meta.EVA > 0 {
		bits = append(bits, BuffAvoidability)
	}
	if meta.Speed > 0 {
		bits = append(bits, BuffSpeed)
	}
	if meta.Jump > 0 {
		bits = append(bits, BuffJump)
	}
	if meta.MAD > 0 {
		bits = append(bits, BuffMagicAttack)
	}
	if meta.MDD > 0 {
		bits = append(bits, BuffMagicDefense)
	}
	if meta.PAD > 0 {
		bits = append(bits, BuffWeaponAttack)
	}
	if meta.PDD > 0 {
		bits = append(bits, BuffWeaponDefense)
	}
	if len(bits) == 0 {
		return
	}

	remainSec := int16(0)
	if expiresAtMs > 0 {
		now := time.Now().UnixMilli()
		if d := expiresAtMs - now; d > 0 {
			if d > 32767*1000 {
				d = 32767 * 1000
			}
			remainSec = int16((d + 500) / 1000)
		}
	}

	maskBytes := buildMaskBytes64(bits)
	sourceID := -itemID
	values := buildItemBuffTriplesWireOrder(meta, maskBytes, remainSec, sourceID)

	// Send packets
	cb.plr.Send(packetPlayerGiveBuff(maskBytes, values, 0, 0))

	// Track in memory and set timer
	m := make([]byte, len(maskBytes))
	copy(m, maskBytes)
	cb.itemMasks[sourceID] = m

	if expiresAtMs > 0 {
		cb.expireAt[sourceID] = expiresAtMs
		cb.scheduleExpiryLocked(sourceID, time.Until(time.UnixMilli(expiresAtMs)))
	}
}

// AddMobDebuff applies a debuff from a mob skill to the player
func (cb *CharacterBuffs) AddMobDebuff(skillID, level byte, durationSec int16) {
	if cb.plr == nil {
		return
	}

	// Map mob skill IDs to buff bit positions
	var bits []int

	switch skillID {
	case skill.Mob.Seal:
		bits = []int{BuffSeal}
	case skill.Mob.Darkness:
		bits = []int{BuffDarkness}
	case skill.Mob.Weakness:
		bits = []int{BuffWeakness}
	case skill.Mob.Stun:
		bits = []int{BuffStun}
	case skill.Mob.Curse:
		bits = []int{BuffCurse}
	case skill.Mob.Poison:
		bits = []int{BuffPoison}
	case skill.Mob.Slow:
		bits = []int{BuffSpeed}
	default:
		return
	}

	if len(bits) == 0 {
		return
	}

	// Pack skill ID and level into R value: skillID | (level << 16)
	// This matches the reference implementation
	rValue := int32(skillID) | (int32(level) << 16)

	// Register the mob debuff in skillBuffBits so expiration can find it
	skillBuffBits[rValue] = bits

	// Build mask bytes
	maskBytes := buildMaskBytes64(bits)

	// Build value triples for the self packet - scan mask in wire order
	out := make([]byte, 0, 32)
	for byteIdx := 0; byteIdx < 8 && byteIdx < len(maskBytes); byteIdx++ {
		b := maskBytes[byteIdx]
		if b == 0 {
			continue
		}
		for bitPos := 0; bitPos < 8; bitPos++ {
			if (b & (1 << uint(bitPos))) != 0 {
				globalBit := byteIdx*8 + bitPos

				var nValue int16
				switch globalBit {
				case BuffSpeed:
					// Slow: negative speed value
					nValue = -int16(level * 10)
				case BuffPoison:
					// Poison: damage value based on level (X value from skill data)
					nValue = int16(level)
				default:
					// Other debuffs: just 1
					nValue = 1
				}

				// short N value
				out = append(out, byte(nValue), byte(nValue>>8))
				// int32 R value (packed skill ID and level)
				out = append(out, byte(rValue), byte(rValue>>8), byte(rValue>>16), byte(rValue>>24))
				// short time (seconds)
				out = append(out, byte(durationSec), byte(durationSec>>8))
			}
		}
	}

	// Send to self
	const extra byte = 0
	const delay int16 = 0
	cb.plr.Send(packetPlayerGiveBuff(maskBytes, out, delay, extra))

	// Build foreign buff values for broadcasting - scan mask in wire order
	fout := make([]byte, 0, 16)
	for byteIdx := 0; byteIdx < 8 && byteIdx < len(maskBytes); byteIdx++ {
		b := maskBytes[byteIdx]
		if b == 0 {
			continue
		}
		for bitPos := 0; bitPos < 8; bitPos++ {
			if (b & (1 << uint(bitPos))) != 0 {
				globalBit := byteIdx*8 + bitPos

				switch globalBit {
				case BuffSpeed:
					// Speed/Slow: write byte N (speed value)
					speedVal := byte(-level * 10) // negative for slow
					fout = append(fout, speedVal)
				case BuffStun, BuffDarkness, BuffSeal, BuffWeakness:
					// These need int32 R (packed skill ID and level)
					fout = append(fout, byte(rValue), byte(rValue>>8), byte(rValue>>16), byte(rValue>>24))
				case BuffPoison:
					// Poison: write short N, int32 R
					n := int16(level)
					fout = append(fout, byte(n), byte(n>>8))
					fout = append(fout, byte(rValue), byte(rValue>>8), byte(rValue>>16), byte(rValue>>24))
				case BuffCurse:
					// Curse: write int32 R (packed skill ID and level)
					fout = append(fout, byte(rValue), byte(rValue>>8), byte(rValue>>16), byte(rValue>>24))
				}
			}
		}
	}

	// Broadcast to others in the instance
	if cb.plr.inst != nil && len(fout) > 0 {
		cb.plr.inst.send(packetPlayerGiveForeignBuff(cb.plr.ID, maskBytes, fout, delay))
	}

	// Track the debuff with expiry using the packed rValue as the key
	if durationSec > 0 {
		expiresAtMs := time.Now().Add(time.Duration(durationSec) * time.Second).UnixMilli()
		cb.expireAt[rValue] = expiresAtMs
		cb.scheduleExpiryLocked(rValue, time.Duration(durationSec)*time.Second)
		cb.activeSkillLevels[rValue] = level
	}
}

func (cb *CharacterBuffs) AddBuffFromCC(charId, skillID int32, expiresAtMs int64, level byte, foreign bool, delay int16) {
	if skillID == 0 || level == 0 {
		return
	}
	cb.check(skillID)

	// Use configured per-skill bit positions -> build 8-byte mask deterministically (LSB-first).
	bits, ok := skillBuffBits[skillID]
	if !ok || len(bits) == 0 {
		return
	}
	maskBytes := buildMaskBytes64(bits)

	// Compute client-visible remaining time in seconds (ceil), clamped to int16 range.
	remainSec := int16(0)
	if expiresAtMs > 0 {
		now := time.Now().UnixMilli()
		if d := expiresAtMs - now; d > 0 {
			if d > 32767*1000 {
				d = 32767 * 1000
			}
			remainSec = int16((d + 500) / 1000)
		}
	}

	// Emit value triples in exactly the same mask byte/bit order for local (self) path.
	values := cb.buildBuffTriplesWireOrder(skillID, level, maskBytes, remainSec)
	if len(values) == 0 {
		log.Printf("BUFF ABORT: no values produced for skillID=%d", skillID)
		return
	}

	// Extra trailing byte only for combo/charges.
	extra := byte(0)

	// Send to self
	cb.plr.Send(packetPlayerGiveBuff(maskBytes, values, delay, extra))

	// Build and broadcast the foreign buff state so others see the effect/state.
	if cb.plr != nil && cb.plr.inst != nil {
		fMask, fVals := cb.buildForeignBuffMaskAndValues(skillID, level, bits)
		cb.plr.inst.send(packetPlayerGiveForeignBuff(cb.plr.ID, fMask, fVals, delay))
	}

	cb.activeSkillLevels[skillID] = level

	cb.expireAt[skillID] = expiresAtMs
	d := time.Until(time.UnixMilli(expiresAtMs))
	cb.scheduleExpiryLocked(skillID, d)

	// Start Recovery skill ticker if this is the Recovery skill
	if skillID == int32(skill.Recovery) {
		cb.startRecoveryTicker(level)
	}

	// If this is a non-puppet summon skill applied to self (e.g., on CC/login restore), spawn the summon entity now.
	if !foreign {
		switch skill.Skill(skillID) {
		case skill.SilverHawk, skill.GoldenEagle, skill.SummonDragon:
			if cb.plr != nil {
				spawn := cb.plr.pos
				if cb.plr.inst != nil {
					if snapped := cb.plr.inst.fhHist.getFinalPosition(newPos(spawn.x, spawn.y, 0)); snapped.foothold != 0 {
						spawn = snapped
					}
				}
				su := &summon{
					OwnerID:    cb.plr.ID,
					SkillID:    skillID,
					Level:      level,
					Pos:        spawn,
					Stance:     0,
					Foothold:   spawn.foothold,
					IsPuppet:   false,
					SummonType: 0,
				}
				cb.plr.addSummon(su)
			}
		}
	}
}

func (cb *CharacterBuffs) buildForeignBuffMaskAndValues(skillID int32, level byte, skillBits []int) ([]byte, []byte) {
	levels, err := nx.GetPlayerSkill(skillID)
	if err != nil || level == 0 || int(level) > len(levels) {
		return nil, nil
	}
	sl := levels[level-1]

	// Helper to check if this skill sets a given bit
	hasBit := func(bit int) bool {
		for _, b := range skillBits {
			if b == bit {
				return true
			}
		}
		return false
	}

	addedBits := make([]int, 0, 8)
	out := make([]byte, 0, 32)

	// Speed: write byte N
	if hasBit(BuffSpeed) {
		addedBits = append(addedBits, BuffSpeed)
		val := byte(1)
		if sl.Speed != 0 {
			val = byte(sl.Speed)
		}
		out = append(out, val)
	}
	// ComboAttack: write byte N
	if hasBit(BuffComboAttack) {
		addedBits = append(addedBits, BuffComboAttack)
		val := byte(1)
		if sl.X != 0 {
			val = byte(sl.X)
		}
		out = append(out, val)
	}
	// Charges: write int32 R (reason/source skill)
	if hasBit(BuffCharges) {
		addedBits = append(addedBits, BuffCharges)
		id := skillID
		out = append(out, byte(id), byte(id>>8), byte(id>>16), byte(id>>24))
	}
	// Stun: write int32 R
	if hasBit(BuffStun) {
		addedBits = append(addedBits, BuffStun)
		id := skillID
		out = append(out, byte(id), byte(id>>8), byte(id>>16), byte(id>>24))
	}
	// Darkness: write int32 R
	if hasBit(BuffDarkness) {
		addedBits = append(addedBits, BuffDarkness)
		id := skillID
		out = append(out, byte(id), byte(id>>8), byte(id>>16), byte(id>>24))
	}
	// Seal: write int32 R
	if hasBit(BuffSeal) {
		addedBits = append(addedBits, BuffSeal)
		id := skillID
		out = append(out, byte(id), byte(id>>8), byte(id>>16), byte(id>>24))
	}
	// Weakness: write int32 R
	if hasBit(BuffWeakness) {
		addedBits = append(addedBits, BuffWeakness)
		id := skillID
		out = append(out, byte(id), byte(id>>8), byte(id>>16), byte(id>>24))
	}
	// Poison: write short N, int32 R
	if hasBit(BuffPoison) {
		addedBits = append(addedBits, BuffPoison)
		n := int16(1)
		if sl.X != 0 {
			n = int16(sl.X)
		}
		out = append(out, byte(n), byte(n>>8))
		id := skillID
		out = append(out, byte(id), byte(id>>8), byte(id>>16), byte(id>>24))
	}
	// SoulArrow: toggle, no payload
	if hasBit(BuffSoulArrow) {
		addedBits = append(addedBits, BuffSoulArrow)
	}
	// ShadowPartner: toggle, no payload
	if hasBit(BuffShadowPartner) {
		addedBits = append(addedBits, BuffShadowPartner)
	}
	// DarkSight: toggle, no payload
	if hasBit(BuffDarkSight) {
		addedBits = append(addedBits, BuffDarkSight)
	}
	// BuffHands: toggle, no payload (used for Recovery skill visual indicator)
	if hasBit(BuffHands) {
		addedBits = append(addedBits, BuffHands)
	}

	if len(addedBits) == 0 {
		return nil, nil
	}
	mask := buildMaskBytes64(addedBits)
	return mask, out
}

// ClearBuff removes a specific buff from Player and DB.
func (cb *CharacterBuffs) ClearBuff(skillID int32, _ uint32) {
	mask := buildBuffMask(skillID)
	if mask != nil && !mask.IsZero() && cb.plr.inst != nil {
		cb.plr.inst.send(packetPlayerCancelForeignBuff(cb.plr.ID, mask.ToByteArray(false)))
	}
	delete(cb.activeSkillLevels, skillID)
	delete(cb.expireAt, skillID)
	if t, ok := cb.expireTimers[skillID]; ok && t != nil {
		t.Stop()
		delete(cb.expireTimers, skillID)
	}
}

// dispelAllBuffs removes all active buffs from the player
func (cb *CharacterBuffs) dispelAllBuffs() {
	for skillID := range cb.activeSkillLevels {
		if skillID >= 1000 {
			cb.expireBuffNow(skillID)
		}
	}
}

func (cb *CharacterBuffs) AuditAndExpireStaleBuffs() {
	now := time.Now().UnixMilli()
	toExpire := make([]int32, 0, 4)

	for src, ts := range cb.expireAt {
		if ts > 0 && ts <= now {
			toExpire = append(toExpire, src)
		}
	}

	for _, src := range toExpire {
		cb.expireBuffNow(src)
	}
}

// Snapshot/Restore across CC/Login

type BuffSnapshot struct {
	SourceID    int32 // skillID or -itemID
	Level       byte  // 0 for Item buffs
	ExpiresAtMs int64 // 0 for toggles/indefinite
}

func (cb *CharacterBuffs) Snapshot() []BuffSnapshot {
	out := make([]BuffSnapshot, 0, len(cb.activeSkillLevels)+len(cb.itemMasks))

	// Skills
	for sid, lvl := range cb.activeSkillLevels {
		out = append(out, BuffSnapshot{
			SourceID:    sid,
			Level:       lvl,
			ExpiresAtMs: cb.expireAt[sid],
		})
	}
	// Items (sourceID is negative)
	for src := range cb.itemMasks {
		out = append(out, BuffSnapshot{
			SourceID:    src,              // negative Item ID
			Level:       0,                // not used
			ExpiresAtMs: cb.expireAt[src], // may be 0 if toggle
		})
	}
	return out
}

func (cb *CharacterBuffs) RestoreFromSnapshot(snaps []BuffSnapshot) {
	if len(snaps) == 0 {
		return
	}
	for _, s := range snaps {
		if s.SourceID > 0 {
			// Skill
			cb.AddBuffFromCC(cb.plr.ID, s.SourceID, s.ExpiresAtMs, s.Level, false, 0)
		} else if s.SourceID < 0 {
			// Item
			itemID := -s.SourceID
			cb.AddItemBuffFromCC(itemID, s.ExpiresAtMs)
		}
	}
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

func (cb *CharacterBuffs) post(fn func()) {
	if cb == nil || cb.plr == nil || cb.plr.inst == nil || cb.plr.inst.dispatch == nil {
		return
	}
	select {
	case cb.plr.inst.dispatch <- fn:
		return
	default:
		fn()
		return
	}
}

func (cb *CharacterBuffs) scheduleExpiryLocked(skillID int32, after time.Duration) {
	// Cancel previous
	if t, ok := cb.expireTimers[skillID]; ok && t != nil {
		t.Stop()
		delete(cb.expireTimers, skillID)
	}

	if after <= 0 {
		cb.post(func() { cb.expireBuffNow(skillID) })
		return
	}

	cb.expireTimers[skillID] = time.AfterFunc(after, func() {
		cb.post(func() { cb.expireBuffNow(skillID) })
	})
}

func (cb *CharacterBuffs) expireBuffNow(skillID int32) {
	if cb == nil || cb.plr == nil {
		return
	}
	if t, ok := cb.expireTimers[skillID]; ok && t != nil {
		t.Stop()
		delete(cb.expireTimers, skillID)
	}
	delete(cb.expireAt, skillID)

	// Stop Recovery skill ticker if this is the Recovery skill
	if skillID == int32(skill.Recovery) {
		cb.stopRecoveryTicker()
	}

	// Item-source (negative) or skill-source (positive)
	bits, ok := skillBuffBits[skillID]
	if !ok || len(bits) == 0 {
		if skillID < 0 {
			if mask, ok2 := cb.itemMasks[skillID]; ok2 {
				cb.plr.Send(packetPlayerCancelBuff(mask))
				if cb.plr.inst != nil {
					cb.plr.inst.send(packetPlayerCancelForeignBuff(cb.plr.ID, mask))
				}
				delete(cb.itemMasks, skillID)
			}
		}
		cb.despawnSummonIfMatches(skillID)
		return
	}
	maskBytes := buildMaskBytes64(bits)

	cb.plr.Send(packetPlayerCancelBuff(maskBytes))
	if cb.plr.inst != nil {
		cb.plr.inst.send(packetPlayerCancelForeignBuff(cb.plr.ID, maskBytes))
	}

	delete(cb.activeSkillLevels, skillID)

	// Clean up mob debuff from skillBuffBits (rValue format: skillID | (level << 16))
	// Only clean up if it looks like a mob debuff (skill ID 100-200 range)
	baseSkillID := skillID & 0xFFFF
	if baseSkillID >= 100 && baseSkillID <= 200 {
		delete(skillBuffBits, skillID)
	}

	cb.despawnSummonIfMatches(skillID)
}

func (cb *CharacterBuffs) despawnSummonIfMatches(skillID int32) {
	p := cb.plr
	if p == nil || p.summons == nil {
		return
	}
	if p.summons.puppet != nil && p.summons.puppet.SkillID == skillID {
		p.removeSummon(true, constant.SummonRemoveReasonKeepBuff)
		return
	}
	if p.summons.summon != nil && p.summons.summon.SkillID == skillID {
		p.removeSummon(false, constant.SummonRemoveReasonKeepBuff)
		return
	}
}

func (cb *CharacterBuffs) check(skillID int32) {
	// Placeholder for conflicting buff cleanup if needed.
}

func (cb *CharacterBuffs) startRecoveryTicker(level byte) {
	// Stop any existing ticker first
	cb.stopRecoveryTicker()

	// Create new 5-second ticker
	cb.recoveryTicker = time.NewTicker(5 * time.Second)

	// Start goroutine to handle recovery ticks
	go func() {
		ticker := cb.recoveryTicker
		if ticker == nil {
			return
		}

		for range ticker.C {
			cb.post(func() {
				// Check if buff is still active and player is alive
				currentLevel, hasRecovery := cb.activeSkillLevels[int32(skill.Recovery)]
				if !hasRecovery || cb.plr == nil || cb.plr.hp <= 0 {
					return
				}

				// Get skill data to retrieve the X value (recovery amount)
				skillData, err := nx.GetPlayerSkill(int32(skill.Recovery))
				if err != nil || currentLevel == 0 || int(currentLevel) > len(skillData) {
					return
				}

				// Get the recovery amount from the X field
				recoveryAmount := int16(skillData[currentLevel-1].X)
				if recoveryAmount > 0 {
					cb.plr.giveHP(recoveryAmount)
				}
			})
		}
	}()
}

func (cb *CharacterBuffs) stopRecoveryTicker() {
	if cb.recoveryTicker != nil {
		cb.recoveryTicker.Stop()
		cb.recoveryTicker = nil
	}
}
