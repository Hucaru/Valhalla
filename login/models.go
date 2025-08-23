package login

import (
	"log"
	"math"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

type player struct {
	id        int32 // Unique identifier of the character
	accountID int32
	worldID   byte

	mapID  int32
	mapPos byte

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
	guild   string

	equip []item
}

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

func (d player) save() error {
	query := `UPDATE characters set skin=?, hair=?, face=?, level=?,
	job=?, str=?, dex=?, intt=?, luk=?, hp=?, maxHP=?, mp=?, maxMP=?,
	ap=?, sp=?, exp=?, fame=?, mapID=?, mapPos=? WHERE id=?`

	d.mapPos = 0

	_, err := common.DB.Exec(query,
		d.skin, d.hair, d.face, d.level, d.job, d.str, d.dex, d.intt, d.luk, d.hp, d.maxHP, d.mp,
		d.maxMP, d.ap, d.sp, d.exp, d.fame, d.mapID, d.mapPos, d.id)

	if err != nil {
		return err
	}

	return err
}

func getCharactersFromAccountWorldID(accountID int32, worldID byte) []player {
	c := []player{}

	filter := "id,accountID,worldID,name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp,exp,fame,mapID,mapPos"

	chars, err := common.DB.Query("SELECT "+filter+" FROM characters WHERE accountID=? AND worldID=?", accountID, worldID)

	if err != nil {
		log.Println(err)
	}

	defer chars.Close()

	for chars.Next() {
		var char player

		err = chars.Scan(&char.id, &char.accountID, &char.worldID, &char.name, &char.gender, &char.skin, &char.hair,
			&char.face, &char.level, &char.job, &char.str, &char.dex, &char.intt, &char.luk, &char.hp, &char.maxHP,
			&char.mp, &char.maxMP, &char.ap, &char.sp, &char.exp, &char.fame, &char.mapID, &char.mapPos)

		if err != nil {
			log.Println(err)
		}

		char.equip = loadEquipsFromDb(char.id)

		c = append(c, char)
	}

	return c
}

func loadPlayerFromID(id int32) player {
	c := player{}
	filter := "id,accountID,worldID,name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp,exp,fame,mapID,mapPos"

	err := common.DB.QueryRow("SELECT "+filter+" FROM characters where id=?", id).Scan(&c.id,
		&c.accountID, &c.worldID, &c.name, &c.gender, &c.skin, &c.hair, &c.face,
		&c.level, &c.job, &c.str, &c.dex, &c.intt, &c.luk, &c.hp, &c.maxHP, &c.mp,
		&c.maxMP, &c.ap, &c.sp, &c.exp, &c.fame, &c.mapID, &c.mapPos)

	if err != nil {
		log.Println(err)
		return c
	}

	c.equip = loadEquipsFromDb(c.id)

	return c
}

type item struct {
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
}

func newAdminItem(id int32, name string) item {
	data, _ := nx.GetItem(id)

	return item{
		id:           id,
		invID:        1,
		amount:       1,
		creatorName:  name,
		flag:         0,
		upgradeSlots: data.Tuc,
		reqLevel:     data.ReqLevel,
		scrollLevel:  0,
		str:          data.IncSTR,
		dex:          data.IncDEX,
		intt:         data.IncINT,
		luk:          data.IncLUK,
		reqStr:       data.ReqSTR,
		reqDex:       data.ReqDEX,
		reqInt:       data.ReqINT,
		reqLuk:       data.ReqLUK,
		hp:           int16(data.IncMHP),
		mp:           int16(data.IncMMP),
		watk:         int16(data.IncPAD),
		matk:         int16(data.IncMAD),
		wdef:         int16(data.IncPDD),
		mdef:         int16(data.IncPAD),
		accuracy:     int16(data.IncACC),
		avoid:        int16(data.IncEVA),
		hands:        0,
		speed:        int16(data.IncSpeed),
		jump:         int16(data.IncJump),
		attackSpeed:  data.AttackSpeed,
	}
}

func newBeginnerItem(id int32) item {
	return newAdminItem(id, "")
}

func (v item) save(charID int32) (bool, error) {
	props := `characterID,inventoryID,itemID,slotNumber,amount,flag,upgradeSlots,level,
				str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,
				expireTime,creatorName`

	query := "INSERT into items (" + props + ") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

	_, err := common.DB.Exec(query,
		charID, v.invID, v.id, v.slotID, v.amount, v.flag, v.upgradeSlots, v.scrollLevel,
		v.str, v.dex, v.intt, v.luk, v.hp, v.mp, v.watk, v.matk, v.wdef, v.mdef, v.accuracy, v.avoid, v.hands, v.speed, v.jump,
		v.expireTime, v.creatorName)

	if err != nil {
		return false, err
	}

	return true, nil
}

func loadEquipsFromDb(charID int32) []item {
	filter := "inventoryID,itemID,slotNumber,amount,flag,upgradeSlots,level,str,dex,intt,luk,hp,mp,watk,matk,wdef,mdef,accuracy,avoid,hands,speed,jump,expireTime,creatorName"
	row, err := common.DB.Query("SELECT "+filter+" FROM items WHERE characterID=?", charID)

	if err != nil {
		panic(err)
	}

	equip := []item{}

	defer row.Close()

	for row.Next() {

		item := item{}

		row.Scan(&item.invID,
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

		equip = append(equip, item)
	}

	return equip
}
