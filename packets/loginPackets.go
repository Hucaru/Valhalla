package packets

import (
	"strconv"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func LoginReturnFromChannel() maplepacket.Packet {
	pac := maplepacket.NewPacket()
	pac.WriteByte(constants.SendLoginRestarter)
	pac.WriteByte(0x01)

	return pac
}

func LoginResponce(result byte, userID int32, gender byte, isAdmin byte, username string, isBanned int) maplepacket.Packet {
	pac := maplepacket.NewPacket()
	pac.WriteByte(constants.SendLoginResponce)
	pac.WriteByte(result)
	pac.WriteByte(0x00)
	pac.WriteInt32(0)

	if result <= 0x01 {
		pac.WriteInt32(userID)
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

func LoginMigrateClient(ip []byte, port int16, charID int32) maplepacket.Packet {
	pac := maplepacket.NewPacket()
	pac.WriteByte(constants.SendLoginCharacterMigrate)
	pac.WriteByte(0x00)
	pac.WriteByte(0x00)
	pac.WriteBytes(ip)
	pac.WriteInt16(port)
	pac.WriteInt32(charID)
	pac.WriteByte(byte(0) | byte(1<<0))
	pac.WriteInt32(1)

	return pac
}

func LoginSendBadMigrate() maplepacket.Packet {
	pac := maplepacket.NewPacket()
	pac.WriteByte(constants.SendLoginCharacterMigrate)
	pac.WriteByte(0x00) // flipping these 2 bytes makes the character select screen do nothing it appears
	pac.WriteByte(0x00)
	pac.WriteBytes([]byte{0, 0, 0, 0})
	pac.WriteInt16(0)
	pac.WriteInt32(8)
	pac.WriteByte(byte(0) | byte(1<<0))
	pac.WriteInt32(1)

	return pac
}

func LoginDisplayCharacters(characters []character.Character) maplepacket.Packet {
	pac := maplepacket.NewPacket()
	pac.WriteByte(constants.SendLoginCharacterData)
	pac.WriteByte(0) // ?

	if len(characters) < 4 && len(characters) > 0 {
		pac.WriteByte(byte(len(characters)))

		for _, c := range characters {
			LoginWritePlayerCharacter(&pac, c.GetCharID(), c)
		}
	} else {
		pac.WriteByte(0)
	}

	return pac
}

func LoginNameCheck(name string, nameFound int) maplepacket.Packet {
	pac := maplepacket.NewPacket()
	pac.WriteByte(constants.SendLoginNameCheckResult)
	pac.WriteString(name)

	if nameFound > 0 {
		pac.WriteByte(0x1) // 0 = good name, 1 = bad name
	} else {
		pac.WriteByte(0x0)
	}

	return pac
}

func LoginCreatedCharacter(success bool, character character.Character) maplepacket.Packet {
	pac := maplepacket.NewPacket()
	pac.WriteByte(constants.SendLoginNewCharacterGood)

	if success {
		pac.WriteByte(0x0) // if creation was sucessfull - 0 = good, 1 = bad
		LoginWritePlayerCharacter(&pac, character.GetCharID(), character)
	} else {
		pac.WriteByte(0x1)
	}

	return pac
}

func LoginDeleteCharacter(charID int32, deleted bool, hacking bool) maplepacket.Packet {
	pac := maplepacket.NewPacket()
	pac.WriteByte(constants.SendLoginDeleteCharacter)
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

func LoginWritePlayerCharacter(pac *maplepacket.Packet, pos int32, char character.Character) {
	pac.WriteInt32(pos)

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
	pac.WriteInt32(char.GetFace())  // face
	pac.WriteInt32(char.GetHair())  // Hair

	pac.WriteInt64(0x0) // Pet cash ID

	pac.WriteByte(char.GetLevel())  // level
	pac.WriteInt16(char.GetJob())   // Job
	pac.WriteInt16(char.GetStr())   // str
	pac.WriteInt16(char.GetDex())   // dex
	pac.WriteInt16(char.GetInt())   // int
	pac.WriteInt16(char.GetLuk())   // luk
	pac.WriteInt16(char.GetHP())    // hp
	pac.WriteInt16(char.GetMaxHP()) // max hp
	pac.WriteInt16(char.GetMP())    // mp
	pac.WriteInt16(char.GetMaxMP()) // max mp
	pac.WriteInt16(char.GetAP())    // ap
	pac.WriteInt16(char.GetSP())    // sp
	pac.WriteInt32(char.GetEXP())   // exp
	pac.WriteInt16(char.GetFame())  // fame

	pac.WriteInt32(char.GetCurrentMap())   // map id
	pac.WriteByte(char.GetCurrentMapPos()) // map

	pac.WriteBytes(writeDisplayCharacter(char))

	pac.WriteInt32(0) // is character is selected and which one
	pac.WriteByte(1)  // Rankings
	pac.WriteInt32(1) // world ranking position
	pac.WriteInt32(2) // increase / decrease amount
	pac.WriteInt32(3) // class ranking position
	pac.WriteInt32(4) // increase / decrease amount
}

func LoginWorldListing(worldIndex byte) maplepacket.Packet {
	pac := maplepacket.NewPacket()
	pac.WriteByte(constants.SendLoginWorldList)
	pac.WriteByte(worldIndex)                          // world id
	pac.WriteString(constants.WORLD_NAMES[worldIndex]) // World name -
	pac.WriteByte(3)                                   // Ribbon on world - 0 = normal, 1 = event, 2 = new, 3 = hot
	pac.WriteString("test")
	pac.WriteByte(0)  // ? exp event notification?
	pac.WriteByte(20) // number of channels

	maxPopulation := 150
	population := 50

	for j := 1; j < 21; j++ {
		pac.WriteString(constants.WORLD_NAMES[worldIndex] + "-" + strconv.Itoa(j))     // channel name
		pac.WriteInt32(int32(1200.0 * (float64(population) / float64(maxPopulation)))) // Population
		pac.WriteByte(worldIndex)                                                      // world id
		pac.WriteByte(byte(j))                                                         // channel id
		pac.WriteByte(byte(j - 1))                                                     //?
	}

	return pac
}

func LoginEndWorldList() maplepacket.Packet {
	pac := maplepacket.NewPacket()
	pac.WriteByte(constants.SendLoginWorldList)
	pac.WriteByte(0xFF)

	return pac
}

func LoginWorldInfo(warning byte, population byte) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SendLoginWorldMeta)
	p.WriteByte(warning)    // Warning - 0 = no warning, 1 - high amount of concurent users, 2 = max uesrs in world
	p.WriteByte(population) // Population marker - 0 = No maker, 1 = Highly populated, 2 = over populated

	return p
}
