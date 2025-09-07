package cashshop

import (
	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

func packetCashShopSet(plr *channel.Player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSetCashShop)

	plr.WriteCharacterInfoPacket(&p)

	p.WriteByte(1)
	p.WriteString(plr.GetAccountName())

	p.WriteInt16(0) // Wishlist

	p.WriteBytes(make([]byte, 121))

	// Featured/Best items: Category (1..8, excluding Quest=9), Gender (0..1), then SN
	for i := 1; i <= 8; i++ { // categories excluding Quest
		for j := 0; j <= 1; j++ { // gender
			for k := 0; k < 5; k++ { // top 5
				p.WriteInt32(int32(i)) // Category
				p.WriteInt32(int32(j)) // Gender
				sn := nx.GetBestSN(i, j, k)
				p.WriteInt32(sn) // 0 if none
			}
		}
	}

	p.WriteInt32(0)
	p.WriteByte(0)

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

func packetCashShopSendCSItemInventory(slotType byte, it channel.Item) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCSAction)
	p.WriteByte(0x2F)
	p.WriteInt16(int16(slotType))
	p.WriteByte(slotType)

	p.WriteBytes(it.InventoryBytes())
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
