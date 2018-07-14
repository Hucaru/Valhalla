package packets

import (
	"crypto/rand"
	"fmt"
	"math"
	"time"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/inventory"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/nx"
)

func PlayerReceivedDmg(charID int32, ammount int32, dmgType byte, mobID int32, hit byte, reduction byte, stance byte) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_TAKE_DMG)
	p.WriteInt32(charID)
	p.WriteByte(dmgType)

	if dmgType == 0xFE {
		p.WriteInt32(ammount)
		p.WriteInt32(ammount)
	} else {
		p.WriteInt32(0) // ?
		p.WriteInt32(mobID)
		p.WriteByte(hit)
		p.WriteByte(stance)
		p.WriteInt32(0)       // ?
		p.WriteInt32(ammount) // skill id of attack?
	}

	return p
}

func PlayerLevelUpAnimation(charID int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_ANIMATION)
	p.WriteInt32(charID)
	p.WriteByte(0x00)

	return p
}

func PlayerMove(charID int32, leftOverBytes maplepacket.Packet) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_MOVEMENT)
	p.WriteInt32(charID)
	p.WriteBytes(leftOverBytes)

	return p
}

func PlayerEmoticon(playerID int32, emotion int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_EMOTION)
	p.WriteInt32(playerID)
	p.WriteInt32(emotion)

	return p
}

func PlayerSkillBookUpdate(skillID int32, level int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_SKILL_RECORD_UPDATE)
	p.WriteByte(0x01)  // time check?
	p.WriteInt16(0x01) // number of skills to update
	p.WriteInt32(skillID)
	p.WriteInt32(level)
	p.WriteByte(0x01)

	return p
}

func PlayerStatChange(byPlayer bool, stat int32, value int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_STAT_CHANGE)
	p.WriteBool(byPlayer)
	p.WriteInt32(stat)
	p.WriteInt32(value)

	return p
}

func PlayerStatNoChange() maplepacket.Packet {
	p := maplepacket.NewPacket()
	// Continue game opcode is part of inventory opcode list?
	p.WriteByte(constants.SEND_CHANNEL_INVENTORY_OPERATION)
	p.WriteByte(0x01)
	p.WriteByte(0x00)
	p.WriteByte(0x00)

	return p
}

func PlayerAvatarSummaryWindow(charID int32, char character.Character, guildName string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_AVATAR_INFO_WINDOW)
	p.WriteInt32(charID)
	p.WriteByte(char.GetLevel())
	p.WriteInt16(char.GetJob())
	p.WriteInt16(char.GetFame())

	p.WriteString(guildName)

	p.WriteBool(false) // if has pet
	p.WriteByte(0)     // wishlist count

	return p
}

func PlayerEnterGame(char character.Character, channelID int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_WARP_TO_MAP)
	p.WriteInt32(channelID)
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
	p.WriteInt32(char.GetCharID())
	p.WritePaddedString(char.GetName(), 13)
	p.WriteByte(char.GetGender())
	p.WriteByte(char.GetSkin())
	p.WriteInt32(char.GetFace())
	p.WriteInt32(char.GetHair())

	p.WriteInt64(0) // Pet Cash ID

	p.WriteByte(char.GetLevel())
	p.WriteInt16(char.GetJob())
	p.WriteInt16(char.GetStr())
	p.WriteInt16(char.GetDex())
	p.WriteInt16(char.GetInt())
	p.WriteInt16(char.GetLuk())
	p.WriteInt16(char.GetHP())
	p.WriteInt16(char.GetMaxHP())
	p.WriteInt16(char.GetMP())
	p.WriteInt16(char.GetMaxMP())
	p.WriteInt16(char.GetAP())
	p.WriteInt16(char.GetSP())
	p.WriteInt32(char.GetEXP())
	p.WriteInt16(char.GetFame())

	p.WriteInt32(char.GetCurrentMap())
	p.WriteByte(char.GetCurrentMapPos())

	p.WriteByte(20) // budy list size
	p.WriteInt32(char.GetMesos())

	p.WriteByte(char.GetEquipSlotSize())
	p.WriteByte(char.GetUsetSlotSize())
	p.WriteByte(char.GetSetupSlotSize())
	p.WriteByte(char.GetEtcSlotSize())
	p.WriteByte(char.GetCashSlotSize())

	for _, v := range char.GetItems() {
		if v.GetSlotID() < 0 && v.GetInvID() == 1 && !nx.IsCashItem(v.GetItemID()) {
			p.WriteBytes(addItem(v, true))
		}
	}

	p.WriteByte(0)

	// Equips
	for _, v := range char.GetItems() {
		if v.GetSlotID() < 0 && v.GetInvID() == 1 && nx.IsCashItem(v.GetItemID()) {
			p.WriteBytes(addItem(v, true))
		}
	}

	p.WriteByte(0)

	// Inventory windows starts
	for _, v := range char.GetItems() {
		if v.GetSlotID() > -1 && v.GetInvID() == 1 {
			p.WriteBytes(addItem(v, true))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.GetInvID() == 2 { // Use
			p.WriteBytes(addItem(v, true))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.GetInvID() == 3 { // Set-up
			p.WriteBytes(addItem(v, true))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.GetInvID() == 4 { // Etc
			p.WriteBytes(addItem(v, true))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.GetInvID() == 5 { // Cash  - not working propery :(
			p.WriteBytes(addItem(v, true))
		}
	}

	p.WriteByte(0)

	// Skills
	p.WriteInt16(int16(len(char.GetSkills()))) // number of skills

	for id, level := range char.GetSkills() {
		p.WriteInt32(id)
		p.WriteInt32(level)
	}

	// Quests
	p.WriteInt16(0) // # of quests?

	// What are these for? Minigame record and some other things?
	p.WriteInt16(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)

	p.WriteUint64(0)
	p.WriteUint64(0)
	p.WriteUint64(0)
	p.WriteUint64(0)
	p.WriteUint64(0)

	p.WriteInt64(time.Now().Unix())

	return p
}

func addItem(item inventory.Item, needInvID bool) maplepacket.Packet {
	p := maplepacket.NewPacket()

	if needInvID {
		if nx.IsCashItem(item.GetItemID()) && item.GetSlotID() < 0 {
			p.WriteByte(byte(math.Abs(float64(item.GetSlotID() + 100))))
		} else {
			p.WriteByte(byte(math.Abs(float64(item.GetSlotID()))))
		}

		p.WriteByte(item.GetInvID())
	} else {
		p.WriteInt16(item.GetSlotID())
		p.WriteByte(item.GetInvID())
	}

	p.WriteInt32(item.GetItemID())

	if nx.IsCashItem(item.GetItemID()) {
		p.WriteByte(1)
		p.WriteUint64(uint64(item.GetItemID()))
	} else {
		p.WriteByte(0)
	}

	p.WriteUint64(item.GetExpirationTime())

	switch item.GetInvID() {
	case 1:
		p.WriteByte(item.GetUpgradeSlots())
		p.WriteByte(item.GetScrollLevel())
		p.WriteInt16(item.GetStr())
		p.WriteInt16(item.GetDex())
		p.WriteInt16(item.GetInt())
		p.WriteInt16(item.GetLuk())
		p.WriteInt16(item.GetHP())
		p.WriteInt16(item.GetMP())
		p.WriteInt16(item.GetWatk())
		p.WriteInt16(item.GetMatk())
		p.WriteInt16(item.GetWdef())
		p.WriteInt16(item.GetMdef())
		p.WriteInt16(item.GetAccuracy())
		p.WriteInt16(item.GetAvoid())
		p.WriteInt16(item.GetHands())
		p.WriteInt16(item.GetSpeed())
		p.WriteInt16(item.GetJump())
		p.WriteString(item.GetCreatorName())
		p.WriteInt16(item.GetFlag()) // lock, show, spikes, cape, cold protection etc ?
	case 2:
		p.WriteInt16(item.GetAmount()) // amount
		p.WriteString(item.GetCreatorName())
		p.WriteInt16(item.GetFlag()) // lock, show, spikes, cape, cold protection etc ?
	case 3:
	case 4:
	case 5:
	default:
		fmt.Println("Unsuported item type")
	}

	return p
}

func writeDisplayCharacter(char character.Character) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(char.GetGender()) // gender
	p.WriteByte(char.GetSkin())   // skin
	p.WriteInt32(char.GetFace())  // face
	p.WriteByte(0x00)             // ?
	p.WriteInt32(char.GetHair())  // hair
	cashWeapon := int32(0)

	for _, b := range char.GetItems() {
		if b.GetSlotID() < 0 && b.GetSlotID() > -20 {
			p.WriteByte(byte(math.Abs(float64(b.GetSlotID()))))
			p.WriteInt32(b.GetItemID())
		}
	}

	for _, b := range char.GetItems() {
		if b.GetSlotID() < -100 {
			if b.GetSlotID() == -111 {
				cashWeapon = b.GetItemID()
			} else {
				p.WriteByte(byte(math.Abs(float64(b.GetSlotID() + 100))))
				p.WriteInt32(b.GetItemID())
			}
		}
	}

	p.WriteByte(0xFF)
	// What items go here?
	p.WriteByte(0xFF)
	p.WriteInt32(cashWeapon)

	return p
}
