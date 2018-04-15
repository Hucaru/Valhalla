package inventory

import (
	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func HandleMoveInventoryItem(conn interfaces.ClientConn, reader maplepacket.Reader) (maplepacket.Packet, uint32, character.Item) {
	invTabID := reader.ReadByte()
	origPos := reader.ReadInt16()
	newPos := reader.ReadInt16()

	reader.ReadInt16() // amount?

	packet := []byte{}
	dropItem := character.Item{}
	dropItem.SetSlotNumber(-1)

	if invTabID < 0 || invTabID > 5 || origPos == 0 {
		conn.Write(doNothing()) // bad packet, hacker?
		return packet, 0, dropItem
	}

	char := charsPtr.GetOnlineCharacterHandle(conn)
	items := char.GetItems()

	foundItems := make([]character.Item, 2)

	for _, i := range items {
		if i.GetInvID() == invTabID {
			if i.GetSlotNumber() == origPos {
				foundItems[0] = i
			} else if i.GetSlotNumber() == newPos {
				foundItems[1] = i
			}
		}
	}

	if len(foundItems) < 1 { // Someone is trying to move/drop an item that does not exist
		return packet, 0, dropItem
	}

	if newPos == 0 {
		// drop item
		dropItem = foundItems[0]
		conn.Write(doNothing())
	} else {
		if len(foundItems) == 1 {
			modified := foundItems[0]
			modified.SetSlotNumber(newPos)
			char.UpdateItem(foundItems[0], modified)

			conn.Write(changeItemSlot(invTabID, origPos, newPos))

			if origPos < 0 || newPos < 0 {
				// unequip/equip
				packet = changeEquip(char)
			}

		} else if len(foundItems) == 2 {
			if foundItems[0].GetItemID() == foundItems[1].GetItemID() &&
				foundItems[1].GetItemID()/1e6 != 5 && // pet item
				foundItems[1].GetInvID() != 1 && // equip
				foundItems[1].GetItemID()/1e4 != 207 { // rechargeable item
				// Handle partial and complete merges
			} else {
				char.SwitchItems(foundItems[0], foundItems[1])
				conn.Write(changeItemSlot(invTabID, origPos, newPos))

			}

		} else {
			// Shouldn't be able to get here, but handle just in case
			conn.Write(doNothing())
		}
	}

	return packet, char.GetCurrentMap(), dropItem
}
