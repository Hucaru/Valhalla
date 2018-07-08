package character

import (
	"log"
)

// Base -
type Base struct {
	invID       byte
	slotNumber  int16
	itemID      int32
	expireTime  uint64
	amount      int16
	creatorName string
	flag        int16
}

func (b *Base) GetInvID() byte {
	return b.invID
}

func (b *Base) SetInvID(val byte) {
	b.invID = val
}

func (b *Base) GetSlotNumber() int16 {
	return b.slotNumber
}

func (b *Base) SetSlotNumber(val int16) {
	b.slotNumber = val
}

func (b *Base) GetItemID() int32 {
	return b.itemID
}

func (b *Base) SetItemID(val int32) {
	b.itemID = val
}

func (b *Base) GetExpirationTime() uint64 {
	return b.expireTime
}

func (b *Base) SetExpirationTime(val uint64) {
	b.expireTime = val
}

func (b *Base) GetAmount() int16 {
	return b.amount
}

func (b *Base) SetAmount(val int16) {
	b.amount = val
}

func (b *Base) GetCreatorName() string {
	return b.creatorName
}

func (b *Base) SetCreatorName(val string) {
	b.creatorName = val
}

func (b *Base) GetFlag() int16 {
	return b.flag
}

func (b *Base) SetFlag(val int16) {
	b.flag = val
}

// Item -
type Item struct {
	Base
	upgradeSlots byte
	level        byte
	str          int16
	dex          int16
	intt         int16
	luk          int16
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

func CreateItemFromID(id int32, isDrop bool) Item {
	log.Println("Implement create item from id:", id)
	return Item{}
}

func (e *Item) GetUpgradeSlots() byte {
	return e.upgradeSlots
}

func (e *Item) SetUpgradeSlots(val byte) {
	e.upgradeSlots = val
}

func (e *Item) GetLevel() byte {
	return e.level
}

func (e *Item) SetLevel(val byte) {
	e.level = val
}

func (e *Item) GetStr() int16 {
	return e.str
}

func (e *Item) SetStr(val int16) {
	e.str = val
}

func (e *Item) GetDex() int16 {
	return e.dex
}

func (e *Item) SetDex(val int16) {
	e.dex = val
}

func (e *Item) GetInt() int16 {
	return e.intt
}

func (e *Item) SetInt(val int16) {
	e.intt = val
}

func (e *Item) GetLuk() int16 {
	return e.luk
}

func (e *Item) SetLuk(val int16) {
	e.luk = val
}

func (e *Item) GetHP() int16 {
	return e.hp
}

func (e *Item) SetHP(val int16) {
	e.hp = val
}

func (e *Item) GetMP() int16 {
	return e.mp
}

func (e *Item) SetMP(val int16) {
	e.mp = val
}

func (e *Item) GetWatk() int16 {
	return e.watk
}

func (e *Item) SetWatk(val int16) {
	e.watk = val
}

func (e *Item) GetMatk() int16 {
	return e.matk
}

func (e *Item) SetMatk(val int16) {
	e.matk = val
}

func (e *Item) GetWdef() int16 {
	return e.wdef
}

func (e *Item) SetWdef(val int16) {
	e.wdef = val
}

func (e *Item) GetMdef() int16 {
	return e.mdef
}

func (e *Item) SetMdef(val int16) {
	e.mdef = val
}

func (e *Item) SetAccuracy(val int16) {
	e.accuracy = val
}

func (e *Item) GetAccuracy() int16 {
	return e.accuracy
}

func (e *Item) GetAvoid() int16 {
	return e.avoid
}

func (e *Item) SetAvoid(val int16) {
	e.avoid = val
}

func (e *Item) GetHands() int16 {
	return e.hands
}

func (e *Item) SetHands(val int16) {
	e.hands = val
}

func (e *Item) GetSpeed() int16 {
	return e.speed
}

func (e *Item) SetSpeed(val int16) {
	e.speed = val
}

func (e *Item) GetJump() int16 {
	return e.jump
}

func (e *Item) SetJump(val int16) {
	e.jump = val
}
