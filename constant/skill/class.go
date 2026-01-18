package skill

type Skill int

type swordsman struct {
	JobID                 int
	ImprovedMaxHpIncrease Skill
	Endure                Skill
	IronBody              Skill
}

type fighter struct {
	JobID        int
	AxeBooster   Skill
	AxeMastery   Skill
	PowerGuard   Skill
	Rage         Skill
	SwordBooster Skill
	SwordMastery Skill
}

type crusader struct {
	JobID              int
	ImprovedMpRecovery Skill
	ArmorCrash         Skill
	AxeComa            Skill
	AxePanic           Skill
	ComboAttack        Skill
	Shout              Skill
	SwordComa          Skill
	SwordPanic         Skill
}

type page struct {
	JobID        int
	BwBooster    Skill
	BwMastery    Skill
	PowerGuard   Skill
	SwordBooster Skill
	SwordMastery Skill
	Threaten     Skill
}

type whiteknight struct {
	JobID              int
	ImprovedMpRecovery Skill
	BwFireCharge       Skill
	BwIceCharge        Skill
	BwLitCharge        Skill
	ChargeBlow         Skill
	MagicCrash         Skill
	SwordFireCharge    Skill
	SwordIceCharge     Skill
	SwordLitCharge     Skill
}

type spearman struct {
	JobID          int
	HyperBody      Skill
	IronWill       Skill
	PolearmBooster Skill
	PolearmMastery Skill
	SpearBooster   Skill
	SpearMastery   Skill
}

type dragonknight struct {
	JobID               int
	DragonBlood         Skill
	DragonRoar          Skill
	ElementalResistance Skill
	PowerCrash          Skill
	Sacrifice           Skill
}

type magician struct {
	JobID                 int
	ImprovedMpRecovery    Skill
	ImprovedMaxMpIncrease Skill
	MagicArmor            Skill
	MagicGuard            Skill
	MagicClaw             Skill
	EnergyBolt            Skill
}

type fpwizard struct {
	JobID        int
	Meditation   Skill
	MpEater      Skill
	PoisonBreath Skill
	FireArrow    Skill
	Slow         Skill
}

type fpmage struct {
	JobID                int
	ElementAmplification Skill
	ElementComposition   Skill
	PartialResistance    Skill
	PoisonMyst           Skill
	Seal                 Skill
	SpellBooster         Skill
	Explosion            Skill
}

type ilwizard struct {
	JobID       int
	ColdBeam    Skill
	Meditation  Skill
	MpEater     Skill
	Slow        Skill
	ThunderBolt Skill
}

type ilmage struct {
	JobID                int
	ElementAmplification Skill
	ElementComposition   Skill
	IceStrike            Skill
	PartialResistance    Skill
	Seal                 Skill
	SpellBooster         Skill
	Lightning            Skill
}

type cleric struct {
	JobID      int
	Bless      Skill
	Heal       Skill
	Invincible Skill
	MpEater    Skill
	HolyArrow  Skill
}

type priest struct {
	JobID               int
	Dispel              Skill
	Doom                Skill
	ElementalResistance Skill
	HolySymbol          Skill
	MysticDoor          Skill
	SummonDragon        Skill
	Resurrection        Skill
}

type archer struct {
	JobID            int
	BlessingOfAmazon Skill
	CriticalShot     Skill
	Focus            Skill
}

type hunter struct {
	JobID          int
	PowerKnockback Skill
	ArrowBomb      Skill
	BowBooster     Skill
	BowMastery     Skill
	SoulArrow      Skill
}

type ranger struct {
	JobID      int
	MortalBlow Skill
	Puppet     Skill
	SilverHawk Skill
	Inferno    Skill
}

type crossbowman struct {
	JobID           int
	PowerKnockback  Skill
	CrossbowBooster Skill
	CrossbowMastery Skill
	SoulArrow       Skill
	IronArrow       Skill
}

type sniper struct {
	JobID       int
	Blizzard    Skill
	GoldenEagle Skill
	MortalBlow  Skill
	Puppet      Skill
}

type rogue struct {
	JobID      int
	NimbleBody Skill
	DarkSight  Skill
	Disorder   Skill
	DoubleStab Skill
	LuckySeven Skill
}

type assassin struct {
	JobID         int
	ClawBooster   Skill
	ClawMastery   Skill
	CriticalThrow Skill
	Endure        Skill
	Drain         Skill
	Haste         Skill
}

type hermit struct {
	JobID         int
	Alchemist     Skill
	Avenger       Skill
	MesoUp        Skill
	ShadowMeso    Skill
	ShadowPartner Skill
	ShadowWeb     Skill
}

type bandit struct {
	JobID         int
	DaggerBooster Skill
	DaggerMastery Skill
	Endure        Skill
	Haste         Skill
	SavageBlow    Skill
	Steal         Skill
}

type chiefbandit struct {
	JobID         int
	Assaulter     Skill
	BandOfThieves Skill
	Chakra        Skill
	MesoExplosion Skill
	MesoGuard     Skill
	Pickpocket    Skill
}

type gm struct {
	JobID           int
	GMSelfHaste     Skill
	SuperDragonRoar Skill
	GMTeleport      Skill
}

type supergm struct {
	JobID               int
	SuperGMHealDispell  Skill
	SuperGMHaste        Skill
	SuperGMHolySymbol   Skill
	SuperGMBless        Skill
	SuperGMHide         Skill
	SuperGMResurrection Skill
}

var (
	Swordsman    swordsman
	Fighter      fighter
	Crusader     crusader
	Page         page
	WhiteKnight  whiteknight
	Spearman     spearman
	DragonKnight dragonknight
	Magician     magician
	FPWizard     fpwizard
	FPMage       fpmage
	ILWizard     ilwizard
	ILMage       ilmage
	Cleric       cleric
	Priest       priest
	Archer       archer
	Hunter       hunter
	Ranger       ranger
	Crossbowman  crossbowman
	Sniper       sniper
	Rogue        rogue
	Assassin     assassin
	Hermit       hermit
	Bandit       bandit
	ChiefBandit  chiefbandit
	GM           gm
	SuperGM      supergm
)

const (
	//Beginner Skills - 0
	Recovery   Skill = 1001
	NimbleFeet Skill = 1002

	//Swordsman Skills - 100
	ImprovedMaxHpIncrease Skill = 1000001
	Endure                Skill = 1000002
	IronBody              Skill = 1001003

	//Fighter Skills - 110
	AxeBooster   Skill = 1101005
	AxeMastery   Skill = 1100001
	PowerGuard   Skill = 1101007
	Rage         Skill = 1101006
	SwordBooster Skill = 1101004
	SwordMastery Skill = 1100000

	//Crusader Skills - 111
	ImprovedMpRecovery Skill = 1110000
	ArmorCrash         Skill = 1111007
	AxeComa            Skill = 1111006
	AxePanic           Skill = 1111004
	ComboAttack        Skill = 1111002
	Shout              Skill = 1111008
	SwordComa          Skill = 1111005
	SwordPanic         Skill = 1111003

	//Page Skills - 120
	BwBooster        Skill = 1201005
	BwMastery        Skill = 1200001
	PagePowerGuard   Skill = 1201007
	PageSwordBooster Skill = 1201004
	PageSwordMastery Skill = 1200000
	Threaten         Skill = 1201006

	//WhiteKnight Skills - 121
	WKImprovedMpRecovery Skill = 1210000
	BwFireCharge         Skill = 1211004
	BwIceCharge          Skill = 1211006
	BwLitCharge          Skill = 1211008
	ChargeBlow           Skill = 1211002
	MagicCrash           Skill = 1211009
	SwordFireCharge      Skill = 1211003
	SwordIceCharge       Skill = 1211005
	SwordLitCharge       Skill = 1211007

	//Spearman Skills - 130
	HyperBody      Skill = 1301007
	IronWill       Skill = 1301006
	PolearmBooster Skill = 1301005
	PolearmMastery Skill = 1300001
	SpearBooster   Skill = 1301004
	SpearMastery   Skill = 1300000

	//DragonKnight Skills - 131
	DragonBlood         Skill = 1311008
	DragonRoar          Skill = 1311006
	ElementalResistance Skill = 1310000
	PowerCrash          Skill = 1311007
	Sacrifice           Skill = 1311005

	//Magician Skills - 200
	MagImprovedMpRecovery Skill = 2000000
	ImprovedMaxMpIncrease Skill = 2000001
	MagicArmor            Skill = 2001003
	MagicGuard            Skill = 2001002
	MagicClaw             Skill = 2001005
	EnergyBolt            Skill = 2001004

	//FPWizard Skills - 210
	Meditation   Skill = 2101001
	MpEater      Skill = 2100000
	PoisonBreath Skill = 2101005
	FireArrow    Skill = 2101004
	Slow         Skill = 2101003

	//FPMage Skills - 211
	ElementAmplification Skill = 2110001
	ElementComposition   Skill = 2111006
	PartialResistance    Skill = 2110000
	PoisonMyst           Skill = 2111003
	Seal                 Skill = 2111004
	SpellBooster         Skill = 2111005
	Explosion            Skill = 2111002

	//ILWizard Skills - 220
	ColdBeam     Skill = 2201004
	ILMeditation Skill = 2201001
	ILMpEater    Skill = 2200000
	ILSlow       Skill = 2201003
	ThunderBolt  Skill = 2201005

	//ILMage Skills - 221
	ILElementAmplification Skill = 2210001
	ILElementComposition   Skill = 2211006
	IceStrike              Skill = 2211002
	ILPartialResistance    Skill = 2210000
	ILSeal                 Skill = 2211004
	ILSpellBooster         Skill = 2211005
	Lightning              Skill = 2211003

	//Cleric Skills - 230
	Bless         Skill = 2301004
	Heal          Skill = 2301002
	Invincible    Skill = 2301003
	ClericMpEater Skill = 2300000
	HolyArrow     Skill = 2301005

	//Priest Skills - 231
	Dispel                    Skill = 2311001
	Doom                      Skill = 2311005
	PriestElementalResistance Skill = 2310000
	HolySymbol                Skill = 2311003
	MysticDoor                Skill = 2311002
	SummonDragon              Skill = 2311006
	Resurrection              Skill = 2311004

	//Archer Skills - 300
	BlessingOfAmazon Skill = 3000000
	CriticalShot     Skill = 3000001
	Focus            Skill = 3001003

	//Hunter Skills - 310
	PowerKnockback Skill = 3101003
	ArrowBomb      Skill = 3101005
	BowBooster     Skill = 3101002
	BowMastery     Skill = 3100000
	SoulArrow      Skill = 3101004

	//Ranger Skills - 311
	MortalBlow Skill = 3110001
	Puppet     Skill = 3111002
	SilverHawk Skill = 3111005
	Inferno    Skill = 3111004

	//Crossbowman Skills - 320
	CBPowerKnockback Skill = 3201003
	IronArrow        Skill = 3201005
	CrossbowBooster  Skill = 3201002
	CrossbowMastery  Skill = 3200000
	CBSoulArrow      Skill = 3201004

	//Sniper Skills - 321
	Blizzard         Skill = 3211003
	GoldenEagle      Skill = 3211005
	SniperMortalBlow Skill = 3210001
	SniperPuppet     Skill = 3211002

	//Rogue Skills - 400
	NimbleBody Skill = 4000000
	DarkSight  Skill = 4001003
	Disorder   Skill = 4001002
	DoubleStab Skill = 4001334
	LuckySeven Skill = 4001344

	//Assassin Skills - 410
	ClawBooster    Skill = 4101003
	ClawMastery    Skill = 4100000
	CriticalThrow  Skill = 4100001
	AssassinEndure Skill = 4100002
	Drain          Skill = 4101005
	Haste          Skill = 4101004

	//Hermit Skills - 411
	Alchemist     Skill = 4110000
	Avenger       Skill = 4111005
	MesoUp        Skill = 4111001
	ShadowMeso    Skill = 4111004
	ShadowPartner Skill = 4111002
	ShadowWeb     Skill = 4111003

	//Bandit Skills - 420
	DaggerBooster Skill = 4201002
	DaggerMastery Skill = 4200000
	BanditEndure  Skill = 4200001
	BanditHaste   Skill = 4201003
	SavageBlow    Skill = 4201005
	Steal         Skill = 4201004

	//ChiefBandit Skills - 421
	Assaulter     Skill = 4211002
	BandOfThieves Skill = 4211004
	Chakra        Skill = 4211001
	MesoExplosion Skill = 4211006
	MesoGuard     Skill = 4211005
	Pickpocket    Skill = 4211003

	//GM Skills - 500
	GMSelfHaste     Skill = 5001000
	SuperDragonRoar Skill = 5001001
	GMTeleport      Skill = 5001002

	//SuperGM SKills - 510
	SuperGMHealDispell  Skill = 5101000
	SuperGMHaste        Skill = 5101001
	SuperGMHolySymbol   Skill = 5101002
	SuperGMBless        Skill = 5101003
	SuperGMHide         Skill = 5101004
	SuperGMResurrection Skill = 5101005
)

const (
	SwordsmanJobID    = 100
	FighterJobID      = 110
	CrusaderJobID     = 111
	PageJobID         = 120
	WhiteKnightJobID  = 121
	SpearmanJobID     = 130
	DragonKnightJobID = 131

	MagicianJobID         = 200
	FirePoisonWizardJobID = 210
	FirePoisonMageJobID   = 211
	IceLightWizardJobID   = 220
	IceLightMageJobID     = 221
	ClericJobID           = 230
	PriestJobID           = 231

	ArcherJobID      = 300
	HunterJobID      = 310
	RangerJobID      = 311
	CrossbowmanJobID = 320
	SniperJobID      = 321

	RogueJobID       = 400
	AssassinJobID    = 410
	HermitJobID      = 411
	BanditJobID      = 420
	ChiefBanditJobID = 421

	GmJobID      = 500
	SuperGmJobID = 510
)

func init() {
	Swordsman = swordsman{
		JobID:                 SwordsmanJobID,
		ImprovedMaxHpIncrease: ImprovedMaxHpIncrease,
		Endure:                Endure,
		IronBody:              IronBody,
	}

	Fighter = fighter{
		JobID:        FighterJobID,
		AxeBooster:   AxeBooster,
		AxeMastery:   AxeMastery,
		PowerGuard:   PowerGuard,
		Rage:         Rage,
		SwordBooster: SwordBooster,
		SwordMastery: SwordMastery,
	}

	Crusader = crusader{
		JobID:              CrusaderJobID,
		ImprovedMpRecovery: ImprovedMpRecovery,
		ArmorCrash:         ArmorCrash,
		AxeComa:            AxeComa,
		AxePanic:           AxePanic,
		ComboAttack:        ComboAttack,
		Shout:              Shout,
		SwordComa:          SwordComa,
		SwordPanic:         SwordPanic,
	}

	Page = page{
		JobID:        PageJobID,
		BwBooster:    BwBooster,
		BwMastery:    BwMastery,
		PowerGuard:   PagePowerGuard,
		SwordBooster: PageSwordBooster,
		SwordMastery: PageSwordMastery,
		Threaten:     Threaten,
	}

	WhiteKnight = whiteknight{
		JobID:              WhiteKnightJobID,
		ImprovedMpRecovery: WKImprovedMpRecovery,
		BwFireCharge:       BwFireCharge,
		BwIceCharge:        BwIceCharge,
		BwLitCharge:        BwLitCharge,
		ChargeBlow:         ChargeBlow,
		MagicCrash:         MagicCrash,
		SwordFireCharge:    SwordFireCharge,
		SwordIceCharge:     SwordIceCharge,
		SwordLitCharge:     SwordLitCharge,
	}

	Spearman = spearman{
		JobID:          SpearmanJobID,
		HyperBody:      HyperBody,
		IronWill:       IronWill,
		PolearmBooster: PolearmBooster,
		PolearmMastery: PolearmMastery,
		SpearBooster:   SpearBooster,
		SpearMastery:   SpearMastery,
	}

	DragonKnight = dragonknight{
		JobID:               DragonKnightJobID,
		DragonBlood:         DragonBlood,
		DragonRoar:          DragonRoar,
		ElementalResistance: ElementalResistance,
		PowerCrash:          PowerCrash,
		Sacrifice:           Sacrifice,
	}

	Magician = magician{
		JobID:                 MagicianJobID,
		ImprovedMpRecovery:    ImprovedMpRecovery,
		ImprovedMaxMpIncrease: ImprovedMaxMpIncrease,
		MagicArmor:            MagicArmor,
		MagicGuard:            MagicGuard,
		MagicClaw:             MagicClaw,
		EnergyBolt:            EnergyBolt,
	}

	FPWizard = fpwizard{
		JobID:        FirePoisonWizardJobID,
		Meditation:   Meditation,
		MpEater:      MpEater,
		PoisonBreath: PoisonBreath,
		FireArrow:    FireArrow,
		Slow:         Slow,
	}

	FPMage = fpmage{
		JobID:                FirePoisonMageJobID,
		ElementAmplification: ElementAmplification,
		ElementComposition:   ElementComposition,
		PartialResistance:    PartialResistance,
		PoisonMyst:           PoisonMyst,
		Seal:                 Seal,
		SpellBooster:         SpellBooster,
		Explosion:            Explosion,
	}

	ILWizard = ilwizard{
		JobID:       IceLightWizardJobID,
		ColdBeam:    ColdBeam,
		Meditation:  ILMeditation,
		MpEater:     ILMpEater,
		Slow:        ILSlow,
		ThunderBolt: ThunderBolt,
	}

	ILMage = ilmage{
		JobID:                IceLightMageJobID,
		ElementAmplification: ILElementAmplification,
		ElementComposition:   ILElementComposition,
		IceStrike:            IceStrike,
		PartialResistance:    ILPartialResistance,
		Seal:                 ILSeal,
		SpellBooster:         ILSpellBooster,
		Lightning:            Lightning,
	}

	Cleric = cleric{
		JobID:      ClericJobID,
		Bless:      Bless,
		Heal:       Heal,
		Invincible: Invincible,
		MpEater:    ClericMpEater,
		HolyArrow:  HolyArrow,
	}

	Priest = priest{
		JobID:               PriestJobID,
		Dispel:              Dispel,
		Doom:                Doom,
		ElementalResistance: PriestElementalResistance,
		HolySymbol:          HolySymbol,
		MysticDoor:          MysticDoor,
		SummonDragon:        SummonDragon,
		Resurrection:        Resurrection,
	}

	Archer = archer{
		JobID:            ArcherJobID,
		BlessingOfAmazon: BlessingOfAmazon,
		CriticalShot:     CriticalShot,
		Focus:            Focus,
	}

	Hunter = hunter{
		JobID:          HunterJobID,
		PowerKnockback: PowerKnockback,
		ArrowBomb:      ArrowBomb,
		BowBooster:     BowBooster,
		BowMastery:     BowMastery,
		SoulArrow:      SoulArrow,
	}

	Ranger = ranger{
		JobID:      RangerJobID,
		MortalBlow: MortalBlow,
		Puppet:     Puppet,
		SilverHawk: SilverHawk,
		Inferno:    Inferno,
	}

	Crossbowman = crossbowman{
		JobID:           CrossbowmanJobID,
		PowerKnockback:  CBPowerKnockback,
		CrossbowBooster: CrossbowBooster,
		CrossbowMastery: CrossbowMastery,
		SoulArrow:       CBSoulArrow,
		IronArrow:       IronArrow,
	}

	Sniper = sniper{
		JobID:       SniperJobID,
		Blizzard:    Blizzard,
		GoldenEagle: GoldenEagle,
		MortalBlow:  SniperMortalBlow,
		Puppet:      SniperPuppet,
	}

	Rogue = rogue{
		JobID:      RogueJobID,
		NimbleBody: NimbleBody,
		DarkSight:  DarkSight,
		Disorder:   Disorder,
		DoubleStab: DoubleStab,
		LuckySeven: LuckySeven,
	}

	Assassin = assassin{
		JobID:         AssassinJobID,
		ClawBooster:   ClawBooster,
		ClawMastery:   ClawMastery,
		CriticalThrow: CriticalThrow,
		Endure:        AssassinEndure,
		Drain:         Drain,
		Haste:         Haste,
	}

	Hermit = hermit{
		JobID:         HermitJobID,
		Alchemist:     Alchemist,
		Avenger:       Avenger,
		MesoUp:        MesoUp,
		ShadowMeso:    ShadowMeso,
		ShadowPartner: ShadowPartner,
		ShadowWeb:     ShadowWeb,
	}

	Bandit = bandit{
		JobID:         BanditJobID,
		DaggerBooster: DaggerBooster,
		DaggerMastery: DaggerMastery,
		Endure:        BanditEndure,
		Haste:         BanditHaste,
		SavageBlow:    SavageBlow,
		Steal:         Steal,
	}

	ChiefBandit = chiefbandit{
		JobID:         ChiefBanditJobID,
		Assaulter:     Assaulter,
		BandOfThieves: BandOfThieves,
		Chakra:        Chakra,
		MesoExplosion: MesoExplosion,
		MesoGuard:     MesoGuard,
		Pickpocket:    Pickpocket,
	}

	GM = gm{
		JobID:           GmJobID,
		GMSelfHaste:     GMSelfHaste,
		SuperDragonRoar: SuperDragonRoar,
		GMTeleport:      GMTeleport,
	}

	SuperGM = supergm{
		JobID:               SuperGmJobID,
		SuperGMHealDispell:  SuperGMHealDispell,
		SuperGMHaste:        SuperGMHaste,
		SuperGMHolySymbol:   SuperGMHolySymbol,
		SuperGMBless:        SuperGMBless,
		SuperGMHide:         SuperGMHide,
		SuperGMResurrection: SuperGMResurrection,
	}
}
