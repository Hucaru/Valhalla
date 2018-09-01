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
	p.WriteByte(constants.SendChannelPlayerTakeDmg)
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
	p.WriteByte(constants.SendChannelPlayerAnimation)
	p.WriteInt32(charID)
	p.WriteByte(0x00)

	return p
}

func PlayerMove(charID int32, leftOverBytes maplepacket.Packet) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SendChannelPlayerMovement)
	p.WriteInt32(charID)
	p.WriteBytes(leftOverBytes)

	return p
}

func PlayerEmoticon(playerID int32, emotion int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SendChannelPlayerEmoticon)
	p.WriteInt32(playerID)
	p.WriteInt32(emotion)

	return p
}

func PlayerSkillBookUpdate(skillID int32, level int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SendChannelSkillRecordUpdate)
	p.WriteByte(0x01)  // time check?
	p.WriteInt16(0x01) // number of skills to update
	p.WriteInt32(skillID)
	p.WriteInt32(level)
	p.WriteByte(0x01)

	return p
}

func PlayerStatChange(byPlayer bool, stat int32, value int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SendChannelStatChange)
	p.WriteBool(byPlayer)
	p.WriteInt32(stat)
	p.WriteInt32(value)

	return p
}

func PlayerStatNoChange() maplepacket.Packet {
	p := maplepacket.NewPacket()
	// Continue game opcode is part of inventory opcode list?
	p.WriteByte(constants.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x00)
	p.WriteByte(0x00)

	return p
}

func PlayerAvatarSummaryWindow(charID int32, char character.Character, guildName string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SendChannelAvatarInfoWindow)
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
	p.WriteByte(constants.SendChannelWarpToMap)
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
		if v.SlotID < 0 && v.InvID == 1 && !nx.IsCashItem(v.ItemID) {
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	// Equips
	for _, v := range char.GetItems() {
		if v.SlotID < 0 && v.InvID == 1 && nx.IsCashItem(v.ItemID) {
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	// Inventory windows starts
	for _, v := range char.GetItems() {
		if v.SlotID > -1 && v.InvID == 1 {
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.InvID == 2 { // Use
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.InvID == 3 { // Set-up
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.InvID == 4 { // Etc
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range char.GetItems() {
		if v.InvID == 5 { // Cash  - not working propery :(
			p.WriteBytes(addItem(v, false))
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

func addItem(item inventory.Item, shortSlot bool) maplepacket.Packet {
	p := maplepacket.NewPacket()

	if !shortSlot {
		if nx.IsCashItem(item.ItemID) && item.SlotID < 0 {
			p.WriteByte(byte(math.Abs(float64(item.SlotID + 100))))
		} else {
			p.WriteByte(byte(math.Abs(float64(item.SlotID))))
		}
	} else {
		p.WriteInt16(item.SlotID)
	}

	switch item.InvID {
	case 1:
		p.WriteByte(0x01)
	default:
		p.WriteByte(0x02)
	}

	p.WriteInt32(item.ItemID)

	if nx.IsCashItem(item.ItemID) {
		p.WriteByte(1)
		p.WriteUint64(uint64(item.ItemID))
	} else {
		p.WriteByte(0)
	}

	p.WriteUint64(item.ExpireTime)

	switch item.InvID {
	case 1:
		p.WriteByte(item.UpgradeSlots)
		p.WriteByte(item.ScrollLevel)
		p.WriteInt16(item.Str)
		p.WriteInt16(item.Dex)
		p.WriteInt16(item.Int)
		p.WriteInt16(item.Luk)
		p.WriteInt16(item.HP)
		p.WriteInt16(item.MP)
		p.WriteInt16(item.Watk)
		p.WriteInt16(item.Matk)
		p.WriteInt16(item.Wdef)
		p.WriteInt16(item.Mdef)
		p.WriteInt16(item.Accuracy)
		p.WriteInt16(item.Avoid)
		p.WriteInt16(item.Hands)
		p.WriteInt16(item.Speed)
		p.WriteInt16(item.Jump)
		p.WriteString(item.CreatorName)
		p.WriteInt16(item.Flag) // lock, show, spikes, cape, cold protection etc ?
	case 2:
		fallthrough
	case 3:
		fallthrough
	case 4:
		fallthrough
	case 5:
		p.WriteInt16(item.Amount)
		p.WriteString(item.CreatorName)
		p.WriteInt16(item.Flag) // lock, show, spikes, cape, cold protection etc ?
	default:
		fmt.Println("Unsuported item type", item.InvID)
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
		if b.SlotID < 0 && b.SlotID > -20 {
			p.WriteByte(byte(math.Abs(float64(b.SlotID))))
			p.WriteInt32(b.ItemID)
		}
	}

	for _, b := range char.GetItems() {
		if b.SlotID < -100 {
			if b.SlotID == -111 {
				cashWeapon = b.ItemID
			} else {
				p.WriteByte(byte(math.Abs(float64(b.SlotID + 100))))
				p.WriteInt32(b.ItemID)
			}
		}
	}

	p.WriteByte(0xFF)
	p.WriteByte(0xFF)
	p.WriteInt32(cashWeapon)

	return p
}
