package skill

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

var MobStat mobStat

type mobStat struct {
	PhysicalDamage  int32
	PhysicalDefense int32
	MagicDamage     int32
	MagicDefense    int32
	Accurrency      int32
	Evasion         int32
	Speed           int32
	Stun            int32
	Freeze          int32
	Poison          int32
	Seal            int32
	Darkness        int32
	PowerUp         int32
	MagicUp         int32
	PowerGuardUp    int32
	MagicGuardUp    int32
	Doom            int32
	Web             int32
	PhysicalImmune  int32
	MagicImmune     int32
	HardSkin        int32
	Ambush          int32
	Venom           int32
	Blind           int32
	SealSkill       int32
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

	MobStat.PhysicalDamage = 0x1
	MobStat.PhysicalDefense = 0x2
	MobStat.MagicDamage = 0x4
	MobStat.MagicDefense = 0x8
	MobStat.Accurrency = 0x10
	MobStat.Evasion = 0x20
	MobStat.Speed = 0x40
	MobStat.Stun = 0x80
	MobStat.Freeze = 0x100
	MobStat.Poison = 0x200
	MobStat.Seal = 0x400
	MobStat.Darkness = 0x800
	MobStat.PowerUp = 0x1000
	MobStat.MagicUp = 0x2000
	MobStat.PowerGuardUp = 0x4000
	MobStat.MagicGuardUp = 0x8000
	MobStat.Doom = 0x10000
	MobStat.Web = 0x20000
	MobStat.PhysicalImmune = 0x40000
	MobStat.MagicImmune = 0x80000
	MobStat.HardSkin = 0x200000
	MobStat.Ambush = 0x400000
	MobStat.Venom = 0x1000000
	MobStat.Blind = 0x2000000
	MobStat.SealSkill = 0x4000000
}
