package channel

import (
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	mathrand "math/rand"
	"os"
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/google/uuid"
)

type dropTableEntry struct {
	IsMesos bool  `json:"isMesos"`
	ItemID  int32 `json:"itemId"`
	Min     int32 `json:"min"`
	Max     int32 `json:"max"`
	QuestID int32 `json:"questId"` // TODO: Validate this
	Chance  int64 `json:"chance"`
}

// DropTable is the global lookup table for drops
var dropTable map[int32][]dropTableEntry

// PopulateDropTable from json file
func populateDropTable(dropJSON string) error {
	jsonFile, err := os.Open(dropJSON)

	if err != nil {
		return err
	}

	defer jsonFile.Close()

	jsonBytes, _ := ioutil.ReadAll(jsonFile)

	return json.Unmarshal(jsonBytes, &dropTable)
}

type Item struct {
	dbID         int64
	uuid         uuid.UUID
	cash         bool
	cashID       int64
	cashSN       int32
	invID        byte
	slotID       int16
	ID           int32
	expireTime   int64
	amount       int16
	creatorName  string
	flag         int16
	upgradeSlots byte
	reqLevel     byte
	scrollLevel  byte
	str          int16
	dex          int16
	intt         int16
	luk          int16
	reqStr       int16
	reqDex       int16
	reqInt       int16
	reqLuk       int16
	hp           int16
	mp           int16
	hpr          int16
	mpr          int16
	watk         int16
	matk         int16
	wdef         int16
	mdef         int16
	accuracy     int16
	avoid        int16
	hands        int16
	speed        int16
	jump         int16
	attackSpeed  int16
	buffTime     int16
	stand        byte // TODO: Investigate this, it doesn't appear to be saved or used anywhere

	weaponType byte
	twoHanded  bool
	pet        bool
	petData    *pet

	spawnMobs map[int32]int32
}

const neverExpire int64 = 150842304000000000

// GenerateCashID generates a unique cash ID using crypto/rand
func GenerateCashID() int64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		log.Println("Warning: crypto/rand failed in GenerateCashID")
		b[0] = 0
		for i := 1; i < 8; i++ {
			b[i] = byte(i * 17)
		}
	}

	cashID := int64(binary.LittleEndian.Uint64(b[:]))
	return cashID & 0x00FFFFFFFFFFFFFF
}

func loadInventoryFromDb(charID int32) ([]Item, []Item, []Item, []Item, []Item) {
	filter := "ID,inventoryID,itemID,slotNumber,amount,flag,upgradeSlots,level,str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,expireTime,creatorName,cashID,cashSN"
	row, err := common.DB.Query("SELECT "+filter+" FROM items WHERE characterID=?", charID)

	if err != nil {
		panic(err)
	}

	equip := []Item{}
	use := []Item{}
	setUp := []Item{}
	etc := []Item{}
	cash := []Item{}

	defer row.Close()

	for row.Next() {

		item := Item{uuid: uuid.New()}
		var cashIDNullable sql.NullInt64
		var cashSNNullable sql.NullInt32

		err := row.Scan(&item.dbID,
			&item.invID,
			&item.ID,
			&item.slotID,
			&item.amount,
			&item.flag,
			&item.upgradeSlots,
			&item.scrollLevel,
			&item.str,
			&item.dex,
			&item.intt,
			&item.luk,
			&item.hp,
			&item.mp,
			&item.watk,
			&item.matk,
			&item.wdef,
			&item.mdef,
			&item.accuracy,
			&item.avoid,
			&item.hands,
			&item.speed,
			&item.jump,
			&item.expireTime,
			&item.creatorName,
			&cashIDNullable,
			&cashSNNullable)

		if err != nil {
			log.Println(err)
			continue
		}

		if cashIDNullable.Valid {
			item.cashID = cashIDNullable.Int64
		}
		if cashSNNullable.Valid {
			item.cashSN = cashSNNullable.Int32
		}

		if nxInfo, err := nx.GetItem(item.ID); err == nil {
			item.cash = nxInfo.Cash

			if item.cash && item.cashID == 0 {
				item.cashID = GenerateCashID()
			}

			if item.cash && item.cashSN == 0 {
				if sn, ok := nx.GetCommoditySNByItemID(item.ID); ok {
					item.cashSN = sn
				}
			}

			item.pet = nxInfo.Pet
			if item.pet {
				petRow := common.DB.QueryRow(`
					SELECT name, sn, level, closeness, fullness,
						   deadDate, spawnDate, lastInteraction
					FROM pets WHERE parentID=?`, item.dbID)

				petData := pet{
					itemID:   item.ID,
					itemDBID: item.dbID,
				}

				if err := petRow.Scan(
					&petData.name,
					&petData.sn,
					&petData.level,
					&petData.closeness,
					&petData.fullness,
					&petData.deadDate,
					&petData.spawnDate,
					&petData.lastInteraction,
				); err == nil {
					item.petData = &petData
				} else if err == sql.ErrNoRows {
					sn, _ := nx.GetCommoditySNByItemID(item.ID)
					item.petData = newPet(item.ID, sn, item.dbID)
				} else {
					log.Println("error loading pet:", err)
				}

				err := savePet(&item)
				if err != nil {
					log.Println(err)
				}
			}
			item.buffTime = nxInfo.Time
			item.spawnMobs = nxInfo.SpawnMobs
		}

		item.calculateWeaponType()

		switch item.invID {
		case 1:
			equip = append(equip, item)
		case 2:
			use = append(use, item)
		case 3:
			setUp = append(setUp, item)
		case 4:
			etc = append(etc, item)
		case 5:
			cash = append(cash, item)
		default:
		}

	}

	return equip, use, setUp, etc, cash
}

func createPerfectItemFromID(id int32, amount int16) (Item, error) {
	return createBiasItemFromID(id, amount, 1, false)
}

func CreateItemFromID(id int32, amount int16) (Item, error) {
	return createBiasItemFromID(id, amount, 0, false)
}

func createItemWorstFromID(id int32, amount int16) (Item, error) {
	return createBiasItemFromID(id, amount, -1, false)
}

func createAverageItemFromID(id int32, amount int16) (Item, error) {
	return createBiasItemFromID(id, amount, 0, true)
}

func createBiasItemFromID(id int32, amount int16, bias int8, average bool) (Item, error) {
	randomStat := func(stat float64, average bool) int16 {
		if average {
			return int16(stat)
		}

		max := int(math.Ceil(stat * 1.1))
		min := int(math.Floor(stat * 0.9))

		if bias == 1 {
			return int16(max)
		} else if bias == -1 {
			return int16(min)
		}

		if max-min == 0 {
			return int16(max)
		}

		mathrand.Seed(time.Now().Unix())

		return int16(mathrand.Intn(max-min) + min)
	}

	newItem := Item{dbID: 0, uuid: uuid.New()}

	nxInfo, err := nx.GetItem(id)

	if err != nil {
		return Item{}, fmt.Errorf("Unable to generate Item of ID: %v", id)
	}

	newItem.cash = nxInfo.Cash
	newItem.invID = byte(id / 1e6)
	newItem.ID = id
	newItem.buffTime = nxInfo.Time
	newItem.accuracy = randomStat(nxInfo.IncACC, average)
	newItem.avoid = randomStat(nxInfo.IncEVA, average)
	newItem.speed = randomStat(nxInfo.IncSpeed, average)

	newItem.matk = randomStat(nxInfo.IncMAD, average)
	newItem.mdef = randomStat(nxInfo.IncMDD, average)
	newItem.watk = randomStat(nxInfo.IncPAD, average)
	newItem.wdef = randomStat(nxInfo.IncPDD, average)

	newItem.hp = nxInfo.HP
	newItem.mp = nxInfo.MP

	newItem.str = nxInfo.IncSTR
	newItem.dex = nxInfo.IncDEX
	newItem.intt = nxInfo.IncINT
	newItem.luk = nxInfo.IncLUK

	newItem.attackSpeed = nxInfo.AttackSpeed
	newItem.reqLevel = nxInfo.ReqLevel
	newItem.upgradeSlots = nxInfo.Tuc
	newItem.pet = nxInfo.Pet
	newItem.spawnMobs = nxInfo.SpawnMobs

	if amount < 1 {
		amount = 1
	}

	newItem.amount = amount
	newItem.stand = byte(nxInfo.Stand)
	newItem.calculateWeaponType()

	newItem.expireTime = neverExpire

	return newItem, nil
}

func (v *Item) calculateWeaponType() {
	switch v.ID / 10000 % 100 {
	case 30:
		v.weaponType = 1 // Sword1H
	case 31:
		v.weaponType = 2 // Axe1H
	case 32:
		v.weaponType = 3 // Blunt1H
	case 33:
		v.weaponType = 4 // Dagger
	case 37:
		v.weaponType = 5 // Wand
	case 38:
		v.weaponType = 6 // Staff
	case 40:
		v.weaponType = 7 // Sword2H
	case 41:
		v.weaponType = 8 // Axe2H
	case 42:
		v.weaponType = 9 // Blunt2H
	case 43:
		v.weaponType = 10 // Spear
	case 44:
		v.weaponType = 11 // PoleArm
	case 45:
		v.weaponType = 12 // Bow
	case 46:
		v.weaponType = 13 // Crossbow
	case 47:
		v.weaponType = 14 // Claw
	case 48:
		v.weaponType = 15 // Knuckle
	case 49:
		v.weaponType = 16 // Gun
	case 9:
		v.weaponType = 17 // Shield
	default:
		v.weaponType = 0 // Not a weapon
	}

	if v.weaponType > 6 && v.weaponType < 15 {
		v.twoHanded = true
	}
}

func (v Item) isStackable() bool {
	bullet := v.ID / 1e4

	if v.invID != 5.0 && // pet Item
		v.invID != 1.0 && // equip
		bullet != 207 && // star/arrow etc
		v.amount <= constant.MaxItemStack {

		return true
	}

	return false
}

func (v Item) getSlots() int {
	return int(v.upgradeSlots)
}

func (v *Item) setSlots(slots int) {
	v.upgradeSlots = byte(slots)
}

// SetCashID sets the cash shop storage ID for tracking items from cash shop
func (v *Item) SetCashID(cashID int64) { v.cashID = cashID }

// SetCashSN sets the commodity serial number for cash shop items
func (v *Item) SetCashSN(sn int32) { v.cashSN = sn }

// GetCashSN returns the commodity serial number
func (v Item) GetCashSN() int32 { return v.cashSN }

func (v Item) GetAmount() int16 { return v.amount }

func (v Item) GetExpireTime() int64 { return v.expireTime }

func (v Item) isRechargeable() bool {
	return float64(v.ID/10000) == 207 // Taken from client
}

func (v Item) shield() bool {
	return v.weaponType == 17
}

func (v *Item) save(charID int32) (bool, error) {
	if v.dbID == 0 {
		props := `characterID,inventoryID,itemID,slotNumber,amount,flag,upgradeSlots,level,
				str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,
				expireTime,creatorName,cashID,cashSN`

		query := "INSERT into items (" + props + ") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

		res, err := common.DB.Exec(query,
			charID, v.invID, v.ID, v.slotID, v.amount, v.flag, v.upgradeSlots, v.scrollLevel,
			v.str, v.dex, v.intt, v.luk, v.hp, v.mp, v.watk, v.matk, v.wdef, v.mdef, v.accuracy, v.avoid, v.hands, v.speed, v.jump,
			v.expireTime, v.creatorName, sql.NullInt64{Int64: v.cashID, Valid: v.cashID != 0}, sql.NullInt32{Int32: v.cashSN, Valid: v.cashSN != 0})

		if err != nil {
			return false, err
		}

		v.dbID, err = res.LastInsertId()

		if err != nil {
			return false, err
		}
	} else {
		props := `slotNumber=?,amount=?,flag=?,upgradeSlots=?,level=?,
			str=?,dex=?,intt=?,luk=?,hp=?,mp=?,watk=?,matk=?,wdef=?,mdef=?,accuracy=?,avoid=?,hands=?,speed=?,jump=?,
			expireTime=?,cashID=?,cashSN=?`

		query := "UPDATE items SET " + props + " WHERE ID=?"

		_, err := common.DB.Exec(query,
			v.slotID, v.amount, v.flag, v.upgradeSlots, v.scrollLevel,
			v.str, v.dex, v.intt, v.luk, v.hp, v.mp, v.watk, v.matk, v.wdef, v.mdef, v.accuracy, v.avoid, v.hands, v.speed, v.jump,
			v.expireTime, sql.NullInt64{Int64: v.cashID, Valid: v.cashID != 0}, sql.NullInt32{Int32: v.cashSN, Valid: v.cashSN != 0}, v.dbID)

		if err != nil {
			return false, err
		}
	}

	if v.pet {
		err := savePet(v)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func (v Item) SaveToCashShopStorage(tx *sql.Tx, accountID int32, slotNumber int16, cashID int64, sn int32) error {
	if v.ID == 0 {
		return nil
	}

	const ins = `
		INSERT INTO account_cashshop_storage_items(
			accountID, itemID, cashID, sn, slotNumber, amount, flag, upgradeSlots, level,
			str, dex, intt, luk, hp, mp, watk, matk, wdef, mdef, accuracy, avoid, hands,
			speed, jump, expireTime, creatorName
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_, err := tx.Exec(
		ins,
		accountID, v.ID, cashID, sn, slotNumber, v.amount,
		v.flag, v.upgradeSlots, v.scrollLevel,
		v.str, v.dex, v.intt, v.luk,
		v.hp, v.mp, v.watk, v.matk,
		v.wdef, v.mdef, v.accuracy, v.avoid,
		v.hands, v.speed, v.jump,
		v.expireTime, v.creatorName,
	)
	return err
}

func (v Item) delete() error {
	query := "DELETE FROM `items` WHERE ID=?"
	_, err := common.DB.Exec(query, v.dbID)

	if err != nil {
		return err
	}

	return nil
}

// InventoryBytes to display in character inventory window
func (v Item) InventoryBytes() []byte {
	return v.bytes(false, false)
}

func (v Item) StorageBytes() []byte {
	return v.bytes(false, true)
}

// ShortBytes e.g. inventory operation, storage window
func (v Item) ShortBytes() []byte {
	return v.bytes(true, false)
}

func (v Item) bytes(shortSlot, storage bool) []byte {
	p := mpacket.NewPacket()

	if !storage {
		if !shortSlot {
			if v.slotID < 0 {
				if v.slotID < -100 {
					p.WriteByte(byte(math.Abs(float64(v.slotID + 100))))
				} else {
					p.WriteByte(byte(math.Abs(float64(v.slotID))))
				}
			} else {
				p.WriteByte(byte(v.slotID))
			}
		} else {
			p.WriteInt16(v.slotID)
		}
	}

	if v.invID == 1 {
		p.WriteByte(0x01)
	} else if v.pet {
		p.WriteByte(0x03)
	} else {
		p.WriteByte(0x02)
	}

	p.WriteInt32(v.ID)

	p.WriteBool(v.cash)
	if v.cash {
		// Write the unique cash ID (not the SN) for cash shop tracking
		p.WriteUint64(uint64(v.cashID))
	}

	p.WriteInt64(v.expireTime)

	if v.invID == 1 {
		p.WriteByte(v.upgradeSlots)
		p.WriteByte(v.scrollLevel)
		p.WriteInt16(v.str)
		p.WriteInt16(v.dex)
		p.WriteInt16(v.intt)
		p.WriteInt16(v.luk)
		p.WriteInt16(v.hp)
		p.WriteInt16(v.mp)
		p.WriteInt16(v.watk)
		p.WriteInt16(v.matk)
		p.WriteInt16(v.wdef)
		p.WriteInt16(v.mdef)
		p.WriteInt16(v.accuracy)
		p.WriteInt16(v.avoid)
		p.WriteInt16(v.hands)
		p.WriteInt16(v.speed)
		p.WriteInt16(v.jump)
		p.WriteString(v.creatorName)
		p.WriteInt16(v.flag) // lock/seal, show, spikes, cape, cold protection etc ?
	} else if v.pet {
		p.WritePaddedString(v.petData.name, 13)
		p.WriteByte(v.petData.level)
		p.WriteInt16(v.petData.closeness)
		p.WriteByte(v.petData.fullness)
		p.WriteInt64(v.petData.deadDate)
		p.WriteInt32(0) // Pet flags?
	} else {
		p.WriteInt16(v.amount)
		p.WriteString(v.creatorName)
		p.WriteInt16(v.flag) // even (normal), odd (sealed) ?
	}

	return p
}

// Use applies stat changes for items
func (v Item) use(plr *Player) {
	if plr.hp < 1 {
		plr.noChange()
		return
	}

	// Let's use NX data as source of truth for useable items
	nxData, err := nx.GetItem(v.ID)
	if err != nil {
		log.Println("could not load item from nx: ", v.ID)
		return
	}

	if nxData.HP > 0 {
		plr.giveHP(nxData.HP)
	}
	if nxData.MP > 0 {
		plr.giveMP(nxData.MP)
	}

	if nxData.HPR > 0 {
		base := int(plr.effectiveMaxHP())
		hpAmt := int(math.Floor(float64(base) * float64(nxData.HPR) / 100.0))
		if hpAmt < 1 {
			hpAmt = 1
		}
		plr.giveHP(int16(hpAmt))
	}

	if nxData.MPR > 0 {
		base := int(plr.effectiveMaxMP())
		mpAmt := int(math.Floor(float64(base) * float64(nxData.MPR) / 100.0))
		if mpAmt < 1 {
			mpAmt = 1
		}
		plr.giveMP(int16(mpAmt))
	}

	if plr.buffs == nil {
		NewCharacterBuffs(plr)
	}

	plr.buffs.plr.inst = plr.inst
	plr.buffs.AddItemBuff(nxData, v.ID)
}

// applyScrollEffects mutates the equip with the scroll increments from NX.
func (v *Item) applyScrollEffects(scroll nx.Item) {
	v.str += scroll.IncSTR
	v.dex += scroll.IncDEX
	v.intt += scroll.IncINT
	v.luk += scroll.IncLUK

	v.hp += int16(scroll.IncMHP)
	v.mp += int16(scroll.IncMMP)

	v.watk += int16(scroll.IncPAD)
	v.wdef += int16(scroll.IncPDD)
	v.matk += int16(scroll.IncMAD)
	v.mdef += int16(scroll.IncMDD)
	v.accuracy += int16(scroll.IncACC)
	v.avoid += int16(scroll.IncEVA)

	v.speed += int16(scroll.IncSpeed)
	v.jump += int16(scroll.IncJump)
}

func (v *Item) incrementScrollCount() {
	v.scrollLevel++
}

// CreateItemFromDBValues creates an Item from database values, used for loading saved items
// This preserves all stat modifications, scrolls, etc. from the database
func CreateItemFromDBValues(itemID int32, slotID int16, amount int16, flag int16, upgradeSlots, scrollLevel byte,
	str, dex, intt, luk, hp, mp, watk, matk, wdef, mdef, accuracy, avoid, hands, speed, jump int16,
	expireTime int64, creatorName string) (Item, error) {

	// Create base item to get metadata
	item, err := CreateItemFromID(itemID, amount)
	if err != nil {
		return item, err
	}

	// Override with saved stat values
	item.slotID = slotID
	item.flag = flag
	item.upgradeSlots = upgradeSlots
	item.scrollLevel = scrollLevel
	item.str = str
	item.dex = dex
	item.intt = intt
	item.luk = luk
	item.hp = hp
	item.mp = mp
	item.watk = watk
	item.matk = matk
	item.wdef = wdef
	item.mdef = mdef
	item.accuracy = accuracy
	item.avoid = avoid
	item.hands = hands
	item.speed = speed
	item.jump = jump
	item.expireTime = expireTime
	item.creatorName = creatorName

	return item, nil
}

func getItemType(itemID int32) int32 {
	return itemID / 10000
}

func itemTypeToScrollType(itemID int32) int32 {
	return (getItemType(itemID) % 100) * 100
}

func getScrollType(itemID int32) int32 {
	return (itemID % 10000) - (itemID % 100)
}

// validateScrollTarget performs basic compatibility checks between the scroll and target equip.
func validateScrollTarget(scrollID int32, equipID int32) bool {
	return itemTypeToScrollType(equipID) == getScrollType(scrollID)
}
