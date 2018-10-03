package handlers

import (
	"fmt"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/consts"
	"github.com/Hucaru/Valhalla/inventory"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/packets"
)

func handleMoveInventoryItem(conn *connection.Channel, reader maplepacket.Reader) {
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
			if item.InvID == invTabID {
				if item.SlotID == origPos {
					items[0] = item
				} else if item.SlotID == newPos {
					items = append(items, item)
				}
			}
		}

		if len(items) == 1 && newPos != 0 {
			items[0].SlotID = newPos
			char.UpdateItem(items[0])

			if origPos < 0 || newPos < 0 && isEquipable(items[0]) {
				// unequip/equip
				channel.Maps.GetMap(char.GetCurrentMap()).SendPacketExcept(packets.InventoryChangeEquip(char.Character), conn)
			}

		} else if len(items) == 1 && newPos == 0 {
			items[0].SlotID = 0
			char.UpdateItem(items[0])

			fmt.Println("drop amount:", amount, reader)

		} else if len(items) == 2 {
			if items[0].ItemID == items[1].ItemID && inventory.IsStackable(items[0].ItemID, items[0].Amount) && inventory.IsStackable(items[1].ItemID, items[1].Amount) {
				// Handle partial and complete merges
				remainder := items[0].Amount + items[1].Amount - consts.MAX_ITEM_STACK

				if remainder > 0 {
					// partial
					items[1].Amount += items[0].Amount - remainder
					char.UpdateItem(items[1])

					items[0].Amount = remainder
					char.UpdateItem(items[0])
				} else {
					// complete
					items[1].Amount += items[0].Amount
					char.UpdateItem(items[1])
					char.TakeItem(items[0], items[0].Amount)
				}

			} else {
				slot := items[0].SlotID
				items[0].SlotID = items[1].SlotID
				char.UpdateItem(items[0])

				cItems := char.GetItems()

				for i, v := range cItems {
					if v.UUID == items[1].UUID {
						cItems[i].SlotID = slot
					}
				}

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
	if item.ItemID/1e6 != 5 || item.InvID != 1 {
		return true
	}

	return false
}
