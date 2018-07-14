package inventory

import (
	"log"
	"math"

	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/nx"
)

func IsRechargeAble(itemID int32) bool {
	return (math.Floor(float64(itemID/10000)) == 207) // Taken from cliet
}

func IsStackable(itemID int32, ammount int16) bool {
	invID := math.Floor(float64(itemID) / 1e6)
	bullet := math.Floor(float64(itemID) / 1e4)

	if invID != 5 && // pet item
		invID != 1.0 && // equip
		bullet != 207 && // star/arrow etc
		ammount < constants.MAX_ITEM_STACK+1 {

		return true
	}

	return false
}

// Item -
type Item struct {
	invID        byte
	slotID       int16
	itemID       int32
	expireTime   uint64
	amount       int16
	creatorName  string
	flag         int16
	upgradeSlots byte
	level        byte
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

// TODO: Fill the rest out, for now this can be used to check functionality
func CreateFromID(id int32, isDrop bool) Item {
	newItem := Item{}

	nxInfo := nx.Items[id]

	newItem.SetInvID(byte(id / 1e6))
	newItem.SetItemID(id)
	newItem.SetAccuracy(nxInfo.Accuracy)
	newItem.SetAvoid(nxInfo.Evasion)

	newItem.SetMatk(nxInfo.MagicAttack)
	newItem.SetMdef(nxInfo.MagicDefence)
	newItem.SetWatk(nxInfo.WeaponAttack)
	newItem.SetWdef(nxInfo.WeaponDefence)

	newItem.SetStr(nxInfo.Str)
	newItem.SetDex(nxInfo.Dex)
	newItem.SetInt(nxInfo.Int)
	newItem.SetLuk(nxInfo.Luk)

	newItem.SetLevel(nxInfo.ReqLevel)
	newItem.SetUpgradeSlots(nxInfo.Upgrades)

	newItem.SetAmount(1)

	log.Println("Finish create item from ID function", newItem)

	return newItem
}

func (i *Item) GetInvID() byte {
	return i.invID
}

func (i *Item) SetInvID(val byte) {
	i.invID = val
}

func (i *Item) GetSlotID() int16 {
	return i.slotID
}

func (i *Item) SetSlotID(val int16) {
	i.slotID = val
}

func (i *Item) GetItemID() int32 {
	return i.itemID
}

func (i *Item) SetItemID(val int32) {
	i.itemID = val
}

func (i *Item) GetExpirationTime() uint64 {
	return i.expireTime
}

func (i *Item) SetExpirationTime(val uint64) {
	i.expireTime = val
}

func (i *Item) GetAmount() int16 {
	return i.amount
}

func (i *Item) SetAmount(val int16) {
	i.amount = val
}

func (i *Item) GetCreatorName() string {
	return i.creatorName
}

func (i *Item) SetCreatorName(val string) {
	i.creatorName = val
}

func (i *Item) GetFlag() int16 {
	return i.flag
}

func (i *Item) SetFlag(val int16) {
	i.flag = val
}

func (i *Item) GetUpgradeSlots() byte {
	return i.upgradeSlots
}

func (i *Item) SetUpgradeSlots(val byte) {
	i.upgradeSlots = val
}

func (i *Item) GetLevel() byte {
	return i.level
}

func (i *Item) SetLevel(val byte) {
	i.level = val
}

func (i *Item) GetStr() int16 {
	return i.str
}

func (i *Item) SetStr(val int16) {
	i.str = val
}

func (i *Item) GetDex() int16 {
	return i.dex
}

func (i *Item) SetDex(val int16) {
	i.dex = val
}

func (i *Item) GetInt() int16 {
	return i.intt
}

func (i *Item) SetInt(val int16) {
	i.intt = val
}

func (i *Item) GetLuk() int16 {
	return i.luk
}

func (i *Item) SetLuk(val int16) {
	i.luk = val
}

func (i *Item) GetHP() int16 {
	return i.hp
}

func (i *Item) SetHP(val int16) {
	i.hp = val
}

func (i *Item) GetMP() int16 {
	return i.mp
}

func (i *Item) SetMP(val int16) {
	i.mp = val
}

func (i *Item) GetWatk() int16 {
	return i.watk
}

func (i *Item) SetWatk(val int16) {
	i.watk = val
}

func (i *Item) GetMatk() int16 {
	return i.matk
}

func (i *Item) SetMatk(val int16) {
	i.matk = val
}

func (i *Item) GetWdef() int16 {
	return i.wdef
}

func (i *Item) SetWdef(val int16) {
	i.wdef = val
}

func (i *Item) GetMdef() int16 {
	return i.mdef
}

func (i *Item) SetMdef(val int16) {
	i.mdef = val
}

func (i *Item) SetAccuracy(val int16) {
	i.accuracy = val
}

func (i *Item) GetAccuracy() int16 {
	return i.accuracy
}

func (i *Item) GetAvoid() int16 {
	return i.avoid
}

func (i *Item) SetAvoid(val int16) {
	i.avoid = val
}

func (i *Item) GetHands() int16 {
	return i.hands
}

func (i *Item) SetHands(val int16) {
	i.hands = val
}

func (i *Item) GetSpeed() int16 {
	return i.speed
}

func (i *Item) SetSpeed(val int16) {
	i.speed = val
}

func (i *Item) GetJump() int16 {
	return i.jump
}

func (i *Item) SetJump(val int16) {
	i.jump = val
}
