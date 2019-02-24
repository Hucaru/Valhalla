package game

import (
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/game/packet"
)

func (p *Player) MoveItem(a def.Item, newPos int16) {
	switch a.InvID {
	case 1:
		for i, item := range p.char.Inventory.Equip {
			if item.SlotID == a.SlotID {
				p.char.Inventory.Equip[i].SlotID = newPos
			}
		}

		p.Send(packet.InventoryChangeItemSlot(a.InvID, a.SlotID, newPos))

		if newPos < 0 || a.SlotID < 0 {
			Maps[p.char.MapID].SendExcept(packet.InventoryChangeEquip(*p.char), p.MConnChannel, p.InstanceID)
		}
	case 2:
		for i, item := range p.char.Inventory.Use {
			if item.SlotID == a.SlotID {
				p.char.Inventory.Use[i].SlotID = newPos
			}
		}

		p.Send(packet.InventoryChangeItemSlot(a.InvID, a.SlotID, newPos))
	case 3:
		for i, item := range p.char.Inventory.SetUp {
			if item.SlotID == a.SlotID {
				p.char.Inventory.SetUp[i].SlotID = newPos
			}
		}

		p.Send(packet.InventoryChangeItemSlot(a.InvID, a.SlotID, newPos))
	case 4:
		for i, item := range p.char.Inventory.Etc {
			if item.SlotID == a.SlotID {
				p.char.Inventory.Etc[i].SlotID = newPos
			}
		}

		p.Send(packet.InventoryChangeItemSlot(a.InvID, a.SlotID, newPos))
	case 5:
		for i, item := range p.char.Inventory.Cash {
			if item.SlotID == a.SlotID {
				p.char.Inventory.Cash[i].SlotID = newPos
			}
		}

		p.Send(packet.InventoryChangeItemSlot(a.InvID, a.SlotID, newPos))
	}
}

func (p *Player) SwapItems(a, b def.Item) {
	if a.InvID != b.InvID {
		return
	}

	switch a.InvID {
	case 1:
		for i, item := range p.char.Inventory.Equip {
			if item.SlotID == a.SlotID {
				p.char.Inventory.Equip[i].SlotID = b.SlotID
			} else if item.SlotID == b.SlotID {
				p.char.Inventory.Equip[i].SlotID = a.SlotID
			}
		}

		p.Send(packet.InventoryChangeItemSlot(a.InvID, a.SlotID, b.SlotID))

		if a.SlotID < 0 || b.SlotID < 0 {
			Maps[p.char.MapID].SendExcept(packet.InventoryChangeEquip(*p.char), p.MConnChannel, p.InstanceID)
		}
	case 2:
		// determine if items can be stacked together
		if a.ItemID == b.ItemID && a.ItemID/1e4 != 207 && b.Amount < constant.MaxItemStack {
			if a.Amount+b.Amount <= constant.MaxItemStack {
				b.Amount += a.Amount
				p.Send(packet.InventoryAddItem(b, false))

				for i, item := range p.char.Inventory.Use {
					if item.SlotID == a.SlotID {
						p.char.Inventory.Use[i] = p.char.Inventory.Use[len(p.char.Inventory.Use)-1]
						p.char.Inventory.Use = p.char.Inventory.Use[:len(p.char.Inventory.Use)-1]
						p.Send(packet.InventoryRemoveItem(a))
						break
					}
				}
			} else {
				a.Amount += b.Amount - constant.MaxItemStack
				b.Amount = constant.MaxItemStack

				for i, item := range p.char.Inventory.Use {
					if item.SlotID == a.SlotID {
						p.char.Inventory.Use[i].Amount = a.Amount
						p.Send(packet.InventoryAddItem(a, false))
					} else if item.SlotID == b.SlotID {
						p.char.Inventory.Use[i].Amount = b.Amount
						p.Send(packet.InventoryAddItem(b, false))
					}
				}
			}
		} else {
			for i, item := range p.char.Inventory.Use {
				if item.SlotID == a.SlotID {
					p.char.Inventory.Use[i].SlotID = b.SlotID
				} else if item.SlotID == b.SlotID {
					p.char.Inventory.Use[i].SlotID = a.SlotID
				}
			}

			p.Send(packet.InventoryChangeItemSlot(a.InvID, a.SlotID, b.SlotID))
		}
	case 3:
		for i, item := range p.char.Inventory.SetUp {
			if item.SlotID == a.SlotID {
				p.char.Inventory.SetUp[i].SlotID = b.SlotID
			} else if item.SlotID == b.SlotID {
				p.char.Inventory.SetUp[i].SlotID = a.SlotID
			}
		}

		p.Send(packet.InventoryChangeItemSlot(a.InvID, a.SlotID, b.SlotID))
	case 4:
		for i, item := range p.char.Inventory.Etc {
			if item.SlotID == a.SlotID {
				p.char.Inventory.Etc[i].SlotID = b.SlotID
			} else if item.SlotID == b.SlotID {
				p.char.Inventory.Etc[i].SlotID = a.SlotID
			}
		}

		p.Send(packet.InventoryChangeItemSlot(a.InvID, a.SlotID, b.SlotID))
	case 5:
		for i, item := range p.char.Inventory.Cash {
			if item.SlotID == a.SlotID {
				p.char.Inventory.Cash[i].SlotID = b.SlotID
			} else if item.SlotID == b.SlotID {
				p.char.Inventory.Cash[i].SlotID = a.SlotID
			}
		}

		p.Send(packet.InventoryChangeItemSlot(a.InvID, a.SlotID, b.SlotID))
	}

}

func (p *Player) GiveItem(a def.Item) {
	var items []def.Item
	var maxInvSize byte

	switch a.InvID {
	case 1:
		items = p.char.Inventory.Equip
		maxInvSize = p.char.EquipSlotSize
	case 2:
		items = p.char.Inventory.Use
		maxInvSize = p.char.UseSlotSize
	case 3:
		items = p.char.Inventory.SetUp
		maxInvSize = p.char.SetupSlotSize
	case 4:
		items = p.char.Inventory.Etc
		maxInvSize = p.char.EtcSlotSize
	case 5:
		items = p.char.Inventory.Cash
		maxInvSize = p.char.CashSlotSize
	}

	filledSlots := make([]bool, maxInvSize)
	for _, item := range items {
		if item.SlotID < 1 {
			continue
		}
		filledSlots[item.SlotID-1] = true
	}

	var availableSlot int16
	for i, slot := range filledSlots {
		if slot == false {
			availableSlot = int16(i + 1)
			break
		}
	}

	if availableSlot == 0 {
		p.Send(packet.MessageRedText("Not enough inventory space"))
	}

	a.SlotID = availableSlot

	switch a.InvID {
	case 1:
		p.char.Inventory.Equip = append(p.char.Inventory.Equip, a)
	case 2:
	case 3:
		p.char.Inventory.SetUp = append(p.char.Inventory.SetUp, a)
	case 4:
	case 5:
		p.char.Inventory.Cash = append(p.char.Inventory.Cash, a)
	}

	p.Send(packet.InventoryAddItem(a, true))

}

func (p *Player) TakeItem() {

}
