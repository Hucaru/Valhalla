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

var ExpTable = [...]int32{15, 34, 57, 92, 135, 372, 560, 840, 1242, 1144, // Beginner

	// 1st Job
	1573, 2144, 2800, 3640, 4700, 5893, 7360, 9144, 11120, 13477, 16268,
	19320, 22880, 27008, 31477, 36600, 42444, 48720, 55813, 63800,

	// 2nd Job
	86784, 98208, 110932, 124432, 139372, 155865, 173280, 192400, 213345, 235372,
	259392, 285532, 312928, 342624, 374760, 408336, 445544, 483532, 524160, 567772,
	598886, 631704, 666321, 702836, 741351, 781976, 824828, 870028, 917625, 967995,
	1021041, 1076994, 1136013, 1198266, 1263930, 1333194, 1406252, 1483314, 1564600,
	1650340,
	// 3rd job
	1740778, 1836173, 1936794, 2042930, 2154882, 2272970, 2397528, 2528912, 2667496,
	2813674, 2967863, 3130502, 3302053, 3483005, 3673873, 3875201, 4087562, 4311559,
	4547832, 4797053, 5059931, 5337215, 5629694, 5938202, 6263614, 6606860, 6968915,
	7350811, 7753635, 8178534, 8626718, 9099462, 9598112, 10124088, 10678888, 11264090,
	11881362, 12532461, 13219239, 13943653, 14707765, 15513750, 16363902, 17260644,
	18206527, 19204245, 20256637, 21366700, 22537594, 23772654,

	// 4th job
	25075395, 26449526, 27898960, 29427822, 31040466, 32741483, 34535716, 36428273, 38424542,
	40530206, 42751262, 45094030, 47565183, 50171755, 52921167, 55821246, 58880250, 62106888,
	65510344, 69100311, 72887008, 76881216, 81094306, 85594273, 90225770, 95170142, 100385466,
	105886589, 111689174, 117809740, 124265714, 131075474, 138258410, 145834970, 153826726,
	162256430, 171148082, 180526997, 190419876, 200854885, 211861732, 223471711, 223471711,
	248635353, 262260570, 276632449, 291791906, 307782102, 324648562, 342439302, 361204976,
	380999008, 401877754, 423900654, 447130410, 471633156, 497478653, 524740482, 553496261,
	583827855, 615821622, 649568646, 685165008, 722712050, 762316670, 804091623, 848155844,
	894634784, 943660770, 995373379, 1049919840, 1107455447, 1168144006, 1232158297, 1299680571,
	1370903066, 1446028554, 1525246918, 1608855764, 1697021059, // 0 is the amount of exp needed for level 200 to level up i.e. never shall
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
	ItemMesoMagnet = 1812000
	ItemItemPouch  = 1812001
)

// Login result codes
const (
	LoginResultSuccess            byte = 0x00
	LoginResultBanned             byte = 0x02
	LoginResultDeletedOrBlocked   byte = 0x03
	LoginResultInvalidPassword    byte = 0x04
	LoginResultNotRegistered      byte = 0x05
	LoginResultSystemError        byte = 0x06
	LoginResultAlreadyOnline      byte = 0x07
	LoginResultSystemError9       byte = 0x09
	LoginResultTooManyRequests    byte = 0x0A
	LoginResultOlderThan20        byte = 0x0B
	LoginResultValidLogin         byte = 0x0C
	LoginResultMasterCannotLogin  byte = 0x0D
	LoginResultWrongGatewayKR     byte = 0x0E
	LoginResultProcessingKR       byte = 0x0F
	LoginResultVerifyEmail        byte = 0x10
	LoginResultGatewayEN          byte = 0x17
	LoginResultVerifyEmail21      byte = 0x15
	LoginResultEULA               byte = 0x17
)

// Auto-registration default values
const (
	AutoRegisterDefaultGender     byte  = 0
	AutoRegisterDefaultDOB        int   = 1111111
	AutoRegisterDefaultEULA       byte  = 1
	AutoRegisterDefaultAdminLevel int   = 0
	AutoRegisterDefaultIsBanned   int   = 0
	AutoRegisterDefaultNX         int   = 0
	AutoRegisterDefaultMaplePoints int  = 0
	AutoRegisterDefaultIsLoggedIn int   = 0
	AutoRegisterDefaultPIN        string = "1111"
)
