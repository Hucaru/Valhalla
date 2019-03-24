package game

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

func PacketInventoryAddItem(item Item, newItem bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelInventoryOperation)
	p.WriteByte(0x01)     // ?
	p.WriteByte(0x01)     // number of operations? // e.g. loop over multiple interweaved operations
	p.WriteBool(!newItem) // operation type
	p.WriteByte(item.InvID)

	if newItem {
		p.WriteBytes(addItem(item, true))
	} else {
		p.WriteInt16(item.SlotID)
		p.WriteInt16(item.Amount) // the new amount value (not a delta)
	}

	return p
}

func PacketInventoryChangeItemSlot(invTabID byte, origPos, newPos int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x02)
	p.WriteByte(invTabID)
	p.WriteInt16(origPos)
	p.WriteInt16(newPos)
	p.WriteByte(0x00) // ?

	return p
}

func PacketInventoryRemoveItem(item Item) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x03)
	p.WriteByte(item.InvID)
	p.WriteInt16(item.SlotID)
	p.WriteUint64(0) //?

	return p
}

func PacketInventoryChangeEquip(char Character) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelPlayerChangeAvatar)
	p.WriteInt32(char.ID)
	p.WriteByte(1)
	p.WriteBytes(writeDisplayCharacter(char))
	p.WriteByte(0xFF)
	p.WriteUint64(0) //?

	return p
}
