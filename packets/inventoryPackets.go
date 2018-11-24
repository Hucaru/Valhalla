package packets

import (
	"github.com/Hucaru/Valhalla/consts/opcodes"
	"github.com/Hucaru/Valhalla/def"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func InventoryAddItem(item def.Item, newItem bool) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelInventoryOperation)
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

func InventoryChangeItemSlot(invTabID byte, origPos, newPos int16) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x02)
	p.WriteByte(invTabID)
	p.WriteInt16(origPos)
	p.WriteInt16(newPos)
	p.WriteByte(0x00) // ?

	return p
}

func InventoryRemoveItem(item def.Item) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x03)
	p.WriteByte(item.InvID)
	p.WriteInt16(item.SlotID)
	p.WriteUint64(0) //?

	return p
}

func InventoryChangeEquip(char def.Character) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelPlayerChangeAvatar)
	p.WriteInt32(char.ID)
	p.WriteByte(1)
	p.WriteBytes(writeDisplayCharacter(char))
	p.WriteByte(0xFF)
	p.WriteUint64(0) //?

	return p
}
