package item

import (
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/google/uuid"
)

type Data struct {
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
	stand        byte

	weaponType byte
	twoHanded  bool
	pet        bool
}

func (v Data) DbID() int64         { return v.dbID }
func (v Data) ID() int32           { return v.id }
func (v Data) Cash() bool          { return v.cash }
func (v Data) SlotID() int16       { return v.slotID }
func (v Data) InvID() byte         { return v.invID }
func (v Data) Pet() bool           { return v.pet }
func (v Data) UpgradeSlots() byte  { return v.upgradeSlots }
func (v Data) ScrollLevel() byte   { return v.scrollLevel }
func (v Data) Str() int16          { return v.str }
func (v Data) Dex() int16          { return v.dex }
func (v Data) Int() int16          { return v.intt }
func (v Data) Luk() int16          { return v.luk }
func (v Data) Hp() int16           { return v.hp }
func (v Data) Mp() int16           { return v.mp }
func (v Data) Watk() int16         { return v.watk }
func (v Data) Matk() int16         { return v.matk }
func (v Data) Wdef() int16         { return v.wdef }
func (v Data) Mdef() int16         { return v.mdef }
func (v Data) Accuracy() int16     { return v.accuracy }
func (v Data) Avoid() int16        { return v.avoid }
func (v Data) Hands() int16        { return v.hands }
func (v Data) Speed() int16        { return v.speed }
func (v Data) Jump() int16         { return v.jump }
func (v Data) CreatorName() string { return v.creatorName }
func (v Data) Flag() int16         { return v.flag }
func (v Data) ExpireTime() int64   { return v.expireTime }
func (v Data) Amount() int16       { return v.amount }

// LoadInventoryFromDb gets the inventory for a given database connection and character id, returning equip, use, set-up, etc and cash slices
func LoadInventoryFromDb(db *sql.DB, charID int32) ([]Data, []Data, []Data, []Data, []Data) {
	filter := "id,inventoryID,itemID,slotNumber,amount,flag,upgradeSlots,level,str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,expireTime,creatorName"
	row, err := db.Query("SELECT "+filter+" FROM items WHERE characterID=?", charID)

	if err != nil {
		panic(err)
	}

	equip := []Data{}
	use := []Data{}
	setUp := []Data{}
	etc := []Data{}
	cash := []Data{}

	defer row.Close()

	for row.Next() {

		item := Data{uuid: uuid.New()}

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

// CreatePerfectFromID creates an item with bis stats
func CreatePerfectFromID(id int32, amount int16) (Data, error) { return createItemFromID(id, amount, 1) }

// CreateFromID creates an item with randomised stats within a predefined percentage range
func CreateFromID(id int32, amount int16) (Data, error) { return createItemFromID(id, amount, 0) }

// CreateWorstFromID creates an item with wis stats
func CreateWorstFromID(id int32, amount int16) (Data, error) { return createItemFromID(id, amount, -1) }

func createItemFromID(id int32, amount int16, bias int8) (Data, error) {
	randomStat := func(min, max int) int16 {
		if bias > 0 {
			return int16(max)
		} else if bias < 1 {
			return int16(min)
		}

		if max-min == 0 {
			return int16(max)
		}

		rand.Seed(time.Now().Unix())

		return int16(rand.Intn(max-min) + min)
	}

	newItem := Data{dbID: 0, uuid: uuid.New()}

	nxInfo, err := nx.GetItem(id)

	if err != nil {
		return Data{}, fmt.Errorf("Unable to generate item of id: %v", id)
	}

	newItem.cash = nxInfo.Cash
	newItem.invID = byte(id / 1e6)
	newItem.id = id
	newItem.accuracy = randomStat(int(math.Floor(nxInfo.IncACC*0.9)), int(math.Ceil(nxInfo.IncACC*1.1)))
	newItem.avoid = randomStat(int(math.Floor(nxInfo.IncEVA*0.9)), int(math.Ceil(nxInfo.IncEVA*1.1)))
	newItem.speed = randomStat(int(math.Floor(nxInfo.IncSpeed*0.9)), int(math.Ceil(nxInfo.IncSpeed*1.1)))

	newItem.matk = randomStat(int(math.Floor(nxInfo.IncMAD*0.9)), int(math.Ceil(nxInfo.IncMAD*1.1)))
	newItem.mdef = randomStat(int(math.Floor(nxInfo.IncMDD*0.9)), int(math.Ceil(nxInfo.IncMDD*1.1)))
	newItem.watk = randomStat(int(math.Floor(nxInfo.IncPAD*0.9)), int(math.Ceil(nxInfo.IncPAD*1.1)))
	newItem.wdef = randomStat(int(math.Floor(nxInfo.IncPDD*0.9)), int(math.Ceil(nxInfo.IncPDD*1.1)))

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

	return newItem, nil
}

func (v *Data) calculateWeaponType() {
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

func (v *Data) SetDbID(id int64) {
	v.dbID = id
}

func (v *Data) SetCreatorName(name string) {
	v.creatorName = name
}

func (v *Data) SetSlotID(id int16) {
	v.slotID = id
}

func (v *Data) SetAmount(value int16) {
	v.amount = value
}

func (v Data) IsStackable() bool {
	bullet := v.id / 1e4

	if v.invID != 5.0 && // pet item
		v.invID != 1.0 && // equip
		bullet != 207 && // star/arrow etc
		v.amount <= constant.MaxItemStack {

		return true
	}

	return false
}

func (v Data) IsRechargeable() bool {
	return (math.Floor(float64(v.id/10000)) == 207) // Taken from cliet
}

func (v Data) TwoHanded() bool {
	return v.twoHanded
}

func (v Data) Shield() bool {
	return v.weaponType == 17
}

// Save item to database
func (v Data) Save(db *sql.DB, charID int32) (bool, error) {
	if v.dbID == 0 {
		props := `characterID,inventoryID,itemID,slotNumber,amount,flag,upgradeSlots,level,
				str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,
				expireTime,creatorName`

		query := "INSERT into items (" + props + ") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

		res, err := db.Exec(query,
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

		_, err := db.Exec(query,
			v.slotID, v.amount, v.flag, v.upgradeSlots, v.scrollLevel,
			v.str, v.dex, v.intt, v.luk, v.hp, v.mp, v.watk, v.matk, v.wdef, v.mdef, v.accuracy, v.avoid, v.hands, v.speed, v.jump,
			v.expireTime, v.dbID)

		if err != nil {
			return false, err
		}
	}
	return true, nil
}

// Delete item from database
func (v Data) Delete(db *sql.DB) error {
	query := "DELETE FROM `items` WHERE id=?"
	_, err := db.Exec(query, v.dbID)

	if err != nil {
		return err
	}

	return nil
}

// InventoryBytes to display in character inventory window
func (v Data) InventoryBytes() []byte {
	return v.bytes(false)
}

// ShortBytes e.g. inventory operation, storage window
func (v Data) ShortBytes() []byte {
	return v.bytes(true)
}

func (v Data) bytes(shortSlot bool) []byte {
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

		if v.IsRechargeable() {
			p.WriteInt32(0) // ?
		}
	}

	return p
}
