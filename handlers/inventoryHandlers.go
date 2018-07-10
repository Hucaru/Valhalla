package handlers

import (
	"fmt"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/interop"
	"github.com/Hucaru/Valhalla/inventory"
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

	items := make([]inventory.Item, 1)

	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
		currentItems := char.GetItems()

		for _, item := range currentItems {
			if item.GetInvID() == invTabID {
				if item.GetSlotID() == origPos {
					items[0] = item
				} else if item.GetSlotID() == newPos {
					items = append(items, item)
				}
			}
		}

		if len(items) == 1 && newPos != 0 {
			modified := items[0]
			modified.SetSlotID(newPos)
			char.UpdateItem(items[0], modified)

			conn.Write(packets.InventoryChangeItemSlot(invTabID, origPos, newPos))

			if origPos < 0 || newPos < 0 && isEquipable(items[0]) {
				// unequip/equip
				channel.Maps.GetMap(char.GetCurrentMap()).SendPacketExcept(packets.InventoryChangeEquip(char.Character), conn)
			}

		} else if len(items) == 1 && newPos == 0 {
			dropItem := items[0]
			dropItem.SetSlotID(0)
			char.UpdateItem(items[0], dropItem)
			conn.Write(packets.InventoryChangeItemSlot(invTabID, origPos, 0)) // successful drop of whole item
			conn.Write(packets.PlayerStatNoChange())
			fmt.Println("drop amount:", amount, reader)

		} else if len(items) == 2 {
			if items[0].GetItemID() == items[1].GetItemID() && inventory.IsStackable(items[0].GetItemID(), items[0].GetAmount()) && inventory.IsStackable(items[1].GetItemID(), items[1].GetAmount()) {
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

func isEquipable(item inventory.Item) bool {
	if item.GetItemID()/1e6 != 5 || item.GetInvID() != 1 {
		return true
	}

	return false
}
