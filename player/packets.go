package player

import (
	"crypto/rand"
	"math"
	"time"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/gopacket"
)

func playerMovePacket(charID uint32, leftOverBytes gopacket.Packet) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_MOVEMENT)
	p.WriteUint32(charID)
	p.WriteBytes(leftOverBytes)

	return p
}

func enterGame(char character.Character, channelID uint32) gopacket.Packet {
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
	p.WriteUint32(char.GetCharID())
	p.WritePaddedString(char.GetName(), 13)
	p.WriteByte(char.GetGender())
	p.WriteByte(char.GetSkin())
	p.WriteUint32(char.GetFace())
	p.WriteUint32(char.GetHair())

	p.WriteInt64(0) // Pet Cash ID

	p.WriteByte(char.GetLevel())
	p.WriteUint16(char.GetJob())
	p.WriteUint16(char.GetStr())
	p.WriteUint16(char.GetDex())
	p.WriteUint16(char.GetInt())
	p.WriteUint16(char.GetLuk())
	p.WriteUint16(char.GetHP())
	p.WriteUint16(char.GetMaxHP())
	p.WriteUint16(char.GetMP())
	p.WriteUint16(char.GetMaxMP())
	p.WriteUint16(char.GetAP())
	p.WriteUint16(char.GetSP())
	p.WriteUint32(char.GetEXP())
	p.WriteUint16(char.GetFame())

	p.WriteUint32(char.GetCurrentMap())
	p.WriteByte(char.GetCurrentMapPos())

	p.WriteByte(20) // budy list size
	p.WriteUint32(char.GetMesos())

	p.WriteByte(char.GetEquipSlotSize())
	p.WriteByte(char.GetUsetSlotSize())
	p.WriteByte(char.GetSetupSlotSize())
	p.WriteByte(char.GetEtcSlotSize())
	p.WriteByte(char.GetCashSlotSize())

	for _, v := range char.GetEquips() {
		if !nx.IsCashItem(v.GetItemID()) {
			p.WriteBytes(addEquip(v))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetEquips() {
		if nx.IsCashItem(v.GetItemID()) {
			p.WriteBytes(addEquip(v))
		}
	}

	p.WriteByte(0)

	// Inventory windows starts
	for _, v := range char.GetEquips() {
		if v.GetSlotID() > -1 {
			p.WriteBytes(addEquip(v))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.GetInvID() == 1 { // Use
			p.WriteBytes(addItem(v))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.GetInvID() == 2 { // Set-up
			p.WriteBytes(addItem(v))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.GetInvID() == 3 { // Etc
			p.WriteBytes(addItem(v))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.GetInvID() == 4 { // Cash  - not working propery :(
			p.WriteBytes(addItem(v))
		}
	}

	p.WriteByte(0)

	// Skills
	p.WriteUint16(uint16(len(char.GetSkills()))) // number of skills

	for _, v := range char.GetSkills() {
		p.WriteUint32(v.GetID())
		p.WriteUint32(uint32(v.GetLevel()))
	}

	// Quests
	p.WriteUint16(0) // # of quests

	// What are these for? Minigame and some other things?
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

	if nx.IsCashItem(item.GetItemID()) {
		p.WriteByte(byte(math.Abs(float64(item.GetSlotID() + 100))))
	} else {
		p.WriteByte(byte(math.Abs(float64(item.GetSlotID()))))
	}
	p.WriteByte(byte(item.GetItemID() / 1000000))
	p.WriteUint32(item.GetItemID())

	if nx.IsCashItem(item.GetItemID()) {
		p.WriteByte(1)
		p.WriteUint64(uint64(item.GetItemID()))
	} else {
		p.WriteByte(0)
	}

	p.WriteUint64(item.GetExpireTime())
	p.WriteByte(item.GetUpgradeSlots())
	p.WriteByte(item.GetLevel())
	p.WriteUint16(item.GetStr())
	p.WriteUint16(item.GetDex())
	p.WriteUint16(item.GetInt())
	p.WriteUint16(item.GetLuk())
	p.WriteUint16(item.GetHP())
	p.WriteUint16(item.GetMP())
	p.WriteUint16(item.GetWatk())
	p.WriteUint16(item.GetMatk())
	p.WriteUint16(item.GetWdef())
	p.WriteUint16(item.GetMdef())
	p.WriteUint16(item.GetAccuracy())
	p.WriteUint16(item.GetAvoid())
	p.WriteUint16(item.GetHands())
	p.WriteUint16(item.GetSpeed())
	p.WriteUint16(item.GetJump())
	p.WriteString(item.GetCreatorName()) // Name of creator
	p.WriteInt16(2)                      // lock, show, spikes, cape, cold protection etc ?
	return p
}

func addItem(item character.Item) gopacket.Packet {
	p := gopacket.NewPacket()

	p.WriteByte(item.GetSlotNumber()) // slot id
	p.WriteByte(2)                    // type of item e.g. equip, has amount, cash
	p.WriteUint32(item.GetItemID())   //  itemID
	p.WriteByte(0)
	p.WriteUint64(item.GetExpiration()) // expiration
	p.WriteUint16(item.GetAmount())     // amount
	p.WriteString(item.GetCreatorName())
	p.WriteUint16(item.GetFlag()) // is it sealed

	return p
}
