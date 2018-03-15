package player

import (
	"log"
	"math/rand"

	"github.com/Hucaru/gopacket"

	"github.com/Hucaru/Valhalla/interfaces"
)

var charsPtr interfaces.Characters

// RegisterCharactersObj -
func RegisterCharactersObj(chars interfaces.Characters) {
	charsPtr = chars
}

func SetJob(conn interfaces.ClientConn, newJob uint16) {
	charsPtr.GetOnlineCharacterHandle(conn).SetJob(jobID)
	conn.Write(statChangePacket(true, jobID, uint32(newJob)))
}

func SetLevel(conn interfaces.ClientConn, newLevel byte) {
	char := charsPtr.GetOnlineCharacterHandle(conn)

	delta := int16(newLevel) - int16(char.GetLevel())

	if delta > 0 {
		newAP := char.GetAP() + 5*uint16(delta)
		newSP := char.GetSP() + 3*uint16(delta)

		char.SetAP(newAP)
		char.SetSP(newSP)

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

		conn.Write(statChangePacket(true, hpID, uint32(newHp)))
		conn.Write(statChangePacket(true, maxHpID, uint32(newHp)))

		conn.Write(statChangePacket(true, mpID, uint32(newHp)))
		conn.Write(statChangePacket(true, maxMpID, uint32(newHp)))

		conn.Write(statChangePacket(true, apID, uint32(newAP)))
		conn.Write(statChangePacket(true, spID, uint32(newSP)))
	}

	char.SetEXP(600)
	conn.Write(statChangePacket(true, expID, 600))

	char.SetLevel(newLevel)
	conn.Write(statChangePacket(true, levelID, uint32(newLevel)))
}

func GiveExp(conn interfaces.ClientConn, exp uint32) gopacket.Packet {
	if conn == nil {
		return []byte{}
	}

	p := gopacket.NewPacket()

	return p
}
