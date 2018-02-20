package character

import (
	"sync"
)

type Character struct {
	charID          uint32
	userID          uint32
	worldID         uint32
	name            string
	gender          byte
	skin            byte
	face            uint32
	hair            uint32
	level           byte
	job             uint16
	str             uint16
	dex             uint16
	intt            uint16
	luk             uint16
	hp              uint16
	maxHP           uint16
	mp              uint16
	maxMP           uint16
	ap              uint16
	sp              uint16
	exp             uint32
	fame            uint16
	currentMap      uint32
	currentMapPos   byte
	previousMap     uint32
	feeMarketReturn uint32
	mesos           uint32
	equipSlotSize   byte
	useSlotSize     byte
	setupSlotSize   byte
	etcSlotSize     byte
	cashSlotSize    byte

	equips []Equip
	skills []Skill
	items  []Item

	x       int16
	y       int16
	fh      uint16
	state   byte
	chairID uint32

	mutex *sync.RWMutex
}

func (c *Character) GetEquips() []Equip {
	c.mutex.RLock()
	val := c.equips
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetEquips(val []Equip) {
	c.mutex.Lock()
	c.equips = val
	c.mutex.Unlock()
}

func (c *Character) AddEquip(val Equip) {
	c.mutex.Lock()
	c.equips = append(c.equips, val)
	c.mutex.Unlock()
}

func (c *Character) GetSkills() []Skill {
	c.mutex.RLock()
	val := c.skills
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetSkills(val []Skill) {
	c.mutex.Lock()
	c.skills = val
	c.mutex.Unlock()
}

func (c *Character) AddSkill(val Skill) {
	c.mutex.Lock()
	c.skills = append(c.skills, val)
	c.mutex.Unlock()
}

func (c *Character) GetItems() []Item {
	c.mutex.RLock()
	val := c.items
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetItems(val []Item) {
	c.mutex.Lock()
	c.items = val
	c.mutex.Unlock()
}

func (c *Character) AddItem(val Item) {
	c.mutex.Lock()
	c.items = append(c.items, val)
	c.mutex.Unlock()
}

func (c *Character) GetCharID() uint32 {
	c.mutex.RLock()
	val := c.charID
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetCharID(val uint32) {
	c.mutex.Lock()
	c.charID = val
	c.mutex.Unlock()
}

func (c *Character) GetUserID() uint32 {
	c.mutex.RLock()
	val := c.userID
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetUserID(val uint32) {
	c.mutex.Lock()
	c.userID = val
	c.mutex.Unlock()
}

func (c *Character) GetWorldID() uint32 {
	c.mutex.RLock()
	val := c.worldID
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetWorldID(val uint32) {
	c.mutex.Lock()
	c.worldID = val
	c.mutex.Unlock()
}

func (c *Character) GetName() string {
	c.mutex.RLock()
	val := c.name
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetName(val string) {
	c.mutex.Lock()
	c.name = val
	c.mutex.Unlock()
}

func (c *Character) GetGender() byte {
	c.mutex.RLock()
	val := c.gender
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetGender(val byte) {
	c.mutex.Lock()
	c.gender = val
	c.mutex.Unlock()
}

func (c *Character) GetSkin() byte {
	c.mutex.RLock()
	val := c.skin
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetSkin(val byte) {
	c.mutex.Lock()
	c.skin = val
	c.mutex.Unlock()
}

func (c *Character) GetFace() uint32 {
	c.mutex.RLock()
	val := c.face
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetFace(val uint32) {
	c.mutex.Lock()
	c.face = val
	c.mutex.Unlock()
}

func (c *Character) GetHair() uint32 {
	c.mutex.RLock()
	val := c.hair
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetHair(val uint32) {
	c.mutex.Lock()
	c.hair = val
	c.mutex.Unlock()
}

func (c *Character) GetLevel() byte {
	c.mutex.RLock()
	val := c.level
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetLevel(val byte) {
	c.mutex.Lock()
	c.level = val
	c.mutex.Unlock()
}

func (c *Character) GetJob() uint16 {
	c.mutex.RLock()
	val := c.job
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetJob(val uint16) {
	c.mutex.Lock()
	c.job = val
	c.mutex.Unlock()
}

func (c *Character) GetStr() uint16 {
	c.mutex.RLock()
	val := c.str
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetStr(val uint16) {
	c.mutex.Lock()
	c.str = val
	c.mutex.Unlock()
}

func (c *Character) GetDex() uint16 {
	c.mutex.RLock()
	val := c.dex
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetDex(val uint16) {
	c.mutex.Lock()
	c.dex = val
	c.mutex.Unlock()
}

func (c *Character) GetInt() uint16 {
	c.mutex.RLock()
	val := c.intt
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetInt(val uint16) {
	c.mutex.Lock()
	c.intt = val
	c.mutex.Unlock()
}

func (c *Character) GetLuk() uint16 {
	c.mutex.RLock()
	val := c.luk
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetLuk(val uint16) {
	c.mutex.Lock()
	c.luk = val
	c.mutex.Unlock()
}

func (c *Character) GetHP() uint16 {
	c.mutex.RLock()
	val := c.hp
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetHP(val uint16) {
	c.mutex.Lock()
	c.hp = val
	c.mutex.Unlock()
}

func (c *Character) GetMaxHP() uint16 {
	c.mutex.RLock()
	val := c.maxHP
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetMaxHP(val uint16) {
	c.mutex.Lock()
	c.maxHP = val
	c.mutex.Unlock()
}

func (c *Character) GetMP() uint16 {
	c.mutex.RLock()
	val := c.mp
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetMP(val uint16) {
	c.mutex.Lock()
	c.mp = val
	c.mutex.Unlock()
}

func (c *Character) GetMaxMP() uint16 {
	c.mutex.RLock()
	val := c.maxMP
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetMaxMp(val uint16) {
	c.mutex.Lock()
	c.maxMP = val
	c.mutex.Unlock()
}

func (c *Character) GetAP() uint16 {
	c.mutex.RLock()
	val := c.ap
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetAP(val uint16) {
	c.mutex.Lock()
	c.ap = val
	c.mutex.Unlock()
}
func (c *Character) GetSP() uint16 {
	c.mutex.RLock()
	val := c.sp
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetSP(val uint16) {
	c.mutex.Lock()
	c.sp = val
	c.mutex.Unlock()
}

func (c *Character) GetEXP() uint32 {
	c.mutex.RLock()
	val := c.exp
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetEXP(val uint32) {
	c.mutex.Lock()
	c.exp = val
	c.mutex.Unlock()
}

func (c *Character) GetFame() uint16 {
	c.mutex.RLock()
	val := c.fame
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetFame(val uint16) {
	c.mutex.Lock()
	c.fame = val
	c.mutex.Unlock()
}

func (c *Character) GetCurrentMap() uint32 {
	c.mutex.RLock()
	val := c.currentMap
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetCurrentMap(val uint32) {
	c.mutex.Lock()
	c.currentMap = val
	c.mutex.Unlock()
}

func (c *Character) GetCurrentMapPos() byte {
	c.mutex.RLock()
	val := c.currentMapPos
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetCurrentMapPos(val byte) {
	c.mutex.Lock()
	c.currentMapPos = val
	c.mutex.Unlock()
}

func (c *Character) GetPreviousMap() uint32 {
	c.mutex.RLock()
	val := c.previousMap
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetPreviousMap(val uint32) {
	c.mutex.Lock()
	c.previousMap = val
	c.mutex.Unlock()
}

func (c *Character) GetFeeMarketReturn() uint32 {
	c.mutex.RLock()
	val := c.feeMarketReturn
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetFreeMarketReturn(val uint32) {
	c.mutex.Lock()
	c.feeMarketReturn = val
	c.mutex.Unlock()
}

func (c *Character) GetMesos() uint32 {
	c.mutex.RLock()
	val := c.mesos
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetMesos(val uint32) {
	c.mutex.Lock()
	c.mesos = val
	c.mutex.Unlock()
}

func (c *Character) GetEquipSlotSize() byte {
	c.mutex.RLock()
	val := c.equipSlotSize
	c.mutex.RUnlock()

	return val
}
func (c *Character) SetEquipSlotSize(val byte) {
	c.mutex.Lock()
	c.equipSlotSize = val
	c.mutex.Unlock()
}

func (c *Character) GetUsetSlotSize() byte {
	c.mutex.RLock()
	val := c.useSlotSize
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetUseSlotSize(val byte) {
	c.mutex.Lock()
	c.useSlotSize = val
	c.mutex.Unlock()
}

func (c *Character) GetSetupSlotSize() byte {
	c.mutex.RLock()
	val := c.setupSlotSize
	c.mutex.RUnlock()

	return val
}
func (c *Character) SetSetupSlotSize(val byte) {
	c.mutex.Lock()
	c.setupSlotSize = val
	c.mutex.Unlock()
}

func (c *Character) GetEtcSlotSize() byte {
	c.mutex.RLock()
	val := c.etcSlotSize
	c.mutex.RUnlock()

	return val
}
func (c *Character) SetEtcSlotSize(val byte) {
	c.mutex.Lock()
	c.etcSlotSize = val
	c.mutex.Unlock()
}

func (c *Character) GetCashSlotSize() byte {
	c.mutex.RLock()
	val := c.cashSlotSize
	c.mutex.RUnlock()

	return val
}
func (c *Character) SetCashSlotSize(val byte) {
	c.mutex.Lock()
	c.cashSlotSize = val
	c.mutex.Unlock()
}

func (c *Character) GetX() int16 {
	c.mutex.RLock()
	val := c.x
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetX(val int16) {
	c.mutex.Lock()
	c.x = val
	c.mutex.Unlock()
}

func (c *Character) GetY() int16 {
	c.mutex.RLock()
	val := c.y
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetY(val int16) {
	c.mutex.Lock()
	c.y = val
	c.mutex.Unlock()
}

func (c *Character) GetFh() uint16 {
	c.mutex.RLock()
	val := c.fh
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetFh(val uint16) {
	c.mutex.Lock()
	c.fh = val
	c.mutex.Unlock()
}

func (c *Character) GetState() byte {
	c.mutex.RLock()
	val := c.state
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetState(val byte) {
	c.mutex.Lock()
	c.state = val
	c.mutex.Unlock()
}

func (c *Character) GetChairID() uint32 {
	c.mutex.RLock()
	val := c.chairID
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetChairID(val uint32) {
	c.mutex.Lock()
	c.chairID = val
	c.mutex.Unlock()
}
