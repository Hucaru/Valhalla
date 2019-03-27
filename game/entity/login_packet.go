package entity

import (
	"strconv"

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

func PacketLoginResponce(result byte, userID int32, gender byte, isAdmin bool, username string, isBanned int) mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginResponce)
	pac.WriteByte(result)
	pac.WriteByte(0x00)
	pac.WriteInt32(0)

	if result <= 0x01 {
		pac.WriteInt32(userID)
		pac.WriteByte(gender)
		// pac.WriteByte(isAdmin)
		pac.WriteBool(isAdmin)
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

func PacketLoginMigrateClient(ip []byte, port int16, charID int32) mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginCharacterMigrate)
	pac.WriteByte(0x00)
	pac.WriteByte(0x00)
	pac.WriteBytes(ip)
	pac.WriteInt16(port)
	pac.WriteInt32(charID)
	pac.WriteByte(byte(0) | byte(1<<0))
	pac.WriteInt32(1)

	return pac
}

func PacketLoginSendBadMigrate() mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginCharacterMigrate)
	pac.WriteByte(0x00) // flipping these 2 bytes makes the character select screen do nothing it appears
	pac.WriteByte(0x00)
	pac.WriteBytes([]byte{0, 0, 0, 0})
	pac.WriteInt16(0)
	pac.WriteInt32(8)
	pac.WriteByte(byte(0) | byte(1<<0))
	pac.WriteInt32(1)

	return pac
}

func PacketLoginDisplayCharacters(characters []Character) mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginCharacterData)
	pac.WriteByte(0) // ?

	if len(characters) < 4 && len(characters) > 0 {
		pac.WriteByte(byte(len(characters)))

		for _, c := range characters {
			loginWritePlayerCharacter(&pac, c.ID, c)
		}
	} else {
		pac.WriteByte(0)
	}

	return pac
}

func PacketLoginNameCheck(name string, nameFound int) mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginNameCheckResult)
	pac.WriteString(name)

	if nameFound > 0 {
		pac.WriteByte(0x1) // 0 = good name, 1 = bad name
	} else {
		pac.WriteByte(0x0)
	}

	return pac
}

func PacketLoginCreatedCharacter(success bool, character Character) mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginNewCharacterGood)

	if success {
		pac.WriteByte(0x0) // if creation was sucessfull - 0 = good, 1 = bad
		loginWritePlayerCharacter(&pac, character.ID, character)
	} else {
		pac.WriteByte(0x1)
	}

	return pac
}

func PacketLoginDeleteCharacter(charID int32, deleted bool, hacking bool) mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginDeleteCharacter)
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

func loginWritePlayerCharacter(pac *mpacket.Packet, pos int32, char Character) {
	pac.WriteInt32(pos)

	name := char.Name

	if len(name) > 13 {
		name = name[:13]
	}

	padSize := 13 - len(name)

	pac.WriteBytes([]byte(name))
	for i := 0; i < padSize; i++ {
		pac.WriteByte(0x0)
	}

	pac.WriteByte(char.Gender) //gender
	pac.WriteByte(char.Skin)   // skin
	pac.WriteInt32(char.Face)  // face
	pac.WriteInt32(char.Hair)  // Hair

	pac.WriteInt64(0x0) // Pet cash ID

	pac.WriteByte(char.Level)  // level
	pac.WriteInt16(char.Job)   // Job
	pac.WriteInt16(char.Str)   // str
	pac.WriteInt16(char.Dex)   // dex
	pac.WriteInt16(char.Int)   // int
	pac.WriteInt16(char.Luk)   // luk
	pac.WriteInt16(char.HP)    // hp
	pac.WriteInt16(char.MaxHP) // max hp
	pac.WriteInt16(char.MP)    // mp
	pac.WriteInt16(char.MaxMP) // max mp
	pac.WriteInt16(char.AP)    // ap
	pac.WriteInt16(char.SP)    // sp
	pac.WriteInt32(char.EXP)   // exp
	pac.WriteInt16(char.Fame)  // fame

	pac.WriteInt32(char.MapID) // map id
	pac.WriteByte(char.MapPos) // map

	pac.WriteBytes(writeDisplayCharacter(char))

	pac.WriteInt32(0) // if character is selected and which one
	pac.WriteByte(1)  // Rankings
	pac.WriteInt32(1) // world ranking position
	pac.WriteInt32(2) // increase / decrease amount
	pac.WriteInt32(3) // class ranking position
	pac.WriteInt32(4) // increase / decrease amount
}

func PacketLoginWorldListing(worldIndex byte, w World) mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginWorldList)
	pac.WriteByte(worldIndex) // world id
	pac.WriteString(w.Name)   // World name -
	pac.WriteByte(w.Ribbon)   // Ribbon on world - 0 = normal, 1 = event, 2 = new, 3 = hot
	pac.WriteString(w.Message)
	pac.WriteByte(0)                     // ? exp event notification?
	pac.WriteByte(byte(len(w.Channels))) // number of channels

	for i, v := range w.Channels {
		pac.WriteString(w.Name + "-" + strconv.Itoa(i+1))
		pac.WriteInt32(int32(1200.0 * (float64(v.Pop) / float64(v.MaxPop))))
		pac.WriteByte(worldIndex)
		pac.WriteByte(byte(i + 1)) // channel id
		pac.WriteByte(0)           // ?
	}

	return pac
}

func PacketLoginEndWorldList() mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginWorldList)
	pac.WriteByte(0xFF)

	return pac
}

func PacketLoginWorldInfo(warning byte, population byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendLoginWorldMeta)
	p.WriteByte(warning)    // Warning - 0 = no warning, 1 - high amount of concurent users, 2 = max uesrs in world
	p.WriteByte(population) // Population marker - 0 = No maker, 1 = Highly populated, 2 = over populated

	return p
}

func PacketLoginReturnFromChannel() mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginRestarter)
	pac.WriteByte(0x01)

	return pac
}
