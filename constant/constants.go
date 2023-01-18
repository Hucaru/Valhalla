package constant

import "strings"

var WORLD_NAMES = [...]string{"metaWorld", "metaSchool", "ihq", "metaBank", "metaInvest", "metaClassRoom", "Bellocan", "Mardia", "Kradia", "Yellonde", "Demethos", "Galicia", "El Nido", "Zenith", "Arcania", "Chaos", "Nova", "Renegates"}
var RandomHair = [...]string{"SKM_Male_Hair001_2", "SKM_Male_Hair002_2", "SKM_Female_Hair003", "SKM_Female_Hair001_2", "SKM_Female_Hair003_2", "SKM_Female_Hair002"}
var RandomClothes = [...]string{"SKM_None_Clothes001", "SKM_Male_Clothes001", "SKM_None_Clothes001_2", "SKM_None_Clothes001_2", "SKM_None_Clothes001_2"}
var RandomBottom = [...]string{"SKM_Female_bottom001", "SKM_Female_bottom002", "SKM_Female_bottom001_2", "SKM_None_bottom001", "SKM_None_bottom002"}
var RandomTop = [...]string{"SKM_Male_Top001", "SKM_None_Top002", "SKM_None_Top003", "SKM_None_Top001", "SKM_None_Top004"}
var RandomBody = [...]string{"SKM_None_Body_upper", "SKM_None_Body_Lower", "SKM_None_Body_foot"}

const (
	C2P_RequestLoginUser = 1
	P2C_ResultLoginUser  = 2
	P2C_ReportLoginUser  = 3

	C2P_RequestMoveStart = 4
	C2P_RequestMove      = 5
	C2P_RequestMoveEnd   = 6

	P2C_ReportMoveStart = 7
	P2C_ReportMove      = 8
	P2C_ReportMoveEnd   = 9

	C2P_RequestLogoutUser = 10
	P2C_ResultLogoutUser  = 11
	P2C_ReportLogoutUser  = 12

	C2P_RequestPlayerInfo = 13
	P2C_ResultPlayerInfo  = 14

	C2P_RequestAllChat = 15
	P2C_ReportAllChat  = 16

	C2P_RequestWhisper = 17
	P2C_ReportWhisper  = 18
	P2C_ResultWhisper  = 19

	C2P_RequestRegionChat = 20
	P2C_ReportRegionChat  = 21

	C2P_RequestPlayMontage = 22
	P2C_ReportPlayMontage  = 23

	C2P_RequestInteractionAttach = 24
	P2C_ReportInteractionAttach  = 25

	C2P_RequestMetaSchoolEnter = 26
	P2C_ReportMetaSchoolEnter  = 27
	P2C_ResultMetaSchoolEnter  = 28

	C2P_RequestRoleChecking = 29
	P2C_ResultRoleChecking  = 30

	C2P_RequestRegionChange = 31
	P2C_ResultRegionChange  = 32
	P2C_ReportRegionChange  = 33

	P2C_ResultRegionChat = 34
	P2C_ResultAllChat    = 35

	P2C_ResultInteractionAttach = 36
	C2P_RequestMetaSchoolLeave  = 37
	P2C_ReportMetaSchoolLeave   = 38
	P2C_ReportRegionLeave       = 39

	P2C_ResultGrid    = 40
	P2C_ReportGridOld = 41
	P2C_ReportGridNew = 42

	//errors
	P2C_ResultLoginUserError = 90

	OnConnected    = 10000
	OnDisconnected = 10001
)

const (
	PosX = -5102.0
	PosY = -3365.0
	PosZ = 670.0
	RotX = 0.0
	RotY = 0.0
	RotZ = 0.0
)

const (
	NoError                = -1
	ErrorUserOffline       = 2
	ErrorCodeDuplicateName = 400
	ErrorCodeDuplicateUID  = 401
	ErrorCodeAlreadyOnline = 403

	ErrorCodeChairNotEmpty = 400
	ErrorCodeDBorServer    = 500
)

const (
	DEFAULT_TIME = 0
	NO_TARGET    = -1
)

const (
	LAND_X1         = -90000
	LAND_Y1         = -90000
	LAND_X2         = 90000
	LAND_Y2         = 90000
	LAND_Z          = 3000
	LAND_VIEW_RANGE = 2048 //number multiplies 2
)

const (
	World         = 1
	MetaSchool    = 2
	Ihq           = 3
	MetaBank      = 4
	MetaInvest    = 5
	MetaClassRoom = 6
	RegionMax     = 7
)

const (
	UNKNOWN   = -1
	User      = 0
	Moderator = 1
	Admin     = 2
)

// Generic Constants
const (
	MapleVersion          = 28
	ClientHeaderSize      = 4
	MetaClientHeaderSize  = 8
	InterserverHeaderSize = 4
	OpcodeLength          = 1
)

const (
	MetaEventLogin    = 101
	MetaEventMovement = 102
)

const (
	MaxItemStack = 200

	HpID    = 0x400
	MaxHpID = 0x800
	MpID    = 0x1000
	MaxMpID = 0x2000

	StrID = 0x40
	DexID = 0x80
	IntID = 0x100
	LukID = 0x200

	LevelID = 0x10
	JobID   = 0x20
	ExpID   = 0x10000

	ApID = 0x4000
	SpID = 0x8000

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
)

var LANGUAGES = map[string]string{
	"english": "en",
	"korean":  "ko",
	"thai":    "th",
}

func GetISOCode(lng string) string {
	return LANGUAGES[strings.ToLower(lng)]
}

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
