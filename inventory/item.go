package inventory

// Base -
type Base struct {
	invID       byte
	slotNumber  int16
	itemID      uint32
	expireTime  uint64
	amount      uint16
	creatorName string
	flag        uint16
	isMesos     bool
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

func (b *Base) GetItemID() uint32 {
	return b.itemID
}

func (b *Base) SetItemID(val uint32) {
	b.itemID = val
}

func (b *Base) GetExpirationTime() uint64 {
	return b.expireTime
}

func (b *Base) SetExpirationTime(val uint64) {
	b.expireTime = val
}

func (b *Base) GetAmount() uint16 {
	return b.amount
}

func (b *Base) SetAmount(val uint16) {
	b.amount = val
}

func (b *Base) GetCreatorName() string {
	return b.creatorName
}

func (b *Base) SetCreatorName(val string) {
	b.creatorName = val
}

func (b *Base) GetFlag() uint16 {
	return b.flag
}

func (b *Base) SetFlag(val uint16) {
	b.flag = val
}

func (b *Base) GetIsMesos() bool {
	return b.isMesos
}

func (b *Base) SetIsMesos(isMesos bool) {
	b.isMesos = isMesos
}

// Item -
type Item struct {
	Base
	upgradeSlots byte
	level        byte
	str          uint16
	dex          uint16
	intt         uint16
	luk          uint16
	hp           uint16
	mp           uint16
	watk         uint16
	matk         uint16
	wdef         uint16
	mdef         uint16
	accuracy     uint16
	avoid        uint16
	hands        uint16
	speed        uint16
	jump         uint16
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

func (e *Item) GetStr() uint16 {
	return e.str
}

func (e *Item) SetStr(val uint16) {
	e.str = val
}

func (e *Item) GetDex() uint16 {
	return e.dex
}

func (e *Item) SetDex(val uint16) {
	e.dex = val
}

func (e *Item) GetInt() uint16 {
	return e.intt
}

func (e *Item) SetInt(val uint16) {
	e.intt = val
}

func (e *Item) GetLuk() uint16 {
	return e.luk
}

func (e *Item) SetLuk(val uint16) {
	e.luk = val
}

func (e *Item) GetHP() uint16 {
	return e.hp
}

func (e *Item) SetHP(val uint16) {
	e.hp = val
}

func (e *Item) GetMP() uint16 {
	return e.mp
}

func (e *Item) SetMP(val uint16) {
	e.mp = val
}

func (e *Item) GetWatk() uint16 {
	return e.watk
}

func (e *Item) SetWatk(val uint16) {
	e.watk = val
}

func (e *Item) GetMatk() uint16 {
	return e.matk
}

func (e *Item) SetMatk(val uint16) {
	e.matk = val
}

func (e *Item) GetWdef() uint16 {
	return e.wdef
}

func (e *Item) SetWdef(val uint16) {
	e.wdef = val
}

func (e *Item) GetMdef() uint16 {
	return e.mdef
}

func (e *Item) SetMdef(val uint16) {
	e.mdef = val
}

func (e *Item) SetAccuracy(val uint16) {
	e.accuracy = val
}

func (e *Item) GetAccuracy() uint16 {
	return e.accuracy
}

func (e *Item) GetAvoid() uint16 {
	return e.avoid
}

func (e *Item) SetAvoid(val uint16) {
	e.avoid = val
}

func (e *Item) GetHands() uint16 {
	return e.hands
}

func (e *Item) SetHands(val uint16) {
	e.hands = val
}

func (e *Item) GetSpeed() uint16 {
	return e.speed
}

func (e *Item) SetSpeed(val uint16) {
	e.speed = val
}

func (e *Item) GetJump() uint16 {
	return e.jump
}

func (e *Item) SetJump(val uint16) {
	e.jump = val
}
