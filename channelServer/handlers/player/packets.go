package player

import (
	"crypto/rand"
	"math"
	"time"

	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/nx"
	"github.com/Hucaru/gopacket"
)

func changeMap(mapID uint32, channelID uint32, mapPos byte, hp uint16) gopacket.Packet {
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

func spawnGame(char character.Character, channelID uint32) gopacket.Packet {
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

	for _, v := range char.Equips {
		if !nx.IsCashItem(v.ItemID) {
			p.WriteBytes(addEquip(v))
		}
	}

	p.WriteByte(0)

	for _, v := range char.Equips {
		if nx.IsCashItem(v.ItemID) {
			p.WriteBytes(addEquip(v))
		}
	}

	p.WriteByte(0)

	// Inventory windows starts
	for _, v := range char.Equips {
		if v.SlotID > -1 {
			p.WriteBytes(addEquip(v))
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

func addEquip(item character.Equip) gopacket.Packet {
	p := gopacket.NewPacket()

	if nx.IsCashItem(item.ItemID) {
		p.WriteByte(byte(math.Abs(float64(item.SlotID + 100))))
	} else {
		p.WriteByte(byte(math.Abs(float64(item.SlotID))))
	}
	p.WriteByte(byte(item.ItemID / 1000000))
	p.WriteUint32(item.ItemID)

	if nx.IsCashItem(item.ItemID) {
		p.WriteByte(1)
		p.WriteUint64(uint64(item.ItemID))
	} else {
		p.WriteByte(0)
	}

	p.WriteUint64(item.ExpireTime)
	p.WriteByte(item.UpgradeSlots)
	p.WriteByte(item.Level)
	p.WriteUint16(item.Str)
	p.WriteUint16(item.Dex)
	p.WriteUint16(item.Intt)
	p.WriteUint16(item.Luk)
	p.WriteUint16(item.HP)
	p.WriteUint16(item.MP)
	p.WriteUint16(item.Watk)
	p.WriteUint16(item.Matk)
	p.WriteUint16(item.Wdef)
	p.WriteUint16(item.Mdef)
	p.WriteUint16(item.Accuracy)
	p.WriteUint16(item.Avoid)
	p.WriteUint16(item.Hands)
	p.WriteUint16(item.Speed)
	p.WriteUint16(item.Jump)
	p.WriteString(item.OwnerName) // Name of creator
	p.WriteInt16(2)               // lock, show, spikes, cape, cold protection etc ?
	return p
}

func addUseItem() gopacket.Packet {
	p := gopacket.NewPacket()

	p.WriteByte(1)         // slot id
	p.WriteByte(2)         // type of item e.g. equip, has amount, cash
	p.WriteUint32(2070006) //  itemID
	p.WriteByte(0)
	p.WriteUint64(0)   // expiration
	p.WriteUint16(200) // amount
	p.WriteUint16(0)   // string with name of creator
	p.WriteUint16(0)   // is it sealed

	return p
}
