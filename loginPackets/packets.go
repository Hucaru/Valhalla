package loginPackets

import (
	"strconv"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/gopacket"
)

func ChannelToLogin() gopacket.Packet {
	pac := gopacket.NewPacket()
	pac.WriteByte(constants.SEND_LOGIN_RESTARTER)
	pac.WriteByte(0x01)

	return pac
}

func LoginResponce(result byte, userID uint32, gender byte, isAdmin byte, username string, isBanned int) gopacket.Packet {
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

func EndWorldList() gopacket.Packet {
	pac := gopacket.NewPacket()
	pac.WriteByte(constants.SEND_LOGIN_SEND_WORLD_LIST)
	pac.WriteByte(0xFF)

	return pac
}

func DisplayCharacters(characters []character.Character) gopacket.Packet {
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

func NameCheck(name string, nameFound int) gopacket.Packet {
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

func CreatedCharacter(success bool, character character.Character) gopacket.Packet {
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

func DeleteCharacter(charID int32, deleted bool, hacking bool) gopacket.Packet {
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

func MigrateClient(ip []byte, port uint16, charID int32) gopacket.Packet {
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

func SendBadMigrate() gopacket.Packet {
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

	pac.WriteInt32(0) // is character is selected and which one
	pac.WriteByte(1)  // Rankings
	pac.WriteInt32(1) // world ranking position
	pac.WriteInt32(2) // increase / decrease amount
	pac.WriteInt32(3) // class ranking position
	pac.WriteInt32(4) // increase / decrease amount
}

var worldNames = [...]string{"Scania", "Bera", "Broa", "Windia", "Khaini", "Bellocan", "Mardia", "Kradia", "Yellonde", "Demethos", "Galicia", "El Nido", "Zenith", "Arcania", "Chaos", "Nova", "Renegates"}

func WorldListing(worldIndex byte) gopacket.Packet {
	pac := gopacket.NewPacket()
	pac.WriteByte(constants.SEND_LOGIN_SEND_WORLD_LIST)
	pac.WriteByte(worldIndex)               // world id
	pac.WriteString(worldNames[worldIndex]) // World name -
	pac.WriteByte(3)                        // Ribbon on world - 0 = normal, 1 = event, 2 = new, 3 = hot
	pac.WriteString("test")
	pac.WriteByte(0)  // ? exp event notification?
	pac.WriteByte(20) // number of channels

	maxPopulation := 150
	population := 50

	for j := 1; j < 21; j++ {
		pac.WriteString(worldNames[worldIndex] + "-" + strconv.Itoa(j))                // channel name
		pac.WriteInt32(int32(1200.0 * (float64(population) / float64(maxPopulation)))) // Population
		pac.WriteByte(worldIndex)                                                      // world id
		pac.WriteByte(byte(j))                                                         // channel id
		pac.WriteByte(byte(j - 1))                                                     //?
	}

	return pac
}

func WorldInfo(warning byte, population byte) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_LOGIN_WORLD_META)
	p.WriteByte(warning)    // Warning - 0 = no warning, 1 - high amount of concurent users, 2 = max uesrs in world
	p.WriteByte(population) // Population marker - 0 = No maker, 1 = Highly populated, 2 = over populated

	return p
}
