package channel

import (
	"sort"
	"time"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

func packetCashShopSet(plr *player, accountName string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSetCashShop)

	p.WriteInt16(-1)

	// Stats
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

	p.WriteByte(plr.buddyListSize)

	// Money
	p.WriteInt32(plr.mesos)

	if plr.equipSlotSize == 0 {
		plr.equipSlotSize = 24
	}
	if plr.useSlotSize == 0 {
		plr.useSlotSize = 24
	}
	if plr.setupSlotSize == 0 {
		plr.setupSlotSize = 24
	}
	if plr.etcSlotSize == 0 {
		plr.etcSlotSize = 24
	}
	if plr.cashSlotSize == 0 {
		plr.cashSlotSize = 24
	}

	p.WriteByte(plr.equipSlotSize)
	p.WriteByte(plr.useSlotSize)
	p.WriteByte(plr.setupSlotSize)
	p.WriteByte(plr.etcSlotSize)
	p.WriteByte(plr.cashSlotSize)

	// Equipped (normal then cash)
	for _, it := range plr.equip {
		if it.slotID < 0 && !it.cash {
			p.WriteBytes(it.inventoryBytes())
		}
	}
	p.WriteByte(0)
	for _, it := range plr.equip {
		if it.slotID < 0 && it.cash {
			p.WriteBytes(it.inventoryBytes())
		}
	}
	p.WriteByte(0)

	// Inventory tabs
	writeInv := func(items []item) {
		cp := make([]item, 0, len(items))
		for _, it := range items {
			if it.slotID > 0 {
				cp = append(cp, it)
			}
		}
		sort.Slice(cp, func(i, j int) bool { return cp[i].slotID < cp[j].slotID })
		for _, it := range cp {
			p.WriteBytes(it.inventoryBytes())
		}
		p.WriteByte(0)
	}
	writeInv(plr.equip)
	writeInv(plr.use)
	writeInv(plr.setUp)
	writeInv(plr.etc)
	writeInv(plr.cash)

	// Skills
	p.WriteInt16(int16(len(plr.skills)))
	skillCooldowns := make(map[int32]int16)

	for _, skill := range plr.skills {
		p.WriteInt32(skill.ID)
		p.WriteInt32(int32(skill.Level))

		if skill.Cooldown > 0 {
			skillCooldowns[skill.ID] = skill.Cooldown
		}
	}

	p.WriteInt16(int16(len(skillCooldowns)))

	for id, cooldown := range skillCooldowns {
		p.WriteInt32(id)
		p.WriteInt16(cooldown)
	}

	// Quests
	writeActiveQuests(&p, plr.quests.inProgressList())
	writeCompletedQuests(&p, plr.quests.completedList())

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

	p.WriteByte(1)
	p.WriteString(accountName)
	p.WriteInt16(0)

	comms := nx.GetCommodities()
	p.WriteInt16(int16(len(comms)))
	for sn, c := range comms {
		p.WriteInt32(sn)
		p.WriteInt32(c.StockState)
	}

	return p
}

func packetCashShopUpdateAmounts(nxCredit, maplePoints int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCSUpdateAmounts)
	p.WriteInt32(nxCredit)
	p.WriteInt32(maplePoints)
	return p
}

func packetCashShopShowBoughtItem(charID int32, cashItemSNHash int64, itemID int32, count int16, itemName string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCSAction)
	p.WriteInt64(cashItemSNHash)
	p.WriteInt32(charID)

	for i := 0; i < 4; i++ {
		p.WriteByte(0x01)
	}

	p.WriteInt32(itemID)

	for i := 0; i < 4; i++ {
		p.WriteByte(0x01)
	}

	p.WriteInt16(count)
	p.WriteString(itemName)
	p.WriteInt64(0) // expiration: 0 for non-expiring
	for i := 0; i < 4; i++ {
		p.WriteByte(0x01)
	}
	return p
}

func packetCashShopShowBoughtQuestItem(position byte, itemID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCSAction)
	p.WriteInt32(365) // sub-op code per reference
	p.WriteByte(0)
	p.WriteInt16(1)
	p.WriteByte(position)
	p.WriteByte(0)
	p.WriteInt32(itemID)
	return p
}

func packetCashShopShowCouponRedeemedItem(itemID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCSAction)
	p.WriteInt16(0x3A)
	p.WriteInt32(0)
	p.WriteInt32(1)
	p.WriteInt16(1)
	p.WriteInt16(0x1A)
	p.WriteInt32(itemID)
	p.WriteInt32(0)
	return p
}

func packetCashShopSendCSItemInventory(slotType byte, it item) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCSAction)
	p.WriteByte(0x2F)
	p.WriteInt16(int16(slotType))
	p.WriteByte(slotType)

	p.WriteBytes(it.inventoryBytes())
	return p
}

func packetCashShopWishList(sns []int32, update bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCSAction)
	if update {
		p.WriteByte(0x39)
	} else {
		p.WriteByte(0x33)
	}
	count := 10
	for i := 0; i < count; i++ {
		var v int32
		if i < len(sns) {
			v = sns[i]
		}
		p.WriteInt32(v)
	}
	return p
}

func packetCashShopWrongCoupon() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCSAction)
	p.WriteByte(0x40)
	p.WriteByte(0x87)
	return p
}
