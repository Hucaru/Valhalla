package constant

var WORLD_NAMES = [...]string{"Scania", "Bera", "Broa", "Windia", "Khaini", "Bellocan", "Mardia", "Kradia", "Yellonde", "Demethos", "Galicia", "El Nido", "Zenith", "Arcania", "Chaos", "Nova", "Renegates"}

// Generic Constants
const (
	MapleVersion          = 28
	ClientHeaderSize      = 4
	InterserverHeaderSize = 4
	OpcodeLength          = 1
)

const (
	MaxItemStack = 200

	SkinID  = 0x01
	FaceID  = 0x02 // Eyes
	HairID  = 0x04
	PetID   = 0x08
	LevelID = 0x10
	JobID   = 0x20
	StrID   = 0x40
	DexID   = 0x80
	IntID   = 0x100
	LukID   = 0x200
	HpID    = 0x400
	MaxHpID = 0x800
	MpID    = 0x1000
	MaxMpID = 0x2000
	ApID    = 0x4000
	SpID    = 0x8000
	ExpID   = 0x10000
	FameID  = 0x20000
	MesosID = 0x40000

	BeginnerHpAdd = int16(12)
	BeginnerMpAdd = int16(10)

	WarriorHpAdd = int16(24)
	WarriorMpAdd = int16(4)

	MagicianHpAdd = int16(10)
	MagicianMpAdd = int16(6)

	BowmanHpAdd = int16(20)
	BowmanMpAdd = int16(14)

	ThiefHpAdd = int16(20)
	ThiefMpAdd = int16(14)

	AdminHpAdd = 150
	AdminMpAdd = 150

	BeginnerJobID = 0

	WarriorJobID      = 100
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

	BowmanJobID      = 300
	HunterJobID      = 310
	RangerJobID      = 311
	CrossbowmanJobID = 320
	SniperJobID      = 321

	ThiefJobID       = 400
	AssassinJobID    = 410
	HermitJobID      = 411
	BanditJobID      = 420
	ChiefBanditJobID = 421

	GmJobID      = 500
	SuperGmJobID = 510

	MaxHpValue = 32767
	MaxMpValue = 32767

	MaxPartySize = 6
	MaxGuildSize = 255

	GuildCreateDialogue   byte = 0x02
	GuildInvite           byte = 0x05
	GuildAcceptInvite     byte = 0x06
	GuildLeave            byte = 0x07
	GuildExpel            byte = 0x08
	GuildNoticeChange     byte = 0x10
	GuildUpdateTitleNames byte = 0x0D
	GuildRankChange       byte = 0x0E
	GuildEmblemChange     byte = 0x0F
	GuildContractSign     byte = 0x1E
	GuildRejectInvite     byte = 0x37

	QuestLostItem  = 0x00
	QuestStarted   = 0x01
	QuestCompleted = 0x02
	QuestForfeit   = 0x03

	FameNotifySource  = 0x00
	FameIncorrectUser = 0x01
	FameUnderLevel    = 0x02
	FameThisDay       = 0x03
	FameThisMonth     = 0x04
	FameNotifyTarget  = 0x05

	SummonRemoveReasonCancel   = 0x01
	SummonRemoveReasonKeepBuff = 0x02
	SummonRemoveReasonReplaced = 0x04
	SummonAttackMob            = 0x06
	SummonTakeDamage           = 0xFF

	StorageEquipTab = 0x04
	StorageUseTab   = 0x08
	StorageSetupTab = 0x10
	StorageEtcTab   = 0x20
	StorageCashTab  = 0x40

	MessengerEnter        byte = 0x00
	MessengerEnterResult  byte = 0x01
	MessengerLeave        byte = 0x02
	MessengerInvite       byte = 0x03
	MessengerInviteResult byte = 0x04
	MessengerBlocked      byte = 0x05
	MessengerChat         byte = 0x06
	MessengerAvatar       byte = 0x07
	MessengerMigrated     byte = 0x08

	ReactorWarp      = 0
	ReactorSpawn     = 1
	ReactorDrop      = 2
	ReactorSpawnNPC  = 6
	ReactorRunScript = 10

	PlayerEffectLevelUp          = 0
	PlayerEffectSkillOnSelf      = 1
	PlayerEffectSkillOnOther     = 2
	PlayerEffectQuestEffect      = 3
	PlayerEffectInventoryChanged = 3
	PlayerEffectPet              = 4
	PlayerEffectExpCharm         = 6
	PlayerEffectPortal           = 7
	PlayerEffectJobChange        = 8

	PetRemoveNone   byte = 0
	PetRemoveHungry byte = 1
	PetRemoveExpire byte = 2
)

const (
	MiniRoomCreate        byte = 0
	MiniRoomInvite        byte = 2
	MiniRoomDeclineInvite byte = 3
	MiniRoomEnter         byte = 4
	MiniRoomEnterResult   byte = 5
	MiniRoomChat          byte = 6
	MiniRoomAvatar        byte = 9
	MiniRoomLeave         byte = 10
	MiniRoomOpen          byte = 11

	MiniRoomTradePutItem  byte = 13
	MiniRoomTradePutMesos byte = 14
	MiniRoomTradeAccept   byte = 15

	MiniRoomAddShopItem          byte = 18
	MiniRoomBuyShopItem          byte = 19
	MiniRoomPlayerShopItemResult byte = 0x14
	MiniRoomPlayerShopSoldItem   byte = 0x16
	MiniRoomMoveItemShopToInv    byte = 23
)

const (
	MiniRoomTypeNone          byte = 0
	MiniRoomTypeOmok          byte = 1
	MiniRoomTypeMatchCards    byte = 2
	MiniRoomTypeTrade         byte = 3
	MiniRoomTypePlayerShop    byte = 4
	MiniRoomTypeEntrustedShop byte = 5
)

const (
	MiniRoomEnterRoomAlreadyClosed     byte = 0x01
	MiniRoomEnterFullCapacity          byte = 0x02
	MiniRoomEnterOtherRequests         byte = 0x03
	MiniRoomEnterCantWhileDead         byte = 0x04
	MiniRoomEnterCantInMiddleEvent     byte = 0x05
	MiniRoomEnterUnableToDoIt          byte = 0x06
	MiniRoomEnterOtherItemsAtPoint     byte = 0x07
	MiniRoomEnterCantEstablishRoom     byte = 0x0A
	MiniRoomEnterTradeOnSameMap        byte = 0x09
	MiniRoomEnterNotEnoughMesos        byte = 0x0F
	MiniRoomEnterCantStartGameHere     byte = 0x0B
	MiniRoomEnterBuiltAtMainTown       byte = 0x0C
	MiniRoomEnterUnableEnterTournament byte = 0x0D
	MiniRoomEnterIncorrectPassword     byte = 0x10
)

const (
	PlayerShopNotEnoughInStock       byte = 1
	PlayerShopNotEnoughMesos         byte = 2
	PlayerShopPriceTooHighForTrade   byte = 3
	PlayerShopBuyerNotEnoughMoney    byte = 4
	PlayerShopCannotCarryMoreThanOne byte = 5
	PlayerShopInventoryFull          byte = 6
)

const (
	MiniRoomLeaveReason          byte = 0
	MiniRoomCantEstablish        byte = 1
	MiniRoomCancel               byte = 2
	MiniRoomClosed               byte = 3
	MiniRoomExpelled             byte = 4
	MiniRoomForcedLeave          byte = 5
	MiniRoomTradeSuccess         byte = 6
	MiniRoomTradeFail            byte = 7
	MiniRoomTradeInventoryFull   byte = 8
	MiniRoomTradeWrongMap        byte = 9
	MiniRoomPlayerShopOutOfStock byte = 10
)

const (
	GameWin     byte = 0
	GameTie     byte = 1
	GameForfeit byte = 2
)

const (
	MatchCardsSizeSmall  byte = 0
	MatchCardsSizeMedium byte = 1
	MatchCardsSizeLarge  byte = 2
)

var ExpTable = [...]int32{15, 34, 57, 92, 135, 372, 560, 840, 1242, 1716, // Beginner

	// 1st Job
	2360, 3216, 4200, 5460, 7050, 8840, 11040, 13716, 16680, 20216, 24402,
	28980, 34320, 40512, 47216, 54900, 63666, 73080, 83720, 95700,

	// 2nd Job
	108480, 122760, 138666, 155540, 174216, 194832, 216600, 240500, 266682, 294216,
	324240, 356916, 391160, 428280, 468450, 510420, 555680, 604416, 655200, 709716,
	748608, 789631, 832902, 878545, 926689, 977471, 1031036, 1087536, 1147032, 1209994,
	1276301, 1346242, 1420016, 1497832, 1579913, 1666492, 1757815, 1854143, 1955750,
	2062925,
	// 3rd job
	2175973, 2295216, 2420993, 2553663, 2693603, 2841212, 2996910, 3161140, 3334370,
	3517093, 3709829, 3913127, 4127566, 4353756, 4592341, 4844001, 5109452, 5389449,
	5684790, 5996316, 6324914, 6671519, 7037118, 7422752, 7829518, 8258575, 8711144,
	9188514, 9692044, 10223168, 10783397, 11374327, 11997640, 12655110, 13348610, 14080113,
	14851703, 15665576, 16524049, 17429566, 18384706, 19392187, 20454878, 21575805,
	22758159, 24005306, 25320796, 26708375, 28171993, 29715818,

	// 4th job
	31344244, 33061908, 34873700, 36784778, 38800583, 40926854, 43169645, 45535341, 48030677,
	50662758, 53439077, 56367538, 59456479, 62714694, 66151459, 69776558, 73600313, 77633610,
	81887931, 86375389, 91108760, 96101520, 101367883, 106992842, 112782213, 118962678, 125481832,
	132358236, 139611467, 147262175, 155332142, 163844343, 172823012, 182293713, 192283408,
	202820538, 213935103, 225658746, 238024845, 251068606, 264827165, 279339639, 294647508,
	310794191, 327825712, 345790561, 364739883, 384727628, 405810702, 428049128, 451506220,
	476248760, 502347192, 529875818, 558913012, 589541445, 621848316, 655925603, 691870326,
	729784819, 769777027, 811960808, 856456260, 903390063, 952895838, 1005114529, 1060194805,
	1118293480, 1179575962, 1244216724, 1312399800, 1384319309, 1460180007, 1540197871, 1624600714,
	1713628833, 1807535693, 1906558648, 2011069705, 2121276324, // 0 is the amount of exp needed for level 200 to level up i.e. never shall
}

const (
	RoomMaxPlayers = 2

	OmokBoardSize = 15

	MatchCardsPairsSmall  = 6
	MatchCardsPairsMedium = 10
	MatchCardsPairsLarge  = 15

	RoomOwnerSlot = 0
	RoomGuestSlot = 1

	RoomLeaveTradeCancelled    = 0x02
	RoomYellowChatExpelled     = 0
	RoomYellowChatMatchedCards = 9
	RoomChatTypeChat           = 8
	RoomChatTypeNotice         = 7
	RoomPacketShowWindow       = 0x05
	RoomPacketJoin             = 0x04
	RoomPacketLeave            = 0x0A
	RoomEnterClosed            = 0x01
	RoomEnterFull              = 0x02
	RoomEnterBusy              = 0x03
	RoomEnterNotAllowedDead    = 0x04
	RoomEnterNotAllowedEvent   = 0x05
	RoomEnterThisCharNotAllow  = 0x06
	RoomEnterNoTradeATM        = 0x07
	RoomEnterTradeSameMap      = 0x09
	RoomEnterCannotCreateHere  = 0x0A
	RoomEnterCannotStartHere   = 0x0B
	RoomEnterStoreFMOnly       = 0x0C
	RoomEnterGarbageFloorFM    = 0x0D
	RoomEnterMayNotEnterStore  = 0x0E
	RoomEnterStoreMaint        = 0x0F
	RoomEnterGarbageTradeMsg   = 0x11
	RoomPacketInvite           = 0x02
	RoomPacketInviteResult     = 0x03
	RoomPacketShowAccept       = 0x0F
	RoomPacketMemoryStart      = 0x0C

	RoomRequestTie            byte = 42
	RoomRequestTieResult      byte = 43
	RoomForfeit               byte = 44
	RoomRequestUndo           byte = 46
	RoomRequestUndoResult     byte = 47
	RoomRequestExitDuringGame byte = 48
	RoomUndoRequestExit       byte = 49
	RoomReadyButtonPressed    byte = 50
	RoomUnready               byte = 51
	RoomOwnerExpell           byte = 52
	RoomGameStart             byte = 53
	RoomGameResult            byte = 54
	RoomChangeTurn            byte = 55
	RoomPlacePiece            byte = 56
	RoomInvalidPlace          byte = 57
	RoomSelectCard            byte = 60
)

const (
	ItemMesoMagnet        = 1812000
	ItemItemPouch         = 1812001
	ItemRegTeleportRock   = 5040000
	ItemVIPTeleportRock   = 5041000
	ItemAPReset           = 5050000
	ItemSPResetFirstJob   = 5050001
	ItemSPResetSecondJob  = 5050002
	ItemSPResetThirdJob   = 5050003
	ItemMegaphone         = 5071000
	ItemSuperMegaphone    = 5072000
	ItemHeartSMegaphone   = 5073000
	ItemSkullSMegaphone   = 5074000
	ItemWeatherSnow       = 5120000
	ItemWeatherFlowers    = 5120001
	ItemWeatherSoap       = 5120002
	ItemWeatherSnowFlakes = 5120003
	ItemWeatherPresents   = 5120004
	ItemWeatherChocolate  = 5120005
	ItemWeatherFlower     = 5120006
	ItemWeatherCandy      = 5120007
	ItemWeatherLeaves     = 5120008
	ItemWeatherFireworks  = 5120009
	ItemSafetyCharm       = 5130000
	ItemPetNameTag        = 5170000
	ItemWaterOfLife       = 5180000
)

// Login result codes
const (
	LoginResultSuccess           byte = 0x00
	LoginResultBanned            byte = 0x02
	LoginResultDeletedOrBlocked  byte = 0x03
	LoginResultInvalidPassword   byte = 0x04
	LoginResultNotRegistered     byte = 0x05
	LoginResultSystemError       byte = 0x06
	LoginResultAlreadyOnline     byte = 0x07
	LoginResultSystemError9      byte = 0x09
	LoginResultTooManyRequests   byte = 0x0A
	LoginResultOlderThan20       byte = 0x0B
	LoginResultValidLogin        byte = 0x0C
	LoginResultMasterCannotLogin byte = 0x0D
	LoginResultWrongGatewayKR    byte = 0x0E
	LoginResultProcessingKR      byte = 0x0F
	LoginResultVerifyEmail       byte = 0x10
	LoginResultGatewayEN         byte = 0x17
	LoginResultVerifyEmail21     byte = 0x15
	LoginResultEULA              byte = 0x17
)

// Auto-registration default values
const (
	AutoRegisterDefaultGender      byte   = 0
	AutoRegisterDefaultDOB         int    = 11111111
	AutoRegisterDefaultEULA        byte   = 0
	AutoRegisterDefaultAdminLevel  int    = 0
	AutoRegisterDefaultIsBanned    int    = 0
	AutoRegisterDefaultNX          int    = 0
	AutoRegisterDefaultMaplePoints int    = 0
	AutoRegisterDefaultIsLoggedIn  int    = 0
	AutoRegisterDefaultPIN         string = "1111"
)

const (
	MobSummonTypeFake     int8 = -4
	MobSummonTypeRevive   int8 = -3
	MobSummonTypeRegen    int8 = -2
	MobSummonTypeInstant  int8 = -1
	MobSummonTypeJrBalrog int8 = 0
	MobSummonTypePoof     int8 = 1
)

const (
	MobBalrog         int32 = 8130100
	MobCrimsonBalrog  int32 = 8150000
	MobPapalatusBall  int32 = 8500000
	MobPapalatus      int32 = 8500001
	MobPapalatusClock int32 = 8500002
	MobPianus         int32 = 8520000
	MobPianusAlt      int32 = 8510000
	MobMushmom        int32 = 6130101
	MobZombieMushmom  int32 = 6300005
	MobZakumArm1      int32 = 8800003
	MobZakumArm2      int32 = 8800004
	MobZakumArm3      int32 = 8800005
	MobZakumArm4      int32 = 8800006
	MobZakumArm5      int32 = 8800007
	MobZakumArm6      int32 = 8800008
	MobZakumArm7      int32 = 8800009
	MobZakumArm8      int32 = 8800010
	MobZakum1Body     int32 = 8800000
	MobZakum2Body     int32 = 8800001
	MobZakum3Body     int32 = 8800002
)

const (
	MapFreeMarket          int32 = 910000000
	MapBossPapulatus       int32 = 220080001
	MapBossPapulatusReturn int32 = 220080000
	MapBossPianus          int32 = 230040420
	MapBossZakum           int32 = 280030000
	MapBossZakumReturn     int32 = 211042300

	MapStationEllinia             int32 = 101000300
	MapStationOrbis               int32 = 200000100
	MapStationOrbisEllinaPlatform int32 = 200000111
	MapStationOrbisLudiPlatform   int32 = 200000121
	MapStationLudi                int32 = 220000100
	MapStationLudiOrbisPlatform   int32 = 220000110

	MapBoatElliniaDeparture   int32 = 101000301
	MapBoatElliniaFlight      int32 = 200090010
	MapBoatElliniaFlightCabin int32 = 200090011

	MapBoatOrbisElliniaDeparture   int32 = 200000112
	MapBoatOrbisElliniaFlight      int32 = 200090000
	MapBoatOrbisElliniaFlightCabin int32 = 200090001

	MapBoatOrbisLudiDeparture int32 = 200000122
	MapBoatOrbisLudiFlight    int32 = 200090100

	MapBoatLudiDeparture int32 = 220000111
	MapBoatLudiFlight    int32 = 200090110
)

const (
	PortalFreeMarketEnter = "market00"
	PortalFreeMarketLeave = "out00"
	PortalPapulatus       = "in00"
	PortalPianus          = "boss00"
	PortalZakum           = "ps00"
)
