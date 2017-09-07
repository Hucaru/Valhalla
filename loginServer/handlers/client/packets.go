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
			writePlayerCharacter(&pac, int(c.CharID), c)
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
		writePlayerCharacter(&pac, int(character.CharID), character)
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

func writePlayerCharacter(pac *gopacket.Packet, pos int, character character.Character) {
	pac.WriteInt32(int32(pos))

	name := character.Name

	if len(name) > 13 {
		name = name[:13]
	}

	padSize := 13 - len(name)

	pac.WriteBytes([]byte(name))
	for i := 0; i < padSize; i++ {
		pac.WriteByte(0x0)
	}

	pac.WriteByte(character.Gender) //gender
	pac.WriteByte(character.Skin)   // skin
	pac.WriteInt32(character.Face)  // face
	pac.WriteInt32(character.Hair)  // Hair

	pac.WriteInt64(0x0) // Pet cash ID

	pac.WriteByte(character.Level)  // level
	pac.WriteInt16(character.Job)   // Job
	pac.WriteInt16(character.Str)   // str
	pac.WriteInt16(character.Dex)   // dex
	pac.WriteInt16(character.Intt)  // int
	pac.WriteInt16(character.Luk)   // luk
	pac.WriteInt16(character.HP)    // hp
	pac.WriteInt16(character.MaxHP) // max hp
	pac.WriteInt16(character.MP)    // mp
	pac.WriteInt16(character.MaxMP) // max mp
	pac.WriteInt16(character.AP)    // ap
	pac.WriteInt16(character.SP)    // sp
	pac.WriteInt32(character.EXP)   // exp
	pac.WriteInt16(character.Fame)  // fame

	pac.WriteInt32(character.CurrentMap)   // map id
	pac.WriteByte(character.CurrentMapPos) // map

	// Why is this shit repeated?
	pac.WriteByte(character.Gender) // gender
	pac.WriteByte(character.Skin)   // skin
	pac.WriteInt32(character.Face)  // face
	pac.WriteByte(0x0)              // ?
	pac.WriteInt32(character.Hair)  // hair

	// hidden equip - byte for type id , int for value

	// shown equip - byte for type id , int for value
	for _, b := range character.Items {
		if b.SlotID < 0 {
			pac.WriteByte(byte(math.Abs(float64(b.SlotID))))
			pac.WriteInt32(b.ItemID)
		}
	}

	pac.WriteByte(0xFF)
	pac.WriteByte(0xFF)

	pac.WriteByte(1)  // Rankings
	pac.WriteInt32(0) // ?
	pac.WriteInt32(0) // world old
}
