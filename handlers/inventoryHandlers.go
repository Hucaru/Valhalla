package handlers

import (
	"fmt"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/interop"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/packets"
)

func handleMoveInventoryItem(conn interop.ClientConn, reader maplepacket.Reader) {
	invTabID := reader.ReadByte()
	origPos := reader.ReadInt16()
	newPos := reader.ReadInt16()

	amount := reader.ReadInt16() // amount?

	if invTabID < 0 || invTabID > 5 || origPos == 0 {
		conn.Write(packets.PlayerStatNoChange()) // bad packet, hacker?
	}

	items := make([]character.Item, 1)

	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
		currentItems := char.GetItems()

		for _, item := range currentItems {
			if item.GetInvID() == invTabID {
				if item.GetSlotNumber() == origPos {
					items[0] = item
				} else if item.GetSlotNumber() == newPos {
					items = append(items, item)
				}
			}
		}

		if len(items) == 1 && newPos != 0 {
			modified := items[0]
			modified.SetSlotNumber(newPos)
			char.UpdateItem(items[0], modified)

			conn.Write(packets.InventoryChangeItemSlot(invTabID, origPos, newPos))

			if origPos < 0 || newPos < 0 && isEquipable(items[0]) {
				// unequip/equip
				channel.Maps.GetMap(char.GetCurrentMap()).SendPacketExcept(packets.InventoryChangeEquip(char.Character), conn)
			}

		} else if len(items) == 1 && newPos == 0 {
			dropItem := items[0]
			dropItem.SetSlotNumber(0)
			char.UpdateItem(items[0], dropItem)
			conn.Write(packets.InventoryChangeItemSlot(invTabID, origPos, 0)) // successful drop of whole item
			conn.Write(packets.PlayerStatNoChange())
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
					conn.Write(packets.InventoryAddItem(modified, false))

					modified = items[1]
					modified.SetAmount(modified.GetAmount() + items[0].GetAmount() - remainder)
					char.UpdateItem(items[1], modified)
					conn.Write(packets.InventoryAddItem(modified, false))
				} else {
					// complete
					modified := items[1]
					modified.SetAmount(modified.GetAmount() + items[0].GetAmount())
					char.UpdateItem(items[1], modified)
					conn.Write(packets.InventoryAddItem(modified, false))

					char.RemoveItem(items[0])
					conn.Write(packets.InventoryChangeItemSlot(invTabID, origPos, 0))
				}

			} else {
				char.SwitchItems(items[0], items[1])

				conn.Write(packets.InventoryChangeItemSlot(invTabID, origPos, newPos))

				if origPos < 0 || newPos < 0 && isEquipable(items[0]) && isEquipable(items[1]) {
					// unequip/equip
					channel.Maps.GetMap(char.GetCurrentMap()).SendPacketExcept(packets.InventoryChangeEquip(char.Character), conn)
				}

			}

		} else {
			// Shouldn't be able to get here, but handle just in case
			conn.Write(packets.PlayerStatNoChange())
		}

	})
}

func isStackable(item character.Item) bool {
	if item.GetItemID()/1e6 != 5 && // pet item
		item.GetInvID() != 1 && // equip
		item.GetItemID()/1e4 != 207 && // star/arrow etc
		item.GetAmount() < constants.MAX_ITEM_STACK {

		return true
	}

	return false
}

func isEquipable(item character.Item) bool {
	if item.GetItemID()/1e6 != 5 || item.GetInvID() != 1 {
		return true
	}

	return false
}
