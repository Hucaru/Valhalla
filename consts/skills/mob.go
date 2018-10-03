package skills

var Mob mob

type mob struct {
	WeaponAttackUp      byte
	WeaponAttackUpAoe   byte
	MagicAttackUp       byte
	MagicAttackUpAoe    byte
	WeaponDefenceUp     byte
	WeaponDefenceUpAoe  byte
	MagicDefenceUp      byte
	MagicDefenceUpAoe   byte
	HealAoe             byte
	speedUpAoe          byte
	Seal                byte
	Darkness            byte
	Weakness            byte
	Stun                byte
	Curse               byte
	Poison              byte
	Slow                byte
	Dispel              byte
	Seduce              byte
	SendToTown          byte
	PoisonMist          byte
	CrazySkull          byte
	Zombify             byte
	WeaponImmunity      byte
	MagicImmunity       byte
	ArmorSkill          byte
	WeaponDamageReflect byte
	MagicDamageReflect  byte
	AnyDamageReflect    byte
	McWeaponAttackUp    byte
	McMagicAttackUp     byte
	McWeaponDefenseUp   byte
	McMagicDefenseUp    byte
	McAccuracyUp        byte
	McAvoidUp           byte
	McSpeedUp           byte
	McSeal              byte // Not actually used in Monster Carnival
	Summon              byte
}

func init() {
	Mob.WeaponAttackUp = 100
	Mob.WeaponAttackUpAoe = 110
	Mob.MagicAttackUp = 101
	Mob.MagicAttackUpAoe = 111
	Mob.WeaponDefenceUp = 102
	Mob.WeaponDefenceUpAoe = 112
	Mob.MagicDefenceUp = 103
	Mob.MagicDefenceUpAoe = 113
	Mob.HealAoe = 114
	Mob.speedUpAoe = 115
	Mob.Seal = 120
	Mob.Darkness = 121
	Mob.Weakness = 122
	Mob.Stun = 123
	Mob.Curse = 124
	Mob.Poison = 125
	Mob.Slow = 126
	Mob.Dispel = 127
	Mob.Seduce = 128
	Mob.SendToTown = 129
	Mob.PoisonMist = 131
	Mob.CrazySkull = 132
	Mob.Zombify = 133
	Mob.WeaponImmunity = 140
	Mob.MagicImmunity = 141
	Mob.ArmorSkill = 142
	Mob.WeaponDamageReflect = 143
	Mob.MagicDamageReflect = 144
	Mob.AnyDamageReflect = 145
	Mob.McWeaponAttackUp = 150
	Mob.McMagicAttackUp = 151
	Mob.McWeaponDefenseUp = 152
	Mob.McMagicDefenseUp = 153
	Mob.McAccuracyUp = 154
	Mob.McAvoidUp = 155
	Mob.McSpeedUp = 156
	Mob.McSeal = 157 // Not actually used in Monster Carnival
	Mob.Summon = 200
}
