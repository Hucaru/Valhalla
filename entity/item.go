package entity

import (
	"fmt"
	"math"

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
	expireTime   uint64
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
}

func itemIsRechargeable(itemID int32) bool {
	return (math.Floor(float64(itemID/10000)) == 207) // Taken from cliet
}

func itemIsStackable(itemID int32, ammount int16) bool {
	invID := itemID / 1e6
	bullet := itemID / 1e4

	if invID != 5 && // pet item
		invID != 1.0 && // equip
		bullet != 207 && // star/arrow etc
		ammount <= constant.MaxItemStack {

		return true
	}

	return false
}

// TODO: Fill the rest out, for now this can be used to check functionality
func createItemFromID(id int32) (item, error) {
	newItem := item{}

	nxInfo, err := nx.GetItem(id)

	if err != nil {
		return item{}, fmt.Errorf("Unable to generate item of id: %v", id)
	}

	newItem.uuid = uuid.Must(uuid.NewRandom())
	newItem.cash = nxInfo.Cash
	newItem.invID = byte(id / 1e6)
	newItem.itemID = id
	newItem.accuracy = nxInfo.IncACC
	newItem.avoid = nxInfo.IncEVA

	newItem.matk = nxInfo.IncMAD
	newItem.mdef = nxInfo.IncMDD
	newItem.watk = nxInfo.IncPAD
	newItem.wdef = nxInfo.IncPDD

	newItem.str = nxInfo.IncSTR
	newItem.dex = nxInfo.IncDEX
	newItem.intt = nxInfo.IncINT
	newItem.luk = nxInfo.IncLUK

	newItem.reqLevel = nxInfo.ReqLevel
	newItem.upgradeSlots = nxInfo.Tuc

	newItem.amount = 1

	return newItem, nil
}
