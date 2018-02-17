package packets

import (
	"crypto/rand"
	"time"

	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/nx"
	"github.com/Hucaru/gopacket"
)

func ChangeMap(mapID uint32, channelID uint32, mapPos byte, hp uint16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_WARP_TO_MAP)
	p.WriteUint32(channelID)
	p.WriteByte(1) // character portal counter
	p.WriteByte(0) // Is connecting
	p.WriteUint32(mapID)
	p.WriteByte(mapPos)
	p.WriteUint16(hp)
	p.WriteByte(0) // ?

	return p
}

func SpawnGame(char character.Character, channelID uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_WARP_TO_MAP)
	p.WriteUint32(channelID)
	p.WriteByte(0) // character portal counter
	p.WriteByte(1) // Is connecting

	randomBytes := make([]byte, 4)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err.Error())
	}
	p.WriteBytes(randomBytes)
	p.WriteBytes(randomBytes)
	p.WriteBytes(randomBytes)
	p.WriteBytes(randomBytes)
	p.WriteBytes([]byte{0xFF, 0xFF}) // seperators? For what?
	p.WriteUint32(char.CharID)
	p.WritePaddedString(char.Name, 13)
	p.WriteByte(char.Gender)
	p.WriteByte(char.Skin)
	p.WriteUint32(char.Face)
	p.WriteUint32(char.Hair)

	p.WriteInt64(0) // Pet Cash ID

	p.WriteByte(char.Level)
	p.WriteUint16(char.Job)
	p.WriteUint16(char.Str)
	p.WriteUint16(char.Dex)
	p.WriteUint16(char.Intt)
	p.WriteUint16(char.Luk)
	p.WriteUint16(char.HP)
	p.WriteUint16(char.MaxHP)
	p.WriteUint16(char.MP)
	p.WriteUint16(char.MaxMP)
	p.WriteUint16(char.AP)
	p.WriteUint16(char.SP)
	p.WriteUint32(char.EXP)
	p.WriteUint16(char.Fame)

	p.WriteUint32(char.CurrentMap)
	p.WriteByte(char.CurrentMapPos)

	p.WriteByte(20) // budy list size
	p.WriteUint32(char.Mesos)

	p.WriteByte(char.EquipSlotSize)
	p.WriteByte(char.UsetSlotSize) // User inv size
	p.WriteByte(char.SetupSlotSize)
	p.WriteByte(char.EtcSlotSize)
	p.WriteByte(char.CashSlotSize)

	// Equips -50 -> -1 normal equips
	for _, v := range char.Equips {
		if v.SlotID < 0 && v.SlotID > -20 {
			p.WriteBytes(addEquip(v))
		}
	}

	p.WriteByte(0)

	// Cash item equip covers -150 to -101 maybe?
	for _, v := range char.Equips {
		if v.SlotID < -100 {
			p.WriteBytes(addEquip(v))
		}
	}

	p.WriteByte(0)

	for _, v := range char.Equips {
		if v.SlotID > -1 {
			p.WriteBytes(addEquip(v)) // there is a caveat for adding a cash item
		}
	}

	p.WriteByte(0)

	// use
	p.WriteByte(1)         // slot id
	p.WriteByte(2)         // type of item e.g. equip, has amount, cash
	p.WriteUint32(2070006) //  itemID
	p.WriteByte(0)
	p.WriteUint64(0)   // expiration
	p.WriteUint16(200) // amount
	p.WriteUint16(0)   // string with name of creator
	p.WriteUint16(0)   // is it sealed

	// use
	p.WriteByte(2)         // slot id
	p.WriteByte(2)         // type of item
	p.WriteUint32(2000003) //  itemID
	p.WriteByte(0)
	p.WriteUint64(0)
	p.WriteUint16(200) // amount
	p.WriteUint16(0)
	p.WriteUint16(0) // is it sealed

	p.WriteByte(0) // Inventory tab move forward swap

	p.WriteByte(1)         // slot id
	p.WriteByte(2)         // type of item
	p.WriteUint32(3010000) //  itemID
	p.WriteByte(0)
	p.WriteUint64(0)
	p.WriteUint16(1) // amount
	p.WriteUint16(0)
	p.WriteUint16(0) // is it sealed

	p.WriteByte(0) // Inventory tab move forward swap

	// etc
	p.WriteByte(1)         // slot id
	p.WriteByte(2)         // type of item
	p.WriteUint32(4000000) //  itemID
	p.WriteByte(0)
	p.WriteUint64(0)
	p.WriteUint16(200) // amount
	p.WriteUint16(0)
	p.WriteUint16(0) // is it sealed

	p.WriteByte(0) // Inventory tab move forward swap

	// cash pet item :( not working atm
	p.WriteByte(1)         // slot id
	p.WriteByte(2)         // Type of item (1 means it is an equip, 2 means inv?, 3 means ?)
	p.WriteUint32(5000004) //  itemID
	p.WriteByte(0)
	// p.WriteUint32(5000004)
	p.WriteUint64(0)
	p.WriteUint16(1) // amount
	p.WriteUint16(0)
	p.WriteUint16(0) // is it sealed

	p.WriteByte(0)

	// Skills
	p.WriteUint16(uint16(len(char.Skills))) // number of skills

	for _, v := range char.Skills {
		p.WriteUint32(v.SkillID)
		p.WriteUint32(uint32(v.Level))
	}

	// Quests
	p.WriteUint16(0) // # of quests

	// Minigame
	p.WriteUint16(0)
	p.WriteUint32(0)
	p.WriteUint32(0)
	p.WriteUint32(0)
	p.WriteUint32(0)
	p.WriteUint32(0)

	p.WriteUint64(0)
	p.WriteUint64(0)
	p.WriteUint64(0)
	p.WriteUint64(0)
	p.WriteUint64(0)

	p.WriteInt64(time.Now().Unix())

	return p
}

func SpawnNPC(index uint32, npc nx.Life) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(0x97)
	p.WriteUint32(index)
	p.WriteUint32(npc.ID)
	p.WriteInt16(npc.X)
	p.WriteInt16(npc.Y)

	if npc.F == 0 {
		p.WriteByte(1)
	} else {
		p.WriteByte(0)
	}

	p.WriteInt16(npc.Fh)
	p.WriteInt16(npc.Rx0)
	p.WriteInt16(npc.Rx1)

	p.WriteByte(0x9B)
	p.WriteByte(0x1)
	p.WriteUint32(npc.ID)
	p.WriteUint32(npc.ID)
	p.WriteInt16(npc.X)
	p.WriteInt16(npc.Y)

	if npc.F == 0 {
		p.WriteByte(1)
	} else {
		p.WriteByte(0)
	}
	if npc.F == 0 {
		p.WriteByte(1)
	} else {
		p.WriteByte(0)
	}
	p.WriteInt16(npc.Fh)
	p.WriteInt16(npc.Rx0)
	p.WriteInt16(npc.Rx1)

	return p
}

func sendLevelUpAnimation(charID byte) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_LEVEL_UP_ANIMATION)
	p.WriteByte(charID) // charid

	return p
}

func spawnDoor(x int16, y int16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_SPAWN_DOOR)
	p.WriteByte(0)  // ?
	p.WriteInt32(0) // ?
	p.WriteInt16(x) // x pos
	p.WriteInt16(y) // y pos

	return p
}

func removeDoor(x int16, y int16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_SPAWN_DOOR)
	p.WriteByte(0)  // ?
	p.WriteInt32(0) // ?
	p.WriteInt16(x) // x pos
	p.WriteInt16(y) // y pos

	return p
}

func quizQuestionAndAnswer(isQuestion bool, questionSet byte, questionNumber int16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_QUIZ_Q_AND_A)
	if isQuestion {
		p.WriteByte(0x01)
	} else {
		p.WriteByte(0x00)
	}
	p.WriteByte(questionSet)
	p.WriteInt16(questionNumber)

	return p
}
