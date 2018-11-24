package def

import (
	"log"
	"math"

	"github.com/Hucaru/Valhalla/consts"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/google/uuid"
)

type Item struct {
	UUID         uuid.UUID
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
		ammount <= consts.MAX_ITEM_STACK {

		return true
	}

	return false
}

// TODO: Fill the rest out, for now this can be used to check functionality
func CreateFromID(id int32) (Item, bool) {
	newItem := Item{}

	if _, ok := nx.Items[id]; !ok {
		return Item{}, false
	}

	nxInfo := nx.Items[id]

	newItem.UUID = uuid.Must(uuid.NewRandom())
	newItem.InvID = byte(id / 1e6)
	newItem.ItemID = id
	newItem.Accuracy = nxInfo.Accuracy
	newItem.Avoid = nxInfo.Evasion

	newItem.Matk = nxInfo.MagicAttack
	newItem.Mdef = nxInfo.MagicDefence
	newItem.Watk = nxInfo.WeaponAttack
	newItem.Wdef = nxInfo.WeaponDefence

	newItem.Str = nxInfo.Str
	newItem.Dex = nxInfo.Dex
	newItem.Int = nxInfo.Int
	newItem.Luk = nxInfo.Luk

	newItem.ReqLevel = nxInfo.ReqLevel
	newItem.UpgradeSlots = nxInfo.Upgrades

	newItem.Amount = 1

	log.Println("Finish create item from ID function", newItem)

	return newItem, true
}
