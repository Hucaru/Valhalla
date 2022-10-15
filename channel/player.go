package channel

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/nx"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type buddy struct {
	id        int32
	name      string
	channelID int32
	status    byte  // 0 - online, 1 - buddy request, 2 - offline
	cashShop  int32 // > 0 means is in cash shop
}

type playerSkill struct {
	ID             int32
	Level, Mastery byte
	Cooldown       int16
	CooldownTime   int16
	TimeLastUsed   int64
}

func createPlayerSkillFromData(ID int32, level byte) (playerSkill, error) {
	skill, err := nx.GetPlayerSkill(ID)

	if err != nil {
		return playerSkill{}, fmt.Errorf("Not a valid skill ID %v level %v", ID, level)
	}

	if int(level) > len(skill) {
		return playerSkill{}, fmt.Errorf("Invalid skill level")
	}

	return playerSkill{ID: ID,
		Level:        level,
		Mastery:      byte(skill[level-1].Mastery),
		Cooldown:     0,
		CooldownTime: int16(skill[level-1].Time),
		TimeLastUsed: 0}, nil
}

func getSkillsFromCharID(id int32) []playerSkill {
	skills := []playerSkill{}

	filter := "skillID, level, cooldown"

	row, err := common.DB.Query("SELECT "+filter+" FROM skills where characterID=?", id)

	if err != nil {
		panic(err)
	}

	defer row.Close()

	for row.Next() {
		skill := playerSkill{}

		row.Scan(&skill.ID, &skill.Level, &skill.Cooldown)

		skillData, err := nx.GetPlayerSkill(skill.ID)

		if err != nil {
			return skills
		}

		skill.CooldownTime = int16(skillData[skill.Level-1].Time)

		skills = append(skills, skill)
	}

	return skills
}

type updatePartyInfoFunc func(partyID, playerID, job, level int32, name string)

type player struct {
	conn       mnet.Client
	instanceID int
	inst       *fieldInstance

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
	pos     pos
	guild   string

	equipSlotSize byte
	useSlotSize   byte
	setupSlotSize byte
	etcSlotSize   byte
	cashSlotSize  byte

	equip []item
	use   []item
	setUp []item
	etc   []item
	cash  []item

	mesos int32

	skills map[int32]playerSkill

	miniGameWins, miniGameDraw, miniGameLoss, miniGamePoints int32

	lastAttackPacketTime int64

	buddyListSize byte
	buddyList     []buddy

	party *party

	UpdatePartyInfo updatePartyInfoFunc

	rates *rates
}

// Send the Data a packet
func (d player) send(packet mpacket.Packet) {
	if d.conn == nil {
		return
	}

	d.conn.Send(packet)
}

func (d *player) setJob(id int16) {
	d.job = id
	d.conn.Send(packetPlayerStatChange(true, constant.JobID, int32(id)))

	if d.party != nil {
		d.party.updateJobLevel(d.id, int32(d.job), int32(d.level))
	}
}

func (d *player) levelUp() {
	d.giveAP(5)
	d.giveSP(3)

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

	d.setHP(d.hp)
	d.setMaxHP(d.hp)

	d.setMP(d.mp)
	d.setMaxMP(d.mp)

	d.giveLevel(1)
}

func (d *player) setEXP(amount int32) {
	if d.level > 199 {
		d.exp = amount
		d.send(packetPlayerStatChange(false, constant.ExpID, int32(amount)))
		return
	}

	remainder := amount - constant.ExpTable[d.level-1]

	if remainder >= 0 {
		d.levelUp()
		d.setEXP(remainder)
	} else {
		d.exp = amount
		d.send(packetPlayerStatChange(false, constant.ExpID, int32(amount)))
	}
}

func (d *player) giveEXP(amount int32, fromMob, fromParty bool) {
	amount = int32(d.rates.exp * float32(amount))
	if fromMob {
		d.send(packetMessageExpGained(true, false, amount))
	} else if fromParty {
		d.send(packetMessageExpGained(false, false, amount))
	} else {
		d.send(packetMessageExpGained(false, true, amount))
	}

	d.setEXP(d.exp + amount)
}

func (d *player) setLevel(amount byte) {
	d.level = amount
	d.send(packetPlayerStatChange(false, constant.LevelID, int32(amount)))
	d.inst.send(packetPlayerLevelUpAnimation(d.id))

	if d.party != nil {
		d.party.updateJobLevel(d.id, int32(d.job), int32(d.level))
	}
}

func (d *player) giveLevel(amount byte) {
	d.setLevel(d.level + amount)
}

func (d *player) setAP(amount int16) {
	d.ap = amount
	d.send(packetPlayerStatChange(false, constant.ApID, int32(amount)))
}

func (d *player) giveAP(amount int16) {
	d.setAP(d.ap + amount)
}

func (d *player) setSP(amount int16) {
	d.sp = amount
	d.send(packetPlayerStatChange(false, constant.SpID, int32(amount)))
}

func (d *player) giveSP(amount int16) {
	d.setSP(d.sp + amount)
}

func (d *player) setStr(amount int16) {
	d.str = amount
	d.send(packetPlayerStatChange(true, constant.StrID, int32(amount)))
}

func (d *player) giveStr(amount int16) {
	d.setStr(d.str + amount)
}

func (d *player) setDex(amount int16) {
	d.dex = amount
	d.send(packetPlayerStatChange(true, constant.DexID, int32(amount)))
}

func (d *player) giveDex(amount int16) {
	d.setDex(d.dex + amount)
}

func (d *player) setInt(amount int16) {
	d.intt = amount
	d.send(packetPlayerStatChange(true, constant.IntID, int32(amount)))
}

func (d *player) giveInt(amount int16) {
	d.setInt(d.intt + amount)
}

func (d *player) setLuk(amount int16) {
	d.luk = amount
	d.send(packetPlayerStatChange(true, constant.LukID, int32(amount)))
}

func (d *player) giveLuk(amount int16) {
	d.setLuk(d.luk + amount)
}

func (d *player) setHP(amount int16) {
	if amount > constant.MaxHpValue {
		amount = constant.MaxHpValue
	}

	d.hp = amount
	d.send(packetPlayerStatChange(true, constant.HpID, int32(amount)))
}

func (d *player) giveHP(amount int16) {
	newHP := d.hp + amount
	if newHP < 0 {
		d.setHP(0)
		return
	}
	if newHP > d.maxHP {
		d.setHP(d.maxHP)
		return
	}
	d.setHP(newHP)
}

func (d *player) setMaxHP(amount int16) {
	if amount > constant.MaxHpValue {
		amount = constant.MaxHpValue
	}

	d.maxHP = amount
	d.send(packetPlayerStatChange(true, constant.MaxHpID, int32(amount)))
}

// SetMP of Data
func (d *player) setMP(amount int16) {
	if amount > constant.MaxMpValue {
		amount = constant.MaxMpValue
	}

	d.mp = amount
	d.send(packetPlayerStatChange(true, constant.MpID, int32(amount)))
}

func (d *player) giveMP(amount int16) {
	newMP := d.mp + amount
	if newMP < 0 {
		d.setMP(0)
		return
	}
	if newMP > d.maxMP {
		d.setMP(d.maxMP)
		return
	}
	d.setMP(newMP)
}

func (d *player) setMaxMP(amount int16) {
	if amount > constant.MaxMpValue {
		amount = constant.MaxMpValue
	}

	d.maxMP = amount
	d.send(packetPlayerStatChange(true, constant.MaxMpID, int32(amount)))
}

func (d *player) setFame(amount int16) {

}

func (d *player) addEquip(item item) {
	d.equip = append(d.equip, item)
}

func (d *player) setMesos(amount int32) {
	d.mesos = amount
	d.send(packetPlayerStatChange(true, constant.MesosID, amount))
	d.saveMesos()
}

func (d *player) giveMesos(amount int32) {
	d.setMesos(d.mesos + int32(d.rates.mesos*float32(amount)))
}

func (d *player) takeMesos(amount int32) {
	d.setMesos(d.mesos - amount)
}

func (d *player) saveMesos() error {
	query := "UPDATE characters SET mesos=? WHERE accountID=? and name=?"

	_, err := common.DB.Exec(query,
		d.mesos,
		d.accountID,
		d.name)

	if err != nil {
		return err
	}

	return nil
}

// UpdateMovement - update Data from position data
func (d *player) UpdateMovement(frag movementFrag) {
	d.pos.x = frag.x
	d.pos.y = frag.y
	d.pos.foothold = frag.foothold
	d.stance = frag.stance
}

// SetPos of Data
func (d *player) SetPos(pos pos) {
	d.pos = pos
}

// checks Data is within a certain range of a position
func (d player) checkPos(pos pos, xRange, yRange int16) bool {
	var xValid, yValid bool

	if xRange == 0 {
		xValid = d.pos.x == pos.x
	} else {
		xValid = (pos.x-xRange < d.pos.x && d.pos.x < pos.x+xRange)
	}

	if yRange == 0 {
		xValid = d.pos.y == pos.y
	} else {
		yValid = (pos.y-yRange < d.pos.y && d.pos.y < pos.y+yRange)
	}

	return xValid && yValid
}

func (d *player) setMapID(id int32) {
	oldMapID := d.mapID
	d.mapID = id

	if d.party != nil {
		d.party.updatePlayerMap(d.id, d.mapID)
	}

	if err := d.saveMapID(id, oldMapID); err != nil {
		log.Println(err)
	}

}

func (d *player) saveMapID(newMapId, oldMapId int32) error {
	query := "UPDATE characters SET mapID=?,previousMapID=? WHERE accountID=? and name=?"

	_, err := common.DB.Exec(query,
		newMapId,
		oldMapId,
		d.accountID,
		d.name)

	if err != nil {
		return err
	}

	return nil
}

func (d player) noChange() {
	d.send(packetInventoryNoChange())
}

func (d *player) giveItem(newItem item) error { // TODO: Refactor
	findFirstEmptySlot := func(items []item, size byte) (int16, error) {
		slotsUsed := make([]bool, size)

		for _, v := range items {
			if v.slotID > 0 {
				slotsUsed[v.slotID-1] = true
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

	switch newItem.invID {
	case 1: // Equip
		slotID, err := findFirstEmptySlot(d.equip, d.equipSlotSize)

		if err != nil {
			return err
		}

		newItem.slotID = slotID
		newItem.amount = 1 // just in case
		newItem.save(d.id)
		d.equip = append(d.equip, newItem)
		d.send(packetInventoryAddItem(newItem, true))
	case 2: // Use
		size := newItem.amount
		for size > 0 {
			var value int16 = 200
			value -= size

			if value < 1 {
				value = 200
			} else {
				value = size
			}
			size -= constant.MaxItemStack

			newItem.amount = value

			var slotID int16
			var index int
			for i, v := range d.use {
				if v.id == newItem.id && v.amount < constant.MaxItemStack {
					slotID = v.slotID
					index = i
					break
				}
			}

			if slotID == 0 {
				slotID, err := findFirstEmptySlot(d.use, d.useSlotSize)

				if err != nil {
					return err
				}

				newItem.slotID = slotID
				newItem.save(d.id)
				d.use = append(d.use, newItem)
				d.send(packetInventoryAddItem(newItem, true))
			} else {
				remainder := newItem.amount - (constant.MaxItemStack - d.use[index].amount)

				if remainder > 0 { //partial merge
					slotID, err := findFirstEmptySlot(d.use, d.useSlotSize)

					if err != nil {
						return err
					}

					newItem.amount = value
					newItem.slotID = slotID
					newItem.save(d.id)

					d.use = append(d.use, newItem)
					d.send(packetInventoryAddItems([]item{d.use[index], newItem}, []bool{false, true}))
				} else { // full merge
					d.use[index].amount = d.use[index].amount + newItem.amount
					d.send(packetInventoryAddItem(d.use[index], false))
					d.use[index].save(d.id)
				}
			}

		}
	case 3: // Set-up
		slotID, err := findFirstEmptySlot(d.setUp, d.setupSlotSize)

		if err != nil {
			return err
		}

		newItem.slotID = slotID
		newItem.save(d.id)
		d.setUp = append(d.setUp, newItem)
		d.send(packetInventoryAddItem(newItem, true))
	case 4: // Etc
		size := newItem.amount
		for size > 0 {
			var value int16 = 200
			value -= size

			if value < 1 {
				value = 200
			} else {
				value = size
			}
			size -= constant.MaxItemStack

			newItem.amount = value

			var slotID int16
			var index int
			for i, v := range d.etc {
				if v.id == newItem.id && v.amount < constant.MaxItemStack {
					slotID = v.slotID
					index = i
					break
				}
			}

			if slotID == 0 {
				slotID, err := findFirstEmptySlot(d.etc, d.etcSlotSize)

				if err != nil {
					return err
				}

				newItem.slotID = slotID
				newItem.save(d.id)
				d.etc = append(d.etc, newItem)
				d.send(packetInventoryAddItem(newItem, true))
			} else {
				remainder := newItem.amount - (constant.MaxItemStack - d.etc[index].amount)

				if remainder > 0 { //partial merge
					slotID, err := findFirstEmptySlot(d.etc, d.etcSlotSize)

					if err != nil {
						return err
					}

					newItem.amount = value
					newItem.slotID = slotID
					newItem.save(d.id)

					d.etc = append(d.etc, newItem)
					d.send(packetInventoryAddItems([]item{d.etc[index], newItem}, []bool{false, true}))
				} else { // full merge
					d.etc[index].amount = d.etc[index].amount + newItem.amount
					d.send(packetInventoryAddItem(d.etc[index], false))
					d.etc[index].save(d.id)
				}
			}

		}
	case 5: // Cash
		// some are stackable, how to tell?
		slotID, err := findFirstEmptySlot(d.cash, d.cashSlotSize)

		if err != nil {
			return err
		}

		newItem.slotID = slotID
		newItem.save(d.id)
		d.cash = append(d.cash, newItem)
		d.send(packetInventoryAddItem(newItem, true))
	default:
		return fmt.Errorf("Unkown inventory id: %d", newItem.invID)
	}

	return nil
}

func (d *player) takeItem(id int32, slot int16, amount int16, invID byte) (item, error) {
	item, err := d.getItem(invID, slot)
	if err != nil {
		return item, err
	}

	if item.id != id {
		return item, fmt.Errorf("item.ID(%d) does not match ID(%d) provided", item.id, id)
	}

	maxRemove := math.Min(float64(item.amount), float64(amount))
	item.amount = item.amount - int16(maxRemove)
	if item.amount == 0 {
		// Delete item
		d.removeItem(item)
	} else {
		// Update item with new stack size
		d.updateItemStack(item)

	}

	return item, nil

}

func (d player) updateItemStack(item item) {
	item.save(d.id)
	d.updateItem(item)
	d.send(packetInventoryModifyItemAmount(item))
}

func (d *player) updateItem(new item) {
	var items = d.findItemInventory(new)

	for i, v := range items {
		if v.dbID == new.dbID {
			items[i] = new
			break
		}
	}
	d.updateItemInventory(new.invID, items)
}

func (d *player) updateItemInventory(invID byte, inventory []item) {
	switch invID {
	case 1:
		d.equip = inventory
	case 2:
		d.use = inventory
	case 3:
		d.setUp = inventory
	case 4:
		d.etc = inventory
	case 5:
		d.cash = inventory
	}
}

func (d *player) findItemInventory(item item) []item {
	switch item.invID {
	case 1:
		return d.equip
	case 2:
		return d.use
	case 3:
		return d.setUp
	case 4:
		return d.etc
	case 5:
		return d.cash
	}

	return nil
}

func (d player) getItem(invID byte, slotID int16) (item, error) {
	var items []item

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
		if v.slotID == slotID {
			return v, nil
		}
	}

	return item{}, fmt.Errorf("Could not find item")
}

func (d *player) swapItems(item1, item2 item, start, end int16) {
	item1.slotID = end
	item1.save(d.id)
	d.updateItem(item1)

	item2.slotID = start
	item2.save(d.id)
	d.updateItem(item2)

	d.send(packetInventoryChangeItemSlot(item1.invID, start, end))
}

func (d *player) removeItem(item item) {
	switch item.invID {
	case 1:
		for i, v := range d.equip {
			if v.dbID == item.dbID {
				d.equip[i] = d.equip[len(d.equip)-1]
				d.equip = d.equip[:len(d.equip)-1]
				break
			}
		}
	case 2:
		for i, v := range d.use {
			if v.dbID == item.dbID {
				d.use[i] = d.use[len(d.use)-1]
				d.use = d.use[:len(d.use)-1]
				break
			}
		}
	case 3:
		for i, v := range d.setUp {
			if v.dbID == item.dbID {
				d.setUp[i] = d.setUp[len(d.setUp)-1]
				d.setUp = d.setUp[:len(d.setUp)-1]
				break
			}
		}
	case 4:
		for i, v := range d.etc {
			if v.dbID == item.dbID {
				d.etc[i] = d.etc[len(d.etc)-1]
				d.etc = d.etc[:len(d.etc)-1]
				break
			}
		}
	case 5:
		for i, v := range d.cash {
			if v.dbID == item.dbID {
				d.cash[i] = d.cash[len(d.cash)-1]
				d.cash = d.cash[:len(d.cash)-1]
				break
			}
		}
	}

	item.delete()
	d.send(packetInventoryRemoveItem(item))
}

func (d *player) dropMesos(amount int32) error {
	if d.mesos < amount {
		return errors.New("not enough mesos")
	}

	d.takeMesos(amount)

	return nil
}

func (d *player) moveItem(start, end, amount int16, invID byte) error {
	if end == 0 { //drop item
		item, err := d.getItem(invID, start)

		if err != nil {
			return fmt.Errorf("Item to move doesn't exist")
		}

		dropItem := item
		dropItem.amount = amount
		dropItem.dbID = 0

		d.takeItem(item.id, item.slotID, amount, item.invID)

		d.inst.dropPool.createDrop(dropSpawnNormal, dropFreeForAll, 0, d.pos, true, d.id, 0, dropItem)
	} else if end < 0 { // Move to equip slot
		item1, err := d.getItem(invID, start)

		if err != nil {
			return fmt.Errorf("Item to move doesn't exist")
		}

		if item1.twoHanded {
			if _, err := d.getItem(invID, -10); err == nil {
				d.send(packetInventoryNoChange()) // Should this do switching if space is available?
				return nil
			}
		} else if item1.shield() {
			if weapon, err := d.getItem(invID, -11); err == nil && weapon.twoHanded {
				d.send(packetInventoryNoChange()) // should this move weapon into into item 1 slot?
				return nil
			}
		}

		item2, err := d.getItem(invID, end)

		if err == nil {
			item2.slotID = start
			item2.save(d.id)
			d.updateItem(item2)
		}

		item1.slotID = end
		item1.save(d.id)
		d.updateItem(item1)

		d.send(packetInventoryChangeItemSlot(invID, start, end))
		d.inst.send(packetInventoryChangeEquip(*d))
	} else { // move within inventory
		item1, err := d.getItem(invID, start)

		if err != nil {
			return fmt.Errorf("Item to move doesn't exist")
		}

		item2, err := d.getItem(invID, end)

		if err != nil { // empty slot
			item1.slotID = end
			item1.save(d.id)
			d.updateItem(item1)

			d.send(packetInventoryChangeItemSlot(invID, start, end))
		} else { // moved onto item
			if (item1.isStackable() && item2.isStackable()) && (item1.id == item2.id) {
				if item1.amount == constant.MaxItemStack || item2.amount == constant.MaxItemStack { // swap items
					d.swapItems(item1, item2, start, end)
				} else if item2.amount < constant.MaxItemStack { // full merge
					if item2.amount+item1.amount <= constant.MaxItemStack {
						item2.amount = item2.amount + item1.amount
						item2.save(d.id)
						d.updateItem(item2)
						d.send(packetInventoryAddItem(item2, false))

						d.removeItem(item1)
					} else { // partial merge is just a swap
						d.swapItems(item1, item2, start, end)
					}
				}
			} else {
				d.swapItems(item1, item2, start, end)
			}
		}

		if start < 0 || end < 0 {
			d.inst.send(packetInventoryChangeEquip(*d))
		}
	}

	return nil
}

func (d *player) updateSkill(updatedSkill playerSkill) {
	d.skills[updatedSkill.ID] = updatedSkill
	d.send(packetPlayerSkillBookUpdate(updatedSkill.ID, int32(updatedSkill.Level)))
}

func (d *player) useSkill(id int32, level byte) error {
	skillInfo, _ := nx.GetPlayerSkill(id)

	for lvl, skill := range skillInfo {
		if lvl == int(level) {

			d.giveMP(-int16(skill.MpCon))

			// Use item
			// d.consumeItem(skill.itemCon, skill.itemConNo)

		}

		// If haste, etc
	}

	return nil
}

func (d player) admin() bool { return d.conn.GetAdminLevel() > 0 }

func (d player) displayBytes() []byte {
	pkt := mpacket.NewPacket()
	pkt.WriteByte(d.gender)
	pkt.WriteByte(d.skin)
	pkt.WriteInt32(d.face)
	pkt.WriteByte(0x00) // ?
	pkt.WriteInt32(d.hair)

	cashWeapon := int32(0)

	for _, b := range d.equip {
		if b.slotID < 0 && b.slotID > -20 {
			pkt.WriteByte(byte(math.Abs(float64(b.slotID))))
			pkt.WriteInt32(b.id)
		}
	}

	for _, b := range d.equip {
		if b.slotID < -100 {
			if b.slotID == -111 {
				cashWeapon = b.id
			} else {
				pkt.WriteByte(byte(math.Abs(float64(b.slotID + 100))))
				pkt.WriteInt32(b.id)
			}
		}
	}

	pkt.WriteByte(0xFF)
	pkt.WriteByte(0xFF)
	pkt.WriteInt32(cashWeapon)

	return pkt
}

// Save data - this needs to be split to occur at relevant points in time
func (d player) save() error {
	query := `UPDATE characters set skin=?, hair=?, face=?, level=?,
	job=?, str=?, dex=?, intt=?, luk=?, hp=?, maxHP=?, mp=?, maxMP=?,
	ap=?, sp=?, exp=?, fame=?, mapID=?, mapPos=?, mesos=?, miniGameWins=?,
	miniGameDraw=?, miniGameLoss=?, miniGamePoints=?, buddyListSize=? WHERE id=?`

	var mapPos byte
	var err error

	if d.inst != nil {
		mapPos, err = d.inst.calculateNearestSpawnPortalID(d.pos)
	}

	if err != nil {
		return err
	}

	d.mapPos = mapPos

	// TODO: Move mesos, to instances of it changing, otherwise items and mesos can become out of sync from
	// any crashes
	_, err = common.DB.Exec(query,
		d.skin, d.hair, d.face, d.level, d.job, d.str, d.dex, d.intt, d.luk, d.hp, d.maxHP, d.mp,
		d.maxMP, d.ap, d.sp, d.exp, d.fame, d.mapID, d.mapPos, d.mesos, d.miniGameWins,
		d.miniGameDraw, d.miniGameLoss, d.miniGamePoints, d.buddyListSize, d.id)

	if err != nil {
		return err
	}

	query = `INSERT INTO skills(characterID,skillID,level,cooldown) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE characterID=?, skillID=?`
	for skillID, skill := range d.skills {
		_, err := common.DB.Exec(query, d.id, skillID, skill.Level, skill.Cooldown, d.id, skillID)

		if err != nil {
			return err
		}
	}

	return err
}

func (d *player) damagePlayer(damage int16) {
	if damage < -1 {
		return
	}
	newHP := d.hp - damage

	if newHP <= -1 {
		d.hp = 0
	} else {
		d.hp = newHP
	}

	d.send(packetPlayerStatChange(true, constant.HpID, int32(d.hp)))
}

// UpdateGuildInfo for the player
func (d *player) UpdateGuildInfo() {
	d.send(packetGuildInfo(0, "[Admins]", 0))
}

// UpdateBuddyInfo for the player
func (d *player) UpdateBuddyInfo() {
	d.send(packetBuddyListSizeUpdate(d.buddyListSize))
	d.send(packetBuddyInfo(d.buddyList))
}

// BuddyListFull checks if buddy list is full
func (d player) buddyListFull() bool {
	count := 0
	for _, v := range d.buddyList {
		if v.status != 1 {
			count++
		}
	}

	if count < int(d.buddyListSize) {
		return false
	}

	return true
}

func (d *player) addOnlineBuddy(id int32, name string, channel int32) {
	if d.buddyListFull() {
		return
	}

	for i, v := range d.buddyList {
		if v.id == id {
			d.buddyList[i].status = 0
			d.buddyList[i].channelID = channel
			d.send(packetBuddyUpdate(id, name, d.buddyList[i].status, channel, false))
			return
		}
	}

	newBuddy := buddy{id: id, name: name, status: 0, channelID: channel}

	d.buddyList = append(d.buddyList, newBuddy)
	d.send(packetBuddyInfo(d.buddyList))

	return
}

func (d *player) addOfflineBuddy(id int32, name string) {
	if d.buddyListFull() {
		return
	}

	for i, v := range d.buddyList {
		if v.id == id {
			d.buddyList[i].status = 2
			d.buddyList[i].channelID = -1
			d.send(packetBuddyUpdate(id, name, d.buddyList[i].status, -1, false))
			return
		}
	}

	newBuddy := buddy{id: id, name: name, status: 2, channelID: -1}

	d.buddyList = append(d.buddyList, newBuddy)
	d.send(packetBuddyInfo(d.buddyList))

	return
}

func (d player) hasBuddy(id int32) bool {
	for _, v := range d.buddyList {
		if v.id == id {
			return true
		}
	}

	return false
}

func (d *player) removeBuddy(id int32) {
	for i, v := range d.buddyList {
		if v.id == id {
			d.buddyList[i] = d.buddyList[len(d.buddyList)-1]
			d.buddyList = d.buddyList[:len(d.buddyList)-1]
			d.send(packetBuddyInfo(d.buddyList))
			return
		}
	}
}

func loadPlayerFromID(id int32, conn mnet.Client) player {
	c := player{}
	filter := "id,accountID,worldID,name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp, exp,fame,mapID,mapPos,previousMapID,mesos," +
		"equipSlotSize,useSlotSize,setupSlotSize,etcSlotSize,cashSlotSize,miniGameWins," +
		"miniGameDraw,miniGameLoss,miniGamePoints,buddyListSize"

	err := common.DB.QueryRow("SELECT "+filter+" FROM characters where id=?", id).Scan(&c.id,
		&c.accountID, &c.worldID, &c.name, &c.gender, &c.skin, &c.hair, &c.face,
		&c.level, &c.job, &c.str, &c.dex, &c.intt, &c.luk, &c.hp, &c.maxHP, &c.mp,
		&c.maxMP, &c.ap, &c.sp, &c.exp, &c.fame, &c.mapID, &c.mapPos,
		&c.previousMap, &c.mesos, &c.equipSlotSize, &c.useSlotSize, &c.setupSlotSize,
		&c.etcSlotSize, &c.cashSlotSize, &c.miniGameWins, &c.miniGameDraw, &c.miniGameLoss,
		&c.miniGamePoints, &c.buddyListSize)

	if err != nil {
		log.Println(err)
		return c
	}

	c.skills = make(map[int32]playerSkill)

	for _, s := range getSkillsFromCharID(c.id) {
		c.skills[s.ID] = s
	}

	nxMap, err := nx.GetMap(c.mapID)

	if err != nil {
		log.Println(err)
		return c
	}

	c.pos.x = nxMap.Portals[c.mapPos].X
	c.pos.y = nxMap.Portals[c.mapPos].Y

	c.equip, c.use, c.setUp, c.etc, c.cash = loadInventoryFromDb(c.id)

	c.buddyList = getBuddyList(c.id, c.buddyListSize)
	c.conn = conn
	return c
}

func getBuddyList(playerID int32, buddySize byte) []buddy {
	buddies := make([]buddy, 0, buddySize)
	filter := "friendID,accepted"
	rows, err := common.DB.Query("SELECT "+filter+" FROM buddy where characterID=?", playerID)

	if err != nil {
		log.Fatal(err)
		return buddies
	}

	defer rows.Close()

	i := 0
	for rows.Next() {
		newBuddy := buddy{}

		var accepted bool
		rows.Scan(&newBuddy.id, &accepted)

		filter := "channelID,name,inCashShop"
		err := common.DB.QueryRow("SELECT "+filter+" FROM characters where id=?", newBuddy.id).Scan(&newBuddy.channelID, &newBuddy.name, &newBuddy.cashShop)

		if err != nil {
			log.Fatal(err)
			return buddies
		}

		if !accepted {
			newBuddy.status = 1 // pending buddy request
		} else if newBuddy.channelID == -1 {
			newBuddy.status = 2 // offline
		} else {
			newBuddy.status = 0 // online
		}

		buddies = append(buddies, newBuddy)

		i++
	}

	return buddies
}

func packetPlayerReceivedDmg(charID int32, attack int8, initalAmmount, reducedAmmount, spawnID, mobID, healSkillID int32,
	stance, reflectAction byte, reflected byte, reflectX, reflectY int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerTakeDmg)
	p.WriteInt32(charID)
	p.WriteInt8(attack)
	p.WriteInt32(initalAmmount)

	p.WriteInt32(spawnID)
	p.WriteInt32(mobID)
	p.WriteByte(stance)
	p.WriteByte(reflected)

	if reflected > 0 {
		p.WriteByte(reflectAction)
		p.WriteInt16(reflectX)
		p.WriteInt16(reflectY)
	}

	p.WriteInt32(reducedAmmount)

	// Check if used
	if reducedAmmount < 0 {
		p.WriteInt32(healSkillID)
	}

	return p
}

func packetPlayerLevelUpAnimation(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerAnimation)
	p.WriteInt32(charID)
	p.WriteByte(0x00)

	return p
}

func packetPlayerMove(charID int32, bytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerMovement)
	p.WriteInt32(charID)
	p.WriteBytes(bytes)

	return p
}

func packetPlayerEmoticon(charID int32, emotion int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerEmoticon)
	p.WriteInt32(charID)
	p.WriteInt32(emotion)

	return p
}

func packetPlayerSkillBookUpdate(skillID int32, level int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSkillRecordUpdate)
	p.WriteByte(0x01)  // time check?
	p.WriteInt16(0x01) // number of skills to update
	p.WriteInt32(skillID)
	p.WriteInt32(level)
	p.WriteByte(0x01)

	return p
}

func packetPlayerStatChange(unknown bool, stat int32, value int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelStatChange)
	p.WriteBool(unknown)
	p.WriteInt32(stat)
	p.WriteInt32(value)

	return p
}

func packetPlayerNoChange() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x00)
	p.WriteByte(0x00)

	return p
}

func packetChangeChannel(ip []byte, port int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelChange)
	p.WriteBool(true)
	p.WriteBytes(ip)
	p.WriteInt16(port)

	return p
}

func packetCannotEnterCashShop() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelChangeServer)
	p.WriteByte(2)

	return p
}

func packetPlayerEnterGame(plr player, channelID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelWarpToMap)
	p.WriteInt32(channelID)
	p.WriteByte(0) // character portal counter
	p.WriteByte(1) // Is connecting

	randomBytes := make([]byte, 4)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err.Error())
	}
	p.WriteBytes(randomBytes)
	p.WriteBytes(randomBytes)
	p.WriteBytes(randomBytes)
	p.WriteBytes(randomBytes)

	// Are active buffs name encoded in here?
	p.WriteByte(0xFF)
	p.WriteByte(0xFF)

	p.WriteInt32(plr.id)
	p.WritePaddedString(plr.name, 13)
	p.WriteByte(plr.gender)
	p.WriteByte(plr.skin)
	p.WriteInt32(plr.face)
	p.WriteInt32(plr.hair)

	p.WriteInt64(0) // Pet Cash ID

	p.WriteByte(plr.level)
	p.WriteInt16(plr.job)
	p.WriteInt16(plr.str)
	p.WriteInt16(plr.dex)
	p.WriteInt16(plr.intt)
	p.WriteInt16(plr.luk)
	p.WriteInt16(plr.hp)
	p.WriteInt16(plr.maxHP)
	p.WriteInt16(plr.mp)
	p.WriteInt16(plr.maxMP)
	p.WriteInt16(plr.ap)
	p.WriteInt16(plr.sp)
	p.WriteInt32(plr.exp)
	p.WriteInt16(plr.fame)

	p.WriteInt32(plr.mapID)
	p.WriteByte(plr.mapPos)

	p.WriteByte(20) // budy list size
	p.WriteInt32(plr.mesos)

	p.WriteByte(plr.equipSlotSize)
	p.WriteByte(plr.useSlotSize)
	p.WriteByte(plr.setupSlotSize)
	p.WriteByte(plr.etcSlotSize)
	p.WriteByte(plr.cashSlotSize)

	for _, v := range plr.equip {
		if v.slotID < 0 && !v.cash {
			p.WriteBytes(v.inventoryBytes())
		}
	}

	p.WriteByte(0)

	// Equips
	for _, v := range plr.equip {
		if v.slotID < 0 && v.cash {
			p.WriteBytes(v.inventoryBytes())
		}
	}

	p.WriteByte(0)

	// Inventory windows starts
	for _, v := range plr.equip {
		if v.slotID > -1 {
			p.WriteBytes(v.inventoryBytes())
		}
	}

	p.WriteByte(0)

	for _, v := range plr.use {
		p.WriteBytes(v.inventoryBytes())
	}

	p.WriteByte(0)

	for _, v := range plr.setUp {
		p.WriteBytes(v.inventoryBytes())
	}

	p.WriteByte(0)

	for _, v := range plr.etc {
		p.WriteBytes(v.inventoryBytes())
	}

	p.WriteByte(0)

	for _, v := range plr.cash {
		p.WriteBytes(v.inventoryBytes())
	}

	p.WriteByte(0)

	// Skills
	p.WriteInt16(int16(len(plr.skills))) // number of skills

	skillCooldowns := make(map[int32]int16)

	for _, skill := range plr.skills {
		p.WriteInt32(skill.ID)
		p.WriteInt32(int32(skill.Level))

		if skill.Cooldown > 0 {
			skillCooldowns[skill.ID] = skill.Cooldown
		}
	}

	p.WriteInt16(int16(len(skillCooldowns))) // number of cooldowns

	for id, cooldown := range skillCooldowns {
		p.WriteInt32(id)
		p.WriteInt16(cooldown)
	}

	// Quests
	p.WriteInt16(3) // Active quest count
	p.WriteInt16(2029)
	p.WriteString("")
	p.WriteInt16(2000)
	p.WriteString("")
	p.WriteInt16(1000)
	p.WriteString("")
	p.WriteInt16(0) // Completed quest count?

	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt64(time.Now().Unix())

	return p
}

func packetInventoryAddItem(item item, newItem bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteBool(!newItem)
	p.WriteByte(item.invID)

	if newItem {
		p.WriteBytes(item.shortBytes())
	} else {
		p.WriteInt16(item.slotID)
		p.WriteInt16(item.amount)
	}

	return p
}

func packetInventoryModifyItemAmount(item item) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(item.invID)
	p.WriteInt16(item.slotID)
	p.WriteInt16(item.amount)

	return p
}

func packetInventoryAddItems(items []item, newItem []bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)

	p.WriteByte(0x01)
	if len(items) != len(newItem) {
		p.WriteByte(0)
		return p
	}

	p.WriteByte(byte(len(items)))

	for i, v := range items {
		p.WriteBool(!newItem[i])
		p.WriteByte(v.invID)

		if newItem[i] {
			p.WriteBytes(v.shortBytes())
		} else {
			p.WriteInt16(v.slotID)
			p.WriteInt16(v.amount)
		}
	}

	return p
}

func packetInventoryChangeItemSlot(invTabID byte, origPos, newPos int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x02)
	p.WriteByte(invTabID)
	p.WriteInt16(origPos)
	p.WriteInt16(newPos)
	p.WriteByte(0x00) // ?

	return p
}

func packetInventoryRemoveItem(item item) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x03)
	p.WriteByte(item.invID)
	p.WriteInt16(item.slotID)
	p.WriteUint64(0) //?

	return p
}

func packetInventoryChangeEquip(char player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerChangeAvatar)
	p.WriteInt32(char.id)
	p.WriteByte(1)
	p.WriteBytes(char.displayBytes())
	p.WriteByte(0xFF)
	p.WriteUint64(0) //?

	return p
}

func packetInventoryNoChange() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x00)
	p.WriteByte(0x00)

	return p
}

func packetInventoryDontTake() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteInt16(1)

	return p
}

func packetGuildInfo(id int32, name string, memberCount byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x1a)

	if len(name) == 0 {
		p.WriteByte(0x00) // removes player from guild
		return p
	}

	p.WriteBool(true) // In guild
	p.WriteInt32(1)   // guild id (value cannot be zero)
	p.WriteString(name)

	// 5 ranks each have a title
	p.WriteString("rank1")
	p.WriteString("rank2")
	p.WriteString("rank3")
	p.WriteString("rank4")
	p.WriteString("rank5")

	capacity := 250             // maximum
	p.WriteByte(byte(capacity)) // member count

	// iterate over all members and output ids
	for i := 0; i < capacity; i++ {
		p.WriteInt32(int32(i + 1))
	}

	// iterate over all members and input their info
	for i := 0; i < capacity; i++ {
		p.WritePaddedString("[GM]Hucaru", 13) // name
		p.WriteInt32(510)                     // job
		p.WriteInt32(255)                     // level

		if i > 4 {
			p.WriteInt32(5) // rank starts at 1
		} else {
			p.WriteInt32(int32(i + 1)) // rank starts at 1
		}

		if i%2 == 0 {
			p.WriteInt32(1) // online or not
		} else {
			p.WriteInt32(0)
		}

		p.WriteInt32(int32(i)) // ?
	}

	p.WriteInt32(int32(capacity)) // capacity
	p.WriteInt16(1030)            // logo background
	p.WriteByte(3)                // logo bg colour
	p.WriteInt16(4017)            // logo
	p.WriteByte(2)                // logo colour
	p.WriteString("notice")       // notice
	p.WriteInt32(9999)            // ?

	return p
}

func packetBuddyInfo(buddyList []buddy) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x12)
	p.WriteByte(byte(len(buddyList)))

	for _, v := range buddyList {
		p.WriteInt32(v.id)
		p.WritePaddedString(v.name, 13)
		p.WriteByte(v.status)
		p.WriteInt32(v.channelID)
	}

	for _, v := range buddyList {
		p.WriteInt32(v.cashShop) // wizet mistake and this should be a bool?
	}

	return p
}

// It is possible to change id's using this packet, however if the id is a request it will crash the users
// client when selecting an option in notification, therefore the id has not been allowed to change
func packetBuddyUpdate(id int32, name string, status byte, channelID int32, cashShop bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x08)
	p.WriteInt32(id) // original id
	p.WriteInt32(id)
	p.WritePaddedString(name, 13)
	p.WriteByte(status)
	p.WriteInt32(channelID)
	p.WriteBool(cashShop)

	return p
}

func packetBuddyListSizeUpdate(size byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x15)
	p.WriteByte(size)

	return p
}

func packetPlayerAvatarSummaryWindow(charID int32, plr player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelAvatarInfoWindow)
	p.WriteInt32(plr.id)
	p.WriteByte(plr.level)
	p.WriteInt16(plr.job)
	p.WriteInt16(plr.fame)

	p.WriteString(plr.guild)

	p.WriteBool(false) // if has pet
	p.WriteByte(0)     // wishlist count

	return p
}

func packetShowCountdown(time int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCountdown)
	p.WriteByte(2)
	p.WriteInt32(time)

	return p
}

func packetHideCountdown() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCountdown)
	p.WriteByte(0)
	p.WriteInt32(0)

	return p
}

func packetBuddyUnkownError() mpacket.Packet {
	return packetBuddyRequestResult(0x16)
}

func packetBuddyPlayerFullList() mpacket.Packet {
	return packetBuddyRequestResult(0x0b)
}

func packetBuddyOtherFullList() mpacket.Packet {
	return packetBuddyRequestResult(0x0c)
}

func packetBuddyAlreadyAdded() mpacket.Packet {
	return packetBuddyRequestResult(0x0d)
}

func packetBuddyIsGM() mpacket.Packet {
	return packetBuddyRequestResult(0x0e)
}

func packetBuddyNameNotRegistered() mpacket.Packet {
	return packetBuddyRequestResult(0x0f)
}

func packetBuddyRequestResult(code byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(code)

	return p
}

func packetBuddyReceiveRequest(fromID int32, fromName string, fromChannelID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x9)
	p.WriteInt32(fromID)
	p.WriteString(fromName)
	p.WriteInt32(fromID)
	p.WritePaddedString(fromName, 13)
	p.WriteByte(1)
	p.WriteInt32(fromChannelID)
	p.WriteBool(false) // sender in cash shop

	return p
}

func packetBuddyOnlineStatus(id int32, channelID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x14)
	p.WriteInt32(id)
	p.WriteInt8(0)
	p.WriteInt32(channelID)

	return p
}

func packetBuddyChangeChannel(id int32, channelID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x14)
	p.WriteInt32(id)
	p.WriteInt8(1)
	p.WriteInt32(channelID)

	return p
}
