package game

import (
	"strconv"

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

func packetLoginResponce(result byte, userID int32, gender byte, isAdmin bool, username string, isBanned int) mpacket.Packet {
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

func packetLoginMigrateClient(ip []byte, port int16, charID int32) mpacket.Packet {
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

func packetLoginSendBadMigrate() mpacket.Packet {
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

func packetLoginDisplayCharacters(characters []character) mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginCharacterData)
	pac.WriteByte(0) // ?

	if len(characters) < 4 && len(characters) > 0 {
		pac.WriteByte(byte(len(characters)))

		for _, c := range characters {
			loginWritePlayerCharacter(&pac, c.id, c)
		}
	} else {
		pac.WriteByte(0)
	}

	return pac
}

func packetLoginNameCheck(name string, nameFound int) mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginNameCheckResult)
	pac.WriteString(name)

	if nameFound > 0 {
		pac.WriteByte(0x1) // 0 = good name, 1 = bad name
	} else {
		pac.WriteByte(0x0)
	}

	return pac
}

func packetLoginCreatedCharacter(success bool, char character) mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginNewCharacterGood)

	if success {
		pac.WriteByte(0x0) // if creation was sucessfull - 0 = good, 1 = bad
		loginWritePlayerCharacter(&pac, char.id, char)
	} else {
		pac.WriteByte(0x1)
	}

	return pac
}

func packetLoginDeleteCharacter(charID int32, deleted bool, hacking bool) mpacket.Packet {
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

func loginWritePlayerCharacter(pac *mpacket.Packet, pos int32, char character) {
	pac.WriteInt32(pos)

	name := char.name

	if len(name) > 13 {
		name = name[:13]
	}

	padSize := 13 - len(name)

	pac.WriteBytes([]byte(name))
	for i := 0; i < padSize; i++ {
		pac.WriteByte(0x0)
	}

	pac.WriteByte(char.gender) //gender
	pac.WriteByte(char.skin)   // skin
	pac.WriteInt32(char.face)  // face
	pac.WriteInt32(char.hair)  // Hair

	pac.WriteInt64(0x0) // Pet cash ID

	pac.WriteByte(char.level)  // level
	pac.WriteInt16(char.job)   // Job
	pac.WriteInt16(char.str)   // str
	pac.WriteInt16(char.dex)   // dex
	pac.WriteInt16(char.intt)  // int
	pac.WriteInt16(char.luk)   // luk
	pac.WriteInt16(char.hp)    // hp
	pac.WriteInt16(char.maxHP) // max hp
	pac.WriteInt16(char.mp)    // mp
	pac.WriteInt16(char.maxMP) // max mp
	pac.WriteInt16(char.ap)    // ap
	pac.WriteInt16(char.sp)    // sp
	pac.WriteInt32(char.exp)   // exp
	pac.WriteInt16(char.fame)  // fame

	pac.WriteInt32(char.mapID) // map id
	pac.WriteByte(char.mapPos) // map

	pac.WriteBytes(writeDisplayCharacter(char))

	pac.WriteInt32(0) // if character is selected and which one
	pac.WriteByte(1)  // Rankings
	pac.WriteInt32(1) // world ranking position
	pac.WriteInt32(2) // increase / decrease amount
	pac.WriteInt32(3) // class ranking position
	pac.WriteInt32(4) // increase / decrease amount
}

func packetLoginWorldListing(worldIndex byte, w world) mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginWorldList)
	pac.WriteByte(worldIndex) // world id
	pac.WriteString(w.name)   // World name -
	pac.WriteByte(w.ribbon)   // Ribbon on world - 0 = normal, 1 = event, 2 = new, 3 = hot
	pac.WriteString(w.message)
	pac.WriteByte(0)                     // ? exp event notification?
	pac.WriteByte(byte(len(w.channels))) // number of channels

	for i, v := range w.channels {
		pac.WriteString(w.name + "-" + strconv.Itoa(i+1))
		if v.maxPop == 0 {
			pac.WriteInt32(0)
		} else {
			pac.WriteInt32(int32(1200.0 * (float64(v.pop) / float64(v.maxPop))))
		}
		pac.WriteByte(worldIndex)
		pac.WriteByte(byte(i + 1)) // channel id
		pac.WriteByte(0)           // ?
	}

	return pac
}

func packetLoginEndWorldList() mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginWorldList)
	pac.WriteByte(0xFF)

	return pac
}

func packetLoginWorldInfo(warning byte, population byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendLoginWorldMeta)
	p.WriteByte(warning)    // Warning - 0 = no warning, 1 - high amount of concurent users, 2 = max users in world
	p.WriteByte(population) // Population marker - 0 = No maker, 1 = Highly populated, 2 = over populated

	return p
}

func packetLoginReturnFromChannel() mpacket.Packet {
	pac := mpacket.CreateWithOpcode(opcode.SendLoginRestarter)
	pac.WriteByte(0x01)

	return pac
}
