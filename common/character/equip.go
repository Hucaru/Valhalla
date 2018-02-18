package character

type Equip struct {
	itemID       uint32
	slotID       int32
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
	expireTime   uint64
	creatorName  string
}

func (e *Equip) GetItemID() uint32 {
	return e.itemID
}

func (e *Equip) SetItemID(val uint32) {
	e.itemID = val
}

func (e *Equip) GetSlotID() int32 {
	return e.slotID
}

func (e *Equip) SetSlotID(val int32) {
	e.slotID = val
}

func (e *Equip) GetUpgradeSlots() byte {
	return e.upgradeSlots
}

func (e *Equip) SetUpgradeSlots(val byte) {
	e.upgradeSlots = val
}

func (e *Equip) GetLevel() byte {
	return e.level
}

func (e *Equip) SetLevel(val byte) {
	e.level = val
}

func (e *Equip) GetStr() uint16 {
	return e.str
}

func (e *Equip) SetStr(val uint16) {
	e.str = val
}

func (e *Equip) GetDex() uint16 {
	return e.dex
}

func (e *Equip) SetDex(val uint16) {
	e.dex = val
}

func (e *Equip) GetInt() uint16 {
	return e.intt
}

func (e *Equip) SetInt(val uint16) {
	e.intt = val
}

func (e *Equip) GetLuk() uint16 {
	return e.luk
}

func (e *Equip) SetLuk(val uint16) {
	e.luk = val
}

func (e *Equip) GetHP() uint16 {
	return e.hp
}

func (e *Equip) SetHP(val uint16) {
	e.hp = val
}

func (e *Equip) GetMP() uint16 {
	return e.mp
}

func (e *Equip) SetMP(val uint16) {
	e.mp = val
}

func (e *Equip) GetWatk() uint16 {
	return e.watk
}

func (e *Equip) SetWatk(val uint16) {
	e.watk = val
}

func (e *Equip) GetMatk() uint16 {
	return e.matk
}

func (e *Equip) SetMatk(val uint16) {
	e.matk = val
}

func (e *Equip) GetWdef() uint16 {
	return e.wdef
}

func (e *Equip) SetWdef(val uint16) {
	e.wdef = val
}

func (e *Equip) GetMdef() uint16 {
	return e.mdef
}

func (e *Equip) SetMdef(val uint16) {
	e.mdef = val
}

func (e *Equip) SetAccuracy(val uint16) {
	e.accuracy = val
}

func (e *Equip) GetAccuracy() uint16 {
	return e.accuracy
}

func (e *Equip) GetAvoid() uint16 {
	return e.avoid
}

func (e *Equip) SetAvoid(val uint16) {
	e.avoid = val
}

func (e *Equip) GetHands() uint16 {
	return e.hands
}

func (e *Equip) SetHands(val uint16) {
	e.hands = val
}

func (e *Equip) GetSpeed() uint16 {
	return e.speed
}

func (e *Equip) SetSpeed(val uint16) {
	e.speed = val
}

func (e *Equip) GetJump() uint16 {
	return e.jump
}

func (e *Equip) SetJump(val uint16) {
	e.jump = val
}

func (e *Equip) GetExpireTime() uint64 {
	return e.expireTime
}

func (e *Equip) SetExpireTime(val uint64) {
	e.expireTime = val
}

func (e *Equip) GetCreatorName() string {
	return e.creatorName
}

func (e *Equip) SetCreatorName(val string) {
	e.creatorName = val
}
