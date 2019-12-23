package player

import (
	"crypto/rand"

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

func packetPlayerReceivedDmg(charID int32, attack int8, initalAmmount, reducedAmmount, spawnID, mobID, healSkillID int32,
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

func packetPlayerLevelUpAnimation(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerAnimation)
	p.WriteInt32(charID)
	p.WriteByte(0x00)

	return p
}

func packetPlayerMove(charID int32, bytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerMovement)
	p.WriteInt32(charID)
	p.WriteBytes(bytes)

	return p
}

func packetPlayerEmoticon(charID int32, emotion int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerEmoticon)
	p.WriteInt32(charID)
	p.WriteInt32(emotion)

	return p
}

func packetPlayerSkillBookUpdate(skillID int32, level int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSkillRecordUpdate)
	p.WriteByte(0x01)  // time check?
	p.WriteInt16(0x01) // number of skills to update
	p.WriteInt32(skillID)
	p.WriteInt32(level)
	p.WriteByte(0x01)

	return p
}

func packetPlayerStatChange(unknown bool, stat int32, value int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelStatChange)
	p.WriteBool(unknown)
	p.WriteInt32(stat)
	p.WriteInt32(value)

	return p
}

func packetPlayerNoChange() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x00)
	p.WriteByte(0x00)

	return p
}

func packetPlayerAvatarSummaryWindow(charID int32, plr Player, guildName string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelAvatarInfoWindow)
	p.WriteInt32(plr.id)
	p.WriteByte(plr.level)
	p.WriteInt16(plr.job)
	p.WriteInt16(plr.fame)

	p.WriteString(guildName)

	p.WriteBool(false) // if has pet
	p.WriteByte(0)     // wishlist count

	return p
}

func packetChangeChannel(ip []byte, port int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelChange)
	p.WriteBool(true)
	p.WriteBytes(ip)
	p.WriteInt16(port)

	return p
}

func packetCannotChangeChannel() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelChangeServer)
	p.WriteByte(1)

	return p
}

func packetCannotEnterCashShop() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelChangeServer)
	p.WriteByte(2)

	return p
}

func packetPlayerEnterGame(plr Player, channelID int32) mpacket.Packet {
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

	p.WriteInt32(plr.id)
	p.WritePaddedString(plr.name, 13)
	p.WriteByte(plr.gender)
	p.WriteByte(plr.skin)
	p.WriteInt32(plr.face)
	p.WriteInt32(plr.hair)

	p.WriteInt64(0) // Pet Cash ID

	p.WriteByte(plr.level)
	p.WriteInt16(plr.job)
	p.WriteInt16(plr.str)
	p.WriteInt16(plr.dex)
	p.WriteInt16(plr.intt)
	p.WriteInt16(plr.luk)
	p.WriteInt16(plr.hp)
	p.WriteInt16(plr.maxHP)
	p.WriteInt16(plr.mp)
	p.WriteInt16(plr.maxMP)
	p.WriteInt16(plr.ap)
	p.WriteInt16(plr.sp)
	p.WriteInt32(plr.exp)
	p.WriteInt16(plr.fame)

	p.WriteInt32(plr.mapID)
	p.WriteByte(plr.mapPos)

	p.WriteByte(20) // budy list size
	p.WriteInt32(plr.mesos)

	p.WriteByte(plr.equipSlotSize)
	p.WriteByte(plr.useSlotSize)
	p.WriteByte(plr.setupSlotSize)
	p.WriteByte(plr.etcSlotSize)
	p.WriteByte(plr.cashSlotSize)

	for _, v := range plr.equip {
		if v.SlotID() < 0 && !v.Cash() {
			p.WriteBytes(v.InventoryBytes())
		}
	}

	p.WriteByte(0)

	// Equips
	for _, v := range plr.equip {
		if v.SlotID() < 0 && v.Cash() {
			p.WriteBytes(v.InventoryBytes())
		}
	}

	p.WriteByte(0)

	// Inventory windows starts
	for _, v := range plr.equip {
		if v.SlotID() > -1 {
			p.WriteBytes(v.InventoryBytes())
		}
	}

	p.WriteByte(0)

	for _, v := range plr.use {
		p.WriteBytes(v.InventoryBytes())
	}

	p.WriteByte(0)

	for _, v := range plr.setUp {
		p.WriteBytes(v.InventoryBytes())
	}

	p.WriteByte(0)

	for _, v := range plr.etc {
		p.WriteBytes(v.InventoryBytes())
	}

	p.WriteByte(0)

	for _, v := range plr.cash {
		p.WriteBytes(v.InventoryBytes())
	}

	p.WriteByte(0)

	// Skills
	p.WriteInt16(int16(len(plr.skills))) // number of skills

	for _, skill := range plr.skills {
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

func packetMessageExpGained(whiteText, appearInChat bool, ammount int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(3)
	p.WriteBool(whiteText)
	p.WriteInt32(ammount)
	p.WriteBool(appearInChat)

	return p
}

func packetInventoryAddItem(item item, newItem bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteBool(!newItem)
	p.WriteByte(item.InvID())

	if newItem {
		p.WriteBytes(item.ShortBytes())
	} else {
		p.WriteInt16(item.SlotID())
		p.WriteInt16(item.Amount())
	}

	return p
}

func packetInventoryAddItems(items []item, newItem []bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)

	p.WriteByte(0x01)
	if len(items) != len(newItem) {
		p.WriteByte(0)
		return p
	}

	p.WriteByte(byte(len(items)))

	for i, v := range items {
		p.WriteBool(!newItem[i])
		p.WriteByte(v.InvID())

		if newItem[i] {
			p.WriteBytes(v.ShortBytes())
		} else {
			p.WriteInt16(v.SlotID())
			p.WriteInt16(v.Amount())
		}
	}

	return p
}
