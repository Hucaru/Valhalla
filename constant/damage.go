package constant

// AttackAction represents different attack animations
type AttackAction int

const (
	AttackActionSwing1H1 AttackAction = 0x05
	AttackActionSwing1H2 AttackAction = 0x06
	AttackActionSwing1H3 AttackAction = 0x07
	AttackActionSwing1H4 AttackAction = 0x08

	AttackActionSwing2H1 AttackAction = 0x09
	AttackActionSwing2H2 AttackAction = 0x0A
	AttackActionSwing2H3 AttackAction = 0x0B
	AttackActionSwing2H4 AttackAction = 0x0C
	AttackActionSwing2H5 AttackAction = 0x0D
	AttackActionSwing2H6 AttackAction = 0x0E
	AttackActionSwing2H7 AttackAction = 0x0F

	AttackActionStab1 AttackAction = 0x10
	AttackActionStab2 AttackAction = 0x11
	AttackActionStab3 AttackAction = 0x12
	AttackActionStab4 AttackAction = 0x13
	AttackActionStab5 AttackAction = 0x14
	AttackActionStab6 AttackAction = 0x15

	AttackActionBullet1 AttackAction = 0x16
	AttackActionBullet2 AttackAction = 0x17
	AttackActionBullet3 AttackAction = 0x18
	AttackActionBullet4 AttackAction = 0x19
	AttackActionBullet5 AttackAction = 0x1A
	AttackActionBullet6 AttackAction = 0x1B

	AttackActionProne AttackAction = 0x20
	AttackActionHeal  AttackAction = 0x28
	AttackActionUnk35 AttackAction = 0x35
)

// AttackOption represents flags for special attack properties
type AttackOption byte

const (
	AttackOptionNormal          AttackOption = 0
	AttackOptionSlashBlastFA    AttackOption = 1
	AttackOptionMortalBlowProp  AttackOption = 4
	AttackOptionShadowPartner   AttackOption = 8
	AttackOptionMortalBlowMelee AttackOption = 16
)

// Damage calculation constants
const (
	DamageMaxHits    = 15
	DamageMaxTargets = 15
	DamageMaxPAD     = 999

	DamageRollsPerTarget = 7

	DamageVarianceTolerance = 0.15 // 15% tolerance

	DamageRngBufferSize = 7
	DamageRngModulo     = 10000000
	DamageRngNormalize  = 1e-7

	// Meso Explosion formula constants
	MesoExplosionMaxDamage              = 999999999 // Max damage cap to prevent false bans
	MesoExplosionLowMesoThreshold       = 1000.0
	MesoExplosionLowMesoMultiplier      = 0.82
	MesoExplosionLowMesoOffset          = 28.0
	MesoExplosionLowMesoDivisor         = 5300.0
	MesoExplosionHighMesoDivisorOffset  = 5250.0
	MesoExplosionDamageVarianceTolerance = 1.5
)

type WeaponType int

const (
	WeaponTypeNone      WeaponType = 0
	WeaponTypeSword1H   WeaponType = 30
	WeaponTypeAxe1H     WeaponType = 31
	WeaponTypeBW1H      WeaponType = 32
	WeaponTypeDagger2   WeaponType = 33
	WeaponTypeWand2     WeaponType = 37
	WeaponTypeStaff2    WeaponType = 38
	WeaponTypeSword2H   WeaponType = 40
	WeaponTypeAxe2H     WeaponType = 41
	WeaponTypeBW2H      WeaponType = 42
	WeaponTypeSpear2    WeaponType = 43
	WeaponTypePolearm2  WeaponType = 44
	WeaponTypeBow2      WeaponType = 45
	WeaponTypeCrossbow2 WeaponType = 46
	WeaponTypeClaw2     WeaponType = 47
)

// ElementModifier represents element resistance/weakness
type ElementModifier int

const (
	ElementModifierNormal     ElementModifier = 0
	ElementModifierNullify    ElementModifier = 1
	ElementModifierHalf       ElementModifier = 2
	ElementModifierOneAndHalf ElementModifier = 3
)

var (
	SlashBlastFAModifiers = [DamageMaxTargets]float64{
		0.666667,
		0.222222,
		0.074074,
		0.024691,
		0.008229999999999,
		0.002743,
		0.000914,
		0.000305,
		0.000102,
		0.000033,
		0.000011,
		0.000004,
		0.000001,
		0.0,
		0.0,
	}

	IronArrowModifiers = [DamageMaxTargets]float64{
		1.0,
		0.8,
		0.64,
		0.512,
		0.4096,
		0.32768,
		0.262144,
		0.209715,
		0.167772,
		0.134218,
		0.107374,
		0.085899,
		0.068719,
		0.054976,
		0.04398,
	}
)

func GetWeaponType(weaponID int32) WeaponType {
	if weaponID/1000000 != 1 {
		return WeaponTypeNone
	}
	weaponType := (weaponID / 10000) % 100

	if weaponType < 30 {
		return WeaponTypeNone
	}
	if weaponType > 33 && weaponType <= 36 {
		return WeaponTypeNone
	}
	if weaponType == 39 {
		return WeaponTypeNone
	}
	if weaponType > 47 {
		return WeaponTypeNone
	}

	return WeaponType(weaponType)
}
