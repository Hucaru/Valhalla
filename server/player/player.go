package player

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/item"
	"github.com/Hucaru/Valhalla/server/pos"
)

type sender interface {
	Send(mpacket.Packet) error
}

type instance interface {
	sender
	CalculateNearestSpawnPortalID(pos.Data) (byte, error)
	ID() int
}

// Data connected to server
type Data struct {
	conn       mnet.Client
	instanceID int
	inst       instance

	id        int32 // Unique identifier of the character
	accountID int32
	worldID   byte

	mapID       int32
	mapPos      byte
	previousMap int32
	portalCount byte

	job int16

	level byte
	str   int16
	dex   int16
	intt  int16
	luk   int16
	hp    int16
	maxHP int16
	mp    int16
	maxMP int16
	ap    int16
	sp    int16
	exp   int32
	fame  int16

	name    string
	gender  byte
	skin    byte
	face    int32
	hair    int32
	chairID int32
	stance  byte
	pos     pos.Data
	guild   string

	equipSlotSize byte
	useSlotSize   byte
	setupSlotSize byte
	etcSlotSize   byte
	cashSlotSize  byte

	equip []item.Data
	use   []item.Data
	setUp []item.Data
	etc   []item.Data
	cash  []item.Data

	mesos int32

	skills map[int32]Skill

	miniGameWins, miniGameDraw, miniGameLoss, miniGamePoints int32

	lastAttackPacketTime int64
}

// Conn - client connection associated with this Data
func (d Data) Conn() mnet.Client {
	return d.conn
}

// InstanceID - field instance id the Data is currently on
func (d Data) InstanceID() int {
	return d.inst.ID()
}

// SetInstance of player
func (d *Data) SetInstance(inst interface{}) {
	d.inst, _ = inst.(instance)
}

// Send the Data a packet
func (d Data) Send(packet mpacket.Packet) {
	d.conn.Send(packet)
}

// SetJob id of the Data
func (d *Data) SetJob(id int16) {
	d.job = id
	d.conn.Send(packetPlayerStatChange(true, constant.JobID, int32(id)))
}

func (d *Data) levelUp(inst sender) {
	d.GiveAP(5)
	d.GiveSP(3)

	levelUpHp := func(classIncrease int16, bonus int16) int16 {
		return int16(rand.Intn(3)+1) + classIncrease + bonus // deterministic rand, maybe seed with time?
	}

	levelUpMp := func(classIncrease int16, bonus int16) int16 {
		return int16(rand.Intn(1)+1) + classIncrease + bonus // deterministic rand, maybe seed with time?
	}

	switch d.job / 100 { // add effects from skills e.g. improve max mp
	case 0:
		d.maxHP += levelUpHp(constant.BeginnerHpAdd, 0)
		d.maxMP += levelUpMp(constant.BeginnerMpAdd, d.intt)
	case 1:
		d.maxHP += levelUpHp(constant.WarriorHpAdd, 0)
		d.maxMP += levelUpMp(constant.WarriorMpAdd, d.intt)
	case 2:
		d.maxHP += levelUpHp(constant.MagicianHpAdd, 0)
		d.maxMP += levelUpMp(constant.MagicianMpAdd, 2*d.intt)
	case 3:
		d.maxHP += levelUpHp(constant.BowmanHpAdd, 0)
		d.maxMP += levelUpMp(constant.BowmanMpAdd, d.intt)
	case 4:
		d.maxHP += levelUpHp(constant.ThiefHpAdd, 0)
		d.maxMP += levelUpMp(constant.ThiefMpAdd, d.intt)
	case 5:
		d.maxHP += constant.AdminHpAdd
		d.maxMP += constant.AdminMpAdd
	default:
		log.Println("Unkown job during level up", d.job)
	}

	d.hp = d.maxHP
	d.mp = d.maxMP

	d.SetHP(d.hp)
	d.SetMaxHP(d.hp)

	d.SetMP(d.mp)
	d.SetMaxMP(d.mp)

	d.GiveLevel(1)
}

// SetEXP of the Data
func (d *Data) SetEXP(amount int32) {
	if d.level > 199 {
		d.exp = amount
		d.Send(packetPlayerStatChange(false, constant.ExpID, int32(amount)))
		return
	}

	remainder := amount - constant.ExpTable[d.level-1]

	if remainder >= 0 {
		d.levelUp(d.inst)
		d.SetEXP(remainder)
	} else {
		d.exp = amount
		d.Send(packetPlayerStatChange(false, constant.ExpID, int32(amount)))
	}
}

// GiveEXP to the Data
func (d *Data) GiveEXP(amount int32, fromMob, fromParty bool) {
	if fromMob {
		d.Send(packetMessageExpGained(!fromParty, false, amount))
	} else {
		d.Send(packetMessageExpGained(true, true, amount))
	}

	d.SetEXP(d.exp + amount)
}

// SetLevel of the Data
func (d *Data) SetLevel(amount byte) {
	d.level = amount
	d.Send(packetPlayerStatChange(false, constant.LevelID, int32(amount)))
	d.inst.Send(packetPlayerLevelUpAnimation(d.id))
}

// GiveLevel amount ot the Data
func (d *Data) GiveLevel(amount byte) {
	d.SetLevel(d.level + amount)
}

// SetAP of Data
func (d *Data) SetAP(amount int16) {
	d.ap = amount
	d.Send(packetPlayerStatChange(false, constant.ApID, int32(amount)))
}

// GiveAP to Data
func (d *Data) GiveAP(amount int16) {
	d.SetAP(d.ap + amount)
}

// SetSP of Data
func (d *Data) SetSP(amount int16) {
	d.sp = amount
	d.Send(packetPlayerStatChange(false, constant.SpID, int32(amount)))
}

// GiveSP to Data
func (d *Data) GiveSP(amount int16) {
	d.SetSP(d.sp + amount)
}

// SetStr of the Data
func (d *Data) SetStr(amount int16) {
	d.str = amount
	d.Send(packetPlayerStatChange(true, constant.StrID, int32(amount)))
}

// GiveStr to Data
func (d *Data) GiveStr(amount int16) {
	d.SetStr(d.str + amount)
}

// SetDex of Data
func (d *Data) SetDex(amount int16) {
	d.dex = amount
	d.Send(packetPlayerStatChange(true, constant.DexID, int32(amount)))
}

// GiveDex to Data
func (d *Data) GiveDex(amount int16) {
	d.SetDex(d.dex + amount)
}

// SetInt of Data
func (d *Data) SetInt(amount int16) {
	d.intt = amount
	d.Send(packetPlayerStatChange(true, constant.IntID, int32(amount)))
}

// GiveInt to Data
func (d *Data) GiveInt(amount int16) {
	d.SetInt(d.intt + amount)
}

// SetLuk of Data
func (d *Data) SetLuk(amount int16) {
	d.luk = amount
	d.Send(packetPlayerStatChange(true, constant.LukID, int32(amount)))
}

// GiveLuk to Data
func (d *Data) GiveLuk(amount int16) {
	d.SetLuk(d.luk + amount)
}

// SetHP of Data
func (d *Data) SetHP(amount int16) {
	d.hp = amount
	d.Send(packetPlayerStatChange(true, constant.HpID, int32(amount)))
}

// GiveHP to Data
func (d *Data) GiveHP(amount int16) {
	newHP := d.hp + amount
	if newHP < 0 {
		d.SetHP(0)
		return
	}
	if newHP > d.MaxHP() {
		d.SetHP(d.MaxHP())
		return
	}
	d.SetHP(newHP)
}

// SetMaxHP of Data
func (d *Data) SetMaxHP(amount int16) {
	d.maxHP = amount
	d.Send(packetPlayerStatChange(true, constant.MaxHpID, int32(amount)))
}

// SetMP of Data
func (d *Data) SetMP(amount int16) {
	d.mp = amount
	d.Send(packetPlayerStatChange(true, constant.MpID, int32(amount)))
}

// GiveMP to Data
func (d *Data) GiveMP(amount int16) {
	newMP := d.mp + amount
	if newMP < 0 {
		d.SetMP(0)
		return
	}
	if newMP > d.MaxMP() {
		d.SetMP(d.MaxMP())
		return
	}
	d.SetMP(newMP)
}

// SetMaxMP of Data
func (d *Data) SetMaxMP(amount int16) {
	d.maxMP = amount
	d.Send(packetPlayerStatChange(true, constant.MaxMpID, int32(amount)))
}

// SetFame of Data
func (d *Data) SetFame(amount int16) {

}

// IncrementPortalCount of player
func (d *Data) IncrementPortalCount() {
	d.portalCount++
}

// AddEquip item to slice
func (d *Data) AddEquip(item item.Data) {
	d.equip = append(d.equip, item)
}

// SetGuild of Data
func (d *Data) SetGuild(name string) {

}

// SetEquipSlotSize of Data
func (d *Data) SetEquipSlotSize(size byte) {

}

// SetUseSlotSize of Data
func (d *Data) SetUseSlotSize(size byte) {

}

// SetSetUpSlotSize of Data
func (d *Data) SetSetUpSlotSize(size byte) {

}

// SetEtcSlotSize of Data
func (d *Data) SetEtcSlotSize(size byte) {

}

// SetCashSlotSize of Data
func (d *Data) SetCashSlotSize(size byte) {

}

// SetMesos of Data
func (d *Data) SetMesos(amount int32) {
	d.mesos = amount
	d.Send(packetPlayerStatChange(false, constant.MesosID, amount))
}

// GiveMesos to Data
func (d *Data) GiveMesos(amount int32) {
	d.SetMesos(d.mesos + amount)
}

// SetMiniGameWins of Data
func (d *Data) SetMiniGameWins(v int32) {
	d.miniGameWins = v
}

// SetMiniGameLoss of Data
func (d *Data) SetMiniGameLoss(v int32) {
	d.miniGameLoss = v
}

// SetMiniGameDraw of Data
func (d *Data) SetMiniGameDraw(v int32) {
	d.miniGameDraw = v
}

// SetMiniGamePoints of data
func (d *Data) SetMiniGamePoints(v int32) {
	d.miniGamePoints = v
}

// LastAttackPacketTime of player
func (d *Data) LastAttackPacketTime() int64 {
	return d.lastAttackPacketTime
}

// SetLastAttackPacketTime of player
func (d *Data) SetLastAttackPacketTime(t int64) {
	d.lastAttackPacketTime = t
}

type movementFrag interface {
	X() int16
	Y() int16
	Foothold() int16
	Stance() byte
}

// UpdateMovement - update Data from position data
func (d *Data) UpdateMovement(frag movementFrag) {
	d.pos.SetX(frag.X())
	d.pos.SetY(frag.Y())
	d.pos.SetFoothold(frag.Foothold())
	d.stance = frag.Stance()
}

// SetPos of Data
func (d *Data) SetPos(pos pos.Data) {
	d.pos = pos
}

// CheckPos - checks Data is within a certain range of a position
func (d Data) CheckPos(pos pos.Data, xRange, yRange int16) bool {
	var xValid, yValid bool

	if xRange == 0 {
		xValid = d.pos.X() == pos.X()
	} else {
		xValid = (pos.X()-xRange < d.pos.X() && d.pos.X() < pos.X()+xRange)
	}

	if yRange == 0 {
		xValid = d.pos.Y() == pos.Y()
	} else {
		yValid = (pos.Y()-yRange < d.pos.Y() && d.pos.Y() < pos.Y()+yRange)
	}

	return xValid && yValid
}

// SetMapID of Data
func (d *Data) SetMapID(id int32) {
	d.mapID = id
}

// SetMapPosID of Data
func (d *Data) SetMapPosID(pos byte) {
	d.mapPos = pos
}

func (d Data) NoChange() {
	d.Send(packetInventoryNoChange())
}

// GiveItem to Data
func (d *Data) GiveItem(newItem item.Data, db *sql.DB) error { // TODO: Refactor
	findFirstEmptySlot := func(items []item.Data, size byte) (int16, error) {
		slotsUsed := make([]bool, size)

		for _, v := range items {
			if v.SlotID() > 0 {
				slotsUsed[v.SlotID()-1] = true
			}
		}

		slot := 0

		for i, v := range slotsUsed {
			if v == false {
				slot = i + 1
				break
			}
		}

		if slot == 0 {
			slot = len(slotsUsed) + 1
		}

		if byte(slot) > size {
			return 0, fmt.Errorf("No empty item slot left")
		}

		return int16(slot), nil
	}

	switch newItem.InvID() {
	case 1: // Equip
		slotID, err := findFirstEmptySlot(d.equip, d.equipSlotSize)

		if err != nil {
			return err
		}

		newItem.SetSlotID(slotID)
		newItem.SetAmount(1) // just in case
		newItem.Save(db, d.id)
		d.equip = append(d.equip, newItem)
		d.Send(packetInventoryAddItem(newItem, true))
	case 2: // Use
		size := newItem.Amount()
		for size > 0 {
			var value int16 = 200
			value -= size

			if value < 1 {
				value = 200
			} else {
				value = size
			}
			size -= constant.MaxItemStack

			newItem.SetAmount(value)

			var slotID int16
			var index int
			for i, v := range d.use {
				if v.ID() == newItem.ID() && v.Amount() < constant.MaxItemStack {
					slotID = v.SlotID()
					index = i
					break
				}
			}

			if slotID == 0 {
				slotID, err := findFirstEmptySlot(d.use, d.useSlotSize)

				if err != nil {
					return err
				}

				newItem.SetSlotID(slotID)
				newItem.Save(db, d.id)
				d.use = append(d.use, newItem)
				d.Send(packetInventoryAddItem(newItem, true))
			} else {
				remainder := newItem.Amount() - (constant.MaxItemStack - d.use[index].Amount())

				if remainder > 0 { //partial merge
					slotID, err := findFirstEmptySlot(d.use, d.useSlotSize)

					if err != nil {
						return err
					}

					newItem.SetAmount(value)
					newItem.SetSlotID(slotID)
					newItem.Save(db, d.id)

					d.use = append(d.use, newItem)
					d.Send(packetInventoryAddItems([]item.Data{d.use[index], newItem}, []bool{false, true}))
				} else { // full merge
					d.use[index].SetAmount(d.use[index].Amount() + newItem.Amount())
					d.Send(packetInventoryAddItem(d.use[index], false))
					d.use[index].Save(db, d.id)
				}
			}

		}
	case 3: // Set-up
		slotID, err := findFirstEmptySlot(d.setUp, d.setupSlotSize)

		if err != nil {
			return err
		}

		newItem.SetSlotID(slotID)
		newItem.Save(db, d.id)
		d.setUp = append(d.setUp, newItem)
		d.Send(packetInventoryAddItem(newItem, true))
	case 4: // Etc
		size := newItem.Amount()
		for size > 0 {
			var value int16 = 200
			value -= size

			if value < 1 {
				value = 200
			} else {
				value = size
			}
			size -= constant.MaxItemStack

			newItem.SetAmount(value)

			var slotID int16
			var index int
			for i, v := range d.etc {
				if v.ID() == newItem.ID() && v.Amount() < constant.MaxItemStack {
					slotID = v.SlotID()
					index = i
					break
				}
			}

			if slotID == 0 {
				slotID, err := findFirstEmptySlot(d.etc, d.etcSlotSize)

				if err != nil {
					return err
				}

				newItem.SetSlotID(slotID)
				newItem.Save(db, d.id)
				d.etc = append(d.etc, newItem)
				d.Send(packetInventoryAddItem(newItem, true))
			} else {
				remainder := newItem.Amount() - (constant.MaxItemStack - d.etc[index].Amount())

				if remainder > 0 { //partial merge
					slotID, err := findFirstEmptySlot(d.etc, d.etcSlotSize)

					if err != nil {
						return err
					}

					newItem.SetAmount(value)
					newItem.SetSlotID(slotID)
					newItem.Save(db, d.id)

					d.etc = append(d.etc, newItem)
					d.Send(packetInventoryAddItems([]item.Data{d.etc[index], newItem}, []bool{false, true}))
				} else { // full merge
					d.etc[index].SetAmount(d.etc[index].Amount() + newItem.Amount())
					d.Send(packetInventoryAddItem(d.etc[index], false))
					d.etc[index].Save(db, d.id)
				}
			}

		}
	case 5: // Cash
		// some are stackable, how to tell?
		slotID, err := findFirstEmptySlot(d.cash, d.cashSlotSize)

		if err != nil {
			return err
		}

		newItem.SetSlotID(slotID)
		newItem.Save(db, d.id)
		d.cash = append(d.cash, newItem)
		d.Send(packetInventoryAddItem(newItem, true))
	default:
		return fmt.Errorf("Unkown inventory id: %d", newItem.InvID())
	}

	return nil
}

// TakeItem from Data
func (d *Data) TakeItem(id int32, slot int16, amount int16, invID byte, db *sql.DB) (item.Data, error) {
	item, err := d.getItem(invID, slot)
	if err != nil {
		return item, err
	}

	if item.ID() != id {
		return item, fmt.Errorf("item.ID(%d) does not match ID(%d) provided", item.ID(), id)
	}

	maxRemove := math.Min(float64(item.Amount()), float64(amount))
	item.UpdateAmount(item.Amount() - int16(maxRemove))
	if item.Amount() == 0 {
		// Delete item
		d.removeItem(item, db)
	} else {
		// Update item with new stack size
		d.updateItemStack(item, db)

	}

	return item, nil

}

func (d Data) updateItemStack(item item.Data, db *sql.DB) {
	item.Save(db, d.id)
	d.updateItem(item)
	d.Send(packetInventoryAddItem(item, false))
}

func (d *Data) updateItem(new item.Data) {
	var items []item.Data

	switch new.InvID() {
	case 1:
		items = d.equip
	case 2:
		items = d.use
	case 3:
		items = d.setUp
	case 4:
		items = d.etc
	case 5:
		items = d.cash
	}

	for i, v := range items {
		if v.DbID() == new.DbID() {
			items[i] = new
			break
		}
	}
}

func (d Data) getItem(invID byte, slotID int16) (item.Data, error) {
	var items []item.Data

	switch invID {
	case 1:
		items = d.equip
	case 2:
		items = d.use
	case 3:
		items = d.setUp
	case 4:
		items = d.etc
	case 5:
		items = d.cash
	}

	for _, v := range items {
		if v.SlotID() == slotID {
			return v, nil
		}
	}

	return item.Data{}, fmt.Errorf("Could not find item")
}

func (d *Data) swapItems(item1, item2 item.Data, start, end int16, db *sql.DB) {
	item1.SetSlotID(end)
	item1.Save(db, d.id)
	d.updateItem(item1)

	item2.SetSlotID(start)
	item2.Save(db, d.id)
	d.updateItem(item2)

	d.Send(packetInventoryChangeItemSlot(item1.InvID(), start, end))
}

func (d *Data) removeItem(item item.Data, db *sql.DB) {
	switch item.InvID() {
	case 1:
		for i, v := range d.equip {
			if v.DbID() == item.DbID() {
				d.equip[i] = d.equip[len(d.equip)-1]
				d.equip = d.equip[:len(d.equip)-1]
				break
			}
		}
	case 2:
		for i, v := range d.use {
			if v.DbID() == item.DbID() {
				d.use[i] = d.use[len(d.use)-1]
				d.use = d.use[:len(d.use)-1]
				break
			}
		}
	case 3:
		for i, v := range d.setUp {
			if v.DbID() == item.DbID() {
				d.setUp[i] = d.setUp[len(d.setUp)-1]
				d.setUp = d.setUp[:len(d.setUp)-1]
				break
			}
		}
	case 4:
		for i, v := range d.etc {
			if v.DbID() == item.DbID() {
				d.etc[i] = d.etc[len(d.etc)-1]
				d.etc = d.etc[:len(d.etc)-1]
				break
			}
		}
	case 5:
		for i, v := range d.cash {
			if v.DbID() == item.DbID() {
				d.cash[i] = d.cash[len(d.cash)-1]
				d.cash = d.cash[:len(d.cash)-1]
				break
			}
		}
	}

	item.Delete(db)
	d.Send(packetInventoryRemoveItem(item))
}

// MoveItem from one slot to another, if the final slot is zero then this is a drop action
func (d *Data) MoveItem(start, end, amount int16, invID byte, inst instance, db *sql.DB) error {
	if end == 0 { //drop item
		fmt.Println("Drop item amount:", amount)
		item, err := d.getItem(invID, start)

		if err != nil {
			return fmt.Errorf("Item to move doesn't exist")
		}

		d.removeItem(item, db)
		// inst.AddDrop()
	} else if end < 0 { // Move to equip slot
		item1, err := d.getItem(invID, start)

		if err != nil {
			return fmt.Errorf("Item to move doesn't exist")
		}

		if item1.TwoHanded() {
			if _, err := d.getItem(invID, -10); err == nil {
				d.Send(packetInventoryNoChange()) // Should this do switching if space is available?
				return nil
			}
		} else if item1.Shield() {
			if weapon, err := d.getItem(invID, -11); err == nil && weapon.TwoHanded() {
				d.Send(packetInventoryNoChange()) // should this move weapon into into item 1 slot?
				return nil
			}
		}

		item2, err := d.getItem(invID, end)

		if err == nil {
			item2.SetSlotID(start)
			item2.Save(db, d.id)
			d.updateItem(item2)
		}

		item1.SetSlotID(end)
		item1.Save(db, d.id)
		d.updateItem(item1)

		d.Send(packetInventoryChangeItemSlot(invID, start, end))
		inst.Send(packetInventoryChangeEquip(*d))
	} else { // move within inventory
		item1, err := d.getItem(invID, start)

		if err != nil {
			return fmt.Errorf("Item to move doesn't exist")
		}

		item2, err := d.getItem(invID, end)

		if err != nil { // empty slot
			item1.SetSlotID(end)
			item1.Save(db, d.id)
			d.updateItem(item1)

			d.Send(packetInventoryChangeItemSlot(invID, start, end))
		} else { // moved onto item
			if (item1.IsStackable() && item2.IsStackable()) && (item1.ID() == item2.ID()) {
				if item1.Amount() == constant.MaxItemStack || item2.Amount() == constant.MaxItemStack { // swap items
					d.swapItems(item1, item2, start, end, db)
				} else if item2.Amount() < constant.MaxItemStack { // full merge
					if item2.Amount()+item1.Amount() <= constant.MaxItemStack {
						item2.SetAmount(item2.Amount() + item1.Amount())
						item2.Save(db, d.id)
						d.updateItem(item2)
						d.Send(packetInventoryAddItem(item2, false))

						d.removeItem(item1, db)
					} else { // partial merge is just a swap
						d.swapItems(item1, item2, start, end, db)
					}
				}
			} else {
				d.swapItems(item1, item2, start, end, db)
			}
		}

		if start < 0 || end < 0 {
			inst.Send(packetInventoryChangeEquip(*d))
		}
	}

	return nil
}

// Use items
func (d Data) Use() []item.Data {
	return d.use
}

// UpdateSkill map entry
func (d *Data) UpdateSkill(updatedSkill Skill) {
	d.skills[updatedSkill.ID] = updatedSkill
	d.Send(packetPlayerSkillBookUpdate(updatedSkill.ID, int32(updatedSkill.Level)))
}

// ID of Data
func (d Data) ID() int32 { return d.id }

// AccountID of Data
func (d Data) AccountID() int32 { return d.accountID }

// WorldID of Data
func (d Data) WorldID() byte { return d.worldID }

// MapID of Data
func (d Data) MapID() int32 { return d.mapID }

// MapPos of Data
func (d Data) MapPos() byte { return d.mapPos }

// PreviousMap the Data was on
func (d Data) PreviousMap() int32 { return d.previousMap }

// PortalCount of Data, used in detecting warp hacking
func (d Data) PortalCount() byte { return d.portalCount }

// Job of Data
func (d Data) Job() int16 { return d.job }

// Level of Data
func (d Data) Level() byte { return d.level }

// Str of Data
func (d Data) Str() int16 { return d.str }

//Dex of Data
func (d Data) Dex() int16 { return d.dex }

// Int of Data
func (d Data) Int() int16 { return d.intt }

// Luk of Data
func (d Data) Luk() int16 { return d.luk }

// HP of Data
func (d Data) HP() int16 { return d.hp }

// MaxHP of Data
func (d Data) MaxHP() int16 { return d.maxHP }

// MP of Data
func (d Data) MP() int16 { return d.mp }

// MaxMP of Data
func (d Data) MaxMP() int16 { return d.maxMP }

// AP of Data
func (d Data) AP() int16 { return d.ap }

// SP of Data
func (d Data) SP() int16 { return d.sp }

// Exp of Data
func (d Data) Exp() int32 { return d.exp }

// Fame of Data
func (d Data) Fame() int16 { return d.fame }

// Name of Data
func (d Data) Name() string { return d.name }

// Gender of Data
func (d Data) Gender() byte { return d.gender }

// Skin id of Data
func (d Data) Skin() byte { return d.skin }

// Face id of Data
func (d Data) Face() int32 { return d.face }

// Hair id of Data
func (d Data) Hair() int32 { return d.hair }

// ChairID of the chair the Data is sitting on
func (d Data) ChairID() int32 { return d.chairID }

// Stance id
func (d Data) Stance() byte { return d.stance }

// Pos of Data
func (d Data) Pos() pos.Data { return d.pos }

// Guild name Data is currenty part of
func (d Data) Guild() string { return d.guild }

// EquipSlotSize in inventory
func (d Data) EquipSlotSize() byte { return d.equipSlotSize }

// UseSlotSize in inventory
func (d Data) UseSlotSize() byte { return d.useSlotSize }

// SetupSlotSize in inventory
func (d Data) SetupSlotSize() byte { return d.setupSlotSize }

// EtcSlotSize in inventory
func (d Data) EtcSlotSize() byte { return d.etcSlotSize }

//CashSlotSize in inventory
func (d Data) CashSlotSize() byte { return d.cashSlotSize }

// Mesos Data currently has
func (d Data) Mesos() int32 { return d.mesos }

// Skills and their levels the Data currently has
func (d Data) Skills() map[int32]Skill { return d.skills }

// MiniGameWins between omok and memory
func (d Data) MiniGameWins() int32 { return d.miniGameWins }

// MiniGameDraw betweeen omok and memory
func (d Data) MiniGameDraw() int32 { return d.miniGameDraw }

// MiniGameLoss between omok and memory
func (d Data) MiniGameLoss() int32 { return d.miniGameLoss }

// MiniGamePoints between omok and memory
func (d Data) MiniGamePoints() int32 { return d.miniGamePoints }

// DisplayBytes used in packets for displaying Data in various situations e.g. in field, in mini game room
func (d Data) DisplayBytes() []byte {
	pkt := mpacket.NewPacket()
	pkt.WriteByte(d.gender)
	pkt.WriteByte(d.skin)
	pkt.WriteInt32(d.face)
	pkt.WriteByte(0x00) // ?
	pkt.WriteInt32(d.hair)

	cashWeapon := int32(0)

	for _, b := range d.equip {
		if b.SlotID() < 0 && b.SlotID() > -20 {
			pkt.WriteByte(byte(math.Abs(float64(b.SlotID()))))
			pkt.WriteInt32(b.ID())
		}
	}

	for _, b := range d.equip {
		if b.SlotID() < -100 {
			if b.SlotID() == -111 {
				cashWeapon = b.ID()
			} else {
				pkt.WriteByte(byte(math.Abs(float64(b.SlotID() + 100))))
				pkt.WriteInt32(b.ID())
			}
		}
	}

	pkt.WriteByte(0xFF)
	pkt.WriteByte(0xFF)
	pkt.WriteInt32(cashWeapon)

	return pkt
}

// Save data
func (d Data) Save(db *sql.DB) error {
	query := `UPDATE characters set skin=?, hair=?, face=?, level=?,
	job=?, str=?, dex=?, intt=?, luk=?, hp=?, maxHP=?, mp=?, maxMP=?,
	ap=?, sp=?, exp=?, fame=?, mapID=?, mapPos=?, mesos=?, miniGameWins=?,
	miniGameDraw=?, miniGameLoss=?, miniGamePoints=? WHERE id=?`

	var mapPos byte
	var err error

	if d.inst != nil {
		mapPos, err = d.inst.CalculateNearestSpawnPortalID(d.pos)
	}

	if err != nil {
		return err
	}

	d.mapPos = mapPos

	// TODO: Move mesos, to instances of it changing, otherwise items and mesos can become out of sync from
	// any crashes
	_, err = db.Exec(query,
		d.skin, d.hair, d.face, d.level, d.job, d.str, d.dex, d.intt, d.luk, d.hp, d.maxHP, d.mp,
		d.maxMP, d.ap, d.sp, d.exp, d.fame, d.mapID, d.mapPos, d.mesos, d.miniGameWins,
		d.miniGameDraw, d.miniGameLoss, d.miniGamePoints, d.id)

	if err != nil {
		return err
	}

	query = `INSERT INTO skills(characterID,skillID,level,cooldown) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE characterID=?, skillID=?`
	for skillID, skill := range d.skills {
		_, err := db.Exec(query, d.id, skillID, skill.Level, skill.Cooldown, d.id, skillID)

		if err != nil {
			return err
		}
	}

	return err
}

// UpdateGuildInfo for the player
func (d *Data) UpdateGuildInfo() {
	d.Send(packetGuildInfo(0, "[Admins]", 0))
}

// DamagePlayer reduces character HP based on damage
func (d *Data) DamagePlayer(damage int16) {
	if damage < -1 {
		return
	}
	newHP := d.hp - damage

	if newHP <= -1 {
		d.hp = 0
	} else {
		d.hp = newHP
	}

	d.Send(packetPlayerStatChange(true, constant.HpID, int32(d.hp)))
}
