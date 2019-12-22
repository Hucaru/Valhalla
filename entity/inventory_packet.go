package entity

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

func PacketInventoryAddItem(item item, newItem bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteBool(!newItem)
	p.WriteByte(item.invID)

	if newItem {
		p.WriteBytes(addItem(item, true))
	} else {
		p.WriteInt16(item.slotID)
		p.WriteInt16(item.amount)
	}

	return p
}

func PacketInventoryAddItems(items []item, newItem []bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)

	p.WriteByte(0x01)
	if len(items) != len(newItem) {
		p.WriteByte(0)
		return p
	}

	p.WriteByte(byte(len(items)))

	for i, v := range items {
		p.WriteBool(!newItem[i])
		p.WriteByte(v.invID)

		if newItem[i] {
			p.WriteBytes(addItem(v, true))
		} else {
			p.WriteInt16(v.slotID)
			p.WriteInt16(v.amount)
		}
	}

	return p
}

func PacketInventoryChangeItemSlot(invTabID byte, origPos, newPos int16) mpacket.Packet {
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

func PacketInventoryRemoveItem(item item) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x03)
	p.WriteByte(item.invID)
	p.WriteInt16(item.slotID)
	p.WriteUint64(0) //?

	return p
}

func PacketInventoryChangeEquip(char Character) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerChangeAvatar)
	p.WriteInt32(char.id)
	p.WriteByte(1)
	p.WriteBytes(char.DisplayBytes())
	p.WriteByte(0xFF)
	p.WriteUint64(0) //?

	return p
}
