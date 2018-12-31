package def

import (
	"fmt"
	"log"
	"math"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/google/uuid"
)

type Item struct {
	UUID         uuid.UUID
	Cash         bool
	InvID        byte
	SlotID       int16
	ItemID       int32
	ExpireTime   uint64
	Amount       int16
	CreatorName  string
	Flag         int16
	UpgradeSlots byte
	ReqLevel     byte
	ScrollLevel  byte
	Str          int16
	Dex          int16
	Int          int16
	Luk          int16
	ReqStr       int16
	ReqDex       int16
	ReqInt       int16
	ReqLuk       int16
	HP           int16
	MP           int16
	Watk         int16
	Matk         int16
	Wdef         int16
	Mdef         int16
	Accuracy     int16
	Avoid        int16
	Hands        int16
	Speed        int16
	Jump         int16
}

func ItemIsRechargeable(itemID int32) bool {
	return (math.Floor(float64(itemID/10000)) == 207) // Taken from cliet
}

func ItemIsStackable(itemID int32, ammount int16) bool {
	invID := itemID / 1e6
	bullet := itemID / 1e4

	if invID != 5 && // pet item
		invID != 1.0 && // equip
		bullet != 207 && // star/arrow etc
		ammount <= constant.MAX_ITEM_STACK {

		return true
	}

	return false
}

// TODO: Fill the rest out, for now this can be used to check functionality
func CreateFromID(id int32) (Item, bool) {
	newItem := Item{}

	nxInfo, err := nx.GetItem(id)

	if err != nil {
		fmt.Println("Unable to generate item of id:", id)
		return Item{}, false
	}

	newItem.UUID = uuid.Must(uuid.NewRandom())
	newItem.Cash = nxInfo.Cash
	newItem.InvID = byte(id / 1e6)
	newItem.ItemID = id
	newItem.Accuracy = nxInfo.IncACC
	newItem.Avoid = nxInfo.IncEVA

	newItem.Matk = nxInfo.IncMAD
	newItem.Mdef = nxInfo.IncMDD
	newItem.Watk = nxInfo.IncPAD
	newItem.Wdef = nxInfo.IncPDD

	newItem.Str = nxInfo.IncSTR
	newItem.Dex = nxInfo.IncDEX
	newItem.Int = nxInfo.IncINT
	newItem.Luk = nxInfo.IncLUK

	newItem.ReqLevel = nxInfo.ReqLevel
	newItem.UpgradeSlots = nxInfo.Tuc

	newItem.Amount = 1

	log.Println("Finish create item from ID function", newItem)

	return newItem, true
}
