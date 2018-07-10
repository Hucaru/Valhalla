package character

import (
	"sync"

	"github.com/Hucaru/Valhalla/inventory"
)

type Character struct {
	charID          int32
	userID          int32
	worldID         int32
	name            string
	gender          byte
	skin            byte
	face            int32
	hair            int32
	level           byte
	job             int16
	str             int16
	dex             int16
	intt            int16
	luk             int16
	hp              int16
	maxHP           int16
	mp              int16
	maxMP           int16
	ap              int16
	sp              int16
	exp             int32
	fame            int16
	currentMap      int32
	currentMapPos   byte
	previousMap     int32
	feeMarketReturn int32
	mesos           int32
	equipSlotSize   byte
	useSlotSize     byte
	setupSlotSize   byte
	etcSlotSize     byte
	cashSlotSize    byte

	items []inventory.Item

	skills map[int32]int32

	x        int16
	y        int16
	foothold int16
	state    byte
	chairID  int32

	mutex *sync.RWMutex // Is this needed anymore? Player character information access is guarded by mutex already
}

func (c *Character) GetSkills() map[int32]int32 {
	c.mutex.RLock()
	val := c.skills
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetSkills(val map[int32]int32) {
	c.mutex.Lock()
	c.skills = val
	c.mutex.Unlock()
}

func (c *Character) UpdateSkill(id, level int32) {
	c.mutex.Lock()
	c.skills[id] = level
	c.mutex.Unlock()
}

func (c *Character) GetItems() []inventory.Item {
	c.mutex.RLock()
	val := c.items
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetItems(val []inventory.Item) {
	c.mutex.Lock()
	c.items = val
	c.mutex.Unlock()
}

func (c *Character) AddItem(val inventory.Item) {
	c.mutex.Lock()
	c.items = append(c.items, val)
	c.mutex.Unlock()
}

func (c *Character) RemoveItem(val inventory.Item) {
	var index int = 0

	c.mutex.RLock()
	for ind, i := range c.items {
		if i.GetSlotID() == val.GetSlotID() &&
			i.GetItemID() == val.GetItemID() &&
			i.GetInvID() == val.GetInvID() {

			index = ind
		}
	}
	c.mutex.RUnlock()

	c.mutex.Lock()
	c.items = append(c.items[:index], c.items[index+1:]...)
	c.mutex.Unlock()
}

func (c *Character) UpdateItem(orig inventory.Item, new inventory.Item) {
	c.mutex.Lock()
	for index, i := range c.items {
		if i.GetSlotID() == orig.GetSlotID() &&
			i.GetInvID() == orig.GetInvID() {
			c.items[index] = new
		}
	}
	c.mutex.Unlock()
}

func (c *Character) SwitchItems(orig inventory.Item, new inventory.Item) {
	var ind1 int = -1
	var slot1 int16 = -1

	var ind2 int = -1
	var slot2 int16 = -1

	c.mutex.RLock()
	for index, i := range c.items {
		if i == orig {
			ind1 = index
			slot1 = i.GetSlotID()
		}

		if i == new {
			ind2 = index
			slot2 = i.GetSlotID()
		}
	}
	c.mutex.RUnlock()

	if ind1 > -1 {
		c.mutex.Lock()
		c.items[ind1].SetSlotID(slot2)
		c.mutex.Unlock()
	}

	if ind2 > -1 {
		c.mutex.Lock()
		c.items[ind2].SetSlotID(slot1)
		c.mutex.Unlock()
	}
}

func (c *Character) GetCharID() int32 {
	c.mutex.RLock()
	val := c.charID
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetCharID(val int32) {
	c.mutex.Lock()
	c.charID = val
	c.mutex.Unlock()
}

func (c *Character) GetUserID() int32 {
	c.mutex.RLock()
	val := c.userID
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetUserID(val int32) {
	c.mutex.Lock()
	c.userID = val
	c.mutex.Unlock()
}

func (c *Character) GetWorldID() int32 {
	c.mutex.RLock()
	val := c.worldID
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetWorldID(val int32) {
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

func (c *Character) GetFace() int32 {
	c.mutex.RLock()
	val := c.face
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetFace(val int32) {
	c.mutex.Lock()
	c.face = val
	c.mutex.Unlock()
}

func (c *Character) GetHair() int32 {
	c.mutex.RLock()
	val := c.hair
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetHair(val int32) {
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

func (c *Character) GetJob() int16 {
	c.mutex.RLock()
	val := c.job
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetJob(val int16) {
	c.mutex.Lock()
	c.job = val
	c.mutex.Unlock()
}

func (c *Character) GetStr() int16 {
	c.mutex.RLock()
	val := c.str
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetStr(val int16) {
	c.mutex.Lock()
	c.str = val
	c.mutex.Unlock()
}

func (c *Character) GetDex() int16 {
	c.mutex.RLock()
	val := c.dex
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetDex(val int16) {
	c.mutex.Lock()
	c.dex = val
	c.mutex.Unlock()
}

func (c *Character) GetInt() int16 {
	c.mutex.RLock()
	val := c.intt
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetInt(val int16) {
	c.mutex.Lock()
	c.intt = val
	c.mutex.Unlock()
}

func (c *Character) GetLuk() int16 {
	c.mutex.RLock()
	val := c.luk
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetLuk(val int16) {
	c.mutex.Lock()
	c.luk = val
	c.mutex.Unlock()
}

func (c *Character) GetHP() int16 {
	c.mutex.RLock()
	val := c.hp
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetHP(val int16) {
	c.mutex.Lock()
	c.hp = val
	c.mutex.Unlock()
}

func (c *Character) GetMaxHP() int16 {
	c.mutex.RLock()
	val := c.maxHP
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetMaxHP(val int16) {
	c.mutex.Lock()
	c.maxHP = val
	c.mutex.Unlock()
}

func (c *Character) GetMP() int16 {
	c.mutex.RLock()
	val := c.mp
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetMP(val int16) {
	c.mutex.Lock()
	c.mp = val
	c.mutex.Unlock()
}

func (c *Character) GetMaxMP() int16 {
	c.mutex.RLock()
	val := c.maxMP
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetMaxMP(val int16) {
	c.mutex.Lock()
	c.maxMP = val
	c.mutex.Unlock()
}

func (c *Character) GetAP() int16 {
	c.mutex.RLock()
	val := c.ap
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetAP(val int16) {
	c.mutex.Lock()
	c.ap = val
	c.mutex.Unlock()
}
func (c *Character) GetSP() int16 {
	c.mutex.RLock()
	val := c.sp
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetSP(val int16) {
	c.mutex.Lock()
	c.sp = val
	c.mutex.Unlock()
}

func (c *Character) GetEXP() int32 {
	c.mutex.RLock()
	val := c.exp
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetEXP(val int32) {
	c.mutex.Lock()
	c.exp = val
	c.mutex.Unlock()
}

func (c *Character) GetFame() int16 {
	c.mutex.RLock()
	val := c.fame
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetFame(val int16) {
	c.mutex.Lock()
	c.fame = val
	c.mutex.Unlock()
}

func (c *Character) GetCurrentMap() int32 {
	c.mutex.RLock()
	val := c.currentMap
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetCurrentMap(val int32) {
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

func (c *Character) GetPreviousMap() int32 {
	c.mutex.RLock()
	val := c.previousMap
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetPreviousMap(val int32) {
	c.mutex.Lock()
	c.previousMap = val
	c.mutex.Unlock()
}

func (c *Character) GetFeeMarketReturn() int32 {
	c.mutex.RLock()
	val := c.feeMarketReturn
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetFreeMarketReturn(val int32) {
	c.mutex.Lock()
	c.feeMarketReturn = val
	c.mutex.Unlock()
}

func (c *Character) GetMesos() int32 {
	c.mutex.RLock()
	val := c.mesos
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetMesos(val int32) {
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

func (c *Character) GetFoothold() int16 {
	c.mutex.RLock()
	val := c.foothold
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetFoothold(val int16) {
	c.mutex.Lock()
	c.foothold = val
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

func (c *Character) GetChairID() int32 {
	c.mutex.RLock()
	val := c.chairID
	c.mutex.RUnlock()

	return val
}

func (c *Character) SetChairID(val int32) {
	c.mutex.Lock()
	c.chairID = val
	c.mutex.Unlock()
}
