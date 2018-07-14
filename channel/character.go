package channel

import (
	"log"
	"math/rand"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/interop"
	"github.com/Hucaru/Valhalla/inventory"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/packets"
)

type MapleCharacter struct {
	character.Character
	conn interop.ClientConn // Might be worth compositing this in?
}

func (c *MapleCharacter) SendPacket(p maplepacket.Packet) {
	if len(p) > 0 {
		c.conn.Write(p)
	}
}

func (c *MapleCharacter) GetConn() interop.ClientConn {
	return c.conn
}

func (c *MapleCharacter) IsAdmin() bool {
	return c.conn.IsAdmin()
}

func (c *MapleCharacter) SetHP(hp int16) {
	c.Character.SetHP(c.GetHP() + hp)

	if c.GetHP() > c.GetMaxHP() {
		c.Character.SetHP(c.GetMaxHP())
	}

	c.conn.Write(packets.PlayerStatChange(true, constants.HP_ID, int32(c.GetHP())))
}

func (c *MapleCharacter) SetMP(mp int16) {
	c.Character.SetMP(c.Character.GetMP() + mp)

	if c.GetMP() > c.GetMaxMP() {
		c.Character.SetMP(c.GetMaxMP())
	}

	c.conn.Write(packets.PlayerStatChange(true, constants.MP_ID, int32(c.GetMP())))
}

func (c *MapleCharacter) SetAP(ap int16) {
	c.Character.SetAP(ap)
	c.conn.Write(packets.PlayerStatChange(true, constants.AP_ID, int32(ap)))
}

func (c *MapleCharacter) SetStr(str int16) {
	var maxValue int16 = 2000

	if c.GetStr() >= maxValue {
		c.conn.Write(packets.PlayerStatNoChange())
		return
	}

	c.Character.SetStr(str)

	c.conn.Write(packets.PlayerStatChange(true, constants.STR_ID, int32(str)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetDex(dex int16) {
	var maxValue int16 = 2000

	if c.GetDex() >= maxValue {
		c.conn.Write(packets.PlayerStatNoChange())
		return
	}

	c.Character.SetDex(dex)

	c.conn.Write(packets.PlayerStatChange(true, constants.DEX_ID, int32(dex)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetInt(intt int16) {
	var maxValue int16 = 2000

	if c.GetInt() >= maxValue {
		c.conn.Write(packets.PlayerStatNoChange())
		return
	}

	c.Character.SetInt(intt)

	c.conn.Write(packets.PlayerStatChange(true, constants.INT_ID, int32(intt)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetLuk(luk int16) {
	var maxValue int16 = 2000

	if c.GetLuk() >= maxValue {
		c.conn.Write(packets.PlayerStatNoChange())
		return
	}

	c.Character.SetLuk(luk)

	c.conn.Write(packets.PlayerStatChange(true, constants.LUK_ID, int32(luk)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetMaxHP(mp int16) {
	var maxValue int16 = 30000

	if c.GetMaxHP() >= maxValue {
		c.conn.Write(packets.PlayerStatNoChange())
		return
	}

	c.Character.SetMaxHP(mp)

	c.conn.Write(packets.PlayerStatChange(true, constants.MAX_HP_ID, int32(mp)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetMaxMP(hp int16) {
	var maxValue int16 = 30000

	if c.GetMaxMP() >= maxValue {
		c.conn.Write(packets.PlayerStatNoChange())
		return
	}

	c.Character.SetMaxMP(hp)

	c.conn.Write(packets.PlayerStatChange(true, constants.MAX_MP_ID, int32(hp)))

	c.SetAP(c.GetAP() - 1)
}

func (c *MapleCharacter) SetSP(sp int16) {
	c.Character.SetSP(sp)
	c.conn.Write(packets.PlayerStatChange(true, constants.SP_ID, int32(sp)))
}

func (c *MapleCharacter) UpdateSkill(id, level int32) {
	c.Character.UpdateSkill(id, level)
	c.SetSP(c.GetSP() - 1)
	c.conn.Write(packets.PlayerSkillBookUpdate(id, level))
}

func (c *MapleCharacter) ChangeMap(mapID int32, portal maplePortal, pID byte) {
	Maps.GetMap(c.GetCurrentMap()).RemovePlayer(c.conn)

	c.SetX(portal.GetX())
	c.SetY(portal.GetY())

	c.conn.Write(packets.MapChange(mapID, 1, pID, c.GetHP())) // replace 1 with channel id
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
		hpToAdd = levelUpHp(constants.BEGGINNER_HP_ADD, 0)
		mpToAdd = levelUpMp(constants.BEGGINNER_MP_ADD, c.Character.GetInt())
	case 1:
		hpToAdd = levelUpHp(constants.WARRIOR_HP_ADD, 0)
		mpToAdd = levelUpMp(constants.WARRIOR_MP_ADD, c.Character.GetInt())
	case 2:
		hpToAdd = levelUpHp(constants.MAGICIAN_HP_ADD, 0)
		mpToAdd = levelUpMp(constants.MAGICIAN_MP_ADD, 2*c.Character.GetInt())
	case 3:
		hpToAdd = levelUpHp(constants.BOWMAN_HP_ADD, 0)
		mpToAdd = levelUpMp(constants.BOWMAN_MP_ADD, c.Character.GetInt())
	case 4:
		hpToAdd = levelUpHp(constants.THIEF_HP_ADD, 0)
		mpToAdd = levelUpMp(constants.THIEF_MP_ADD, c.Character.GetInt())
	case 5:
		hpToAdd = constants.ADMIN_HP_ADD
		mpToAdd = constants.ADMIN_MP_ADD
	default:
		log.Println("Unknown Job ID:", c.Character.GetJob())
	}

	newHp := c.Character.GetMaxHP() + hpToAdd
	c.Character.SetMaxHP(newHp)
	c.Character.SetHP(newHp)

	newMp := c.Character.GetMaxMP() + mpToAdd
	c.Character.SetMaxMP(newMp)
	c.Character.SetMP(newMp)

	c.conn.Write(packets.PlayerStatChange(false, constants.HP_ID, int32(newHp)))
	c.conn.Write(packets.PlayerStatChange(false, constants.MAX_HP_ID, int32(newHp)))

	c.conn.Write(packets.PlayerStatChange(false, constants.MP_ID, int32(newHp)))
	c.conn.Write(packets.PlayerStatChange(false, constants.MAX_MP_ID, int32(newHp)))

	c.conn.Write(packets.PlayerStatChange(false, constants.AP_ID, int32(newAP)))
	c.conn.Write(packets.PlayerStatChange(false, constants.SP_ID, int32(newSP)))
}

func (c *MapleCharacter) SetLevel(level byte) {
	Maps.GetMap(c.GetCurrentMap()).SendPacket(packets.PlayerLevelUpAnimation(c.GetCharID()))
	delta := level - c.Character.GetLevel()

	if delta > 0 {
		for i := byte(0); i < delta; i++ {
			c.LevelUP()
		}
	}

	c.Character.SetLevel(level)
	c.conn.Write(packets.PlayerStatChange(true, constants.LEVEL_ID, int32(level)))

}

func (c *MapleCharacter) SetJob(jobID int16) {
	c.Character.SetJob(jobID)
	c.conn.Write(packets.PlayerStatChange(true, constants.JOB_ID, int32(jobID)))
}

func (c *MapleCharacter) SetMesos(val int32) {
	c.Character.SetMesos(val)
	c.conn.Write(packets.PlayerStatChange(true, constants.MESOS_ID, val))
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
			c.conn.Write(packets.PlayerStatChange(true, constants.EXP_ID, 0))
		} else if c.GetEXP()+val >= ExpTable[c.GetLevel()-1] { // bug here
			leftOver := c.GetEXP() + val - ExpTable[c.GetLevel()-1]
			c.SetLevel(c.GetLevel() + 1)
			c.SetEXP(leftOver)
			giveEXP(leftOver)
		} else {
			c.SetEXP(c.GetEXP() + val)
			c.conn.Write(packets.PlayerStatChange(true, constants.EXP_ID, c.GetEXP()))
		}
	}

	giveEXP(val)

	c.conn.Write(packets.MessageExpGained(whiteText, appearInChat, val))
}

func (c *MapleCharacter) TakeEXP(val int32) {
	if c.GetEXP() < val {
		c.SetEXP(0)
	} else {
		c.SetEXP(c.GetEXP() - val)
	}
}

func (c *MapleCharacter) UpdateItem(item inventory.Item) bool {
	for _, currentItem := range c.GetItems() {

		if currentItem.GetItemID() == item.GetItemID() && currentItem.GetInvID() == item.GetInvID() &&
			currentItem.GetSlotID() == item.GetSlotID() {

			c.Character.UpdateItem(currentItem, item)
			c.conn.Write(packets.InventoryAddItem(item, false))
			return true
		}
	}

	return false
}

func (c *MapleCharacter) GiveItem(item inventory.Item) bool {
	update := false

	var activeSlots []int16

	switch item.GetInvID() {
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
		log.Println("Trying to add item with unkown inv id:", item.GetInvID())
	}

	activeSlots[0] = 1

	for _, currentItem := range c.GetItems() {
		if currentItem.GetSlotID() < 1 || currentItem.GetInvID() != item.GetInvID() {
			continue
		}

		if inventory.IsStackable(currentItem.GetItemID(), currentItem.GetAmount()+item.GetAmount()) &&
			currentItem.GetItemID() == item.GetItemID() {

			tmp := currentItem
			tmp.SetAmount(tmp.GetAmount() + item.GetAmount())
			c.Character.UpdateItem(currentItem, tmp)
			c.conn.Write(packets.InventoryAddItem(tmp, false))
			update = true
			break
		}

		activeSlots[currentItem.GetSlotID()] = 1
	}

	if !update {
		for index, v := range activeSlots {
			if v == 0 {
				item.SetSlotID(int16(index))
				break
			}
		}

		c.AddItem(item)
		c.conn.Write(packets.InventoryAddItem(item, true))
		update = true
	}

	return update
}

func (c *MapleCharacter) TakeItem(invID byte, slotID int16, ammount int16) {
	for _, item := range c.GetItems() {
		if item.GetInvID() == invID &&
			item.GetSlotID() == slotID {
			if ammount < item.GetAmount() {
				updatedItem := item
				updatedItem.SetAmount(item.GetAmount() - ammount)
				c.UpdateItem(updatedItem)
				c.conn.Write(packets.InventoryAddItem(updatedItem, false))
			} else {
				c.RemoveItem(item)
				c.conn.Write(packets.InventoryChangeItemSlot(invID, slotID, 0))
			}
		}
	}
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
	c.conn.Write(packets.PlayerStatChange(false, constants.HP_ID, newHp))
}
