package player

import (
	"crypto/rand"
	"math"
	"time"

	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/Valhalla/inventory"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/gopacket"
)

func avatarSummaryWindow(charID uint32, char *character.Character, handle interfaces.ClientConn) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_AVATAR_INFO_WINDOW)
	p.WriteUint32(charID)
	p.WriteByte(char.GetLevel())
	p.WriteUint16(char.GetJob())
	p.WriteUint16(char.GetFame())

	if handle.IsAdmin() {
		p.WriteString("[Administrator]")
	} else {
		// This is player guild name
		p.WriteString("")
	}

	p.WriteBool(false) // if has pet
	p.WriteByte(0)     // wishlist count

	return p
}

func redTextMessage(msg string) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(9)
	p.WriteString(msg)

	return p
}

func guildPointsChangeMessage(ammount int32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(6)
	p.WriteInt32(ammount)

	return p
}

func fameChangeMessage(ammount int32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(4)
	p.WriteInt32(ammount)

	return p
}

// sends the [item name] has passed its expeiration date and will be removed from your inventory
func itemExpiredMessage(itemID uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(2)
	p.WriteUint32(itemID)
	return p
}

func itemExpiredMessage2(itemID uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(8)
	p.WriteByte(1)
	p.WriteUint32(itemID)
	return p
}

func mesosChangeChatMessage(ammount int32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(5)
	p.WriteInt32(ammount)

	return p
}

func unableToPickUpMessage(itemNotAvailable bool) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(0)
	if itemNotAvailable {
		p.WriteByte(0xFE)
	} else {
		p.WriteByte(0xFF)
	}

	return p
}

func dropPickUpMessage(isMesos bool, itemID, ammount uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(0)

	if isMesos {
		p.WriteUint32(ammount)
		p.WriteUint32(0)
	} else {
		p.WriteUint32(itemID)
		p.WriteUint32(ammount)
	}

	return p
}

func expGainedMessage(whiteText, appearInChat bool, ammount uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(3)
	p.WriteBool(whiteText)
	p.WriteUint32(ammount)
	p.WriteBool(appearInChat)

	return p
}

func levelUpAnimationPacket(charID uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_ANIMATION)
	p.WriteUint32(charID)
	p.WriteByte(0x00)

	return p
}

func skillBookUpdatePacket(skillID uint32, level uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_SKILL_RECORD_UPDATE)
	p.WriteByte(0x01)   // time check?
	p.WriteUint16(0x01) // number of skills to update
	p.WriteUint32(skillID)
	p.WriteUint32(level)
	p.WriteByte(0x01)

	return p
}

func receivedDmgPacket(charID uint32, ammount uint32, dmgType byte, mobID uint32, hit byte, reduction byte, stance byte) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_TAKE_DMG)
	p.WriteUint32(charID)
	p.WriteByte(dmgType)

	if dmgType == 0xFE {
		p.WriteUint32(ammount)
		p.WriteUint32(ammount)
	} else {
		p.WriteUint32(0) // ?
		p.WriteUint32(mobID)
		p.WriteByte(hit)
		p.WriteByte(stance)
		p.WriteUint32(0)       // ?
		p.WriteUint32(ammount) // skill id of attack?
	}

	return p
}

// Maybe split this into byte, uint16 & uint32 forms by taking interace{} and reflecting value type
func statChangePacket(byPlayer bool, stat uint32, value uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_STAT_CHANGE)
	p.WriteBool(byPlayer)
	p.WriteUint32(stat)
	p.WriteUint32(value)

	return p
}

func statNoChangePacket() gopacket.Packet {
	p := gopacket.NewPacket()
	// Continue game opcode is part of inventory opcode list?
	p.WriteByte(constants.SEND_CHANNEL_INVENTORY_OPERATION)
	p.WriteByte(0x01)
	p.WriteByte(0x00)
	p.WriteByte(0x00)

	return p
}

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

	for _, v := range char.GetItems() {
		if v.GetSlotNumber() < 0 && v.GetInvID() == 1 && !nx.IsCashItem(v.GetItemID()) {
			p.WriteBytes(addEquip(v))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.GetSlotNumber() < 0 && v.GetInvID() == 1 && nx.IsCashItem(v.GetItemID()) {
			p.WriteBytes(addEquip(v))
		}
	}

	p.WriteByte(0)

	// Inventory windows starts
	for _, v := range char.GetItems() {
		if v.GetSlotNumber() > -1 && v.GetInvID() == 1 {
			p.WriteBytes(addEquip(v))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.GetInvID() == 2 { // Use
			p.WriteBytes(addItem(v))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.GetInvID() == 3 { // Set-up
			p.WriteBytes(addItem(v))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.GetInvID() == 4 { // Etc
			p.WriteBytes(addItem(v))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.GetInvID() == 5 { // Cash  - not working propery :(
			p.WriteBytes(addItem(v))
		}
	}

	p.WriteByte(0)

	// Skills
	p.WriteUint16(uint16(len(char.GetSkills()))) // number of skills

	for id, level := range char.GetSkills() {
		p.WriteUint32(id)
		p.WriteUint32(level)
	}

	// Quests
	p.WriteUint16(0) // # of quests?

	// What are these for? Minigame record and some other things?
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

func addEquip(item inventory.Item) gopacket.Packet {
	p := gopacket.NewPacket()

	if nx.IsCashItem(item.GetItemID()) {
		p.WriteByte(byte(math.Abs(float64(item.GetSlotNumber() + 100))))
	} else {
		p.WriteByte(byte(math.Abs(float64(item.GetSlotNumber()))))
	}
	p.WriteByte(byte(item.GetItemID() / 1000000))
	p.WriteUint32(item.GetItemID())

	if nx.IsCashItem(item.GetItemID()) {
		p.WriteByte(1)
		p.WriteUint64(uint64(item.GetItemID()))
	} else {
		p.WriteByte(0)
	}

	p.WriteUint64(item.GetExpirationTime())
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

func addItem(item inventory.Item) gopacket.Packet {
	p := gopacket.NewPacket()

	p.WriteByte(byte(item.GetSlotNumber())) // slot id
	p.WriteByte(2)                          // type of item e.g. equip, has amount, cash
	p.WriteUint32(item.GetItemID())         //  itemID
	p.WriteByte(0)
	p.WriteUint64(item.GetExpirationTime()) // expiration
	p.WriteUint16(item.GetAmount())         // amount
	p.WriteString(item.GetCreatorName())
	p.WriteUint16(item.GetFlag()) // is it sealed

	return p
}
