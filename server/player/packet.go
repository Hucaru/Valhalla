package player

import (
	"crypto/rand"
	"math"

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

	for _, v := range plr.inventory.equip {
		if v.slotID < 0 && v.invID == 1 && !v.cash {
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	// Equips
	for _, v := range plr.inventory.equip {
		if v.slotID < 0 && v.invID == 1 && v.cash {
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	// Inventory windows starts
	for _, v := range plr.inventory.equip {
		if v.slotID > -1 && v.invID == 1 {
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range plr.inventory.use {
		if v.invID == 2 { // Use
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range plr.inventory.setUp {
		if v.invID == 3 { // Set-up
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range plr.inventory.etc {
		if v.invID == 4 { // Etc
			p.WriteBytes(addItem(v, false))
		}
	}

	p.WriteByte(0)

	for _, v := range plr.inventory.cash {
		if v.invID == 5 { // Cash  - not working propery :(
			p.WriteBytes(addItem(v, false))
		}
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

type item interface {
	ID() int32
	Cash() bool
	SlotID() int16
	InvID() byte
	Pet() bool
	UpgradeSlots() byte
	ScrollLevel() byte
	Str() int16
	Dex() int16
	Int() int16
	Luk() int16
	Hp() int16
	Mp() int16
	Watk() int16
	Matk() int16
	Wdef() int16
	Mdef() int16
	Accuracy() int16
	Avoid() int16
	Hands() int16
	Speed() int16
	Jump() int16
	CreatorName() string
	Flag() int16
	ExpireTime() int64
	Amount() int16
	IsRechargeable() bool
}

func addItem(item item, shortSlot bool) mpacket.Packet {
	p := mpacket.NewPacket()

	if !shortSlot {
		if item.Cash() && item.SlotID() < 0 {
			p.WriteByte(byte(math.Abs(float64(item.SlotID() + 100))))
		} else {
			p.WriteByte(byte(math.Abs(float64(item.SlotID()))))
		}
	} else {
		p.WriteInt16(item.SlotID())
	}

	if item.InvID() == 1 {
		p.WriteByte(0x01)
	} else if item.Pet() {
		p.WriteByte(0x03)
	} else {
		p.WriteByte(0x02)
	}

	p.WriteInt32(item.ID())

	p.WriteBool(item.Cash())
	if item.Cash() {
		p.WriteUint64(uint64(item.ID())) // I think this is somekind of cashshop transaction ID for the item
	}

	p.WriteInt64(item.ExpireTime())

	if item.InvID() == 1 {
		p.WriteByte(item.UpgradeSlots())
		p.WriteByte(item.ScrollLevel())
		p.WriteInt16(item.Str())
		p.WriteInt16(item.Dex())
		p.WriteInt16(item.Int())
		p.WriteInt16(item.Luk())
		p.WriteInt16(item.Hp())
		p.WriteInt16(item.Mp())
		p.WriteInt16(item.Watk())
		p.WriteInt16(item.Matk())
		p.WriteInt16(item.Wdef())
		p.WriteInt16(item.Mdef())
		p.WriteInt16(item.Accuracy())
		p.WriteInt16(item.Avoid())
		p.WriteInt16(item.Hands())
		p.WriteInt16(item.Speed())
		p.WriteInt16(item.Jump())
		p.WriteString(item.CreatorName())
		p.WriteInt16(item.Flag()) // lock/seal, show, spikes, cape, cold protection etc ?
	} else if item.Pet() {
		p.WritePaddedString(item.CreatorName(), 13)
		p.WriteByte(0)
		p.WriteInt16(0)
		p.WriteByte(0)
		p.WriteInt64(item.ExpireTime())
		p.WriteInt32(0) // ?
	} else {
		p.WriteInt16(item.Amount())
		p.WriteString(item.CreatorName())
		p.WriteInt16(item.Flag()) // even (normal), odd (sealed) ?

		if item.IsRechargeable() {
			p.WriteInt32(0) // ?
		}
	}

	return p
}
