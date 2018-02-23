package player

import (
	"log"

	"github.com/Hucaru/Valhalla/channelServer/playerConn"
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
		char.SetMaxMp(value)
	default:
		log.Println("Unknown stat ID:", stat)
	}

	newAP := char.GetAP() - 1
	conn.Write(statChange(true, stat, value))
	conn.Write(statChange(true, apID, newAP))
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

		conn.Write(statChange(true, hpID, char.GetHP()))
	} else if mp > 0 {
		char.SetMP(char.GetMP() + mp)

		if char.GetMP() > char.GetMaxMP() {
			char.SetMP(char.GetMaxMP())
		}

		conn.Write(statChange(true, mpID, char.GetMP()))
	}

	// If in party send update to party members in map
}

func PlayerChangeJob(conn *playerConn.Conn, jobValue uint16) {
	conn.GetCharacter().SetJob(jobValue)
	conn.Write(statChange(false, jobID, jobValue))
	// Send map change job
}

// TODO: Go change base underlying type from byte to uint16
func PlayerSetLevel(conn *playerConn.Conn, level uint16) {
	delta := int16(level) - int16(conn.GetCharacter().GetLevel())
	char := conn.GetCharacter()

	if delta > 0 {
		newAP := char.GetAP() + 5*uint16(delta)
		newSP := char.GetSP() + 3*uint16(delta)

		char.SetAP(newAP)
		char.SetSP(newSP)

		conn.Write(statChange(true, apID, newAP))
		conn.Write(statChange(true, spID, newSP))
	}

	char.SetLevel(byte(level))
	conn.Write(statChange(false, levelID, level))
}

func PlayerAddExp() {

}

func PlayerSetAP() {

}

func PlayerSetSP() {

}
