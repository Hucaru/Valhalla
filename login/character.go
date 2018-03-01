package login

import (
	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/gopacket"
)

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
