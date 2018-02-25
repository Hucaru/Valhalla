package player

import (
	"log"
	"math/rand"

	"github.com/Hucaru/Valhalla/channelServer/playerConn"
	"github.com/Hucaru/Valhalla/channelServer/server"
	"github.com/Hucaru/gopacket"
)

const (
	hpID    = 0x400
	maxHpID = 0x800
	mpID    = 0x1000
	maxMpID = 0x2000

	strID = 0x40
	dexID = 0x80
	intID = 0x100
	lukID = 0x200

	levelID = 0x10
	jobID   = 0x20
	expID   = 0x10000

	apID = 0x4000
	spID = 0x8000

	fameID  = 0x20000
	mesosID = 0x40000

	begginnerHpAdd = uint16(12)
	begginnerMpAdd = uint16(10)

	warriorHpAdd = uint16(24)
	warriorMpAdd = uint16(4)

	magicianHpAdd = uint16(10)
	magicianMpAdd = uint16(6)

	bowmanHpAdd = uint16(20)
	bowmanMpAdd = uint16(14)

	thiefHpAdd = uint16(20)
	thiefMpAdd = uint16(14)

	adminHpAdd = 150
	adminMpAdd = 150
)

func HandlePlayerChangeStat(reader gopacket.Reader, conn *playerConn.Conn) {
	if conn.GetCharacter().GetAP() == 0 {
		return
	}

	stat := reader.ReadUint32()
	var value uint16

	maxDice := uint16(2000)
	maxHpMp := uint16(30000)

	char := conn.GetCharacter()

	switch stat {
	case strID:
		if char.GetStr() >= maxDice {
			conn.Write(statNoChange())
			return
		}

		value = char.GetStr() + 1
		char.SetStr(value)
	case dexID:
		if char.GetDex() >= maxDice {
			conn.Write(statNoChange())
			return
		}

		value = char.GetDex() + 1
		char.SetDex(value)
	case intID:
		if char.GetInt() >= maxDice {
			conn.Write(statNoChange())
			return
		}

		value = char.GetInt() + 1
		char.SetInt(value)
	case lukID:
		if char.GetLuk() >= maxDice {
			conn.Write(statNoChange())
			return
		}

		value = char.GetLuk() + 1
		char.SetLuk(value)
	case maxHpID:
		if char.GetMaxHP() >= maxHpMp {
			conn.Write(statNoChange())
			return
		}

		value = char.GetMaxHP() + 1
		char.SetMaxHP(value)
	case maxMpID:
		if char.GetMaxHP() >= maxHpMp {
			conn.Write(statNoChange())
			return
		}

		value = char.GetMaxMP() + 1
		char.SetMaxMP(value)
	default:
		log.Println("Unknown stat ID:", stat)
	}

	newAP := char.GetAP() - 1
	conn.Write(statChangeUint16(true, stat, value))
	conn.Write(statChangeUint16(true, apID, newAP))
	char.SetAP(newAP)
}

func HandlePlayerPassiveRegen(reader gopacket.Reader, conn *playerConn.Conn) {
	reader.ReadBytes(4) // Client - Server validation?

	hp := reader.ReadUint16()
	mp := reader.ReadUint16()

	char := conn.GetCharacter()

	if char.GetHP() == 0 || hp > 400 || mp > 1000 || (hp > 0 && mp > 0) {
		return
	}

	if hp > 0 {
		char.SetHP(char.GetHP() + hp)

		if char.GetHP() > char.GetMaxHP() {
			char.SetHP(char.GetMaxHP())
		}

		conn.Write(statChangeUint16(true, hpID, char.GetHP()))
	} else if mp > 0 {
		char.SetMP(char.GetMP() + mp)

		if char.GetMP() > char.GetMaxMP() {
			char.SetMP(char.GetMaxMP())
		}

		conn.Write(statChangeUint16(true, mpID, char.GetMP()))
	}

	// If in party send update to party members in map
}

func PlayerChangeJob(conn *playerConn.Conn, jobValue uint16) {
	conn.GetCharacter().SetJob(jobValue)
	conn.Write(statChangeUint16(false, jobID, jobValue))
	// Send map change job
}

func PlayerSetHP(conn *playerConn.Conn, newHp uint16) {
	char := conn.GetCharacter()
	char.SetHP(newHp)

	conn.Write(statChangeUint16(true, hpID, newHp))
}

func PlayerSetMP(conn *playerConn.Conn, newMp uint16) {
	char := conn.GetCharacter()
	char.SetHP(newMp)

	conn.Write(statChangeUint16(true, mpID, newMp))
}

func PlayerSetLevel(conn *playerConn.Conn, level byte) {
	delta := int16(level) - int16(conn.GetCharacter().GetLevel())
	char := conn.GetCharacter()

	if delta > 0 {
		newAP := char.GetAP() + 5*uint16(delta)
		newSP := char.GetSP() + 3*uint16(delta)

		char.SetAP(newAP)
		char.SetSP(newSP)

		char.SetEXP(0)

		levelUpHp := func(classIncrease uint16, bonus uint16) uint16 {
			return uint16(rand.Intn(3)+1) + classIncrease + bonus // deterministic rand, maybe seed with time?
		}

		levelUpMp := func(classIncrease uint16, bonus uint16) uint16 {
			return uint16(rand.Intn(1)+1) + classIncrease + bonus // deterministic rand, maybe seed with time?
		}

		var hpToAdd uint16
		var mpToAdd uint16

		switch int(char.GetJob() / 100) {
		case 0:
			hpToAdd = levelUpHp(begginnerHpAdd, 0)
			mpToAdd = levelUpMp(begginnerMpAdd, char.GetInt())
		case 1:
			hpToAdd = levelUpHp(warriorHpAdd, 0)
			mpToAdd = levelUpMp(warriorMpAdd, char.GetInt())
		case 2:
			hpToAdd = levelUpHp(magicianHpAdd, 0)
			mpToAdd = levelUpMp(magicianMpAdd, 2*char.GetInt())
		case 3:
			hpToAdd = levelUpHp(bowmanHpAdd, 0)
			mpToAdd = levelUpMp(bowmanHpAdd, char.GetInt())
		case 4:
			hpToAdd = levelUpHp(thiefHpAdd, 0)
			mpToAdd = levelUpMp(thiefMpAdd, char.GetInt())
		case 5:
			hpToAdd = adminHpAdd
			mpToAdd = adminMpAdd
		default:
			log.Println("Unknown Job ID:", char.GetJob())
		}

		newHp := char.GetMaxHP() + hpToAdd*uint16(delta)
		char.SetMaxHP(newHp)
		char.SetHP(newHp)

		newMp := char.GetMaxMP() + mpToAdd*uint16(delta)
		char.SetMaxMP(newMp)
		char.SetMP(newMp)

		conn.Write(statChangeUint16(true, hpID, newHp))
		conn.Write(statChangeUint16(true, maxHpID, newHp))

		conn.Write(statChangeUint16(true, mpID, newHp))
		conn.Write(statChangeUint16(true, maxMpID, newHp))

		conn.Write(statChangeUint16(true, apID, newAP))
		conn.Write(statChangeUint16(true, spID, newSP))
		conn.Write(statChangeUint32(true, expID, 0))
	}

	char.SetLevel(byte(level))
	conn.Write(statChangeByte(false, levelID, level))
	server.SendPacketToMap(char.GetCurrentMap(), playerLevelUpAnimation(char.GetCharID()), nil)
}

func PlayerAddExp(conn *playerConn.Conn, exp uint32) {
	char := conn.GetCharacter()

	if char.GetLevel() > 199 {
		return
	}

	if exp+char.GetEXP() >= expTable[char.GetLevel()-1] {
		PlayerSetLevel(conn, char.GetLevel()+1)
	} else {
		newExp := char.GetEXP() + exp
		char.SetEXP(newExp)
		conn.Write(statChangeUint32(true, expID, newExp))
	}
}

func PlayerAddFame(conn *playerConn.Conn, fame uint16, charToFame uint32) {
	// Send famed player new fame value

	// Send fame up down to both players
}

// https://bbb.hidden-street.net/experience-table
var expTable = [200]uint32{15, 34, 57, 92, 135, 372, 560, 840, 1242, 1144, // Begginer

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
	1370903066, 1446028554, 1525246918, 1608855764, 1697021059,
}
