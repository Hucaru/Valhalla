package constants

var WORLD_NAMES = [...]string{"Scania", "Bera", "Broa", "Windia", "Khaini", "Bellocan", "Mardia", "Kradia", "Yellonde", "Demethos", "Galicia", "El Nido", "Zenith", "Arcania", "Chaos", "Nova", "Renegates"}

const (
	MAX_ITEM_STACK = 200

	HP_ID     = 0x400
	MAX_HP_ID = 0x800
	MP_ID     = 0x1000
	MAX_MP_ID = 0x2000

	STR_ID = 0x40
	DEX_ID = 0x80
	INT_ID = 0x100
	LUK_ID = 0x200

	LEVEL_ID = 0x10
	JOB_ID   = 0x20
	EXP_ID   = 0x10000

	AP_ID = 0x4000
	SP_ID = 0x8000

	FAME_ID  = 0x20000
	MESOS_ID = 0x40000

	BEGGINNER_HP_ADD = int16(12)
	BEGGINNER_MP_ADD = int16(10)

	WARRIOR_HP_ADD = int16(24)
	WARRIOR_MP_ADD = int16(4)

	MAGICIAN_HP_ADD = int16(10)
	MAGICIAN_MP_ADD = int16(6)

	BOWMAN_HP_ADD = int16(20)
	BOWMAN_MP_ADD = int16(14)

	THIEF_HP_ADD = int16(20)
	THIEF_MP_ADD = int16(14)

	ADMIN_HP_ADD = 150
	ADMIN_MP_ADD = 150
)
