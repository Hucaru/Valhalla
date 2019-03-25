package game

import (
	"crypto/rand"
	"fmt"
	"math"

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

func PacketPlayerReceivedDmg(charID int32, attack int8, initalAmmount, reducedAmmount, spawnID, mobID, healSkillID int32,
	stance, reflectAction byte, reflected byte, reflectX, reflectY int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerTakeDmg)
	p.WriteInt32(charID)
	p.WriteInt8(attack)
	p.WriteInt32(initalAmmount)

	p.WriteInt32(spawnID)
	p.WriteInt32(mobID)
	p.WriteByte(stance)
	p.WriteByte(reflected)

	if reflected > 0 {
		p.WriteByte(reflectAction)
		p.WriteInt16(reflectX)
		p.WriteInt16(reflectY)
	}

	p.WriteInt32(reducedAmmount)

	// Check if used
	if reducedAmmount < 0 {
		p.WriteInt32(healSkillID)
	}

	return p
}

func PacketPlayerLevelUpAnimation(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerAnimation)
	p.WriteInt32(charID)
	p.WriteByte(0x00)

	return p
}

func PacketPlayerMove(charID int32, bytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerMovement)
	p.WriteInt32(charID)
	p.WriteBytes(bytes)

	return p
}

func PacketPlayerEmoticon(charID int32, emotion int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerEmoticon)
	p.WriteInt32(charID)
	p.WriteInt32(emotion)

	return p
}

func PacketPlayerSkillBookUpdate(skillID int32, level int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSkillRecordUpdate)
	p.WriteByte(0x01)  // time check?
	p.WriteInt16(0x01) // number of skills to update
	p.WriteInt32(skillID)
	p.WriteInt32(level)
	p.WriteByte(0x01)

	return p
}

func PacketPlayerStatChange(unknown bool, stat int32, value int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelStatChange)
	p.WriteBool(unknown)
	p.WriteInt32(stat)
	p.WriteInt32(value)

	return p
}

func PacketPlayerNoChange() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x00)
	p.WriteByte(0x00)

	return p
}

func PacketPlayerAvatarSummaryWindow(charID int32, char Character, guildName string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelAvatarInfoWindow)
	p.WriteInt32(charID)
	p.WriteByte(char.Level)
	p.WriteInt16(char.Job)
	p.WriteInt16(char.Fame)

	p.WriteString(guildName)

	p.WriteBool(false) // if has pet
	p.WriteByte(0)     // wishlist count

	return p
}

func PacketPlayerEnterGame(char Character, channelID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelWarpToMap)
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

	// Are active buffs name encoded in here?
	p.WriteByte(0xFF)
	p.WriteByte(0xFF)

	p.WriteInt32(char.ID)
	p.WritePaddedString(char.Name, 13)
	p.WriteByte(char.Gender)
	p.WriteByte(char.Skin)
	p.WriteInt32(char.Face)
	p.WriteInt32(char.Hair)

	p.WriteInt64(0) // Pet Cash ID

	p.WriteByte(char.Level)
	p.WriteInt16(char.Job)
	p.WriteInt16(char.Str)
	p.WriteInt16(char.Dex)
	p.WriteInt16(char.Int)
	p.WriteInt16(char.Luk)
	p.WriteInt16(char.HP)
	p.WriteInt16(char.MaxHP)
	p.WriteInt16(char.MP)
	p.WriteInt16(char.MaxMP)
	p.WriteInt16(char.AP)
	p.WriteInt16(char.SP)
	p.WriteInt32(char.EXP)
	p.WriteInt16(char.Fame)

	p.WriteInt32(char.MapID)
	p.WriteByte(char.MapPos)

	p.WriteByte(20) // budy list size
	p.WriteInt32(char.Mesos)

	p.WriteByte(char.EquipSlotSize)
	p.WriteByte(char.UseSlotSize)
	p.WriteByte(char.SetupSlotSize)
	p.WriteByte(char.EtcSlotSize)
	p.WriteByte(char.CashSlotSize)

	for _, v := range char.Equip {
		if v.SlotID < 0 && v.InvID == 1 && !v.Cash {
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	// Equips
	for _, v := range char.Equip {
		if v.SlotID < 0 && v.InvID == 1 && v.Cash {
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	// Inventory windows starts
	for _, v := range char.Equip {
		if v.SlotID > -1 && v.InvID == 1 {
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range char.Use {
		if v.InvID == 2 { // Use
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range char.SetUp {
		if v.InvID == 3 { // Set-up
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range char.Etc {
		if v.InvID == 4 { // Etc
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range char.Cash {
		if v.InvID == 5 { // Cash  - not working propery :(
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	// Skills
	p.WriteInt16(int16(len(char.Skills))) // number of skills

	for _, skill := range char.Skills {
		p.WriteInt32(skill.ID)
		p.WriteInt32(int32(skill.Level))
	}

	// Quests
	p.WriteInt16(0) // # of quests?

	// What are these for?
	p.WriteInt16(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	// Setting these appears to do nothing
	p.WriteInt32(1)
	p.WriteInt32(1)
	p.WriteInt32(1)

	p.WriteUint64(1)
	p.WriteUint64(1)
	p.WriteUint64(1)
	p.WriteUint64(1)
	p.WriteUint64(1)
	p.WriteInt64(1)

	return p
}

func addItem(item Item, shortSlot bool) mpacket.Packet {
	p := mpacket.NewPacket()

	if !shortSlot {
		if item.Cash && item.SlotID < 0 {
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

	if item.Cash {
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

func writeDisplayCharacter(char Character) mpacket.Packet {
	p := mpacket.NewPacket()
	p.WriteByte(char.Gender) // gender
	p.WriteByte(char.Skin)   // skin
	p.WriteInt32(char.Face)  // face
	p.WriteByte(0x00)        // ?
	p.WriteInt32(char.Hair)  // hair

	cashWeapon := int32(0)

	for _, b := range char.Equip {
		if b.SlotID < 0 && b.SlotID > -20 {
			p.WriteByte(byte(math.Abs(float64(b.SlotID))))
			p.WriteInt32(b.ItemID)
		}
	}

	for _, b := range char.Equip {
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
