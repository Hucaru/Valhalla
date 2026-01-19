package channel

import (
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/constant/skill"
	"github.com/Hucaru/Valhalla/nx"
)

// DamageRange represents the min and max damage for validation
type DamageRange struct {
	Min float64
	Max float64
}

// CalcHitResult represents the result of a hit calculation
type CalcHitResult struct {
	IsCrit       bool
	IsMiss       bool
	MinDamage    float64
	MaxDamage    float64
	ExpectedDmg  float64
	ClientDamage int32
	IsValid      bool
}

type Roller struct {
	rollIndex int
	rolls     []uint32
}

func NewRoller(randomizer *rand.Rand, numRolls int) *Roller {
	if randomizer == nil {
		randomizer = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	}

	rolls := make([]uint32, numRolls)
	for i := 0; i < numRolls; i++ {
		rolls[i] = randomizer.Uint32()
	}
	return &Roller{
		rollIndex: 0,
		rolls:     rolls,
	}
}

func (r *Roller) Roll(modifier float64) float64 {
	if r == nil || len(r.rolls) == 0 {
		return 0.5 // Return middle of the typical 0-1 range
	}

	idx := r.rollIndex % len(r.rolls)
	r.rollIndex++
	roll := r.rolls[idx]
	rollValue := float64(roll%10000000) * modifier
	return rollValue
}

type ElementAmpData struct {
	Magic int
	Mana  int
}

type DamageCalculator struct {
	player       *Player
	data         *attackData
	attackType   int
	weaponType   constant.WeaponType
	skill        *nx.PlayerSkill
	skillID      int32
	skillLevel   byte
	isRanged     bool
	masteryMod   float64
	critSkill    *nx.PlayerSkill
	critLevel    byte
	watk         int16
	projectileID int32
	attackAction constant.AttackAction
	attackOption constant.AttackOption
}

// NewDamageCalculator creates a new damage calculator
func NewDamageCalculator(plr *Player, data *attackData, attackType int) *DamageCalculator {
	calc := &DamageCalculator{
		player:       plr,
		data:         data,
		attackType:   attackType,
		isRanged:     attackType == attackRanged,
		skillID:      data.skillID,
		skillLevel:   data.skillLevel,
		projectileID: data.projectileID,
		attackAction: constant.AttackAction(data.action),
		attackOption: constant.AttackOption(data.option),
	}

	weaponID := int32(0)
	for _, item := range plr.equip {
		if item.slotID == -11 {
			weaponID = item.ID
			break
		}
	}
	calc.weaponType = constant.GetWeaponType(weaponID)

	if data.skillID > 0 {
		if skillData, err := nx.GetPlayerSkill(data.skillID); err == nil && len(skillData) > 0 {
			if data.skillLevel > 0 && int(data.skillLevel) <= len(skillData) {
				calc.skill = &skillData[data.skillLevel-1]
			}
		}
	}

	calc.masteryMod = calc.GetMasteryModifier()
	calc.critLevel, calc.critSkill = calc.GetCritSkill()
	calc.watk = calc.GetTotalWatk()

	return calc
}

// ValidateAttack validates all hits in an attack and determines critical hits
func (calc *DamageCalculator) ValidateAttack() [][]CalcHitResult {
	results := make([][]CalcHitResult, len(calc.data.attackInfo))

	for targetIdx := range calc.data.attackInfo {
		info := &calc.data.attackInfo[targetIdx]

		if calc.player.inst == nil {
			continue
		}
		mob, err := calc.player.inst.lifePool.getMobFromID(info.spawnID)
		if err != nil {
			continue
		}

		roller := NewRoller(calc.player.rng, constant.DamageRollsPerTarget)
		ampData := calc.GetElementAmplification()
		targetAccuracy := calc.GetTargetAccuracy(&mob)

		results[targetIdx] = make([]CalcHitResult, len(info.damages))
		for hitIdx := range info.damages {
			results[targetIdx][hitIdx] = calc.CalculateHit(
				&mob,
				info,
				roller,
				ampData,
				targetAccuracy,
				hitIdx,
				targetIdx,
			)
		}
	}

	return results
}

func (calc *DamageCalculator) CalculateHit(
	mob *monster,
	info *attackInfo,
	roller *Roller,
	ampData *ElementAmpData,
	targetAccuracy float64,
	hitIdx int,
	targetIdx int,
) CalcHitResult {
	result := CalcHitResult{
		ClientDamage: info.damages[hitIdx],
		IsValid:      false,
	}

	if calc.handleSpecialSkillDamage(&result, mob, roller) {
		return result
	}
	if calc.attackType == attackMagic && mob.invincible {
		result.MinDamage = 1
		result.MaxDamage = 1
		result.IsValid = (result.ClientDamage == 1)
		return result
	}

	if calc.GetIsMiss(roller, targetAccuracy, mob) {
		result.IsMiss = true
		result.MinDamage = 0
		result.MaxDamage = 0
		result.IsValid = (result.ClientDamage == 0)
		return result
	}

	minDmg, maxDmg := calc.CalculateBaseDamageRange(mob, hitIdx)

	redMin, redMax := calc.CalculateDefenseReductionBounds(mob)
	minDmg -= redMax
	maxDmg -= redMin

	baseMinInt := math.Floor(minDmg)
	baseMaxInt := math.Floor(maxDmg)
	baseDmg := (minDmg + maxDmg) / 2.0

	multiplier := 1.0
	if calc.skill != nil && calc.skill.Damage > 0 {
		multiplier = float64(calc.skill.Damage) / 100.0
	}

	minDmg *= multiplier
	maxDmg *= multiplier
	baseDmg *= multiplier

	result.IsCrit = calc.CheckCritical(roller)
	if result.IsCrit && calc.critSkill != nil {
		critBonus := float64(calc.critSkill.Damage-100) / 100.0

		minDmg += critBonus * baseMinInt
		maxDmg += critBonus * baseMaxInt
		baseDmg += critBonus * math.Floor((baseMinInt+baseMaxInt)/2.0)
	}

	afterMod := calc.GetAfterModifier(targetIdx, baseDmg)
	minDmg *= afterMod
	maxDmg *= afterMod
	baseDmg *= afterMod

	minDmg = math.Floor(minDmg)
	maxDmg = math.Floor(maxDmg)

	result.MinDamage = minDmg
	result.MaxDamage = maxDmg
	result.ExpectedDmg = baseDmg

	tolerance := constant.DamageVarianceTolerance
	toleranceMax := maxDmg * (1.0 + tolerance)

	clientDmgFloat := float64(result.ClientDamage)
	result.IsValid = (clientDmgFloat <= toleranceMax)

	if !result.IsValid {
		log.Printf("Suspicious high damage from player %s (ID: %d): client=%d, max expected=%.0f (with tolerance), skill=%d",
			calc.player.Name, calc.player.ID, result.ClientDamage, toleranceMax, calc.skillID)
	}

	return result
}

func (calc *DamageCalculator) handleSpecialSkillDamage(result *CalcHitResult, mob *monster, roller *Roller) bool {
	str := float64(calc.player.str)
	dex := float64(calc.player.dex)
	luk := float64(calc.player.luk)

	if skill.Skill(calc.skillID) == skill.ShadowMeso {
		if calc.skill != nil {
			mesoCount := float64(calc.skill.X)
			result.MinDamage = 10.0 * mesoCount
			result.MaxDamage = 10.0 * mesoCount
			result.ExpectedDmg = 10.0 * mesoCount

			if roller != nil && calc.skill.Prop > 0 {
				roll := roller.Roll(constant.DamagePropModifier)
				if roll < float64(calc.skill.Prop) {
					result.IsCrit = true
					bonusDmg := float64(100 + calc.skill.X)
					result.MinDamage *= bonusDmg * 0.01
					result.MaxDamage *= bonusDmg * 0.01
					result.ExpectedDmg *= bonusDmg * 0.01
				}
			}

			result.IsValid = float64(result.ClientDamage) <= result.MaxDamage*(1.0+constant.DamageVarianceTolerance)
			return true
		}
	}

	if skill.Skill(calc.skillID) == skill.ShadowWeb {
		if calc.skill != nil && calc.skillLevel > 0 {
			divisor := 50.0 - float64(calc.skillLevel)
			if divisor <= 0 {
				divisor = 1.0
			}
			dmg := float64(mob.maxHP) / divisor
			result.MinDamage = dmg
			result.MaxDamage = dmg
			result.ExpectedDmg = dmg
			result.IsValid = float64(result.ClientDamage) <= result.MaxDamage*(1.0+constant.DamageVarianceTolerance)
			return true
		}
	}

	if skill.Skill(calc.skillID) == skill.Drain {
		basicAttack := float64(calc.watk)
		result.MinDamage = (8.0*(str+luk) + dex*2.0) / 100.0 * basicAttack
		result.MaxDamage = (18.5*(str+luk) + dex*2.0) / 100.0 * basicAttack
		result.ExpectedDmg = (result.MinDamage + result.MaxDamage) / 2.0
		result.IsValid = float64(result.ClientDamage) <= result.MaxDamage*(1.0+constant.DamageVarianceTolerance)
		return true
	}

	if skill.Skill(calc.skillID) == skill.PoisonMyst {
		if calc.skillLevel > 0 {
			divisor := 70.0 - float64(calc.skillLevel)
			if divisor <= 0 {
				divisor = 1.0
			}
			dmg := float64(mob.maxHP) / divisor
			result.MinDamage = dmg
			result.MaxDamage = dmg
			result.ExpectedDmg = dmg
			result.IsValid = float64(result.ClientDamage) <= result.MaxDamage*(1.0+constant.DamageVarianceTolerance)
			return true
		}
	}

	if calc.attackType == attackSummon {
		attackRate := float64(100)
		if calc.skill != nil {
			attackRate = float64(calc.skill.Damage)
		}
		result.MinDamage = (dex*2.5*0.7 + str) * attackRate / 100.0
		result.MaxDamage = (dex*2.5 + str) * attackRate / 100.0
		result.ExpectedDmg = (result.MinDamage + result.MaxDamage) / 2.0
		result.IsValid = float64(result.ClientDamage) <= result.MaxDamage*(1.0+constant.DamageVarianceTolerance)
		return true
	}

	return false
}

func (calc *DamageCalculator) CalculateBaseDamageRange(mob *monster, hitIdx int) (float64, float64) {
	str := float64(calc.GetTotalStr())
	dex := float64(calc.GetTotalDex())
	luk := float64(calc.GetTotalLuk())
	watk := float64(calc.GetTotalWatk())

	if calc.attackType == attackMagic {
		return calc.CalculateMagicDamageRange()
	}
	masteryMin := calc.masteryMod
	masteryMax := 1.0

	var minStatMod, maxStatMod float64

	isSwing := calc.attackAction >= constant.AttackActionSwing1H1 && calc.attackAction <= constant.AttackActionSwing2H7

	switch calc.weaponType {
	case constant.WeaponTypeBow2:
		if skill.Skill(calc.skillID) == skill.PowerKnockback || skill.Skill(calc.skillID) == skill.CBPowerKnockback {
			minStatMod = dex*3.4*0.1*0.9 + str
			maxStatMod = dex*3.4 + str
			minDmg := minStatMod * watk / 150.0
			maxDmg := maxStatMod * watk / 150.0
			return minDmg, maxDmg
		}
		minStatMod = dex*masteryMin*3.4 + str
		maxStatMod = dex*masteryMax*3.4 + str

	case constant.WeaponTypeCrossbow2:
		if skill.Skill(calc.skillID) == skill.PowerKnockback || skill.Skill(calc.skillID) == skill.CBPowerKnockback {
			minStatMod = dex*3.4*0.1*0.9 + str
			maxStatMod = dex*3.4 + str
			minDmg := minStatMod * watk / 150.0
			maxDmg := maxStatMod * watk / 150.0
			return minDmg, maxDmg
		}
		minStatMod = dex*masteryMin*3.6 + str
		maxStatMod = dex*masteryMax*3.6 + str

	case constant.WeaponTypeAxe2H, constant.WeaponTypeBW2H:
		if isSwing {
			minStatMod = str*masteryMin*4.8 + dex
			maxStatMod = str*masteryMax*4.8 + dex
		} else {
			minStatMod = str*masteryMin*3.4 + dex
			maxStatMod = str*masteryMax*3.4 + dex
		}

	case constant.WeaponTypeSpear2, constant.WeaponTypePolearm2:
		if skill.Skill(calc.skillID) == skill.DragonRoar {
			minStatMod = str*4.0*calc.masteryMod*0.9 + dex
			maxStatMod = str*4.0 + dex
		} else if isSwing != (calc.weaponType == constant.WeaponTypeSpear2) {
			minStatMod = str*masteryMin*5.0 + dex
			maxStatMod = str*masteryMax*5.0 + dex
		} else {
			minStatMod = str*masteryMin*3.0 + dex
			maxStatMod = str*masteryMax*3.0 + dex
		}

	case constant.WeaponTypeSword2H:
		minStatMod = str*masteryMin*4.6 + dex
		maxStatMod = str*masteryMax*4.6 + dex

	case constant.WeaponTypeAxe1H, constant.WeaponTypeBW1H, constant.WeaponTypeWand2, constant.WeaponTypeStaff2:
		if isSwing {
			minStatMod = str*masteryMin*4.4 + dex
			maxStatMod = str*masteryMax*4.4 + dex
		} else {
			minStatMod = str*masteryMin*3.2 + dex
			maxStatMod = str*masteryMax*3.2 + dex
		}

	case constant.WeaponTypeSword1H, constant.WeaponTypeDagger2:
		if calc.player.job/100 == 4 && calc.weaponType == constant.WeaponTypeDagger2 {
			minStatMod = luk*masteryMin*3.6 + str + dex
			maxStatMod = luk*masteryMax*3.6 + str + dex
		} else {
			minStatMod = str*masteryMin*4.0 + dex
			maxStatMod = str*masteryMax*4.0 + dex
		}

	case constant.WeaponTypeClaw2:
		if skill.Skill(calc.skillID) == skill.LuckySeven {
			projectileWatk := float64(calc.GetProjectileWatk())
			totalWatk := watk + projectileWatk
			minStatMod = luk * 2.5
			maxStatMod = luk * 5.0

			minDmg := minStatMod * totalWatk / 100.0
			maxDmg := maxStatMod * totalWatk / 100.0
			return minDmg, maxDmg
		} else if calc.attackAction == constant.AttackActionProne ||
			(calc.attackAction >= constant.AttackActionSwing1H1 && calc.attackAction <= constant.AttackActionSwing2H7) {
			minStatMod = luk*0.1 + str + dex
			maxStatMod = luk*1.0 + str + dex

			minDmg := minStatMod * watk / 150.0
			maxDmg := maxStatMod * watk / 150.0
			return minDmg, maxDmg
		} else {
			minStatMod = luk*masteryMin*3.6 + str + dex
			maxStatMod = luk*masteryMax*3.6 + str + dex
		}

	default:
		if calc.weaponType == constant.WeaponTypeNone {
			level := float64(calc.player.level)
			bareHandsATT := math.Floor((2.0*level + 31.0) / 3.0)
			if bareHandsATT > 31 {
				bareHandsATT = 31
			}

			J := 3.0
			if calc.player.job >= 500 && calc.player.job < 600 {
				J = 4.2
			}

			minStatMod = str*J*0.1*0.9 + dex
			maxStatMod = str*J + dex

			minDmg := minStatMod * bareHandsATT / 100.0
			maxDmg := maxStatMod * bareHandsATT / 100.0
			return minDmg, maxDmg
		}
		return 0, 0
	}

	minDmg := minStatMod * watk * 0.01
	maxDmg := maxStatMod * watk * 0.01

	if int(calc.player.level) < int(mob.level) {
		levelPenalty := (100.0 - float64(int(mob.level)-int(calc.player.level))) / 100.0
		minDmg *= levelPenalty
		maxDmg *= levelPenalty
	}

	return minDmg, maxDmg
}

func (calc *DamageCalculator) CalculateMagicDamageRange() (float64, float64) {
	totalMAD := float64(calc.GetTotalMatk())
	intl := float64(calc.player.intt)
	luk := float64(calc.player.luk)

	if skill.Skill(calc.skillID) == skill.Heal {
		numTargets := float64(len(calc.data.attackInfo) + 1)
		targetMultiplier := 1.5 + 5.0/numTargets

		minDmg := (intl*0.3 + luk) * totalMAD / 1000.0 * targetMultiplier
		maxDmg := (intl*1.2 + luk) * totalMAD / 1000.0 * targetMultiplier

		return minDmg, maxDmg
	}

	minMAD := totalMAD * calc.masteryMod
	maxMAD := totalMAD

	minDmg := (intl*0.5 + totalMAD*0.058*totalMAD*0.058 + minMAD*3.3) * float64(calc.skill.Damage) * 0.01
	maxDmg := (intl*0.5 + totalMAD*0.058*totalMAD*0.058 + maxMAD*3.3) * float64(calc.skill.Damage) * 0.01

	return minDmg, maxDmg
}

func (calc *DamageCalculator) ApplySkillModifiers(minDmg, maxDmg float64, ampData *ElementAmpData, mob *monster) (float64, float64) {
	if calc.skill == nil {
		return minDmg, maxDmg
	}

	if calc.attackType == attackMagic {
		elemMod := float64(ampData.Magic) / 100.0
		minDmg *= elemMod
		maxDmg *= elemMod
	}

	return minDmg, maxDmg
}

func (calc *DamageCalculator) CalculateDefenseReductionBounds(mob *monster) (float64, float64) {
	if skill.Skill(calc.skillID) == skill.Sacrifice ||
		skill.Skill(calc.skillID) == skill.Assaulter {
		return 0, 0
	}

	var mobDef float64
	if calc.attackType == attackMagic {
		mobDef = float64(mob.mdDamage)
	} else {
		mobDef = float64(mob.pdDamage)
	}

	redMin := mobDef * 0.5
	redMax := mobDef * 0.6
	return redMin, redMax
}

func (calc *DamageCalculator) CheckCritical(roller *Roller) bool {
	if !calc.isRanged || calc.critSkill == nil {
		return false
	}

	if skill.Skill(calc.skillID) == skill.Blizzard {
		return false
	}

	roll := roller.Roll(constant.DamagePropModifier)
	return roll < float64(calc.critSkill.Prop)
}

func (calc *DamageCalculator) GetAfterModifier(targetIdx int, baseDmg float64) float64 {
	if calc.skill == nil {
		return 1.0
	}

	if calc.attackOption == constant.AttackOptionSlashBlastFA {
		return constant.SlashBlastFAModifiers[targetIdx]
	}

	if calc.skillID == int32(skill.ArrowBomb) {
		if targetIdx > 0 {
			return float64(calc.skill.X) * 0.01
		}
		if baseDmg > 0 {
			return 0.5
		}
		return 0
	}

	if calc.skillID == int32(skill.IronArrow) {
		return constant.IronArrowModifiers[targetIdx]
	}

	return 1.0
}

func (calc *DamageCalculator) GetIsMiss(roller *Roller, targetAccuracy float64, mob *monster) bool {
	roll := roller.Roll(constant.DamageStatModifier)

	var minModifier, maxModifier float64
	if calc.attackType == attackMagic {
		minModifier = 0.5
		maxModifier = 1.2
	} else {
		minModifier = 0.7
		maxModifier = 1.3
	}

	minTACC := targetAccuracy * minModifier
	randTACC := minTACC + (targetAccuracy*maxModifier-minTACC)*roll
	mobAvoid := float64(mob.eva)

	return randTACC < mobAvoid
}

func (calc *DamageCalculator) GetElementAmplification() *ElementAmpData {
	jobID := calc.player.job
	ampSkillID := int32(0)

	if jobID/10 == 21 {
		ampSkillID = int32(skill.ElementAmplification)
	} else if jobID/10 == 22 {
		ampSkillID = int32(skill.ILElementAmplification)
	}

	ampData := &ElementAmpData{Magic: 100, Mana: 100}
	if ampSkillID > 0 {
		if ampSkillInfo, ok := calc.player.skills[ampSkillID]; ok {
			skillData, err := nx.GetPlayerSkill(ampSkillID)
			if err == nil && len(skillData) > 0 && ampSkillInfo.Level > 0 {
				idx := int(ampSkillInfo.Level) - 1
				if idx < len(skillData) {
					ampData.Mana = int(skillData[idx].X)
					ampData.Magic = int(skillData[idx].Y)
				}
			}
		}
	}
	return ampData
}

func (calc *DamageCalculator) GetTargetAccuracy(mob *monster) float64 {
	levelDiff := int(mob.level) - int(calc.player.level)
	if levelDiff < 0 {
		levelDiff = 0
	}

	var accuracy int
	if calc.attackType == attackMagic {
		accuracy = int(5 * (calc.player.intt/10 + calc.player.luk/10))
	} else {
		accuracy = int(calc.player.dex)
	}

	return float64(accuracy*100) / (float64(levelDiff*10) + 255.0)
}

func (calc *DamageCalculator) GetMasteryModifier() float64 {
	var mastery int
	if calc.attackType == attackMagic {
		if calc.skill != nil {
			mastery = int(calc.skill.Mastery)
		}
	} else {
		mastery = calc.GetWeaponMastery()
	}
	return (float64(mastery)*5.0 + 10.0) * 0.009000000000000001
}

func (calc *DamageCalculator) GetWeaponMastery() int {
	switch calc.weaponType {
	case constant.WeaponTypeBow2, constant.WeaponTypeCrossbow2, constant.WeaponTypeClaw2:
		if !calc.isRanged {
			return 0
		}
	default:
		if calc.isRanged {
			return 0
		}
	}

	var skillID int32
	switch calc.weaponType {
	case constant.WeaponTypeSword1H, constant.WeaponTypeSword2H:
		if calc.player.job/10 == 11 {
			skillID = int32(skill.SwordMastery)
		} else {
			skillID = int32(skill.PageSwordMastery)
		}
	case constant.WeaponTypeAxe1H, constant.WeaponTypeAxe2H:
		skillID = int32(skill.AxeMastery)
	case constant.WeaponTypeBW1H, constant.WeaponTypeBW2H:
		skillID = int32(skill.BwMastery)
	case constant.WeaponTypeDagger2:
		skillID = int32(skill.DaggerMastery)
	case constant.WeaponTypeSpear2:
		skillID = int32(skill.SpearMastery)
	case constant.WeaponTypePolearm2:
		skillID = int32(skill.PolearmMastery)
	case constant.WeaponTypeBow2:
		skillID = int32(skill.BowMastery)
	case constant.WeaponTypeCrossbow2:
		skillID = int32(skill.CrossbowMastery)
	case constant.WeaponTypeClaw2:
		skillID = int32(skill.ClawMastery)
	default:
		return 0
	}

	if skillID != 0 {
		if skillInfo, ok := calc.player.skills[skillID]; ok {
			if skillData, err := nx.GetPlayerSkill(skillID); err == nil && len(skillData) > 0 {
				if skillInfo.Level > 0 && int(skillInfo.Level) <= len(skillData) {
					return int(skillData[skillInfo.Level-1].Mastery)
				}
			}
		}
	}
	return 0
}

func (calc *DamageCalculator) GetCritSkill() (byte, *nx.PlayerSkill) {
	if !calc.isRanged {
		return 0, nil
	}

	var skillID int32
	switch calc.weaponType {
	case constant.WeaponTypeBow2, constant.WeaponTypeCrossbow2:
		skillID = int32(skill.CriticalShot)
	case constant.WeaponTypeClaw2:
		skillID = int32(skill.CriticalThrow)
	default:
		return 0, nil
	}

	if skillInfo, ok := calc.player.skills[skillID]; ok {
		if skillData, err := nx.GetPlayerSkill(skillID); err == nil && len(skillData) > 0 {
			if skillInfo.Level > 0 && int(skillInfo.Level) <= len(skillData) {
				return skillInfo.Level, &skillData[skillInfo.Level-1]
			}
		}
	}
	return 0, nil
}

func (calc *DamageCalculator) GetTotalWatk() int16 {
	return calc.player.totalWatk
}

func (calc *DamageCalculator) GetTotalMatk() int16 {
	return calc.player.totalMatk
}

func (calc *DamageCalculator) GetTotalAccuracy() int16 {
	return calc.player.totalAccuracy
}

func (calc *DamageCalculator) GetProjectileWatk() int16 {
	if calc.projectileID == 0 {
		return 0
	}

	for _, item := range calc.player.use {
		if item.ID == calc.projectileID {
			return item.watk
		}
	}

	return 0
}

func (calc *DamageCalculator) GetTotalStr() int16 {
	return calc.player.totalStr
}

func (calc *DamageCalculator) GetTotalDex() int16 {
	return calc.player.totalDex
}

func (calc *DamageCalculator) GetTotalInt() int16 {
	return calc.player.totalInt
}

func (calc *DamageCalculator) GetTotalLuk() int16 {
	return calc.player.totalLuk
}
