package channel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
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
func PopulateDropTable(dropJSON string) error {
	jsonFile, err := os.Open(dropJSON)

	if err != nil {
		return fmt.Errorf(err.Error())
	}

	defer jsonFile.Close()

	jsonBytes, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(jsonBytes, &dropTable)

	return nil
}

type item struct {
	dbID         int64
	uuid         uuid.UUID
	cash         bool
	invID        byte
	slotID       int16
	id           int32
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
	stand        byte // TODO: Investigate this, it doesn't appear to be saved or used anywhere

	weaponType byte
	twoHanded  bool
	pet        bool
}

const neverExpire int64 = 150842304000000000

func loadInventoryFromDb(charID int32) ([]item, []item, []item, []item, []item) {
	filter := "id,inventoryID,itemID,slotNumber,amount,flag,upgradeSlots,level,str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,expireTime,creatorName"
	row, err := common.DB.Query("SELECT "+filter+" FROM items WHERE characterID=?", charID)

	if err != nil {
		panic(err)
	}

	equip := []item{}
	use := []item{}
	setUp := []item{}
	etc := []item{}
	cash := []item{}

	defer row.Close()

	for row.Next() {

		item := item{uuid: uuid.New()}

		row.Scan(&item.dbID,
			&item.invID,
			&item.id,
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
			&item.creatorName)

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

func createPerfectItemFromID(id int32, amount int16) (item, error) {
	return createBiasItemFromID(id, amount, 1, false)
}

func createItemFromID(id int32, amount int16) (item, error) {
	return createBiasItemFromID(id, amount, 0, false)
}

func createItemWorstFromID(id int32, amount int16) (item, error) {
	return createBiasItemFromID(id, amount, -1, false)
}

func createAverageItemFromID(id int32, amount int16) (item, error) {
	return createBiasItemFromID(id, amount, 0, true)
}

func createBiasItemFromID(id int32, amount int16, bias int8, average bool) (item, error) {
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

		rand.Seed(time.Now().Unix())

		return int16(rand.Intn(max-min) + min)
	}

	newItem := item{dbID: 0, uuid: uuid.New()}

	nxInfo, err := nx.GetItem(id)

	if err != nil {
		return item{}, fmt.Errorf("Unable to generate item of id: %v", id)
	}

	newItem.cash = nxInfo.Cash
	newItem.invID = byte(id / 1e6)
	newItem.id = id
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

	if amount < 1 {
		amount = 1
	}

	newItem.amount = amount
	newItem.stand = byte(nxInfo.Stand)
	newItem.calculateWeaponType()

	newItem.expireTime = neverExpire

	return newItem, nil
}

func (v *item) calculateWeaponType() {
	switch v.id / 10000 % 100 {
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

func (v item) isStackable() bool {
	bullet := v.id / 1e4

	if v.invID != 5.0 && // pet item
		v.invID != 1.0 && // equip
		bullet != 207 && // star/arrow etc
		v.amount <= constant.MaxItemStack {

		return true
	}

	return false
}

func (v item) isRechargeable() bool {
	return (math.Floor(float64(v.id/10000)) == 207) // Taken from cliet
}

func (v item) shield() bool {
	return v.weaponType == 17
}

func (v *item) save(charID int32) (bool, error) {
	if v.dbID == 0 {
		props := `characterID,inventoryID,itemID,slotNumber,amount,flag,upgradeSlots,level,
				str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,
				expireTime,creatorName`

		query := "INSERT into items (" + props + ") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

		res, err := common.DB.Exec(query,
			charID, v.invID, v.id, v.slotID, v.amount, v.flag, v.upgradeSlots, v.scrollLevel,
			v.str, v.dex, v.intt, v.luk, v.hp, v.mp, v.watk, v.matk, v.wdef, v.mdef, v.accuracy, v.avoid, v.hands, v.speed, v.jump,
			v.expireTime, v.creatorName)

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
			expireTime=?`

		query := "UPDATE items SET " + props + " WHERE id=?"

		_, err := common.DB.Exec(query,
			v.slotID, v.amount, v.flag, v.upgradeSlots, v.scrollLevel,
			v.str, v.dex, v.intt, v.luk, v.hp, v.mp, v.watk, v.matk, v.wdef, v.mdef, v.accuracy, v.avoid, v.hands, v.speed, v.jump,
			v.expireTime, v.dbID)

		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (v item) delete() error {
	query := "DELETE FROM `items` WHERE id=?"
	_, err := common.DB.Exec(query, v.dbID)

	if err != nil {
		return err
	}

	return nil
}

// InventoryBytes to display in character inventory window
func (v item) inventoryBytes() []byte {
	return v.bytes(false)
}

// ShortBytes e.g. inventory operation, storage window
func (v item) shortBytes() []byte {
	return v.bytes(true)
}

func (v item) bytes(shortSlot bool) []byte {
	p := mpacket.NewPacket()

	if !shortSlot {
		if v.cash && v.slotID < 0 {
			p.WriteByte(byte(math.Abs(float64(v.slotID + 100))))
		} else {
			p.WriteByte(byte(math.Abs(float64(v.slotID))))
		}
	} else {
		p.WriteInt16(v.slotID)
	}

	if v.invID == 1 {
		p.WriteByte(0x01)
	} else if v.pet {
		p.WriteByte(0x03)
	} else {
		p.WriteByte(0x02)
	}

	p.WriteInt32(v.id)

	p.WriteBool(v.cash)
	if v.cash {
		p.WriteUint64(uint64(v.id)) // I think this is somekind of cashshop transaction ID for the item
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
		p.WritePaddedString(v.creatorName, 13)
		p.WriteByte(0)
		p.WriteInt16(0)
		p.WriteByte(0)
		p.WriteInt64(v.expireTime)
		p.WriteInt32(0) // ?
	} else {
		p.WriteInt16(v.amount)
		p.WriteString(v.creatorName)
		p.WriteInt16(v.flag) // even (normal), odd (sealed) ?

		if v.isRechargeable() {
			p.WriteInt64(0) // ?
		}
	}

	return p
}

// Use applies stat changes for items
func (v item) use(plr *player) {
	if plr.hp < 1 {
		plr.noChange()
		return
	}
	if v.hp > 0 {
		plr.giveHP(v.hp)
	}
	if v.mp > 0 {
		plr.giveMP(v.mp)
	}

	// Need to add stat buffs (W.ATT, M.ATT, etc)
	// This will require timers to ensure buffs are removed once finished

}
