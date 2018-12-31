package status

type mob struct {
	Watk                int64
	Wdef                int64
	Matk                int64
	Mdef                int64
	Acc                 int64
	Avoid               int64
	Speed               int64
	Stun                int64
	Freeze              int64
	Poison              int64
	Seal                int64
	NoClue1             int64
	WeaponAttackUp      int64
	WeaponDefenseUp     int64
	MagicAttackUp       int64
	MagicDefenseUp      int64
	Doom                int64
	ShadowWeb           int64
	WeaponImmunity      int64
	MagicImmunity       int64
	NoClue2             int64
	NoClue3             int64
	NinjaAmbush         int64
	NoClue4             int64
	VenomousWeapon      int64
	NoClue5             int64
	NoClue6             int64
	Empty               int64 // All mobs have this when they spawn
	Hypnotize           int64
	WeaponDamageReflect int64
	MagicDamageReflect  int64
	NoClue7             int64 // Last bit you can use with 4 bytes
}

func (m *mob) populate() {
	m.Watk = 0x01
	m.Wdef = 0x02
	m.Matk = 0x04
	m.Mdef = 0x08
	m.Acc = 0x10
	m.Avoid = 0x20
	m.Speed = 0x40
	m.Stun = 0x80
	m.Freeze = 0x100
	m.Poison = 0x200
	m.Seal = 0x400
	m.NoClue1 = 0x800
	m.WeaponAttackUp = 0x1000
	m.WeaponDefenseUp = 0x2000
	m.MagicAttackUp = 0x4000
	m.MagicDefenseUp = 0x8000
	m.Doom = 0x10000
	m.ShadowWeb = 0x20000
	m.WeaponImmunity = 0x40000
	m.MagicImmunity = 0x80000
	m.NoClue2 = 0x100000
	m.NoClue3 = 0x200000
	m.NinjaAmbush = 0x400000
	m.NoClue4 = 0x800000
	m.VenomousWeapon = 0x1000000
	m.NoClue5 = 0x2000000
	m.NoClue6 = 0x4000000
	m.Empty = 0x8000000 // All mobs have this when they spawn
	m.Hypnotize = 0x10000000
	m.WeaponDamageReflect = 0x20000000
	m.MagicDamageReflect = 0x40000000
	m.NoClue7 = 0x80000000 // Last bit you can use with 4 bytes
}
