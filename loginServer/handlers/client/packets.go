package client

import (
	"math"

	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

func channelToLogin() gopacket.Packet {
	pac := gopacket.NewPacket()
	pac.WriteByte(constants.SEND_LOGIN_RESTARTER)
	pac.WriteByte(0x01)

	return pac
}

func loginResponce(result byte, userID uint32, gender byte, isAdmin byte, username string, isBanned int) gopacket.Packet {
	pac := gopacket.NewPacket()
	pac.WriteByte(constants.SEND_LOGIN_RESPONCE)
	pac.WriteByte(result)
	pac.WriteByte(0x00)
	pac.WriteInt32(0)

	if result <= 0x01 {
		pac.WriteUint32(userID)
		pac.WriteByte(gender)
		pac.WriteByte(isAdmin)
		pac.WriteByte(0x01)
		pac.WriteString(username)
	} else if result == 0x02 {
		pac.WriteByte(byte(isBanned))
		pac.WriteInt64(0) // Expire time, for now let set this to epoch
	}

	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)

	return pac
}

func endWorldList() gopacket.Packet {
	pac := gopacket.NewPacket()
	pac.WriteByte(constants.SEND_LOGIN_SEND_WORLD_LIST)
	pac.WriteByte(0xFF)

	return pac
}

func displayCharacters(characters []character.Character) gopacket.Packet {
	pac := gopacket.NewPacket()
	pac.WriteByte(constants.SEND_LOGIN_CHARACTER_DATA)
	pac.WriteByte(0) // ?

	if len(characters) < 4 && len(characters) > 0 {
		pac.WriteByte(byte(len(characters)))

		for _, c := range characters {
			writePlayerCharacter(&pac, c.GetCharID(), c)
		}
	} else {
		pac.WriteByte(0)
	}

	return pac
}

func nameCheck(name string, nameFound int) gopacket.Packet {
	pac := gopacket.NewPacket()
	pac.WriteByte(constants.SEND_LOGIN_NAME_CHECK_RESULT)
	pac.WriteString(name)

	if nameFound > 0 {
		pac.WriteByte(0x1) // 0 = good name, 1 = bad name
	} else {
		pac.WriteByte(0x0)
	}

	return pac
}

func createdCharacter(success bool, character character.Character) gopacket.Packet {
	pac := gopacket.NewPacket()
	pac.WriteByte(constants.SEND_LOGIN_NEW_CHARACTER_GOOD)

	if success {
		pac.WriteByte(0x0) // if creation was sucessfull - 0 = good, 1 = bad
		writePlayerCharacter(&pac, character.GetCharID(), character)
	} else {
		pac.WriteByte(0x1)
	}

	return pac
}

func deleteCharacter(charID int32, deleted bool, hacking bool) gopacket.Packet {
	pac := gopacket.NewPacket()
	pac.WriteByte(constants.SEND_LOGIN_DELETE_CHARACTER)
	pac.WriteInt32(charID)

	if deleted {
		pac.WriteByte(0x0)
	} else if hacking {
		pac.WriteByte(0x0A) // Could not be processed due to server load
	} else {
		pac.WriteByte(0x12)
	}

	return pac
}

func migrateClient(ip []byte, port uint16, charID int32) gopacket.Packet {
	pac := gopacket.NewPacket()
	pac.WriteByte(constants.SEND_LOGIN_CHARACTER_MIGRATE)
	pac.WriteByte(0x00)
	pac.WriteByte(0x00)
	pac.WriteBytes(ip)
	pac.WriteUint16(port)
	pac.WriteInt32(charID)
	pac.WriteByte(byte(0) | byte(1<<0))
	pac.WriteInt32(1)

	return pac
}

func sendBadMigrate() gopacket.Packet {
	pac := gopacket.NewPacket()
	pac.WriteByte(constants.SEND_LOGIN_CHARACTER_MIGRATE)
	pac.WriteByte(0x00) // flipping these 2 bytes makes the character select screen do nothing it appears
	pac.WriteByte(0x00)
	pac.WriteBytes([]byte{0, 0, 0, 0})
	pac.WriteUint16(0)
	pac.WriteInt32(8)
	pac.WriteByte(byte(0) | byte(1<<0))
	pac.WriteInt32(1)

	return pac
}

func writePlayerCharacter(pac *gopacket.Packet, pos uint32, character character.Character) {
	pac.WriteUint32(pos)

	name := character.GetName()

	if len(name) > 13 {
		name = name[:13]
	}

	padSize := 13 - len(name)

	pac.WriteBytes([]byte(name))
	for i := 0; i < padSize; i++ {
		pac.WriteByte(0x0)
	}

	pac.WriteByte(character.GetGender()) //gender
	pac.WriteByte(character.GetSkin())   // skin
	pac.WriteUint32(character.GetFace()) // face
	pac.WriteUint32(character.GetHair()) // Hair

	pac.WriteInt64(0x0) // Pet cash ID

	pac.WriteByte(character.GetLevel())   // level
	pac.WriteUint16(character.GetJob())   // Job
	pac.WriteUint16(character.GetStr())   // str
	pac.WriteUint16(character.GetDex())   // dex
	pac.WriteUint16(character.GetInt())   // int
	pac.WriteUint16(character.GetLuk())   // luk
	pac.WriteUint16(character.GetHP())    // hp
	pac.WriteUint16(character.GetMaxHP()) // max hp
	pac.WriteUint16(character.GetMP())    // mp
	pac.WriteUint16(character.GetMaxMP()) // max mp
	pac.WriteUint16(character.GetAP())    // ap
	pac.WriteUint16(character.GetSP())    // sp
	pac.WriteUint32(character.GetEXP())   // exp
	pac.WriteUint16(character.GetFame())  // fame

	pac.WriteUint32(character.GetCurrentMap())  // map id
	pac.WriteByte(character.GetCurrentMapPos()) // map

	// Why is this shit repeated?
	pac.WriteByte(character.GetGender()) // gender
	pac.WriteByte(character.GetSkin())   // skin
	pac.WriteUint32(character.GetFace()) // face
	pac.WriteByte(0x00)                  // ?
	pac.WriteUint32(character.GetHair()) // hair

	cashWeapon := uint32(0)

	for _, b := range character.GetEquips() {
		if b.GetSlotID() < 0 && b.GetSlotID() > -20 {
			pac.WriteByte(byte(math.Abs(float64(b.GetSlotID()))))
			pac.WriteUint32(b.GetItemID())
		}
	}

	for _, b := range character.GetEquips() {
		if b.GetSlotID() < -100 {
			if b.GetSlotID() == -111 {
				cashWeapon = b.GetItemID()
			} else {
				pac.WriteByte(byte(math.Abs(float64(b.GetSlotID() + 100))))
				pac.WriteUint32(b.GetItemID())
			}
		}
	}

	pac.WriteByte(0xFF)
	// Another set of items go here? if anything. double 0xFF seems weird. I would imagine it is some sort of seperator each
	pac.WriteByte(0xFF)
	pac.WriteUint32(cashWeapon)

	pac.WriteByte(1)  // Rankings
	pac.WriteInt32(1) // ?
	pac.WriteInt32(2) // ?
	pac.WriteInt32(3) // ?
	pac.WriteInt32(4) // ?
	pac.WriteInt32(5) // ?
	pac.WriteInt32(6) // ?
}
