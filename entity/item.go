package entity

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/google/uuid"
)

type item struct {
	uuid         uuid.UUID
	cash         bool
	invID        byte
	slotID       int16
	itemID       int32
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
}

func (v item) Clone() item {
	return v
}

func (v item) IsPet() bool {
	nxInfo, err := nx.GetItem(v.itemID)

	if err != nil {
		return false
	}

	return nxInfo.Pet
}

func (v item) PreventsShield() bool {
	return false
}

func (v *item) SetCreatorName(name string) {
	v.creatorName = name
}

func (v *item) SetSlotID(id int16) {
	v.slotID = id
}

func (v item) Amount() int16 {
	return v.amount
}

func (v *item) SetAmount(value int16) {
	v.amount = value
}

func (v item) IsStackable() bool {
	invID := v.itemID / 1e6
	bullet := v.itemID / 1e4

	if invID != 5.0 && // pet item
		invID != 1.0 && // equip
		bullet != 207 && // star/arrow etc
		v.amount <= constant.MaxItemStack {

		return true
	}

	return false
}

func randomStat(min, max int) int16 {
	if max-min == 0 {
		return int16(max)
	}
	rand.Seed(time.Now().Unix())
	return int16(rand.Intn(max-min) + min)
}

func CreateItemFromID(id int32, amount int16) (item, error) {
	newItem := item{uuid: uuid.New()}

	nxInfo, err := nx.GetItem(id)

	if err != nil {
		return item{}, fmt.Errorf("Unable to generate item of id: %v", id)
	}

	newItem.cash = nxInfo.Cash
	newItem.invID = byte(id / 1e6)
	newItem.itemID = id
	newItem.accuracy = randomStat(int(math.Floor(nxInfo.IncACC-(nxInfo.IncACC*0.1))), int(math.Ceil(nxInfo.IncACC*1.1)))
	newItem.avoid = randomStat(int(math.Floor(nxInfo.IncEVA-(nxInfo.IncEVA*0.1))), int(math.Ceil(nxInfo.IncEVA*1.1)))
	newItem.speed = randomStat(int(math.Floor(nxInfo.IncSpeed-(nxInfo.IncSpeed*0.1))), int(math.Ceil(nxInfo.IncSpeed*1.1)))

	newItem.matk = randomStat(int(math.Floor(nxInfo.IncMAD-(nxInfo.IncMAD*0.1))), int(math.Ceil(nxInfo.IncMAD*1.1)))
	newItem.mdef = randomStat(int(math.Floor(nxInfo.IncMDD-(nxInfo.IncMDD*0.1))), int(math.Ceil(nxInfo.IncMDD*1.1)))
	newItem.watk = randomStat(int(math.Floor(nxInfo.IncPAD-(nxInfo.IncPAD*0.1))), int(math.Ceil(nxInfo.IncPAD*1.1)))
	newItem.wdef = randomStat(int(math.Floor(nxInfo.IncPDD-(nxInfo.IncPDD*0.1))), int(math.Ceil(nxInfo.IncPDD*1.1)))

	newItem.str = nxInfo.IncSTR
	newItem.dex = nxInfo.IncDEX
	newItem.intt = nxInfo.IncINT
	newItem.luk = nxInfo.IncLUK

	newItem.attackSpeed = nxInfo.AttackSpeed
	newItem.reqLevel = nxInfo.ReqLevel
	newItem.upgradeSlots = nxInfo.Tuc

	if amount < 1 {
		amount = 1
	}

	newItem.amount = amount
	newItem.stand = byte(nxInfo.Stand)

	return newItem, nil
}

func itemIsRechargeable(itemID int32) bool {
	return (math.Floor(float64(itemID/10000)) == 207) // Taken from cliet
}