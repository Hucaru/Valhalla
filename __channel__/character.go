package channel

import (
	"log"
	"math/rand"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/consts"
	"github.com/Hucaru/Valhalla/inventory"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/game/packet"
)

type MapleCharacter struct {
	character.Character
	conn mnet.MConnChannel // Might be worth compositing this in?
}

func (c *MapleCharacter) SendPacket(p mpacket.Packet) {
	if len(p) > 0 {
		c.conn.Send(p)
	}
}

func (c *MapleCharacter) GetConn() mnet.MConnChannel {
	return c.conn
}

func (c *MapleCharacter) GetAdmin() bool {
	return c.conn.GetAdmin()
}

func (c *MapleCharacter) SetHP(hp int16) {
	c.Character.SetHP(c.GetHP() + hp)

	if c.GetHP() > c.GetMaxHP() {
		c.Character.SetHP(c.GetMaxHP())
	}

	c.conn.Send(packet.PlayerStatChange(true, consts.HP_ID, int32(c.GetHP())))
}

func (c *MapleCharacter) SetMP(mp int16) {
	c.Character.SetMP(c.Character.GetMP() + mp)

	if c.GetMP() > c.GetMaxMP() {
		c.Character.SetMP(c.GetMaxMP())
	}

	c.conn.Send(packet.PlayerStatChange(true, consts.MP_ID, int32(c.GetMP())))
}

func (c *MapleCharacter) SetAP(ap int16) {
	c.Character.SetAP(ap)
	c.conn.Send(packet.PlayerStatChange(true, consts.AP_ID, int32(ap)))
}

func (c *MapleCharacter) SetStr(str int16) {
	var maxValue int16 = 2000

	if c.GetStr() >= maxValue {
		c.conn.Send(packet.PlayerStatNoChange())
		return
	}

	c.Character.SetStr(str)

	c.conn.Send(packet.PlayerStatChange(true, consts.STR_ID, int32(str)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetDex(dex int16) {
	var maxValue int16 = 2000

	if c.GetDex() >= maxValue {
		c.conn.Send(packet.PlayerStatNoChange())
		return
	}

	c.Character.SetDex(dex)

	c.conn.Send(packet.PlayerStatChange(true, consts.DEX_ID, int32(dex)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetInt(intt int16) {
	var maxValue int16 = 2000

	if c.GetInt() >= maxValue {
		c.conn.Send(packet.PlayerStatNoChange())
		return
	}

	c.Character.SetInt(intt)

	c.conn.Send(packet.PlayerStatChange(true, consts.INT_ID, int32(intt)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetLuk(luk int16) {
	var maxValue int16 = 2000

	if c.GetLuk() >= maxValue {
		c.conn.Send(packet.PlayerStatNoChange())
		return
	}

	c.Character.SetLuk(luk)

	c.conn.Send(packet.PlayerStatChange(true, consts.LUK_ID, int32(luk)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetMaxHP(mp int16) {
	var maxValue int16 = 30000

	if c.GetMaxHP() >= maxValue {
		c.conn.Send(packet.PlayerStatNoChange())
		return
	}

	c.Character.SetMaxHP(mp)

	c.conn.Send(packet.PlayerStatChange(true, consts.MAX_HP_ID, int32(mp)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetMaxMP(hp int16) {
	var maxValue int16 = 30000

	if c.GetMaxMP() >= maxValue {
		c.conn.Send(packet.PlayerStatNoChange())
		return
	}

	c.Character.SetMaxMP(hp)

	c.conn.Send(packet.PlayerStatChange(true, consts.MAX_MP_ID, int32(hp)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetSP(sp int16) {
	c.Character.SetSP(sp)
	c.conn.Send(packet.PlayerStatChange(true, consts.SP_ID, int32(sp)))
}

func (c *MapleCharacter) UpdateSkill(id, level int32) {
	c.Character.UpdateSkill(id, level)
	c.SetSP(c.GetSP() - 1)
	c.conn.Send(packet.PlayerSkillBookUpdate(id, level))
}

func (c *MapleCharacter) ChangeMap(mapID int32, portal maplePortal, pID byte) {
	Maps.GetMap(c.GetCurrentMap()).RemovePlayer(c.conn)

	c.SetX(portal.GetX())
	c.SetY(portal.GetY())

	c.conn.Send(packet.MapChange(mapID, 1, pID, c.GetHP())) // replace 1 with channel id
	c.SetCurrentMap(mapID)
	Maps.GetMap(mapID).AddPlayer(c.conn)
}

func (c *MapleCharacter) LevelUP() {
	newAP := c.Character.GetAP() + 5
	c.Character.SetAP(newAP)

	newSP := c.Character.GetSP() + 3
	c.Character.SetSP(newSP)

	var hpToAdd int16
	var mpToAdd int16

	levelUpHp := func(classIncrease int16, bonus int16) int16 {
		return int16(rand.Intn(3)+1) + classIncrease + bonus // deterministic rand, maybe seed with time?
	}

	levelUpMp := func(classIncrease int16, bonus int16) int16 {
		return int16(rand.Intn(1)+1) + classIncrease + bonus // deterministic rand, maybe seed with time?
	}

	switch int(c.Character.GetJob() / 100) {
	case 0:
		hpToAdd = levelUpHp(consts.BEGGINNER_HP_ADD, 0)
		mpToAdd = levelUpMp(consts.BEGGINNER_MP_ADD, c.Character.GetInt())
	case 1:
		hpToAdd = levelUpHp(consts.WARRIOR_HP_ADD, 0)
		mpToAdd = levelUpMp(consts.WARRIOR_MP_ADD, c.Character.GetInt())
	case 2:
		hpToAdd = levelUpHp(consts.MAGICIAN_HP_ADD, 0)
		mpToAdd = levelUpMp(consts.MAGICIAN_MP_ADD, 2*c.Character.GetInt())
	case 3:
		hpToAdd = levelUpHp(consts.BOWMAN_HP_ADD, 0)
		mpToAdd = levelUpMp(consts.BOWMAN_MP_ADD, c.Character.GetInt())
	case 4:
		hpToAdd = levelUpHp(consts.THIEF_HP_ADD, 0)
		mpToAdd = levelUpMp(consts.THIEF_MP_ADD, c.Character.GetInt())
	case 5:
		hpToAdd = consts.ADMIN_HP_ADD
		mpToAdd = consts.ADMIN_MP_ADD
	default:
		log.Println("Unknown Job ID:", c.Character.GetJob())
	}

	newHp := c.Character.GetMaxHP() + hpToAdd
	c.Character.SetMaxHP(newHp)
	c.Character.SetHP(newHp)

	newMp := c.Character.GetMaxMP() + mpToAdd
	c.Character.SetMaxMP(newMp)
	c.Character.SetMP(newMp)

	c.conn.Send(packet.PlayerStatChange(false, consts.HP_ID, int32(newHp)))
	c.conn.Send(packet.PlayerStatChange(false, consts.MAX_HP_ID, int32(newHp)))

	c.conn.Send(packet.PlayerStatChange(false, consts.MP_ID, int32(newHp)))
	c.conn.Send(packet.PlayerStatChange(false, consts.MAX_MP_ID, int32(newHp)))

	c.conn.Send(packet.PlayerStatChange(false, consts.AP_ID, int32(newAP)))
	c.conn.Send(packet.PlayerStatChange(false, consts.SP_ID, int32(newSP)))
}

func (c *MapleCharacter) SetLevel(level byte) {
	Maps.GetMap(c.GetCurrentMap()).SendPacket(packet.PlayerLevelUpAnimation(c.GetCharID()))
	delta := level - c.Character.GetLevel()

	if delta > 0 {
		for i := byte(0); i < delta; i++ {
			c.LevelUP()
		}
	}

	c.Character.SetLevel(level)
	c.conn.Send(packet.PlayerStatChange(true, consts.LEVEL_ID, int32(level)))

}

func (c *MapleCharacter) SetJob(jobID int16) {
	c.Character.SetJob(jobID)
	c.conn.Send(packet.PlayerStatChange(true, consts.JOB_ID, int32(jobID)))
}

func (c *MapleCharacter) SetMesos(val int32) {
	c.Character.SetMesos(val)
	c.conn.Send(packet.PlayerStatChange(true, consts.MESOS_ID, val))
}

func (c *MapleCharacter) GiveMesos(val int32) {
	c.SetMesos(c.GetMesos() + val)
}

func (c *MapleCharacter) TakeMesos(val int32) {
	c.SetMesos(c.GetMesos() - val)
}

func (c *MapleCharacter) GiveEXP(val int32, whiteText, appearInChat bool) {
	var giveEXP func(val int32)

	giveEXP = func(val int32) {
		if c.GetLevel() > 199 {
			c.SetEXP(0)
			c.conn.Send(packet.PlayerStatChange(true, consts.EXP_ID, 0))
		} else if c.GetEXP()+val >= ExpTable[c.GetLevel()-1] { // bug here
			leftOver := c.GetEXP() + val - ExpTable[c.GetLevel()-1]
			c.SetLevel(c.GetLevel() + 1)
			c.SetEXP(leftOver)
			giveEXP(leftOver)
		} else {
			c.SetEXP(c.GetEXP() + val)
			c.conn.Send(packet.PlayerStatChange(true, consts.EXP_ID, c.GetEXP()))
		}
	}

	giveEXP(val)

	c.conn.Send(packet.MessageExpGained(whiteText, appearInChat, val))
}

func (c *MapleCharacter) TakeEXP(val int32) {
	if c.GetEXP() < val {
		c.SetEXP(0)
	} else {
		c.SetEXP(c.GetEXP() - val)
	}
}

func (c *MapleCharacter) UpdateItem(modified inventory.Item) {
	items := c.GetItems()
	for i, curItem := range items {

		if curItem.UUID == modified.UUID {
			if curItem.Amount != modified.Amount {
				c.conn.Send(packet.InventoryAddItem(modified, false))
			} else if curItem.SlotID != modified.SlotID {
				c.conn.Send(packet.InventoryChangeItemSlot(modified.InvID, curItem.SlotID, modified.SlotID))
			}

			// Add stat change packets
			items[i] = modified

			break
		}
	}

	c.conn.Send(packet.PlayerStatNoChange()) // Figure out why partial stackable item merge appears to needs this
}

func (c *MapleCharacter) GiveItem(item inventory.Item) bool {
	update := false

	var activeSlots []int16

	switch item.InvID {
	case 1:
		activeSlots = make([]int16, c.GetEquipSlotSize()+1)
	case 2:
		activeSlots = make([]int16, c.GetUsetSlotSize()+1)
	case 3:
		activeSlots = make([]int16, c.GetSetupSlotSize()+1)
	case 4:
		activeSlots = make([]int16, c.GetEtcSlotSize()+1)
	case 5:
		activeSlots = make([]int16, c.GetCashSlotSize()+1)
	default:
		log.Println("Trying to add item with unkown inv id:", item.InvID)
	}

	activeSlots[0] = 1

	items := c.GetItems()
	newItem := inventory.Item{}

	for _, curItem := range items {
		if curItem.SlotID < 1 || curItem.InvID != item.InvID {
			continue
		}

		if curItem.ItemID == item.ItemID &&
			inventory.IsStackable(curItem.ItemID, curItem.Amount+item.Amount) { // change to allow stack splitting

			ammount := item.Amount

			// if item.Amount > (consts.MAX_ITEM_STACK - curItem.Amount) {
			// 	ammount = consts.MAX_ITEM_STACK - curItem.Amount
			// }

			update = true

			curItem.Amount += ammount
			c.UpdateItem(curItem)
			break
		}

		activeSlots[curItem.SlotID] = 1
	}

	if !update {
		for index, v := range activeSlots {
			if v == 0 {
				item.SlotID = int16(index)
				break
			}
		}

		newItem = item
		c.conn.Send(packet.InventoryAddItem(newItem, true))
		update = true
	}

	c.SetItems(append(items, newItem))

	return update
}

func (c *MapleCharacter) TakeItem(modified inventory.Item, amount int16) bool {
	items := c.GetItems()

	for i, item := range items {

		if modified.InvID == item.InvID && modified.SlotID == item.SlotID {
			if amount == item.Amount {
				c.SetItems(append(items[:i], items[i+1:]...))
				c.conn.Send(packet.InventoryRemoveItem(item))
				return true
			} else if amount < item.Amount {
				item.Amount -= amount
				c.UpdateItem(item)
				return true
			}
		}

		// Redo following logic
		if modified.SlotID == 0 && modified.ItemID == item.ItemID {
			// Handle case where something has requested I would like to remove x number of item id y, e.g. remove Kerning PQ tickets
			remainder := amount
			inds := []int{}
			k := -1

			for j := range items {
				if items[j].ItemID == modified.ItemID {
					remainder -= modified.Amount

					if remainder < 1 {
						k = j
						break
					}

					inds = append(inds, j)
				}
			}

			if remainder < 1 {
				for _, v := range inds {
					c.SetItems(append(items[:v], items[v+1:]...))
					c.conn.Send(packet.InventoryRemoveItem(items[v]))
				}

				c.SetItems(append(items[:k], items[k+1:]...))
				c.conn.Send(packet.InventoryRemoveItem(items[k]))
				return true
			}
		}
	}

	c.conn.Send(packet.PlayerStatNoChange()) // find out if needed
	return false
}

func (c *MapleCharacter) TakeDamage(ammount int32) {
	delta := int32(c.Character.GetHP()) - int32(ammount)

	var newHp int32

	if delta < 1 {
		newHp = 0
	} else {
		newHp = delta
	}

	c.Character.SetHP(int16(newHp))
	c.conn.Send(packet.PlayerStatChange(false, consts.HP_ID, newHp))
}
