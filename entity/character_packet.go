package entity

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
	p.WriteByte(char.level)
	p.WriteInt16(char.job)
	p.WriteInt16(char.fame)

	p.WriteString(guildName)

	p.WriteBool(false) // if has pet
	p.WriteByte(0)     // wishlist count

	return p
}

func PacketChangeChannel(ip []byte, port int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelChange)
	p.WriteBool(true)
	p.WriteBytes(ip)
	p.WriteInt16(port)

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

	p.WriteInt32(char.id)
	p.WritePaddedString(char.name, 13)
	p.WriteByte(char.gender)
	p.WriteByte(char.skin)
	p.WriteInt32(char.face)
	p.WriteInt32(char.hair)

	p.WriteInt64(0) // Pet Cash ID

	p.WriteByte(char.level)
	p.WriteInt16(char.job)
	p.WriteInt16(char.str)
	p.WriteInt16(char.dex)
	p.WriteInt16(char.intt)
	p.WriteInt16(char.luk)
	p.WriteInt16(char.hp)
	p.WriteInt16(char.maxHP)
	p.WriteInt16(char.mp)
	p.WriteInt16(char.maxMP)
	p.WriteInt16(char.ap)
	p.WriteInt16(char.sp)
	p.WriteInt32(char.exp)
	p.WriteInt16(char.fame)

	p.WriteInt32(char.mapID)
	p.WriteByte(char.mapPos)

	p.WriteByte(20) // budy list size
	p.WriteInt32(char.mesos)

	p.WriteByte(char.equipSlotSize)
	p.WriteByte(char.useSlotSize)
	p.WriteByte(char.setupSlotSize)
	p.WriteByte(char.etcSlotSize)
	p.WriteByte(char.cashSlotSize)

	for _, v := range char.inventory.equip {
		if v.slotID < 0 && v.invID == 1 && !v.cash {
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	// Equips
	for _, v := range char.inventory.equip {
		if v.slotID < 0 && v.invID == 1 && v.cash {
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	// Inventory windows starts
	for _, v := range char.inventory.equip {
		if v.slotID > -1 && v.invID == 1 {
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range char.inventory.use {
		if v.invID == 2 { // Use
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range char.inventory.setUp {
		if v.invID == 3 { // Set-up
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range char.inventory.etc {
		if v.invID == 4 { // Etc
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range char.inventory.cash {
		if v.invID == 5 { // Cash  - not working propery :(
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	// Skills
	p.WriteInt16(int16(len(char.skills))) // number of skills

	for _, skill := range char.skills {
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

func addItem(item item, shortSlot bool) mpacket.Packet {
	p := mpacket.NewPacket()

	if !shortSlot {
		if item.cash && item.slotID < 0 {
			p.WriteByte(byte(math.Abs(float64(item.slotID + 100))))
		} else {
			p.WriteByte(byte(math.Abs(float64(item.slotID))))
		}
	} else {
		p.WriteInt16(item.slotID)
	}

	switch item.invID {
	case 1:
		p.WriteByte(0x01)
	default:
		p.WriteByte(0x02)
	}

	p.WriteInt32(item.itemID)

	if item.cash {
		p.WriteByte(1)
		p.WriteUint64(uint64(item.itemID))
	} else {
		p.WriteByte(0)
	}

	p.WriteUint64(item.expireTime)

	switch item.invID {
	case 1:
		p.WriteByte(item.upgradeSlots)
		p.WriteByte(item.scrollLevel)
		p.WriteInt16(item.str)
		p.WriteInt16(item.dex)
		p.WriteInt16(item.intt)
		p.WriteInt16(item.luk)
		p.WriteInt16(item.hp)
		p.WriteInt16(item.mp)
		p.WriteInt16(item.watk)
		p.WriteInt16(item.matk)
		p.WriteInt16(item.wdef)
		p.WriteInt16(item.mdef)
		p.WriteInt16(item.accuracy)
		p.WriteInt16(item.avoid)
		p.WriteInt16(item.hands)
		p.WriteInt16(item.speed)
		p.WriteInt16(item.jump)
		p.WriteString(item.creatorName)
		p.WriteInt16(item.flag) // lock, show, spikes, cape, cold protection etc ?
	case 2:
		fallthrough
	case 3:
		fallthrough
	case 4:
		fallthrough
	case 5:
		p.WriteInt16(item.amount)
		p.WriteString(item.creatorName)
		p.WriteInt16(item.flag) // lock, show, spikes, cape, cold protection etc ?
	default:
		fmt.Println("Unsuported item type", item.invID)
	}

	return p
}

func WriteDisplayCharacter(char Character) mpacket.Packet {
	p := mpacket.NewPacket()
	p.WriteByte(char.gender) // gender
	p.WriteByte(char.skin)   // skin
	p.WriteInt32(char.face)  // face
	p.WriteByte(0x00)        // ?
	p.WriteInt32(char.hair)  // hair

	cashWeapon := int32(0)

	for _, b := range char.inventory.equip {
		if b.slotID < 0 && b.slotID > -20 {
			p.WriteByte(byte(math.Abs(float64(b.slotID))))
			p.WriteInt32(b.itemID)
		}
	}

	for _, b := range char.inventory.equip {
		if b.slotID < -100 {
			if b.slotID == -111 {
				cashWeapon = b.itemID
			} else {
				p.WriteByte(byte(math.Abs(float64(b.slotID + 100))))
				p.WriteInt32(b.itemID)
			}
		}
	}

	p.WriteByte(0xFF)
	p.WriteByte(0xFF)
	p.WriteInt32(cashWeapon)

	return p
}
