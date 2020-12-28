package player

import (
	"crypto/rand"
	"time"

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/item"
)

func PlayerReceivedDmg(charID int32, attack int8, initalAmmount, reducedAmmount, spawnID, mobID, healSkillID int32,
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

func packetPlayerAvatarSummaryWindow(charID int32, plr Data, guildName string) mpacket.Packet {
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

// PacketPlayerEnterGame - packet that is sent to player when connecting to the channel server
func PacketPlayerEnterGame(plr Data, channelID int32) mpacket.Packet {
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

	skillCooldowns := make(map[int32]int16)

	for _, skill := range plr.skills {
		p.WriteInt32(skill.ID)
		p.WriteInt32(int32(skill.Level))

		if skill.Cooldown > 0 {
			skillCooldowns[skill.ID] = skill.Cooldown
		}
	}

	p.WriteInt16(int16(len(skillCooldowns))) // number of cooldowns

	for id, cooldown := range skillCooldowns {
		p.WriteInt32(id)
		p.WriteInt16(cooldown)
	}

	// Quests
	p.WriteInt16(3) // Active quest count
	p.WriteInt16(2029)
	p.WriteString("")
	p.WriteInt16(2000)
	p.WriteString("")
	p.WriteInt16(1000)
	p.WriteString("")
	p.WriteInt16(0) // Completed quest count?

	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt64(time.Now().Unix())

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

func packetInventoryAddItem(item item.Data, newItem bool) mpacket.Packet {
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

func packetInventoryAddItems(items []item.Data, newItem []bool) mpacket.Packet {
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

func packetInventoryChangeItemSlot(invTabID byte, origPos, newPos int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x02)
	p.WriteByte(invTabID)
	p.WriteInt16(origPos)
	p.WriteInt16(newPos)
	p.WriteByte(0x00) // ?

	return p
}

func packetInventoryRemoveItem(item item.Data) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x03)
	p.WriteByte(item.InvID())
	p.WriteInt16(item.SlotID())
	p.WriteUint64(0) //?

	return p
}

func packetInventoryChangeEquip(char Data) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerChangeAvatar)
	p.WriteInt32(char.id)
	p.WriteByte(1)
	p.WriteBytes(char.DisplayBytes())
	p.WriteByte(0xFF)
	p.WriteUint64(0) //?

	return p
}

func packetInventoryNoChange() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x00)
	p.WriteByte(0x00)

	return p
}

func packetGuildInfo(id int32, name string, memberCount byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x1a)

	if len(name) == 0 {
		p.WriteByte(0x00) // removes player from guild
		return p
	}

	p.WriteBool(true) // In guild
	p.WriteInt32(1)   // guild id (value cannot be zero)
	p.WriteString(name)

	// 5 ranks each have a title
	p.WriteString("rank1")
	p.WriteString("rank2")
	p.WriteString("rank3")
	p.WriteString("rank4")
	p.WriteString("rank5")

	capacity := 250             // maximum
	p.WriteByte(byte(capacity)) // member count

	// iterate over all members and output ids
	for i := 0; i < capacity; i++ {
		p.WriteInt32(int32(i + 1))
	}

	// iterate over all members and input their info
	for i := 0; i < capacity; i++ {
		p.WritePaddedString("[GM]Hucaru", 13) // name
		p.WriteInt32(510)                     // job
		p.WriteInt32(255)                     // level

		if i > 4 {
			p.WriteInt32(5) // rank starts at 1
		} else {
			p.WriteInt32(int32(i + 1)) // rank starts at 1
		}

		if i%2 == 0 {
			p.WriteInt32(1) // online or not
		} else {
			p.WriteInt32(0)
		}

		p.WriteInt32(int32(i)) // ?
	}

	p.WriteInt32(int32(capacity)) // capacity
	p.WriteInt16(1030)            // logo background
	p.WriteByte(3)                // logo bg colour
	p.WriteInt16(4017)            // logo
	p.WriteByte(2)                // logo colour
	p.WriteString("notice")       // notice
	p.WriteInt32(9999)            // ?

	return p
}

func packetBuddyInfo(buddyList []buddy) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x12)
	p.WriteByte(byte(len(buddyList)))

	for _, v := range buddyList {
		p.WriteInt32(v.id)
		p.WritePaddedString(v.name, 13)
		p.WriteByte(v.status)
		p.WriteInt32(v.channelID)
	}

	for _, v := range buddyList {
		p.WriteInt32(v.cashShop)
	}

	// for i := 0; i < 10; i++ {
	// 	p.WriteInt32(int32(i + 1))
	// 	p.WritePaddedString("test"+strconv.Itoa(i), 13)
	// 	p.WriteByte(0)  // 0 - online, 1 - buddy request, 2 - offline
	// 	p.WriteInt32(0) // channel id
	// }

	// for i := 0; i < 10; i++ {
	// 	p.WriteInt32(0) // > 0 means is in cash shop?
	// }

	return p
}

// Move this messages into the messages package
func packetBuddyUnkownError() mpacket.Packet {
	return packetBuddyRequestResult(0x16)
}

func packetBuddyPlayerFullList() mpacket.Packet {
	return packetBuddyRequestResult(0x0b)
}

func packetBuddyOtherFullList() mpacket.Packet {
	return packetBuddyRequestResult(0x0c)
}

func packetBuddyAlreadyAdded() mpacket.Packet {
	return packetBuddyRequestResult(0x0d)
}

func packetBuddyIsGM() mpacket.Packet {
	return packetBuddyRequestResult(0x0e)
}

func packetBuddyInvalidName() mpacket.Packet {
	return packetBuddyRequestResult(0x0f)
}

func packetBuddyRequestResult(code byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(code)

	return p
}

func packetBuddyListSizeUpdate(size byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x15)
	p.WriteByte(size)

	return p
}

func packetBuddyReceiveRequest(from string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x9)
	p.WriteInt32(0) // ?
	p.WriteString(from)

	// Missing more data

	return p
}

// buddy operations left - 0x8 (int32, int8), 0x14
