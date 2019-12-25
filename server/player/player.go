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

type instance interface {
	Send(mpacket.Packet) error
	CalculateNearestSpawnPortalID(pos.Data) (byte, error)
}

// Data connected to server
type Data struct {
	conn       mnet.Client
	instanceID int

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

	name     string
	gender   byte
	skin     byte
	face     int32
	hair     int32
	chairID  int32
	stance   byte
	pos      pos.Data
	foothold int16
	guild    string

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

	miniGameWins, miniGameDraw, miniGameLoss int32
}

// Conn - client connection associated with this Data
func (d Data) Conn() mnet.Client {
	return d.conn
}

// InstanceID - field instance id the Data is currently on
func (d Data) InstanceID() int {
	return d.instanceID
}

// SetInstanceID - assign the instance id for the Data
func (d *Data) SetInstanceID(id int) {
	d.instanceID = id
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

func (d *Data) levelUp(inst instance) {
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

	d.GiveLevel(1, inst)
}

// SetEXP of the Data
func (d *Data) SetEXP(amount int32, inst instance) {
	if d.level > 199 {
		return
	}

	remainder := amount - constant.ExpTable[d.level-1]

	if remainder >= 0 {
		d.levelUp(inst)
		d.SetEXP(remainder, inst)
	} else {
		d.exp = amount
		d.Send(packetPlayerStatChange(false, constant.ExpID, int32(amount)))
	}
}

// GiveEXP to the Data
func (d *Data) GiveEXP(amount int32, fromMob, fromParty bool, inst instance) {
	if fromMob {
		d.Send(packetMessageExpGained(!fromParty, false, amount))
	} else {
		d.Send(packetMessageExpGained(true, true, amount))
	}

	d.SetEXP(d.exp+amount, inst)
}

// SetLevel of the Data
func (d *Data) SetLevel(amount byte, inst instance) {
	d.level = amount
	d.Send(packetPlayerStatChange(false, constant.LevelID, int32(amount)))
	inst.Send(packetPlayerLevelUpAnimation(d.id))
}

// GiveLevel amount ot the Data
func (d *Data) GiveLevel(amount byte, inst instance) {
	d.SetLevel(d.level+amount, inst)
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
	d.SetHP(d.hp + amount)
	if d.hp < 0 {
		d.SetHP(0)
	}
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
	d.SetMP(d.mp + amount)
	if d.mp < 0 {
		d.SetMP(0)
	}
}

// SetMaxMP of Data
func (d *Data) SetMaxMP(amount int16) {
	d.maxMP = amount
	d.Send(packetPlayerStatChange(true, constant.MaxMpID, int32(amount)))
}

// SetFame of Data
func (d *Data) SetFame(amount int16) {

}

// AddEquip item to slice
func (d *Data) AddEquip(item item.Data) {
	d.equip = append(d.equip, item)
}

// SetGuild of Data
func (d *Data) SetGuild(name string, inst instance) {

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
	d.foothold = frag.Foothold()
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

// SetFoothold of Data
func (d *Data) SetFoothold(fh int16) {
	d.foothold = fh
}

// SetMapID of Data
func (d *Data) SetMapID(id int32) {
	d.mapID = id
}

// SetMapPosID of Data
func (d *Data) SetMapPosID(pos byte) {
	d.mapPos = pos
}

// GiveItem to Data
func (d *Data) GiveItem(newItem item.Data) error {
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
		d.equip = append(d.equip, newItem)
		d.Send(packetInventoryAddItem(newItem, true))
	case 2: // Use
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
			d.use = append(d.use, newItem)
			d.Send(packetInventoryAddItem(newItem, true))
		} else {
			remainder := newItem.Amount() - (constant.MaxItemStack - d.use[index].Amount())

			if remainder > 0 { //partial merge
				slotID, err := findFirstEmptySlot(d.use, d.useSlotSize)

				if err != nil {
					return err
				}

				newItem.SetAmount(remainder)
				newItem.SetSlotID(slotID)
				d.use = append(d.use, newItem)
				d.use[index].SetAmount(constant.MaxItemStack)

				d.Send(packetInventoryAddItems([]item.Data{d.use[index], newItem}, []bool{false, true}))
			} else { // full merge
				d.use[index].SetAmount(d.use[index].Amount() + newItem.Amount())
				d.Send(packetInventoryAddItem(d.use[index], false))
			}
		}
	case 3: // Set-up
		slotID, err := findFirstEmptySlot(d.setUp, d.setupSlotSize)

		if err != nil {
			return err
		}

		newItem.SetSlotID(slotID)
		d.setUp = append(d.setUp, newItem)
		d.Send(packetInventoryAddItem(newItem, true))
	case 4: // Etc
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
			d.etc = append(d.etc, newItem)
			d.Send(packetInventoryAddItem(newItem, true))
		} else {
			remainder := newItem.Amount() - (constant.MaxItemStack - d.etc[index].Amount())

			if remainder > 0 { //partial merge
				slotID, err := findFirstEmptySlot(d.etc, d.etcSlotSize)

				if err != nil {
					return err
				}

				newItem.SetAmount(remainder)
				newItem.SetSlotID(slotID)
				d.etc = append(d.etc, newItem)
				d.etc[index].SetAmount(constant.MaxItemStack)

				d.Send(packetInventoryAddItems([]item.Data{d.etc[index], newItem}, []bool{false, true}))
			} else { // full merge
				d.etc[index].SetAmount(d.etc[index].Amount() + newItem.Amount())
				d.Send(packetInventoryAddItem(d.etc[index], false))
			}
		}
	case 5: // Cash
		// some are stackable, how to tell?
		slotID, err := findFirstEmptySlot(d.cash, d.cashSlotSize)

		if err != nil {
			return err
		}

		newItem.SetSlotID(slotID)
		d.cash = append(d.cash, newItem)
		d.Send(packetInventoryAddItem(newItem, true))
	default:
		return fmt.Errorf("Unkown inventory id: %d", newItem.InvID())
	}
	return nil
}

// TakeItem from Data
func (d *Data) TakeItem(itemID int32, amount int16) (item.Data, error) {
	return item.Data{}, nil
}

// RemoveItem from Data
func (d *Data) RemoveItem(remove item.Data) {
	// TODO(Hucaru): change function signature to (id int32, count int16) (invID, slotID, error)

	// findIndex := func(items []item, item item) int {
	// 	for i, v := range items {
	// 		if v.uuid == remove.uuid {
	// 			return i
	// 		}
	// 	}

	// 	return 0
	// }

	// switch remove.invID {
	// case 1:
	// 	if i := findIndex(d.inventory.equip, remove); i != 0 {
	// 		d.inventory.equip[i] = d.inventory.equip[len(d.inventory.equip)-1]
	// 		d.inventory.equip = d.inventory.equip[:len(d.inventory.equip)-1]
	// 	}
	// case 2:
	// 	if i := findIndex(d.inventory.use, remove); i != 0 {
	// 		d.inventory.use[i] = d.inventory.use[len(d.inventory.use)-1]
	// 		d.inventory.use = d.inventory.use[:len(d.inventory.use)-1]
	// 	}
	// case 3:
	// 	if i := findIndex(d.inventory.setUp, remove); i != 0 {
	// 		d.inventory.setUp[i] = d.inventory.setUp[len(d.inventory.setUp)-1]
	// 		d.inventory.setUp = d.inventory.setUp[:len(d.inventory.setUp)-1]
	// 	}
	// case 4:
	// 	if i := findIndex(d.inventory.etc, remove); i != 0 {
	// 		d.inventory.etc[i] = d.inventory.etc[len(d.inventory.etc)-1]
	// 		d.inventory.etc = d.inventory.etc[:len(d.inventory.etc)-1]
	// 	}
	// case 5:
	// 	if i := findIndex(d.inventory.cash, remove); i != 0 {
	// 		d.inventory.cash[i] = d.inventory.cash[len(d.inventory.cash)-1]
	// 		d.inventory.cash = d.inventory.cash[:len(d.inventory.cash)-1]
	// 	}
	// }
}

// GetItem from Data
func (d Data) GetItem(invID byte, slotID int16) (item.Data, error) {
	var result item.Data
	var err error

	findItem := func(items []item.Data, slotID int16) (item.Data, error) {
		for _, v := range items {
			if v.SlotID() == slotID {
				return v, nil
			}
		}

		return item.Data{}, fmt.Errorf("Unable to get item")
	}

	switch invID {
	case 1:
		result, err = findItem(d.equip, slotID)
	case 2:
		result, err = findItem(d.use, slotID)
	case 3:
		result, err = findItem(d.setUp, slotID)
	case 4:
		result, err = findItem(d.etc, slotID)
	case 5:
		result, err = findItem(d.cash, slotID)
	}

	return result, err
}

// UpdateItem with the same database id
func (d *Data) UpdateItem(orig, new item.Data) {
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

// Foothold Data is currently tied to
func (d Data) Foothold() int16 { return d.foothold }

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
func (d Data) Save(db *sql.DB, inst instance) error {
	query := `UPDATE characters set skin=?, hair=?, face=?, level=?,
	job=?, str=?, dex=?, intt=?, luk=?, hp=?, maxHP=?, mp=?, maxMP=?,
	ap=?, sp=?, exp=?, fame=?, mapID=?, mapPos=?, mesos=? WHERE id=?`

	var mapPos byte
	var err error

	if inst != nil {
		mapPos, err = inst.CalculateNearestSpawnPortalID(d.pos)
	}

	if err != nil {
		return err
	}

	d.mapPos = mapPos

	_, err = db.Exec(query,
		d.skin, d.hair, d.face, d.level, d.job, d.str, d.dex, d.intt, d.luk, d.hp, d.maxHP, d.mp,
		d.maxMP, d.ap, d.sp, d.exp, d.fame, d.mapID, d.mapPos, d.mesos, d.id)

	// TODO: Move these out into relevant item operations, add item, move item etc
	// send sql queries to a dedicated green thread for item updates once a db id is acquired
	for _, v := range d.equip {
		v.Save(db, d.id)
	}

	for _, v := range d.use {
		v.Save(db, d.id)
	}

	for _, v := range d.setUp {
		v.Save(db, d.id)
	}

	for _, v := range d.etc {
		v.Save(db, d.id)
	}

	for _, v := range d.cash {
		v.Save(db, d.id)
	}

	// TODO: Move this into skill book update, this happens 3 times every level (or 15 at a time for min maxers)
	// There has to be a better way of doing this in mysql
	for skillID, skill := range d.skills {
		query = `UPDATE skills SET level=?, cooldown=? WHERE skillID=? AND characterID=?`
		result, err := db.Exec(query, skill.Level, skill.Cooldown, skillID, d.id)

		if rows, _ := result.RowsAffected(); rows < 1 || err != nil {
			query = `INSERT INTO skills (characterID, skillID, level, cooldown) VALUES (?, ?, ?, ?)`
			_, err = db.Exec(query, d.id, skillID, skill.Level, 0)
		}
	}

	return err
}
