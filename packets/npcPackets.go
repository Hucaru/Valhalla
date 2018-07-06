package packets

import (
	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/interop"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/nx"
)

func NpcShow(npc interop.Npc) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_SHOW)
	p.WriteUint32(npc.GetSpawnID())
	p.WriteUint32(npc.GetID())
	p.WriteInt16(npc.GetX())
	p.WriteInt16(npc.GetY())

	p.WriteByte(1 - npc.GetFace())

	p.WriteInt16(npc.GetFoothold())
	p.WriteInt16(npc.GetRx0())
	p.WriteInt16(npc.GetRx1())

	return p
}

func NPCRemove(npcID uint32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_REMOVE)
	p.WriteUint32(npcID)

	return p
}

func NPCSetController(npcID uint32, isLocal bool) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_CONTROL)
	p.WriteBool(isLocal)
	p.WriteUint32(npcID)

	return p
}

func NPCMovement(bytes []byte) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_MOVEMENT)
	p.WriteBytes(bytes)

	return p
}

func NPCChatBackNext(npcID uint32, msg string, front, back bool) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_DIALOGUE_BOX)
	p.WriteByte(4)
	p.WriteUint32(npcID)
	p.WriteByte(0)
	p.WriteString(msg)
	p.WriteBool(front)
	p.WriteBool(back)

	return p
}

func NPCChatYesNo(npcID uint32, msg string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_DIALOGUE_BOX)
	p.WriteByte(4)
	p.WriteUint32(npcID)
	p.WriteByte(1)
	p.WriteString(msg)

	return p
}

func NPCChatUserString(npcID uint32, msg string, defaultInput string, minLength, maxLength uint16) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_DIALOGUE_BOX)
	p.WriteByte(4)
	p.WriteUint32(npcID)
	p.WriteByte(2)
	p.WriteString(msg)
	p.WriteString(defaultInput)
	p.WriteUint16(minLength)
	p.WriteUint16(maxLength)

	return p
}

func NPCChatUserNumber(npcID uint32, msg string, defaultInput, minLength, maxLength uint32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_DIALOGUE_BOX)
	p.WriteByte(4)
	p.WriteUint32(npcID)
	p.WriteByte(3)
	p.WriteString(msg)
	p.WriteUint32(defaultInput)
	p.WriteUint32(minLength)
	p.WriteUint32(maxLength)

	return p
}

func NPCChatSelection(npcID uint32, msg string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_DIALOGUE_BOX)
	p.WriteByte(4)
	p.WriteUint32(npcID)
	p.WriteByte(4)
	p.WriteString(msg)

	return p
}

func NPCChatStyleWindow(npcID uint32, msg string, styles []uint32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_DIALOGUE_BOX)
	p.WriteByte(4)
	p.WriteUint32(npcID)
	p.WriteByte(5)
	p.WriteString(msg)
	p.WriteByte(byte(len(styles)))

	for _, style := range styles {
		p.WriteUint32(style)
	}

	return p
}

func NPCChatUnkown1(npcID uint32, msg string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_DIALOGUE_BOX)
	p.WriteByte(4)
	p.WriteUint32(npcID)
	p.WriteByte(6)
	p.WriteString(msg)
	// Unkown from here
	p.WriteByte(0)
	p.WriteBytes([]byte{}) // buffer for something to be memcopy in client
	p.WriteByte(0)

	return p
}

func NPCChatUnkown2(npcID uint32, msg string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_DIALOGUE_BOX)
	p.WriteByte(4)
	p.WriteUint32(npcID)
	p.WriteByte(6)
	p.WriteString(msg)
	// Unkown from here
	p.WriteByte(0)
	p.WriteByte(0)
	p.WriteBytes([]byte{}) // buffer for something to be memcopy in client
	p.WriteByte(0)

	return p
}

func NPCShop(npcID uint32, items [][]uint32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_SHOP)
	p.WriteUint32(npcID)
	p.WriteUint16(uint16(len(items)))

	for _, item := range items {
		p.WriteUint32(item[0])

		if len(item) == 2 {
			p.WriteUint32(item[1])

		} else {
			p.WriteUint32(nx.Items[item[0]].Price)
		}

		if nx.IsRechargeAble(item[0]) {
			p.WriteUint64(uint64(nx.Items[item[0]].UnitPrice * float64(nx.Items[item[0]].SlotMax)))
		}

		if nx.Items[item[0]].SlotMax == 0 {
			p.WriteUint16(100)
		} else {
			p.WriteUint16(nx.Items[item[0]].SlotMax)
		}
	}

	return p
}

func nPCShopResult(code byte) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_SHOP_RESULT)
	p.WriteByte(code)

	return p
}

func NPCShopNotEnoughStock() maplepacket.Packet {
	return nPCShopResult(0x09)
}

func NPCShopNotEnoughMesos() maplepacket.Packet {
	return nPCShopResult(0x0A)
}

func NPCTradeError() maplepacket.Packet {
	return nPCShopResult(0xFF)
}

func NPCStorageShow(npcID, storageMesos uint32, storageSlots byte, items []character.Item) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_STORAGE)
	p.WriteUint32(npcID)
	p.WriteByte(storageSlots)
	p.WriteInt16(0x7e)
	p.WriteUint32(storageMesos)
	for _, item := range items {
		addItem(item)
	}

	return p
}
