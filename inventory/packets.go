package inventory

import (
	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func addItem(item character.Item, newItem bool) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INVENTORY_OPERATION)
	p.WriteByte(0x01) // something to do with time
	p.WriteByte(0x01)
	p.WriteByte(0x01) // operation type
	p.WriteByte(item.GetInvID())

	if newItem {
		// add item data
	} else {
		p.WriteInt16(item.GetSlotNumber())
		p.WriteInt16(item.GetAmount()) // the new amount value (not a delta)
	}

	return p
}

func changeItemSlot(invTabID byte, origPos, newPos int16) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INVENTORY_OPERATION)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x02)
	p.WriteByte(invTabID)
	p.WriteInt16(origPos)
	p.WriteInt16(newPos)
	p.WriteByte(0x00) // ?

	return p
}

func changeEquip(char *character.Character) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_CHANGE_AVATAR)
	p.WriteUint32(char.GetCharID())
	p.WriteByte(1)
	character.WriteDisplayCharacter(char, &p)
	p.WriteByte(0xFF)
	p.WriteUint64(0) //?

	return p
}

func doNothing() maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INVENTORY_OPERATION)
	p.WriteByte(0x01)
	p.WriteByte(0x00)

	return p
}
