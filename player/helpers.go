package player

import (
	"log"
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/maplepacket"

	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/interfaces"
)

var charsPtr interfaces.Characters

// RegisterCharactersObj -
func RegisterCharactersObj(chars interfaces.Characters) {
	charsPtr = chars
	go charactersAutoSave()
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

		conn.Write(statChangePacket(false, hpID, uint32(newHp)))
		conn.Write(statChangePacket(false, maxHpID, uint32(newHp)))

		conn.Write(statChangePacket(false, mpID, uint32(newHp)))
		conn.Write(statChangePacket(false, maxMpID, uint32(newHp)))

		conn.Write(statChangePacket(false, apID, uint32(newAP)))
		conn.Write(statChangePacket(false, spID, uint32(newSP)))
	}

	char.SetLevel(newLevel)
	conn.Write(statChangePacket(false, levelID, uint32(newLevel)))
}

func GiveExp(conn interfaces.ClientConn, exp uint32) maplepacket.Packet {
	char := charsPtr.GetOnlineCharacterHandle(conn)

	if conn == nil || char.GetLevel() < 1 || char.GetLevel() > 199 {
		return []byte{}
	}

	newExp := int32(exp * constants.GetRate(constants.ExpRate))

	result := []byte{}

	conn.Write(expGainedMessage(true, false, uint32(newExp)))

	for { // allow character to level up multiple times from exp
		reqExp := constants.ExpTable[char.GetLevel()-1]

		if char.GetEXP()+uint32(newExp) >= reqExp {
			SetLevel(conn, char.GetLevel()+1)
			newExp -= int32(reqExp)

			if newExp < 0 {
				newExp = 0
			}
			char.SetEXP(0)

			result = levelUpAnimationPacket(char.GetCharID())
		} else {
			newExp += int32(char.GetEXP())
			break
		}
	}

	conn.Write(statChangePacket(false, expID, uint32(newExp)))
	char.SetEXP(uint32(newExp))

	return result
}

func charactersAutoSave() {
	// Save character data every 15 mins
	ticker := time.NewTicker(15 * time.Minute)

	for {
		<-ticker.C
		charMap := charsPtr.GetChars()

		if charMap == nil {
			return
		}

		for _, char := range charMap {
			// save changes to db
			err := character.SaveCharacter(char)

			if err != nil {
				log.Println("Unable to save character data")
			}
		}
	}
}
