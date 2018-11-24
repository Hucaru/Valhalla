package packet

import (
	"github.com/Hucaru/Valhalla/consts/opcodes"
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

func NpcShow(npc def.NPC) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelNpcShow)
	p.WriteInt32(npc.SpawnID)
	p.WriteInt32(npc.ID)
	p.WriteInt16(npc.X)
	p.WriteInt16(npc.Y)

	p.WriteBool(!npc.FacesLeft)

	p.WriteInt16(npc.Foothold)
	p.WriteInt16(npc.Rx0)
	p.WriteInt16(npc.Rx1)

	return p
}

func NPCRemove(npcID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelNpcRemove)
	p.WriteInt32(npcID)

	return p
}

func NPCSetController(npcID int32, isLocal bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelNpcControl)
	p.WriteBool(isLocal)
	p.WriteInt32(npcID)

	return p
}

func NPCMovement(bytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelNpcMovement)
	p.WriteBytes(bytes)

	return p
}

func NPCChatBackNext(npcID int32, msg string, front, back bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelNpcDialogueBox)
	p.WriteByte(4)
	p.WriteInt32(npcID)
	p.WriteByte(0)
	p.WriteString(msg)
	p.WriteBool(front)
	p.WriteBool(back)

	return p
}

func NPCChatYesNo(npcID int32, msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelNpcDialogueBox)
	p.WriteByte(4)
	p.WriteInt32(npcID)
	p.WriteByte(1)
	p.WriteString(msg)

	return p
}

func NPCChatUserString(npcID int32, msg string, defaultInput string, minLength, maxLength int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelNpcDialogueBox)
	p.WriteByte(4)
	p.WriteInt32(npcID)
	p.WriteByte(2)
	p.WriteString(msg)
	p.WriteString(defaultInput)
	p.WriteInt16(minLength)
	p.WriteInt16(maxLength)

	return p
}

func NPCChatUserNumber(npcID int32, msg string, defaultInput, minLength, maxLength int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelNpcDialogueBox)
	p.WriteByte(4)
	p.WriteInt32(npcID)
	p.WriteByte(3)
	p.WriteString(msg)
	p.WriteInt32(defaultInput)
	p.WriteInt32(minLength)
	p.WriteInt32(maxLength)

	return p
}

func NPCChatSelection(npcID int32, msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelNpcDialogueBox)
	p.WriteByte(4)
	p.WriteInt32(npcID)
	p.WriteByte(4)
	p.WriteString(msg)

	return p
}

func NPCChatStyleWindow(npcID int32, msg string, styles []int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelNpcDialogueBox)
	p.WriteByte(4)
	p.WriteInt32(npcID)
	p.WriteByte(5)
	p.WriteString(msg)
	p.WriteByte(byte(len(styles)))

	for _, style := range styles {
		p.WriteInt32(style)
	}

	return p
}

func NPCChatUnkown1(npcID int32, msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelNpcDialogueBox)
	p.WriteByte(4)
	p.WriteInt32(npcID)
	p.WriteByte(6)
	p.WriteString(msg)
	// Unkown from here
	p.WriteByte(0)
	p.WriteBytes([]byte{}) // buffer for something to be memcopy in client
	p.WriteByte(0)

	return p
}

func NPCChatUnkown2(npcID int32, msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelNpcDialogueBox)
	p.WriteByte(4)
	p.WriteInt32(npcID)
	p.WriteByte(6)
	p.WriteString(msg)
	// Unkown from here
	p.WriteByte(0)
	p.WriteByte(0)
	p.WriteBytes([]byte{}) // buffer for something to be memcopy in client
	p.WriteByte(0)

	return p
}

func NPCShop(npcID int32, items [][]int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelNpcShop)
	p.WriteInt32(npcID)
	p.WriteInt16(int16(len(items)))

	for _, currentItem := range items {
		p.WriteInt32(currentItem[0])

		if len(currentItem) == 2 {
			p.WriteInt32(currentItem[1])

		} else {
			p.WriteInt32(nx.Items[currentItem[0]].Price)
		}

		if def.ItemIsRechargeable(currentItem[0]) {
			p.WriteUint64(uint64(nx.Items[currentItem[0]].UnitPrice * float64(nx.Items[currentItem[0]].SlotMax)))
		}

		if nx.Items[currentItem[0]].SlotMax == 0 {
			p.WriteInt16(100)
		} else {
			p.WriteInt16(nx.Items[currentItem[0]].SlotMax)
		}
	}

	return p
}

func nPCShopResult(code byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelNpcShopResult)
	p.WriteByte(code)

	return p
}

func NPCShopContinue() mpacket.Packet {
	return nPCShopResult(0x08)
}

func NPCShopNotEnoughStock() mpacket.Packet {
	return nPCShopResult(0x09)
}

func NPCShopNotEnoughMesos() mpacket.Packet {
	return nPCShopResult(0x0A)
}

func NPCTradeError() mpacket.Packet {
	return nPCShopResult(0xFF)
}

func NPCStorageShow(npcID, storageMesos int32, storageSlots byte, items []def.Item) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelNpcStorage)
	p.WriteInt32(npcID)
	p.WriteByte(storageSlots)
	p.WriteInt16(0x7e)
	p.WriteInt32(storageMesos)
	for _, item := range items {
		addItem(item, true)
	}

	return p
}
