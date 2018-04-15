package inventory

import (
	"fmt"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func HandleMoveInventoryItem(conn interfaces.ClientConn, reader maplepacket.Reader) (maplepacket.Packet, uint32, character.Item) {
	invTabID := reader.ReadByte()
	origPos := reader.ReadInt16()
	newPos := reader.ReadInt16()

	amount := reader.ReadInt16() // amount?

	packet := maplepacket.NewPacket()
	dropItem := character.Item{}
	dropItem.SetSlotNumber(-1)

	if invTabID < 0 || invTabID > 5 || origPos == 0 {
		conn.Write(doNothing()) // bad packet, hacker?
		return packet, 0, dropItem
	}

	char := charsPtr.GetOnlineCharacterHandle(conn)
	currentItems := char.GetItems()

	items := make([]character.Item, 1)

	for _, i := range currentItems {
		if i.GetInvID() == invTabID {
			if i.GetSlotNumber() == origPos {
				items[0] = i
			} else if i.GetSlotNumber() == newPos {
				items = append(items, i)
			}
		}
	}

	if len(items) == 1 && newPos != 0 {
		modified := items[0]
		modified.SetSlotNumber(newPos)
		char.UpdateItem(items[0], modified)

		conn.Write(changeItemSlot(invTabID, origPos, newPos))

		if origPos < 0 || newPos < 0 && isEquipable(items[0]) {
			// unequip/equip
			packet = changeEquip(char)
		}

	} else if len(items) == 1 && newPos == 0 {
		dropItem = items[0]
		dropItem.SetSlotNumber(0)
		char.UpdateItem(items[0], dropItem)
		// conn.Write(changeItemSlot(invTabID, origPos, 0)) // successful drop of whole item
		conn.Write(doNothing())
		fmt.Println("drop amount:", amount, reader)
	} else if len(items) == 2 {
		if items[0].GetItemID() == items[1].GetItemID() && isStackable(items[0]) && isStackable(items[1]) {
			// Handle partial and complete merges
			remainder := items[0].GetAmount() + items[1].GetAmount() - constants.MAX_ITEM_STACK

			if remainder > 0 {
				// partial
				modified := items[0]
				modified.SetAmount(remainder)
				char.UpdateItem(items[0], modified)
				conn.Write(addItem(modified, false))

				modified = items[1]
				modified.SetAmount(modified.GetAmount() + items[0].GetAmount() - remainder)
				char.UpdateItem(items[1], modified)
				conn.Write(addItem(modified, false))
			} else {
				// complete
				modified := items[1]
				modified.SetAmount(modified.GetAmount() + items[0].GetAmount())
				char.UpdateItem(items[1], modified)
				conn.Write(addItem(modified, false))

				modified = items[0]
				modified.SetAmount(0)
				char.UpdateItem(items[0], modified)
				conn.Write(changeItemSlot(invTabID, origPos, 0))
			}

		} else {
			// char.SwitchItems(items[0], items[1])

			conn.Write(changeItemSlot(invTabID, origPos, newPos))

			if origPos < 0 || newPos < 0 && isEquipable(items[0]) && isEquipable(items[1]) {
				// unequip/equip
				packet = changeEquip(char)
			}

		}

	} else {
		// Shouldn't be able to get here, but handle just in case
		conn.Write(doNothing())
	}

	return packet, char.GetCurrentMap(), dropItem
}
