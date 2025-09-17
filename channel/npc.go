package channel

import (
	"math"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

type npc struct {
	controller *Player
	id         int32
	spawnID    int32
	pos        pos
	faceLeft   bool
	rx0, rx1   int16
}

func createNpcFromData(spawnID int32, life nx.Life) npc {
	return npc{id: life.ID,
		spawnID:  spawnID,
		pos:      newPos(life.X, life.Y, life.Foothold),
		faceLeft: life.FaceLeft,
		rx0:      life.Rx0,
		rx1:      life.Rx1}
}

func (d *npc) setController(controller *Player) {
	d.controller = controller
	controller.Send(packetNpcSetController(d.spawnID, true))
}

func (d *npc) removeController() {
	if d.controller != nil {
		d.controller.Send(packetNpcSetController(d.spawnID, false))
		d.controller = nil
	}
}

func (d npc) acknowledgeController(plr *Player, inst *fieldInstance, data []byte) {
	if d.controller.Conn != plr.Conn {
		plr.Send(packetNpcSetController(d.spawnID, false))
		return
	}

	plr.Send(packetNpcSetController(d.spawnID, true))
	inst.send(packetNpcMovement(data))
}

func packetNpcSetController(npcID int32, isLocal bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcControl)
	p.WriteBool(isLocal)
	p.WriteInt32(npcID)

	return p
}

func packetNpcMovement(bytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcMovement)
	p.WriteBytes(bytes)

	return p
}

func packetNpcChatBackNext(npcID int32, msg string, next, back bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcDialogueBox)
	p.WriteByte(4)
	p.WriteInt32(npcID)
	p.WriteByte(0)
	p.WriteString(msg)
	p.WriteBool(back)
	p.WriteBool(next)

	return p
}

func packetNpcChatOk(npcID int32, msg string) mpacket.Packet {
	return packetNpcChatBackNext(npcID, msg, false, false)
}

func packetNpcChatYesNo(npcID int32, msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcDialogueBox)
	p.WriteByte(4)
	p.WriteInt32(npcID)
	p.WriteByte(1)
	p.WriteString(msg)

	return p
}

func packetNpcChatUserString(npcID int32, msg string, defaultInput string, minLength, maxLength int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcDialogueBox)
	p.WriteByte(4)
	p.WriteInt32(npcID)
	p.WriteByte(2)
	p.WriteString(msg)
	p.WriteString(defaultInput)
	p.WriteInt16(minLength)
	p.WriteInt16(maxLength)

	return p
}

func packetNpcChatUserNumber(npcID int32, msg string, defaultInput, minLength, maxLength int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcDialogueBox)
	p.WriteByte(4)
	p.WriteInt32(npcID)
	p.WriteByte(3)
	p.WriteString(msg)
	p.WriteInt32(defaultInput)
	p.WriteInt32(minLength)
	p.WriteInt32(maxLength)

	return p
}

func packetNpcChatSelection(npcID int32, msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcDialogueBox)
	p.WriteByte(4)
	p.WriteInt32(npcID)
	p.WriteByte(4)
	p.WriteString(msg)

	return p
}

func packetNpcChatStyleWindow(npcID int32, msg string, styles []int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcDialogueBox)
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

func packetNpcChatPet(npcID int32, msg string, pets map[int64]byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcDialogueBox)
	p.WriteByte(4)
	p.WriteInt32(npcID)
	p.WriteByte(6)
	p.WriteString(msg)
	p.WriteByte(byte(len(pets)))

	for cashID, invSlot := range pets {
		p.WriteInt64(cashID)
		p.WriteByte(invSlot)
	}

	return p
}

func packetNpcChatUnkown(npcID int32, msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcDialogueBox)
	p.WriteByte(4)
	p.WriteInt32(npcID)
	p.WriteByte(7)
	p.WriteString(msg)
	p.WriteByte(1)
	p.WriteByte(1)
	p.WriteInt32(0) // decode buffer
	p.WriteByte(1)

	return p
}

func packetNpcShop(npcID int32, items [][]int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcShop)
	p.WriteInt32(npcID)
	p.WriteInt16(int16(len(items)))

	for _, currentItem := range items {
		id := currentItem[0]
		p.WriteInt32(id)

		item, err := nx.GetItem(id)

		// Price: explicit when provided
		if len(currentItem) == 2 {
			p.WriteInt32(currentItem[1])
		} else {
			if err != nil {
				p.WriteInt32(math.MaxInt32)
			} else {
				p.WriteInt32(item.Price)
			}
		}

		// Rechargeable
		cat := id / 10000
		if cat == 207 || cat == 233 {
			unitPrice := 0.0
			if err == nil {
				unitPrice = item.UnitPrice
			}
			// Write as double raw bits
			p.WriteUint64(math.Float64bits(unitPrice))
		}

		// Slot max (unsigned 16-bit), default 100 when 0 or missing
		var slotMax uint16 = 100
		if err == nil && item.SlotMax != 0 {
			// item.SlotMax is likely int16; clamp to uint16 range just in case
			if item.SlotMax > 0 {
				slotMax = uint16(item.SlotMax)
			}
		}
		p.WriteUint16(slotMax)
	}

	// Trailing padding used by clients
	p.WriteInt64(0)
	p.WriteInt64(0)
	p.WriteInt64(0)

	return p
}

func packetNpcShopResult(code byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcShopResult)
	p.WriteByte(code)

	return p
}

func packetNpcShopContinue() mpacket.Packet {
	return packetNpcShopResult(0x08)
}

func packetShopNotEnoughStock() mpacket.Packet {
	return packetNpcShopResult(0x09)
}

func packetShopNotEnoughMesos() mpacket.Packet {
	return packetNpcShopResult(0x0A)
}

// TODO: Move this into rooms?
func packetTradeError() mpacket.Packet {
	return packetNpcShopResult(0xFF)
}

func packetNpcStorageShow(npcID, storageMesos int32, storageSlots byte, items []Item) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcStorage)
	p.WriteInt32(npcID)
	p.WriteByte(storageSlots)
	// flag for if to show mesos, and Item tabs 1 - 5
	// mesos = 0x02
	// equip = 0x04
	// use = 0x08
	// setup = 0x10
	// etc = equip (old version bug)/0x20
	// pet = 0x40
	p.WriteInt16(0x7e) // allow everything
	p.WriteInt32(storageMesos)
	// loop over valid tabs and show items
	// p.WriteByte(length of items in this inventory slot)
	for _, item := range items {
		p.WriteBytes(item.shortBytes())
	}

	return p
}
