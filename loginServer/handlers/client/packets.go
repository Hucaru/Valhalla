package client

import (
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

func writePlayerCharacter(pac *gopacket.Packet, pos uint32, char character.Character) {
	pac.WriteUint32(pos)

	name := char.GetName()

	if len(name) > 13 {
		name = name[:13]
	}

	padSize := 13 - len(name)

	pac.WriteBytes([]byte(name))
	for i := 0; i < padSize; i++ {
		pac.WriteByte(0x0)
	}

	pac.WriteByte(char.GetGender()) //gender
	pac.WriteByte(char.GetSkin())   // skin
	pac.WriteUint32(char.GetFace()) // face
	pac.WriteUint32(char.GetHair()) // Hair

	pac.WriteInt64(0x0) // Pet cash ID

	pac.WriteByte(char.GetLevel())   // level
	pac.WriteUint16(char.GetJob())   // Job
	pac.WriteUint16(char.GetStr())   // str
	pac.WriteUint16(char.GetDex())   // dex
	pac.WriteUint16(char.GetInt())   // int
	pac.WriteUint16(char.GetLuk())   // luk
	pac.WriteUint16(char.GetHP())    // hp
	pac.WriteUint16(char.GetMaxHP()) // max hp
	pac.WriteUint16(char.GetMP())    // mp
	pac.WriteUint16(char.GetMaxMP()) // max mp
	pac.WriteUint16(char.GetAP())    // ap
	pac.WriteUint16(char.GetSP())    // sp
	pac.WriteUint32(char.GetEXP())   // exp
	pac.WriteUint16(char.GetFame())  // fame

	pac.WriteUint32(char.GetCurrentMap())  // map id
	pac.WriteByte(char.GetCurrentMapPos()) // map

	character.WriteDisplayCharacter(&char, pac)

	pac.WriteByte(0)  // Rankings
	pac.WriteInt32(0) // ?
}
