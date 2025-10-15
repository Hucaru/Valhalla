package channel

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/constant/skill"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

type buddy struct {
	id        int32
	name      string
	channelID int32
	status    byte  // 0 - online, 1 - buddy request, 2 - offline
	cashShop  int32 // > 0 means is in cash shop
}

type playerSkill struct {
	ID             int32
	Level, Mastery byte
	Cooldown       int16
	CooldownTime   int16
	TimeLastUsed   int64
}

func createPlayerSkillFromData(ID int32, level byte) (playerSkill, error) {
	skill, err := nx.GetPlayerSkill(ID)
	if err != nil {
		return playerSkill{}, fmt.Errorf("invalid skill ID %d (level %d): %w", ID, level, err)
	}
	if level == 0 || int(level) > len(skill) {
		return playerSkill{}, fmt.Errorf("invalid skill level %d for skill ID %d (max %d)", level, ID, len(skill))
	}

	return playerSkill{
		ID:           ID,
		Level:        level,
		Mastery:      byte(skill[level-1].Mastery),
		Cooldown:     0,
		CooldownTime: int16(skill[level-1].Time),
		TimeLastUsed: 0,
	}, nil
}

func getSkillsFromCharID(id int32) []playerSkill {
	skills := []playerSkill{}

	const filter = "skillID, level, cooldown"
	row, err := common.DB.Query("SELECT "+filter+" FROM skills WHERE characterID=?", id)
	if err != nil {
		log.Printf("getSkillsFromCharID: query failed for character %d: %v", id, err)
		return skills
	}
	defer row.Close()

	for row.Next() {
		var ps playerSkill
		if err := row.Scan(&ps.ID, &ps.Level, &ps.Cooldown); err != nil {
			log.Printf("getSkillsFromCharID: scan failed for character %d: %v", id, err)
			continue
		}

		skillData, err := nx.GetPlayerSkill(ps.ID)
		if err != nil {
			log.Printf("getSkillsFromCharID: missing nx data for skill %d: %v", ps.ID, err)
			continue
		}
		if ps.Level == 0 || int(ps.Level) > len(skillData) {
			log.Printf("getSkillsFromCharID: invalid level %d for skill %d (max %d), skipping", ps.Level, ps.ID, len(skillData))
			continue
		}

		ps.CooldownTime = int16(skillData[ps.Level-1].Time)
		skills = append(skills, ps)
	}

	if err := row.Err(); err != nil {
		log.Printf("getSkillsFromCharID: rows error for character %d: %v", id, err)
	}

	return skills
}

type updatePartyInfoFunc func(partyID, playerID, job, level, mapID int32, name string)

type Player struct {
	Conn mnet.Client
	inst *fieldInstance

	ID          int32 // Unique identifier of the character
	accountID   int32
	accountName string
	worldID     byte
	ChannelID   byte

	mapID       int32
	mapPos      byte
	previousMap int32
	portalCount byte

	job int16

	level byte
	str   int16
	dex   int16
	intt  int16
	luk   int16
	hp    int16
	maxHP int16
	mp    int16
	maxMP int16
	ap    int16
	sp    int16
	exp   int32
	fame  int16

	Name      string
	gender    byte
	skin      byte
	face      int32
	hair      int32
	chairID   int32
	petCashID int64
	stance    byte
	pos       pos

	equipSlotSize byte
	useSlotSize   byte
	setupSlotSize byte
	etcSlotSize   byte
	cashSlotSize  byte

	equip []Item
	use   []Item
	setUp []Item
	etc   []Item
	cash  []Item

	mesos       int32
	nx          int32
	maplepoints int32

	storageInventory *storage

	skills map[int32]playerSkill

	miniGameWins, miniGameDraw, miniGameLoss, miniGamePoints int32

	lastAttackPacketTime int64

	buddyListSize byte
	buddyList     []buddy

	party *party
	guild *guild

	UpdatePartyInfo updatePartyInfoFunc

	rates *rates

	buffs *CharacterBuffs

	quests quests

	summons *summonState
	pet     *pet

	// Per-Player RNG for deterministic randomness
	rng *rand.Rand

	// write-behind persistence
	dirty DirtyBits

	lastChairHeal time.Time
}

// Helper: mark dirty and schedule debounced save.
func (d *Player) MarkDirty(bits DirtyBits, debounce time.Duration) {
	d.dirty |= bits
	scheduleSave(d, debounce)
}

// Helper: clear dirty bits after successful flush (kept for future; saver currently doesn't feed back)
func (d *Player) clearDirty(bits DirtyBits) {
	d.dirty &^= bits
}

func (d *Player) FlushNow() {
	FlushNow(d)
}

// SeedRNGDeterministic seeds the per-Player RNG using stable identifiers so
// gain sequences are reproducible across restarts and processes.
func (d *Player) SeedRNGDeterministic() {
	// Compose as uint64 to avoid int64 constant overflow, then cast at runtime.
	const gamma uint64 = 0x9e3779b97f4a7c15
	seed64 := gamma ^
		(uint64(uint32(d.ID)) << 1) ^
		(uint64(uint32(d.accountID)) << 33) ^
		(uint64(d.worldID) << 52)

	seed := int64(seed64) // two's complement wrapping is fine for rand.Source
	d.rng = rand.New(rand.NewSource(seed))
}

// ensureRNG guarantees d.rng is initialized. If a deterministic seed has not
// been set yet, it will use a time-based seed (non-deterministic).
func (d *Player) ensureRNG() {
	if d.rng == nil {
		// Default to deterministic seeding for stability unless you want variability:
		d.SeedRNGDeterministic()
	}
}

func (d *Player) randIntn(n int) int {
	d.ensureRNG()
	return d.rng.Intn(n)
}

// levelUpGains returns (hpGain, mpGain) using per-Player RNG and job family.
// The random component uses a small range similar to legacy behavior.
// Tweak the constants to match your balance targets if needed.
func (d *Player) levelUpGains() (int16, int16) {
	r := int16(d.randIntn(3) + 1) // legacy-style variance 1..3

	mainClass := d.job / 100
	switch {
	case d.job == 0 || mainClass == 0: // Beginner and pre-advancement
		// Balanced but modest growth
		return r + 12, r + 10
	case mainClass == 1: // Warrior
		// High HP, low MP growth
		return r + 24, r + 4
	case mainClass == 2: // Magician
		// Low HP, high MP growth
		return r + 10, r + 22
	case mainClass == 3 || mainClass == 4: // Bowman / Thief
		// Moderate HP/MP growth
		return r + 20, r + 14
	default:
		// Fallback for any other jobs/classes
		return r + 16, r + 12
	}
}

// Send the Data a packet
func (d *Player) Send(packet mpacket.Packet) {
	if d == nil || d.Conn == nil {
		return
	}
	d.Conn.Send(packet)
}

func (d *Player) setJob(id int16) {
	d.job = id
	d.Conn.Send(packetPlayerStatChange(true, constant.JobID, int32(id)))
	d.MarkDirty(DirtyJob, 300*time.Millisecond)

	if d.party != nil {
		d.UpdatePartyInfo(d.party.ID, d.ID, int32(d.job), int32(d.level), d.mapID, d.Name)
	}
}

func (d *Player) levelUp() {
	d.giveAP(5)
	if d.level < 10 {
		d.giveSP(1)
	} else {
		d.giveSP(3)
	}

	// Use per-Player RNG and job-based helper for deterministic gains.
	hpGain, mpGain := d.levelUpGains()

	newMaxHP := d.maxHP + hpGain
	newMaxMP := d.maxMP + mpGain
	if newMaxHP < 1 {
		newMaxHP = 1
	}
	if newMaxMP < 0 {
		newMaxMP = 0
	}

	d.setMaxHP(newMaxHP)
	d.setMaxMP(newMaxMP)

	newHP := int16(math.Min(float64(newMaxHP), float64(d.hp+hpGain)))
	newMP := int16(math.Min(float64(newMaxMP), float64(d.mp+mpGain)))
	d.setHP(newHP)
	d.setMP(newMP)

	d.giveLevel(1)
}

func (d *Player) setEXP(amount int32) {
	if d.level >= 200 {
		d.exp = amount
		d.Send(packetPlayerStatChange(false, constant.ExpID, int32(amount)))
		d.MarkDirty(DirtyEXP, 800*time.Millisecond)
		return
	}

	for {
		if d.level >= 200 {
			d.exp = amount
			break
		}

		expForLevel := constant.ExpTable[d.level-1]
		remainder := amount - expForLevel
		if remainder >= 0 {
			d.levelUp()
			amount = remainder
		} else {
			d.exp = amount
			break
		}
	}

	d.Send(packetPlayerStatChange(false, constant.ExpID, d.exp))
	d.MarkDirty(DirtyEXP, 800*time.Millisecond)
}

func (d *Player) giveEXP(amount int32, fromMob, fromParty bool) {
	amount = int32(d.rates.exp * float32(amount))

	switch {
	case fromMob:
		d.Send(packetMessageExpGained(true, false, amount))
	case fromParty:
		d.Send(packetMessageExpGained(false, false, amount))
	default:
		d.Send(packetMessageExpGained(false, true, amount))
	}

	d.setEXP(d.exp + amount)
}

func (d *Player) GetAccountName() string {
	return d.accountName
}

func (d *Player) setLevel(amount byte) {
	d.level = amount
	d.Send(packetPlayerStatChange(false, constant.LevelID, int32(amount)))
	d.inst.send(packetPlayerLevelUpAnimation(d.ID))
	d.MarkDirty(DirtyLevel, 300*time.Millisecond)

	if d.party != nil {
		d.UpdatePartyInfo(d.party.ID, d.ID, int32(d.job), int32(d.level), d.mapID, d.Name)
	}
}

func (d *Player) giveLevel(amount byte) {
	d.setLevel(d.level + amount)
}

func (d *Player) setAP(amount int16) {
	d.ap = amount
	d.Send(packetPlayerStatChange(false, constant.ApID, int32(amount)))
	d.MarkDirty(DirtyAP, 300*time.Millisecond)
}

func (d *Player) giveAP(amount int16) {
	d.setAP(d.ap + amount)
}

func (d *Player) setSP(amount int16) {
	d.sp = amount
	d.Send(packetPlayerStatChange(false, constant.SpID, int32(amount)))
	d.MarkDirty(DirtySP, 300*time.Millisecond)
}

func (d *Player) giveSP(amount int16) {
	d.setSP(d.sp + amount)
}

func (d *Player) setStr(amount int16) {
	d.str = amount
	d.Send(packetPlayerStatChange(true, constant.StrID, int32(amount)))
	d.MarkDirty(DirtyStr, 500*time.Millisecond)
}

func (d *Player) giveStr(amount int16) {
	d.setStr(d.str + amount)
}

func (d *Player) setDex(amount int16) {
	d.dex = amount
	d.Send(packetPlayerStatChange(true, constant.DexID, int32(amount)))
	d.MarkDirty(DirtyDex, 500*time.Millisecond)
}

func (d *Player) giveDex(amount int16) {
	d.setDex(d.dex + amount)
}

func (d *Player) setInt(amount int16) {
	d.intt = amount
	d.Send(packetPlayerStatChange(true, constant.IntID, int32(amount)))
	d.MarkDirty(DirtyInt, 500*time.Millisecond)
}

func (d *Player) giveInt(amount int16) {
	d.setInt(d.intt + amount)
}

func (d *Player) setLuk(amount int16) {
	d.luk = amount
	d.Send(packetPlayerStatChange(true, constant.LukID, int32(amount)))
	d.MarkDirty(DirtyLuk, 500*time.Millisecond)
}

func (d *Player) giveLuk(amount int16) {
	d.setLuk(d.luk + amount)
}

func (d *Player) giveHP(amount int16) {
	newHP := int(d.hp) + int(amount)
	d.setHP(int16(newHP))
}

func (d *Player) setHP(amount int16) {
	if amount < 0 {
		amount = 0
	}
	if amount > constant.MaxHpValue {
		amount = constant.MaxHpValue
	}
	if amount > d.maxHP {
		amount = d.maxHP
	}
	d.hp = amount
	d.Send(packetPlayerStatChange(true, constant.HpID, int32(amount)))
	d.MarkDirty(DirtyHP, 500*time.Millisecond)
}

func (d *Player) setMaxHP(amount int16) {
	if amount > constant.MaxHpValue {
		amount = constant.MaxHpValue
	}
	d.maxHP = amount
	d.Send(packetPlayerStatChange(true, constant.MaxHpID, int32(amount)))
	d.MarkDirty(DirtyMaxHP, 500*time.Millisecond)
}

func (d *Player) giveMP(amount int16) {
	newMP := int(d.mp) + int(amount)
	d.setMP(int16(newMP))
}

func (d *Player) setMP(amount int16) {
	if amount < 0 {
		amount = 0
	}
	if amount > constant.MaxMpValue {
		amount = constant.MaxMpValue
	}
	if amount > d.maxMP {
		amount = d.maxMP
	}
	d.mp = amount
	d.Send(packetPlayerStatChange(true, constant.MpID, int32(amount)))
	d.MarkDirty(DirtyMP, 500*time.Millisecond)
}

func (d *Player) setMaxMP(amount int16) {
	if amount > constant.MaxMpValue {
		amount = constant.MaxMpValue
	}
	d.maxMP = amount
	d.Send(packetPlayerStatChange(true, constant.MaxMpID, int32(amount)))
	d.MarkDirty(DirtyMaxMP, 500*time.Millisecond)
}

func (d *Player) setFame(amount int16) {
	d.fame = amount
	d.Send(packetPlayerStatChange(true, constant.FameID, int32(amount)))

	_, err := common.DB.Exec("UPDATE characters SET fame=? WHERE ID=?", d.fame, d.ID)
	if err != nil {
		log.Printf("setFame: failed to save fame for character %d: %v", d.ID, err)
	}
}

func (d *Player) setMesos(amount int32) {
	d.mesos = amount
	d.Send(packetPlayerStatChange(true, constant.MesosID, amount))
	// write-behind instead of immediate DB write
	d.MarkDirty(DirtyMesos, 200*time.Millisecond)
}

func (d *Player) giveMesos(amount int32) {
	d.setMesos(d.mesos + amount)
}

func (d *Player) takeMesos(amount int32) {
	d.setMesos(d.mesos - amount)
}

func (d *Player) saveMesos() error {
	query := "UPDATE characters SET mesos=? WHERE accountID=? and Name=?"
	_, err := common.DB.Exec(query, d.mesos, d.accountID, d.Name)
	return err
}

func (d *Player) setHair(id int32) error {
	query := "UPDATE characters SET hair=? WHERE ID=?"
	_, err := common.DB.Exec(query, id, d.ID)
	d.hair = id
	d.Send(packetPlayerStatChange(true, constant.HairID, id))
	return err
}

func (d *Player) setFace(id int32) error {
	query := "UPDATE characters SET face=? WHERE ID=?"
	_, err := common.DB.Exec(query, id, d.ID)
	d.face = id
	d.Send(packetPlayerStatChange(true, constant.FaceID, id))
	return err
}

func (d *Player) setSkin(id byte) error {
	query := "UPDATE characters SET skin=? WHERE ID=?"
	_, err := common.DB.Exec(query, id, d.ID)
	d.skin = id
	d.Send(packetPlayerStatChange(true, constant.SkinID, int32(id)))
	return err
}

// UpdateMovement - update Data from position data
func (d *Player) UpdateMovement(frag movementFrag) {
	d.pos.x = frag.x
	d.pos.y = frag.y
	d.pos.foothold = frag.foothold
	d.stance = frag.stance
}

// SetPos of Data
func (d *Player) SetPos(pos pos) {
	d.pos = pos
}

// checks Data is within a certain range of a position
func (d Player) checkPos(pos pos, xRange, yRange int16) bool {
	var xValid, yValid bool

	if xRange == 0 {
		xValid = d.pos.x == pos.x
	} else {
		xValid = (pos.x-xRange < d.pos.x && d.pos.x < pos.x+xRange)
	}

	if yRange == 0 {
		xValid = d.pos.y == pos.y
	} else {
		yValid = (pos.y-yRange < d.pos.y && d.pos.y < pos.y+yRange)
	}

	return xValid && yValid
}

func isExcludedMap(id int32) bool {
	// Free Market range (inclusive)
	return id >= 910000000 && id <= 910000022
}

func (d *Player) setMapID(id int32) {
	// Never set previousMap to a FM ID
	if !isExcludedMap(d.mapID) {
		d.previousMap = d.mapID
	}

	d.mapID = id

	if d.party != nil {
		d.UpdatePartyInfo(d.party.ID, d.ID, int32(d.job), int32(d.level), d.mapID, d.Name)
	}

	// write-behind for mapID/pos (mapPos updated on save())
	d.MarkDirty(DirtyMap, 500*time.Millisecond)
	d.MarkDirty(DirtyPrevMap, 500*time.Millisecond)
}

func (d Player) noChange() {
	d.Send(packetInventoryNoChange())
}

func (d *Player) GetNX() int32 {
	return d.nx
}

func (d *Player) SetNX(nx int32) {
	d.nx = nx
	d.MarkDirty(DirtyNX, 300*time.Millisecond)
}

func (d *Player) GetMaplePoints() int32 {
	return d.maplepoints
}

func (d *Player) SetMaplePoints(points int32) {
	d.maplepoints = points
	d.MarkDirty(DirtyMaplePoints, 300*time.Millisecond)
}

// GiveItem grants the given item to a player and returns the item
func (d *Player) GiveItem(newItem Item) (error, Item) { // TODO: Refactor
	isRechargeable := func(itemID int32) bool {
		base := itemID / 10000
		return base == 207
	}

	newItem.dbID = 0

	findFirstEmptySlot := func(items []Item, size byte) (int16, error) {
		slotsUsed := make([]bool, size)
		for _, v := range items {
			if v.slotID > 0 {
				slotsUsed[v.slotID-1] = true
			}
		}
		slot := 0
		for i, v := range slotsUsed {
			if !v {
				slot = i + 1
				break
			}
		}
		if slot == 0 {
			slot = len(slotsUsed) + 1
		}
		if byte(slot) > size {
			return 0, fmt.Errorf("No empty Item slot left")
		}
		return int16(slot), nil
	}

	switch newItem.invID {
	case 1: // Equip
		slotID, err := findFirstEmptySlot(d.equip, d.equipSlotSize)
		if err != nil {
			return err, Item{}
		}
		newItem.slotID = slotID
		newItem.amount = 1
		newItem.save(d.ID)
		d.equip = append(d.equip, newItem)
		d.Send(packetInventoryAddItem(newItem, true))

	case 2: // Use
		if isRechargeable(newItem.ID) {
			slotID, err := findFirstEmptySlot(d.use, d.useSlotSize)
			if err != nil {
				return err, Item{}
			}
			newItem.slotID = slotID
			newItem.save(d.ID)
			d.use = append(d.use, newItem)
			d.Send(packetInventoryAddItem(newItem, true))
			return nil, newItem
		}

		// Non-rechargeable
		size := newItem.amount
		for size > 0 {
			var value int16 = 200
			value -= size
			if value < 1 {
				value = 200
			} else {
				value = size
			}
			size -= constant.MaxItemStack
			newItem.amount = value

			var slotID int16
			var index int
			for i, v := range d.use {
				if v.ID == newItem.ID && v.amount < constant.MaxItemStack {
					slotID = v.slotID
					index = i
					break
				}
			}

			if slotID == 0 {
				slotID, err := findFirstEmptySlot(d.use, d.useSlotSize)
				if err != nil {
					return err, Item{}
				}
				newItem.slotID = slotID
				newItem.save(d.ID)
				d.use = append(d.use, newItem)
				d.Send(packetInventoryAddItem(newItem, true))
			} else {
				remainder := newItem.amount - (constant.MaxItemStack - d.use[index].amount)
				if remainder > 0 { // partial merge -> place remainder to new slot
					slotID, err := findFirstEmptySlot(d.use, d.useSlotSize)
					if err != nil {
						return err, Item{}
					}
					newItem.amount = value
					newItem.slotID = slotID
					newItem.save(d.ID)
					d.use = append(d.use, newItem)
					d.Send(packetInventoryAddItems([]Item{d.use[index], newItem}, []bool{false, true}))
				} else { // full merge
					d.use[index].amount = d.use[index].amount + newItem.amount
					d.Send(packetInventoryAddItem(d.use[index], false))
					d.use[index].save(d.ID)
				}
			}
		}

	case 3: // Set-up
		slotID, err := findFirstEmptySlot(d.setUp, d.setupSlotSize)
		if err != nil {
			return err, Item{}
		}
		newItem.slotID = slotID
		newItem.save(d.ID)
		d.setUp = append(d.setUp, newItem)
		d.Send(packetInventoryAddItem(newItem, true))

	case 4: // Etc
		size := newItem.amount
		for size > 0 {
			var value int16 = 200
			value -= size
			if value < 1 {
				value = 200
			} else {
				value = size
			}
			size -= constant.MaxItemStack
			newItem.amount = value

			var slotID int16
			var index int
			for i, v := range d.etc {
				if v.ID == newItem.ID && v.amount < constant.MaxItemStack {
					slotID = v.slotID
					index = i
					break
				}
			}

			if slotID == 0 {
				slotID, err := findFirstEmptySlot(d.etc, d.etcSlotSize)
				if err != nil {
					return err, Item{}
				}
				newItem.slotID = slotID
				newItem.save(d.ID)
				d.etc = append(d.etc, newItem)
				d.Send(packetInventoryAddItem(newItem, true))
			} else {
				remainder := newItem.amount - (constant.MaxItemStack - d.etc[index].amount)
				if remainder > 0 {
					slotID, err := findFirstEmptySlot(d.etc, d.etcSlotSize)
					if err != nil {
						return err, Item{}
					}
					newItem.amount = value
					newItem.slotID = slotID
					newItem.save(d.ID)
					d.etc = append(d.etc, newItem)
					d.Send(packetInventoryAddItems([]Item{d.etc[index], newItem}, []bool{false, true}))
				} else {
					d.etc[index].amount = d.etc[index].amount + newItem.amount
					d.Send(packetInventoryAddItem(d.etc[index], false))
					d.etc[index].save(d.ID)
				}
			}
		}

	case 5: // Cash
		slotID, err := findFirstEmptySlot(d.cash, d.cashSlotSize)
		if err != nil {
			return err, Item{}
		}
		newItem.slotID = slotID
		newItem.save(d.ID)
		d.cash = append(d.cash, newItem)
		d.Send(packetInventoryAddItem(newItem, true))

	default:
		return fmt.Errorf("Unknown inventory ID: %d", newItem.invID), Item{}
	}

	return nil, newItem
}

func (d *Player) takeItem(id int32, slot int16, amount int16, invID byte) (Item, error) {
	item, err := d.getItem(invID, slot)
	if err != nil {
		return item, fmt.Errorf("item not found at inv=%d slot=%d", invID, slot)
	}

	if item.ID != id {
		return item, fmt.Errorf("Item.ID(%d) does not match ID(%d) provided", item.ID, id)
	}
	if item.invID != invID {
		return item, fmt.Errorf("inventory ID mismatch: item.invID(%d) vs provided invID(%d)", item.invID, invID)
	}
	if amount <= 0 {
		return item, fmt.Errorf("invalid amount requested: %d", amount)
	}

	if amount > item.amount {
		return item, fmt.Errorf("insufficient quantity: have=%d requested=%d", item.amount, amount)
	}

	item.amount -= amount
	if item.amount == 0 {
		// Delete item
		d.removeItem(item)
	} else {
		// Update item with new stack size
		d.updateItemStack(item)
	}

	return item, nil
}

func (d Player) updateItemStack(item Item) {
	item.save(d.ID)
	d.updateItem(item)
	d.Send(packetInventoryModifyItemAmount(item))
}

func (d *Player) updateItem(new Item) {
	var items = d.findItemInventory(new)

	for i, v := range items {
		if v.dbID == new.dbID {
			items[i] = new
			break
		}
	}
	d.updateItemInventory(new.invID, items)
}

func (d *Player) updateItemInventory(invID byte, inventory []Item) {
	switch invID {
	case 1:
		d.equip = inventory
	case 2:
		d.use = inventory
	case 3:
		d.setUp = inventory
	case 4:
		d.etc = inventory
	case 5:
		d.cash = inventory
	}
}

func (d *Player) findItemInventory(item Item) []Item {
	switch item.invID {
	case 1:
		return d.equip
	case 2:
		return d.use
	case 3:
		return d.setUp
	case 4:
		return d.etc
	case 5:
		return d.cash
	}

	return nil
}

func (d Player) getItem(invID byte, slotID int16) (Item, error) {
	var items []Item

	switch invID {
	case 1:
		items = d.equip
	case 2:
		items = d.use
	case 3:
		items = d.setUp
	case 4:
		items = d.etc
	case 5:
		items = d.cash
	}

	for _, v := range items {
		if v.slotID == slotID {
			return v, nil
		}
	}

	return Item{}, fmt.Errorf("Could not find Item")
}

func (d *Player) swapItems(item1, item2 Item, start, end int16) {
	item1.slotID = end
	item1.save(d.ID)
	d.updateItem(item1)

	item2.slotID = start
	item2.save(d.ID)
	d.updateItem(item2)

	d.Send(packetInventoryChangeItemSlot(item1.invID, start, end))
}

func (d *Player) removeItem(item Item) {
	switch item.invID {
	case 1:
		for i, v := range d.equip {
			if v.dbID == item.dbID {
				d.equip[i] = d.equip[len(d.equip)-1]
				d.equip = d.equip[:len(d.equip)-1]
				break
			}
		}
	case 2:
		for i, v := range d.use {
			if v.dbID == item.dbID {
				d.use[i] = d.use[len(d.use)-1]
				d.use = d.use[:len(d.use)-1]
				break
			}
		}
	case 3:
		for i, v := range d.setUp {
			if v.dbID == item.dbID {
				d.setUp[i] = d.setUp[len(d.setUp)-1]
				d.setUp = d.setUp[:len(d.setUp)-1]
				break
			}
		}
	case 4:
		for i, v := range d.etc {
			if v.dbID == item.dbID {
				d.etc[i] = d.etc[len(d.etc)-1]
				d.etc = d.etc[:len(d.etc)-1]
				break
			}
		}
	case 5:
		for i, v := range d.cash {
			if v.dbID == item.dbID {
				d.cash[i] = d.cash[len(d.cash)-1]
				d.cash = d.cash[:len(d.cash)-1]
				break
			}
		}
	}

	err := item.delete()
	if err != nil {
		log.Println(err)
		return
	}
	d.Send(packetInventoryRemoveItem(item))
}

func (d *Player) dropMesos(amount int32) error {
	if d.mesos < amount {
		return errors.New("not enough mesos")
	}

	d.takeMesos(amount)
	d.inst.dropPool.createDrop(dropSpawnNormal, dropFreeForAll, amount, d.pos, true, d.ID, d.ID)

	return nil
}

func (d *Player) moveItem(start, end, amount int16, invID byte) error {
	isRechargeable := func(itemID int32) bool {
		base := itemID / 10000
		return base == 207
	}

	if end == 0 { // drop item
		item, err := d.getItem(invID, start)
		if err != nil {
			return fmt.Errorf("Item to move doesn't exist")
		}

		if isRechargeable(item.ID) {
			amount = item.amount
		}

		dropItem := item
		dropItem.amount = amount
		dropItem.dbID = 0

		_, err = d.takeItem(item.ID, item.slotID, amount, item.invID)
		if err != nil {
			return fmt.Errorf("unable to take Item")
		}

		d.inst.dropPool.createDrop(dropSpawnNormal, dropFreeForAll, 0, d.pos, true, d.ID, 0, dropItem)
		return nil
	}

	if end < 0 { // move to equip slot
		item1, err := d.getItem(invID, start)
		if err != nil {
			return fmt.Errorf("Item to move doesn't exist")
		}

		if item1.twoHanded {
			if _, err := d.getItem(invID, -10); err == nil {
				d.Send(packetInventoryNoChange())
				return nil
			}
		} else if item1.shield() {
			if weapon, err := d.getItem(invID, -11); err == nil && weapon.twoHanded {
				d.Send(packetInventoryNoChange())
				return nil
			}
		}

		item2, err := d.getItem(invID, end)
		if err == nil {
			item2.slotID = start
			item2.save(d.ID)
			d.updateItem(item2)
		}

		item1.slotID = end
		item1.save(d.ID)
		d.updateItem(item1)

		d.Send(packetInventoryChangeItemSlot(invID, start, end))
		d.inst.send(packetInventoryChangeEquip(*d))
		return nil
	}

	item1, err := d.getItem(invID, start)
	if err != nil {
		return fmt.Errorf("Item to move doesn't exist")
	}

	item2, err := d.getItem(invID, end)
	if err != nil { // empty slot, simple move
		item1.slotID = end
		item1.save(d.ID)
		d.updateItem(item1)
		d.Send(packetInventoryChangeItemSlot(invID, start, end))
	} else { // destination occupied
		if (item1.isStackable() && item2.isStackable()) && (item1.ID == item2.ID) {
			if item1.amount == constant.MaxItemStack || item2.amount == constant.MaxItemStack { // swap
				d.swapItems(item1, item2, start, end)
			} else if item2.amount < constant.MaxItemStack { // try full merge
				if item2.amount+item1.amount <= constant.MaxItemStack {
					item2.amount = item2.amount + item1.amount
					item2.save(d.ID)
					d.updateItem(item2)
					d.Send(packetInventoryAddItem(item2, false))

					d.removeItem(item1)
				} else {
					d.swapItems(item1, item2, start, end)
				}
			}
		} else {
			d.swapItems(item1, item2, start, end)
		}
	}

	if start < 0 || end < 0 {
		d.inst.send(packetInventoryChangeEquip(*d))
	}

	return nil
}

func (d *Player) updateSkill(updatedSkill playerSkill) {
	if d.skills == nil {
		d.skills = make(map[int32]playerSkill)
	}
	d.skills[updatedSkill.ID] = updatedSkill
	d.Send(packetPlayerSkillBookUpdate(updatedSkill.ID, int32(updatedSkill.Level)))
	d.MarkDirty(DirtySkills, 800*time.Millisecond)
}

func (d *Player) useSkill(id int32, level byte, projectileID int32) error {
	skillInfo, _ := nx.GetPlayerSkill(id)

	skillUsed, ok := d.skills[id]
	if !ok {
		return nil
	}

	if skillUsed.Level != level {
		d.Conn.Send(packetMessageRedText("skill level mismatch"))
		return errors.New("skill level mismatch")
	}

	idx := int(skillUsed.Level) - 1
	if idx < 0 || idx >= len(skillInfo) {
		d.Conn.Send(packetMessageRedText("invalid skill data"))
		return errors.New("invalid skill data index")
	}
	si := skillInfo[idx]

	// Resource costs
	if si.MpCon > 0 {
		d.giveMP(-int16(si.MpCon))
	}
	if si.HpCon > 0 {
		d.giveHP(-int16(si.HpCon))
	}
	if si.MoneyConsume > 0 {
		d.takeMesos(int32(si.MoneyConsume))
	}

	if si.ItemCon > 0 {
		itemID := int32(si.ItemCon)
		need := int32(si.ItemConNo)
		if need <= 0 {
			need = 1
		}
		if !d.consumeItemsByID(itemID, need) {
			d.Conn.Send(packetMessageRedText("not enough items to use this skill"))
			return errors.New("not enough required items")
		}
	}

	if projectileID > 0 {
		need := int32(si.BulletConsume)
		if need <= 0 {
			need = int32(si.BulletCount)
		}
		if need > 0 {
			if !d.consumeItemsByID(projectileID, need) {
				d.Conn.Send(packetMessageRedText("not enough projectiles to use this skill"))
				return errors.New("not enough projectiles")
			}
		}
	}

	return nil
}

func (d *Player) consumeItemsByID(itemID int32, reqCount int32) bool {
	if reqCount <= 0 {
		return true
	}
	remaining := reqCount

	drain := func(invID byte, items []Item) {
		for i := range items {
			if remaining == 0 {
				return
			}
			it := items[i]
			if it.ID != itemID || it.amount <= 0 {
				continue
			}
			take := int16(it.amount)
			if int32(take) > remaining {
				take = int16(remaining)
			}
			if _, err := d.takeItem(itemID, it.slotID, take, invID); err == nil {
				remaining -= int32(take)
			}
		}
	}
	// Order: USE, SETUP, ETC, CASH
	drain(2, d.use)
	drain(3, d.setUp)
	drain(4, d.etc)
	drain(5, d.cash)

	return remaining == 0
}

func (d Player) admin() bool { return d.Conn.GetAdminLevel() > 0 }

func (d Player) displayBytes() []byte {
	pkt := mpacket.NewPacket()
	pkt.WriteByte(d.gender)
	pkt.WriteByte(d.skin)
	pkt.WriteInt32(d.face)
	pkt.WriteByte(0x00) // Messenger
	pkt.WriteInt32(d.hair)

	cashWeapon := int32(0)

	for _, b := range d.equip {
		if b.slotID < 0 && b.slotID > -20 {
			pkt.WriteByte(byte(math.Abs(float64(b.slotID))))
			pkt.WriteInt32(b.ID)
		}
	}

	for _, b := range d.equip {
		if b.slotID < -100 {
			if b.slotID == -111 {
				cashWeapon = b.ID
			} else {
				pkt.WriteByte(byte(math.Abs(float64(b.slotID + 100))))
				pkt.WriteInt32(b.ID)
			}
		}
	}

	pkt.WriteByte(0xFF)
	pkt.WriteByte(0xFF)
	pkt.WriteInt32(cashWeapon)
	pkt.WriteInt32(0) // Pet acc

	return pkt
}

// Logout flushes coalesced state and does a full checkpoint save.
func (d Player) Logout() {
	if d.inst != nil {
		if pos, err := d.inst.calculateNearestSpawnPortalID(d.pos); err == nil {
			d.mapPos = pos
		}
	}

	FlushNow(&d)

	if err := d.save(); err != nil {
		log.Printf("Player(%d) logout save failed: %v", d.ID, err)
	}

}

// Save data - this needs to be split to occur at relevant points in time
func (d Player) save() error {
	query := `UPDATE characters set skin=?, hair=?, face=?, level=?,
	job=?, str=?, dex=?, intt=?, luk=?, hp=?, maxHP=?, mp=?, maxMP=?,
	ap=?, sp=?, exp=?, fame=?, mapID=?, mapPos=?, mesos=?, miniGameWins=?,
	miniGameDraw=?, miniGameLoss=?, miniGamePoints=?, buddyListSize=? WHERE ID=?`

	var mapPos byte
	var err error

	if d.inst != nil {
		mapPos, err = d.inst.calculateNearestSpawnPortalID(d.pos)
	}
	if err != nil {
		return err
	}
	d.mapPos = mapPos

	_, err = common.DB.Exec(query,
		d.skin, d.hair, d.face, d.level, d.job, d.str, d.dex, d.intt, d.luk, d.hp, d.maxHP, d.mp,
		d.maxMP, d.ap, d.sp, d.exp, d.fame, d.mapID, d.mapPos, d.mesos, d.miniGameWins,
		d.miniGameDraw, d.miniGameLoss, d.miniGamePoints, d.buddyListSize, d.ID)
	if err != nil {
		return err
	}

	query = `INSERT INTO skills(characterID,skillID,level,cooldown)
	         VALUES(?,?,?,?)
	         ON DUPLICATE KEY UPDATE level=VALUES(level), cooldown=VALUES(cooldown)`
	for skillID, skill := range d.skills {
		if _, err := common.DB.Exec(query, d.ID, skillID, skill.Level, skill.Cooldown); err != nil {
			return err
		}
	}

	return nil
}

func (d *Player) damagePlayer(damage int16) {
	if damage <= 0 {
		return
	}

	newHP := d.hp - damage
	if newHP < 0 {
		newHP = 0
	}

	d.setHP(newHP)
}

func (d *Player) setInventorySlotSizes(equip, use, setup, etc, cash byte) {
	changed := (d.equipSlotSize != equip) || (d.useSlotSize != use) ||
		(d.setupSlotSize != setup) || (d.etcSlotSize != etc) || (d.cashSlotSize != cash)
	if !changed {
		return
	}
	d.equipSlotSize = equip
	d.useSlotSize = use
	d.setupSlotSize = setup
	d.etcSlotSize = etc
	d.cashSlotSize = cash
	d.MarkDirty(DirtyInvSlotSizes, 2*time.Second)
}

func (d *Player) setBuddyListSize(size byte) {
	if d.buddyListSize == size {
		return
	}
	d.buddyListSize = size
	d.MarkDirty(DirtyBuddySize, 1*time.Second)
}

func (d *Player) addMiniGameWin() {
	d.miniGameWins++
	d.MarkDirty(DirtyMiniGame, 1*time.Second)
}

func (d *Player) addMiniGameDraw() {
	d.miniGameDraw++
	d.MarkDirty(DirtyMiniGame, 1*time.Second)
}

func (d *Player) addMiniGameLoss() {
	d.miniGameLoss++
	d.MarkDirty(DirtyMiniGame, 1*time.Second)
}

func (d *Player) addMiniGamePoints(delta int32) {
	d.miniGamePoints += delta
	d.MarkDirty(DirtyMiniGame, 1*time.Second)
}

func (d *Player) sendBuddyList() {
	d.Send(packetBuddyListSizeUpdate(d.buddyListSize))
	d.Send(packetBuddyInfo(d.buddyList))
}

func (d Player) buddyListFull() bool {
	count := 0
	for _, v := range d.buddyList {
		if v.status != 1 {
			count++
		}
	}

	return count >= int(d.buddyListSize)
}

func (d *Player) addOnlineBuddy(id int32, name string, channel int32) {
	if d.buddyListFull() {
		return
	}

	for i, v := range d.buddyList {
		if v.id == id {
			d.buddyList[i].status = 0
			d.buddyList[i].channelID = channel
			d.Send(packetBuddyUpdate(id, name, d.buddyList[i].status, channel, false))
			return
		}
	}

	newBuddy := buddy{id: id, name: name, status: 0, channelID: channel}

	d.buddyList = append(d.buddyList, newBuddy)
	d.Send(packetBuddyInfo(d.buddyList))
}

func (d *Player) addOfflineBuddy(id int32, name string) {
	if d.buddyListFull() {
		return
	}

	for i, v := range d.buddyList {
		if v.id == id {
			d.buddyList[i].status = 2
			d.buddyList[i].channelID = -1
			d.Send(packetBuddyUpdate(id, name, d.buddyList[i].status, -1, false))
			return
		}
	}

	newBuddy := buddy{id: id, name: name, status: 2, channelID: -1}

	d.buddyList = append(d.buddyList, newBuddy)
	d.Send(packetBuddyInfo(d.buddyList))
}

func (d Player) hasBuddy(id int32) bool {
	for _, v := range d.buddyList {
		if v.id == id {
			return true
		}
	}

	return false
}

func (d *Player) removeBuddy(id int32) {
	for i, v := range d.buddyList {
		if v.id == id {
			d.buddyList[i] = d.buddyList[len(d.buddyList)-1]
			d.buddyList = d.buddyList[:len(d.buddyList)-1]
			d.Send(packetBuddyInfo(d.buddyList))
			return
		}
	}
}

// removeEquipAtSlot removes the equip from the given slot (equipped negative or inventory positive).
func (d *Player) removeEquipAtSlot(slot int16) bool {
	if slot < 0 {
		// Equipped Item; find and clear
		for i := range d.equip {
			if d.equip[i].slotID == slot {
				// Remove equipped Item
				d.equip[i].amount = 0
				return true
			}
		}
		return false
	}

	// Inventory equip; remove from inventory
	for i := range d.equip {
		if d.equip[i].slotID == slot {
			if d.equip[i].amount != 1 {
				return false
			}
			d.equip[i].amount = 0
			return true
		}
	}
	return false
}

// findUseItemBySlot returns the use Item (scroll) at the given slot from USE inventory.
func (d *Player) findUseItemBySlot(slot int16) *Item {
	for i := range d.use {
		if d.use[i].slotID == slot {
			return &d.use[i]
		}
	}
	return nil
}

// findEquipBySlot returns the equip by slot (negative = equipped, positive = inventory slot).
func (d *Player) findEquipBySlot(slot int16) *Item {
	for i := range d.equip {
		if d.equip[i].slotID == slot {
			return &d.equip[i]
		}
	}
	return nil
}

func LoadPlayerFromID(id int32, conn mnet.Client) Player {
	c := Player{}
	filter := "ID,accountID,worldID,Name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp, exp,fame,mapID,mapPos,previousMapID,mesos," +
		"equipSlotSize,useSlotSize,setupSlotSize,etcSlotSize,cashSlotSize,miniGameWins," +
		"miniGameDraw,miniGameLoss,miniGamePoints,buddyListSize"

	err := common.DB.QueryRow("SELECT "+filter+" FROM characters where ID=?", id).Scan(&c.ID,
		&c.accountID, &c.worldID, &c.Name, &c.gender, &c.skin, &c.hair, &c.face,
		&c.level, &c.job, &c.str, &c.dex, &c.intt, &c.luk, &c.hp, &c.maxHP, &c.mp,
		&c.maxMP, &c.ap, &c.sp, &c.exp, &c.fame, &c.mapID, &c.mapPos,
		&c.previousMap, &c.mesos, &c.equipSlotSize, &c.useSlotSize, &c.setupSlotSize,
		&c.etcSlotSize, &c.cashSlotSize, &c.miniGameWins, &c.miniGameDraw, &c.miniGameLoss,
		&c.miniGamePoints, &c.buddyListSize)

	if err != nil {
		log.Println(err)
		return c
	}

	c.petCashID = 0

	if err := common.DB.QueryRow("SELECT username, nx, maplepoints FROM accounts WHERE accountID=?", c.accountID).Scan(&c.accountName, &c.nx, &c.maplepoints); err != nil {
		log.Printf("loadPlayerFromID: failed to fetch accountName for accountID=%d: %v", c.accountID, err)
	}

	c.skills = make(map[int32]playerSkill)

	for _, s := range getSkillsFromCharID(c.ID) {
		c.skills[s.ID] = s
	}

	nxMap, err := nx.GetMap(c.mapID)

	if err != nil {
		log.Println(err)
		return c
	}

	c.pos.x = nxMap.Portals[c.mapPos].X
	c.pos.y = nxMap.Portals[c.mapPos].Y

	c.equip, c.use, c.setUp, c.etc, c.cash = loadInventoryFromDb(c.ID)

	c.buddyList = getBuddyList(c.ID, c.buddyListSize)

	c.quests = loadQuestsFromDB(c.ID)
	c.quests.init()
	c.quests.mobKills = loadQuestMobKillsFromDB(c.ID)

	// Initialize the per-Player buff manager so handlers can call plr.addBuff(...)
	c.buffs = NewCharacterBuffs(&c)
	c.buffs.plr.inst = c.inst

	c.storageInventory = new(storage)

	if err := c.storageInventory.load(c.accountID); err != nil {
		log.Printf("loadPlayerFromID: failed to load storage inventory for accountID=%d: %v", c.accountID, err)
	}

	c.Conn = conn

	return c
}

func getBuddyList(playerID int32, buddySize byte) []buddy {
	buddies := make([]buddy, 0, buddySize)
	filter := "friendID,accepted"
	rows, err := common.DB.Query("SELECT "+filter+" FROM buddy where characterID=?", playerID)

	if err != nil {
		log.Fatal(err)
		return buddies
	}

	defer rows.Close()

	i := 0
	for rows.Next() {
		newBuddy := buddy{}

		var accepted bool
		rows.Scan(&newBuddy.id, &accepted)

		filter := "channelID,Name,inCashShop"
		err := common.DB.QueryRow("SELECT "+filter+" FROM characters where ID=?", newBuddy.id).Scan(&newBuddy.channelID, &newBuddy.name, &newBuddy.cashShop)

		if err != nil {
			log.Fatal(err)
			return buddies
		}

		if !accepted {
			newBuddy.status = 1 // pending buddy request
		} else if newBuddy.channelID == -1 {
			newBuddy.status = 2 // offline
		} else {
			newBuddy.status = 0 // online
		}

		buddies = append(buddies, newBuddy)

		i++
	}

	return buddies
}

// Convenience helper used by handlers to apply a skill buff.
// Keeps your call sites (“plr.addBuff(...)”) simple.
func (d *Player) addBuff(skillID int32, level byte, delay int16) {
	if d == nil {
		return
	}
	if d.buffs == nil {
		d.buffs = NewCharacterBuffs(d)
	}
	d.buffs.plr.inst = d.inst
	d.buffs.AddBuff(d.ID, skillID, level, false, delay)
}

func (d *Player) addForeignBuff(charId, skillID int32, level byte, delay int16) {
	d.buffs.AddBuff(charId, skillID, level, true, delay)
}

func (d *Player) addMobDebuff(skillID, level byte, durationSec int16) {
	if d == nil || d.buffs == nil {
		return
	}
	d.buffs.plr.inst = d.inst
	d.buffs.AddMobDebuff(skillID, level, durationSec)
}

func (d *Player) removeAllCooldowns() {
	if d == nil || d.skills == nil {
		return
	}
	for _, ps := range d.skills {
		ps.Cooldown = 0
		ps.TimeLastUsed = 0
		d.updateSkill(ps)
	}
}

func (d *Player) saveBuffSnapshot() {
	if d.buffs == nil {
		return
	}

	// Ensure we don't snapshot already-stale buffs
	d.buffs.plr.inst = d.inst
	d.buffs.AuditAndExpireStaleBuffs()

	snaps := d.buffs.Snapshot()
	if len(snaps) == 0 {
		_, _ = common.DB.Exec("DELETE FROM character_buffs WHERE characterID=?", d.ID)
		return
	}

	tx, err := common.DB.Begin()
	if err != nil {
		log.Println("saveBuffSnapshot: begin tx:", err)
		return
	}
	defer func() { _ = tx.Commit() }()

	if _, err := tx.Exec("DELETE FROM character_buffs WHERE characterID=?", d.ID); err != nil {
		log.Println("saveBuffSnapshot: clear rows:", err)
		return
	}

	stmt, err := tx.Prepare("INSERT INTO character_buffs(characterID, sourceID, level, expiresAtMs) VALUES(?,?,?,?)")
	if err != nil {
		log.Println("saveBuffSnapshot: prepare:", err)
		return
	}
	defer stmt.Close()

	for _, s := range snaps {
		if _, err := stmt.Exec(d.ID, s.SourceID, s.Level, s.ExpiresAtMs); err != nil {
			log.Println("saveBuffSnapshot: insert:", err)
			return
		}
	}
}

func (d *Player) loadAndApplyBuffSnapshot() {
	rows, err := common.DB.Query("SELECT sourceID, level, expiresAtMs FROM character_buffs WHERE characterID=?", d.ID)
	if err != nil {
		log.Println("loadBuffSnapshot:", err)
		return
	}
	defer rows.Close()

	snaps := make([]BuffSnapshot, 0, 8)
	toDelete := make([]int32, 0, 8)

	now := time.Now().UnixMilli()
	for rows.Next() {
		var s BuffSnapshot
		if err := rows.Scan(&s.SourceID, &s.Level, &s.ExpiresAtMs); err != nil {
			log.Println("loadBuffSnapshot scan:", err)
			return
		}

		if s.ExpiresAtMs == 0 {
			toDelete = append(toDelete, s.SourceID)
			continue
		}

		normalized := s.ExpiresAtMs
		if normalized > 0 && normalized < 1000000000000 {
			normalized *= 1000
		}

		if normalized <= 0 || normalized <= now {
			toDelete = append(toDelete, s.SourceID)
			continue
		}

		s.ExpiresAtMs = normalized
		snaps = append(snaps, s)
	}
	if err := rows.Err(); err != nil {
		log.Println("loadBuffSnapshot rows:", err)
		return
	}

	if len(toDelete) > 0 {
		var b strings.Builder
		b.WriteString("DELETE FROM character_buffs WHERE characterID=? AND sourceID IN (")
		for i := range toDelete {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteByte('?')
		}
		b.WriteByte(')')

		args := make([]interface{}, 0, 1+len(toDelete))
		args = append(args, d.ID)
		for _, sid := range toDelete {
			args = append(args, sid)
		}
		if _, err := common.DB.Exec(b.String(), args...); err != nil {
			log.Println("loadBuffSnapshot delete expired:", err)
		}
	}

	if len(snaps) > 0 {
		if d.buffs == nil {
			d.buffs = NewCharacterBuffs(d)
		}
		d.buffs.plr.inst = d.inst
		d.buffs.RestoreFromSnapshot(snaps)
	}
}

// countItem returns total count across USE/SETUP/ETC for an Item ID.
func (d *Player) countItem(itemID int32) int32 {
	var total int32
	for i := range d.use {
		if d.use[i].ID == itemID {
			total += int32(d.use[i].amount)
		}
	}
	for i := range d.setUp {
		if d.setUp[i].ID == itemID {
			total += int32(d.setUp[i].amount)
		}
	}
	for i := range d.etc {
		if d.etc[i].ID == itemID {
			total += int32(d.etc[i].amount)
		}
	}
	for i := range d.cash {
		if d.cash[i].ID == itemID {
			total += int32(d.cash[i].amount)
		}
	}
	return total
}

// removeItemsByID removes up to reqCount across USE/SETUP/ETC. Returns true if fully removed.
func (d *Player) removeItemsByID(itemID int32, reqCount int32) bool {
	if reqCount <= 0 {
		return true
	}
	remaining := reqCount

	drain := func(invID byte, items []Item) {
		for i := range items {
			if remaining == 0 {
				return
			}
			it := items[i]
			if it.ID != itemID || it.amount <= 0 {
				continue
			}
			take := int16(it.amount)
			if int32(take) > remaining {
				take = int16(remaining)
			}
			if _, err := d.takeItem(itemID, it.slotID, take, invID); err == nil {
				remaining -= int32(take)
			}
		}
	}
	drain(2, d.use)
	drain(3, d.setUp)
	drain(4, d.etc)
	drain(5, d.cash)

	return remaining == 0
}

func (d *Player) meetsPrevQuestState(req nx.QuestStateReq) bool {
	switch req.State {
	case 2: // completed
		return d.quests.hasCompleted(req.ID)
	case 1: // in progress
		return d.quests.hasInProgress(req.ID)
	default:
		return true
	}
}

// meetsQuestBlock validates prereqs/Item counts.
func (d *Player) meetsQuestBlock(blk nx.CheckBlock) bool {
	// Previous quest states
	for _, rq := range blk.PrevQuests {
		if rq.State > 0 && !d.meetsPrevQuestState(rq) {
			return false
		}
	}

	// Item possession/turn-in counts
	for _, it := range blk.Items {
		if it.Count > 0 && d.countItem(it.ID) < it.Count {
			return false
		}
	}
	return true
}

// applyQuestAct grants EXP/Mesos and applies Item +/- from NX Act block.
func (d *Player) applyQuestAct(act nx.ActBlock) {
	if act.Exp > 0 {
		d.giveEXP(act.Exp, false, false)
	}
	if act.Money != 0 {
		if act.Money > 0 {
			d.giveMesos(act.Money)
		} else {
			d.takeMesos(-act.Money)
		}
	}

	if act.Pop != 0 {
		d.setFame(d.fame + int16(act.Pop))
	}

	for _, ai := range act.Items {
		switch {
		case ai.Count > 0:
			if it, err := CreateItemFromID(ai.ID, int16(ai.Count)); err == nil {
				_, _ = d.GiveItem(it)
			}
		case ai.Count < 0:
			_ = d.removeItemsByID(ai.ID, -ai.Count)
		}
	}
}

// tryStartQuest validates NX Start requirements, starts quest, applies Act(0).
func (d *Player) tryStartQuest(questID int16) bool {
	q, err := nx.GetQuest(questID)
	if err != nil {
		log.Printf("[Quest] start fail nx lookup: char=%s ID=%d err=%v", d.Name, questID, err)
		return false
	}

	if !d.meetsQuestBlock(q.Start) {
		return false
	}

	d.quests.add(questID, "")
	upsertQuestRecord(d.ID, questID, "")
	d.Send(packetQuestUpdate(questID, ""))

	d.applyQuestAct(q.ActOnStart)
	return true
}

func (d *Player) tryCompleteQuest(questID int16) bool {
	q, err := nx.GetQuest(questID)
	if err != nil {
		log.Printf("[Quest] complete fail nx lookup: char=%s ID=%d err=%v", d.Name, questID, err)
		return false
	}

	if !d.meetsQuestBlock(q.Complete) {
		return false
	}

	if !d.meetsMobKills(q.ID, q.Complete.Mobs) {
		return false
	}

	d.quests.remove(questID)
	nowMs := time.Now().UnixMilli()
	d.quests.complete(questID, nowMs)
	setQuestCompleted(d.ID, questID, nowMs)
	clearQuestMobKills(d.ID, q.ID)

	d.Send(packetQuestUpdate(questID, ""))
	d.Send(packetQuestComplete(questID))

	d.applyQuestAct(q.ActOnComplete)

	if q.ActOnComplete.NextQuest != 0 {
		_ = d.tryStartQuest(q.ActOnComplete.NextQuest)
	}
	return true
}

func (d *Player) buildQuestKillString(q nx.Quest) string {
	if d.quests.mobKills == nil {
		return ""
	}
	counts := d.quests.mobKills[q.ID]
	if counts == nil {
		return ""
	}

	out := make([]byte, 0, len(q.Complete.Mobs)*3)
	for _, req := range q.Complete.Mobs {
		val := counts[req.ID]
		if val < 0 {
			val = 0
		}
		if val > 999 {
			val = 999
		}

		a := byte('0' + (val/100)%10)
		b := byte('0' + (val/10)%10)
		c := byte('0' + (val % 10))
		out = append(out, a, b, c)
	}
	return string(out)
}

func (d *Player) onMobKilled(mobID int32) {
	if d == nil {
		return
	}
	for qid := range d.quests.inProgress {
		q, err := nx.GetQuest(qid)
		if err != nil {
			continue
		}

		var needed int32
		for _, rm := range q.Complete.Mobs {
			if rm.ID == mobID {
				needed = rm.Count
				break
			}
		}
		if needed == 0 {
			continue
		}

		// Init maps
		if d.quests.mobKills == nil {
			d.quests.mobKills = make(map[int16]map[int32]int32)
		}
		if d.quests.mobKills[qid] == nil {
			d.quests.mobKills[qid] = make(map[int32]int32)
		}

		cur := d.quests.mobKills[qid][mobID]
		if cur < needed {
			d.quests.mobKills[qid][mobID] = cur + 1
			upsertQuestMobKill(d.ID, qid, mobID, 1)
		}

		d.Send(packetQuestUpdateMobKills(qid, d.buildQuestKillString(q)))
	}
}

func (d *Player) meetsMobKills(questID int16, reqs []nx.ReqMob) bool {
	if len(reqs) == 0 {
		return true
	}
	m := d.quests.mobKills[questID]
	if m == nil {
		return false
	}
	for _, r := range reqs {
		if m[r.ID] < r.Count {
			return false
		}
	}
	return true
}

func (d *Player) allowsQuestDrop(qid int32) bool {
	if qid == 0 {
		return true
	}
	if d == nil {
		return false
	}
	return d.quests.hasInProgress(int16(qid))
}

func (p *Player) ensureSummonState() {
	if p.summons == nil {
		p.summons = &summonState{}
	}
}

func (p *Player) addSummon(su *summon) {
	p.ensureSummonState()

	if su.IsPuppet {
		if p.summons.puppet != nil {
			p.removeSummon(true, constant.SummonRemoveReasonReplaced)
		}
		p.summons.puppet = su
	} else {
		if p.summons.summon != nil {
			p.removeSummon(false, constant.SummonRemoveReasonReplaced)
		}
		p.summons.summon = su
	}

	p.broadcastShowSummon(su)
}

func (p *Player) removeSummon(puppet bool, reason byte) {
	p.ensureSummonState()

	shouldCancelBuff := func(r byte) bool {
		return r != constant.SummonRemoveReasonKeepBuff && r != constant.SummonRemoveReasonReplaced
	}

	if puppet {
		if p.summons.puppet == nil {
			return
		}
		su := p.summons.puppet
		p.broadcastRemoveSummon(su.SkillID, reason)
		if shouldCancelBuff(reason) && p.buffs != nil {
			p.buffs.expireBuffNow(su.SkillID)
		}
		p.summons.puppet = nil
	} else {
		if p.summons.summon == nil {
			return
		}
		su := p.summons.summon
		p.broadcastRemoveSummon(su.SkillID, reason)
		if shouldCancelBuff(reason) && p.buffs != nil {
			p.buffs.expireBuffNow(su.SkillID)
		}
		p.summons.summon = nil
	}
}

func (p *Player) getSummon(skillID int32) *summon {
	p.ensureSummonState()
	if p.summons.summon != nil && p.summons.summon.SkillID == skillID {
		return p.summons.summon
	}
	if p.summons.puppet != nil && p.summons.puppet.SkillID == skillID {
		return p.summons.puppet
	}
	return nil
}

func (p *Player) expireSummons() {
	if p != nil && p.buffs != nil {
		for sid := range p.buffs.activeSkillLevels {
			switch skill.Skill(sid) {
			case skill.SilverHawk, skill.GoldenEagle, skill.SummonDragon, skill.Puppet, skill.SniperPuppet:
				p.buffs.expireBuffNow(sid)
			}
		}
	}
}

func (p *Player) broadcastShowSummon(su *summon) {
	if p == nil || p.inst == nil {
		return
	}

	p.inst.send(packetShowSummon(p.ID, su))
}

func (p *Player) broadcastRemoveSummon(summonSkillID int32, reason byte) {
	if p == nil || p.inst == nil {
		return
	}

	p.inst.send(packetRemoveSummon(p.ID, summonSkillID, reason))
}

func (p *Player) updatePet() {
	p.MarkDirty(DirtyPet, time.Millisecond*300)
	p.inst.send(packetPlayerPetUpdate(p.pet.sn))
}

func (p *Player) petCanTakeDrop(drop fieldDrop) bool {
	if p.pet == nil {
		return false
	}

	if drop.mesos > 0 {
		if p.hasEquipped(constant.ItemMesoMagnet) {
			return true
		}
		return false
	} else {
		if p.hasEquipped(constant.ItemItemPouch) {
			return true
		}
		return false
	}
}

func (p *Player) hasEquipped(itemID int32) bool {
	if p == nil || itemID <= 0 {
		return false
	}
	for i := range p.equip {
		it := p.equip[i]
		if it.slotID < 0 && it.amount > 0 && it.ID == itemID {
			return true
		}
	}
	return false
}

func packetPlayerReceivedDmg(charID int32, attack int8, initalAmmount, reducedAmmount, spawnID, mobID, healSkillID int32,
	stance, reflectAction byte, reflected byte, reflectX, reflectY int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerTakeDmg)
	p.WriteInt32(charID)
	p.WriteInt8(attack)
	p.WriteInt32(initalAmmount)

	p.WriteInt32(spawnID)
	p.WriteInt32(mobID)
	p.WriteByte(stance)
	p.WriteByte(reflected)

	if reflected > 0 {
		p.WriteByte(reflectAction)
		p.WriteInt16(reflectX)
		p.WriteInt16(reflectY)
	}

	p.WriteInt32(reducedAmmount)

	// Check if used
	if reducedAmmount < 0 {
		p.WriteInt32(healSkillID)
	}

	return p
}

func packetPlayerLevelUpAnimation(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerAnimation)
	p.WriteInt32(charID)
	p.WriteByte(constant.PlayerEffectLevelUp)

	return p
}

func packetPlayerEffectSkill(onOther bool, skillID int32, level byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerEffect)
	if onOther {
		p.WriteByte(constant.PlayerEffectSkillOnOther)
	} else {
		p.WriteByte(constant.PlayerEffectSkillOnSelf)
	}
	p.WriteInt32(skillID)
	p.WriteByte(level)
	return p
}

func packetPlayerSkillAnimation(charID int32, party bool, skillID int32, level byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerAnimation)
	p.WriteInt32(charID)
	if party {
		p.WriteByte(constant.PlayerEffectSkillOnOther)
	} else {
		p.WriteByte(constant.PlayerEffectSkillOnSelf)
	}
	p.WriteInt32(skillID)
	p.WriteByte(level)
	return p
}

func packetPlayerGiveBuff(mask []byte, values []byte, delay int16, extra byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelTempStatChange)

	// Normalize to 8 bytes (low dword, high dword)
	if len(mask) < 8 {
		tmp := make([]byte, 8)
		copy(tmp[8-len(mask):], mask)
		mask = tmp
	} else if len(mask) > 8 {
		mask = mask[len(mask)-8:]
	}
	p.WriteBytes(mask)

	// Per-stat value triples (short value, int32 skill, short time)
	p.WriteBytes(values)

	// Self path: 2-byte delay
	p.WriteInt16(delay)

	// Optional extra (only if specific bits are present)

	writeExtra := buffMaskNeedsExtraByte(mask)
	if writeExtra {
		p.WriteByte(extra)
	}

	p.WriteInt64(0)
	p.WriteInt64(0)

	return p
}

func buffMaskNeedsExtraByte(mask []byte) bool {
	isSetLSB := func(bit int) bool {
		idx := bit / 8
		if idx < 0 || idx >= len(mask) {
			return false
		}
		shift := uint(bit % 8)
		return (mask[idx] & (1 << shift)) != 0
	}
	return isSetLSB(BuffComboAttack) || isSetLSB(BuffCharges)
}

// Self-cancel using 8-byte mask
func packetPlayerCancelBuff(mask []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRemoveTempStat)

	// Normalize to 8 bytes
	if len(mask) < 8 {
		tmp := make([]byte, 8)
		copy(tmp[8-len(mask):], mask)
		mask = tmp
	} else if len(mask) > 8 {
		mask = mask[len(mask)-8:]
	}
	p.WriteBytes(mask)
	p.WriteUint64(0)
	return p
}

func packetPlayerCancelForeignBuff(charID int32, mask []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerResetForeignBuff)
	p.WriteInt32(charID)
	p.WriteBytes(mask)
	return p
}

func packetPlayerMove(charID int32, bytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerMovement)
	p.WriteInt32(charID)
	p.WriteBytes(bytes)

	return p
}

func packetPlayerEmoticon(charID int32, emotion int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerEmoticon)
	p.WriteInt32(charID)
	p.WriteInt32(emotion)

	return p
}

func packetPlayerSkillBookUpdate(skillID int32, level int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSkillRecordUpdate)
	p.WriteByte(0x01)  // time check?
	p.WriteInt16(0x01) // number of skills to update
	p.WriteInt32(skillID)
	p.WriteInt32(level)
	p.WriteByte(0x01)

	return p
}

func packetPlayerStatChange(flag bool, stat int32, value int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelStatChange)
	p.WriteBool(flag)
	p.WriteInt32(stat)
	p.WriteInt32(value)

	return p
}

func packetPlayerNoChange() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x00)
	p.WriteByte(0x00)

	return p
}

func packetChangeChannel(ip []byte, port int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelChange)
	p.WriteBool(true)
	p.WriteBytes(ip)
	p.WriteInt16(port)

	return p
}

func packetCannotEnterCashShop() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelChangeServer)
	p.WriteByte(2)

	return p
}

func packetPlayerEnterGame(plr Player, channelID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelWarpToMap)
	p.WriteInt32(channelID)
	p.WriteByte(0) // character portal counter
	p.WriteByte(1) // Is connecting

	randomBytes := make([]byte, 4)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err.Error())
	}
	p.WriteBytes(randomBytes)
	p.WriteBytes(randomBytes)
	p.WriteBytes(randomBytes)
	p.WriteBytes(randomBytes)

	// Are active buffs Name encoded in here?
	p.WriteByte(0xFF)
	p.WriteByte(0xFF)

	p.WriteInt32(plr.ID)
	p.WritePaddedString(plr.Name, 13)
	p.WriteByte(plr.gender)
	p.WriteByte(plr.skin)
	p.WriteInt32(plr.face)
	p.WriteInt32(plr.hair)

	p.WriteInt64(plr.petCashID)

	p.WriteByte(plr.level)
	p.WriteInt16(plr.job)
	p.WriteInt16(plr.str)
	p.WriteInt16(plr.dex)
	p.WriteInt16(plr.intt)
	p.WriteInt16(plr.luk)
	p.WriteInt16(plr.hp)
	p.WriteInt16(plr.maxHP)
	p.WriteInt16(plr.mp)
	p.WriteInt16(plr.maxMP)
	p.WriteInt16(plr.ap)
	p.WriteInt16(plr.sp)
	p.WriteInt32(plr.exp)
	p.WriteInt16(plr.fame)

	p.WriteInt32(plr.mapID)
	p.WriteByte(plr.mapPos)

	p.WriteByte(20) // budy list size
	p.WriteInt32(plr.mesos)

	p.WriteByte(plr.equipSlotSize)
	p.WriteByte(plr.useSlotSize)
	p.WriteByte(plr.setupSlotSize)
	p.WriteByte(plr.etcSlotSize)
	p.WriteByte(plr.cashSlotSize)

	for _, v := range plr.equip {
		if v.slotID < 0 && !v.cash {
			p.WriteBytes(v.InventoryBytes())
		}
	}

	p.WriteByte(0)

	// Equips
	for _, v := range plr.equip {
		if v.slotID < 0 && v.cash {
			p.WriteBytes(v.InventoryBytes())
		}
	}

	p.WriteByte(0)

	// Inventory windows starts
	for _, v := range plr.equip {
		if v.slotID > -1 {
			p.WriteBytes(v.InventoryBytes())
		}
	}

	p.WriteByte(0)

	for _, v := range plr.use {
		p.WriteBytes(v.InventoryBytes())
	}

	p.WriteByte(0)

	for _, v := range plr.setUp {
		p.WriteBytes(v.InventoryBytes())
	}

	p.WriteByte(0)

	for _, v := range plr.etc {
		p.WriteBytes(v.InventoryBytes())
	}

	p.WriteByte(0)

	for _, v := range plr.cash {
		p.WriteBytes(v.InventoryBytes())
	}

	p.WriteByte(0)

	// Skills
	p.WriteInt16(int16(len(plr.skills))) // number of skills

	skillCooldowns := make(map[int32]int16)

	for _, skill := range plr.skills {
		p.WriteInt32(skill.ID)
		p.WriteInt32(int32(skill.Level))

		if skill.Cooldown > 0 {
			skillCooldowns[skill.ID] = skill.Cooldown
		}
	}

	p.WriteInt16(int16(len(skillCooldowns))) // number of cooldowns

	for id, cooldown := range skillCooldowns {
		p.WriteInt32(id)
		p.WriteInt16(cooldown)
	}

	// Quests
	writeActiveQuests(&p, plr.quests.inProgressList())
	writeCompletedQuests(&p, plr.quests.completedList())

	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt64(time.Now().Unix())

	return p
}

func packetInventoryAddItem(item Item, newItem bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteBool(!newItem)
	p.WriteByte(item.invID)

	if newItem {
		p.WriteBytes(item.shortBytes())
	} else {
		p.WriteInt16(item.slotID)
		p.WriteInt16(item.amount)
	}

	return p
}

func packetInventoryModifyItemAmount(item Item) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(item.invID)
	p.WriteInt16(item.slotID)
	p.WriteInt16(item.amount)

	return p
}

func packetInventoryAddItems(items []Item, newItem []bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)

	p.WriteByte(0x01)
	if len(items) != len(newItem) {
		p.WriteByte(0)
		return p
	}

	p.WriteByte(byte(len(items)))

	for i, v := range items {
		p.WriteBool(!newItem[i])
		p.WriteByte(v.invID)

		if newItem[i] {
			p.WriteBytes(v.shortBytes())
		} else {
			p.WriteInt16(v.slotID)
			p.WriteInt16(v.amount)
		}
	}

	return p
}

func packetInventoryChangeItemSlot(invTabID byte, origPos, newPos int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x02)
	p.WriteByte(invTabID)
	p.WriteInt16(origPos)
	p.WriteInt16(newPos)
	p.WriteByte(0x00) // ?

	return p
}

func packetInventoryRemoveItem(item Item) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x03)
	p.WriteByte(item.invID)
	p.WriteInt16(item.slotID)
	p.WriteUint64(0) //?

	return p
}

func packetInventoryChangeEquip(chr Player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerChangeAvatar)
	p.WriteInt32(chr.ID)
	p.WriteByte(1)
	p.WriteBytes(chr.displayBytes())
	p.WriteInt32(0)
	p.WriteInt32(0)
	p.WriteInt32(chr.chairID)

	// 15 x long(0) placeholders
	for i := 0; i < 15; i++ {
		p.WriteUint64(0)
	}

	return p
}

func packetInventoryNoChange() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x00)
	p.WriteByte(0x00)

	return p
}

func packetInventoryDontTake() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteInt16(1)

	return p
}

func packetBuddyInfo(buddyList []buddy) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x12)
	p.WriteByte(byte(len(buddyList)))

	for _, v := range buddyList {
		p.WriteInt32(v.id)
		p.WritePaddedString(v.name, 13)
		p.WriteByte(v.status)
		p.WriteInt32(v.channelID)
	}

	for _, v := range buddyList {
		p.WriteInt32(v.cashShop) // wizet mistake and this should be a bool?
	}

	return p
}

// It is possible to change ID's using this packet, however if the ID is a request it will crash the users
// client when selecting an option in notification, therefore the ID has not been allowed to change
func packetBuddyUpdate(id int32, name string, status byte, channelID int32, cashShop bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x08)
	p.WriteInt32(id) // original ID
	p.WriteInt32(id)
	p.WritePaddedString(name, 13)
	p.WriteByte(status)
	p.WriteInt32(channelID)
	p.WriteBool(cashShop)

	return p
}

func packetBuddyListSizeUpdate(size byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x15)
	p.WriteByte(size)

	return p
}

func packetPlayerAvatarSummaryWindow(charID int32, plr Player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelAvatarInfoWindow)
	p.WriteInt32(plr.ID)
	p.WriteByte(plr.level)
	p.WriteInt16(plr.job)
	p.WriteInt16(plr.fame)

	if plr.guild != nil {
		p.WriteString(plr.guild.name)
	} else {
		p.WriteString("")
	}

	if plr.petCashID != 0 {
		p.WriteBool(true)
		p.WriteInt32(plr.pet.itemID)
		p.WriteString(plr.pet.name)
		p.WriteByte(plr.pet.level)
		p.WriteInt16(plr.pet.closeness)
		p.WriteByte(plr.pet.fullness)
		p.WriteInt32(0) // equipped items
	} else {
		p.WriteBool(false)
	}
	p.WriteByte(0) // wishlist count

	return p
}

func packetShowCountdown(time int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCountdown)
	p.WriteByte(2)
	p.WriteInt32(time)

	return p
}

func packetHideCountdown() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCountdown)
	p.WriteByte(0)
	p.WriteInt32(0)

	return p
}

func packetBuddyUnkownError() mpacket.Packet {
	return packetBuddyRequestResult(0x16)
}

func packetBuddyPlayerFullList() mpacket.Packet {
	return packetBuddyRequestResult(0x0b)
}

func packetBuddyOtherFullList() mpacket.Packet {
	return packetBuddyRequestResult(0x0c)
}

func packetBuddyAlreadyAdded() mpacket.Packet {
	return packetBuddyRequestResult(0x0d)
}

func packetBuddyIsGM() mpacket.Packet {
	return packetBuddyRequestResult(0x0e)
}

func packetBuddyNameNotRegistered() mpacket.Packet {
	return packetBuddyRequestResult(0x0f)
}

func packetBuddyRequestResult(code byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(code)

	return p
}

func packetBuddyReceiveRequest(fromID int32, fromName string, fromChannelID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x9)
	p.WriteInt32(fromID)
	p.WriteString(fromName)
	p.WriteInt32(fromID)
	p.WritePaddedString(fromName, 13)
	p.WriteByte(1)
	p.WriteInt32(fromChannelID)
	p.WriteBool(false) // sender in cash shop

	return p
}

func packetBuddyOnlineStatus(id int32, channelID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x14)
	p.WriteInt32(id)
	p.WriteInt8(0)
	p.WriteInt32(channelID)

	return p
}

func packetBuddyChangeChannel(id int32, channelID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x14)
	p.WriteInt32(id)
	p.WriteInt8(1)
	p.WriteInt32(channelID)

	return p
}

func packetMapChange(mapID int32, channelID int32, mapPos byte, hp int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelWarpToMap)
	p.WriteInt32(channelID)
	p.WriteByte(0) // character portal counter
	p.WriteByte(0) // Is connecting
	p.WriteInt32(mapID)
	p.WriteByte(mapPos)
	p.WriteInt16(hp)
	p.WriteByte(0) // flag for more reading

	return p
}

func (plr *Player) WriteCharacterInfoPacket(p *mpacket.Packet) {
	p.WriteInt16(-1)

	// Stats
	p.WriteInt32(plr.ID)
	p.WritePaddedString(plr.Name, 13)
	p.WriteByte(plr.gender)
	p.WriteByte(plr.skin)
	p.WriteInt32(plr.face)
	p.WriteInt32(plr.hair)
	p.WriteInt64(plr.petCashID) // Pet Cash ID

	p.WriteByte(plr.level)
	p.WriteInt16(plr.job)
	p.WriteInt16(plr.str)
	p.WriteInt16(plr.dex)
	p.WriteInt16(plr.intt)
	p.WriteInt16(plr.luk)
	p.WriteInt16(plr.hp)
	p.WriteInt16(plr.maxHP)
	p.WriteInt16(plr.mp)
	p.WriteInt16(plr.maxMP)
	p.WriteInt16(plr.ap)
	p.WriteInt16(plr.sp)
	p.WriteInt32(plr.exp)
	p.WriteInt16(plr.fame)

	p.WriteInt32(plr.mapID)
	p.WriteByte(plr.mapPos)

	p.WriteByte(plr.buddyListSize)

	// Money
	p.WriteInt32(plr.mesos)

	if plr.equipSlotSize == 0 {
		plr.equipSlotSize = 24
	}
	if plr.useSlotSize == 0 {
		plr.useSlotSize = 24
	}
	if plr.setupSlotSize == 0 {
		plr.setupSlotSize = 24
	}
	if plr.etcSlotSize == 0 {
		plr.etcSlotSize = 24
	}
	if plr.cashSlotSize == 0 {
		plr.cashSlotSize = 24
	}

	p.WriteByte(plr.equipSlotSize)
	p.WriteByte(plr.useSlotSize)
	p.WriteByte(plr.setupSlotSize)
	p.WriteByte(plr.etcSlotSize)
	p.WriteByte(plr.cashSlotSize)

	// Equipped (normal then cash)
	for _, it := range plr.equip {
		if it.slotID < 0 && !it.cash {
			p.WriteBytes(it.InventoryBytes())
		}
	}
	p.WriteByte(0)
	for _, it := range plr.equip {
		if it.slotID < 0 && it.cash {
			p.WriteBytes(it.InventoryBytes())
		}
	}
	p.WriteByte(0)

	// Inventory tabs
	writeInv := func(items []Item) {
		cp := make([]Item, 0, len(items))
		for _, it := range items {
			if it.slotID > 0 {
				cp = append(cp, it)
			}
		}
		sort.Slice(cp, func(i, j int) bool { return cp[i].slotID < cp[j].slotID })
		for _, it := range cp {
			p.WriteBytes(it.InventoryBytes())
		}
		p.WriteByte(0)
	}
	writeInv(plr.equip)
	writeInv(plr.use)
	writeInv(plr.setUp)
	writeInv(plr.etc)
	writeInv(plr.cash)

	// Skills
	p.WriteInt16(int16(len(plr.skills)))
	skillCooldowns := make(map[int32]int16)

	for _, skill := range plr.skills {
		p.WriteInt32(skill.ID)
		p.WriteInt32(int32(skill.Level))

		if skill.Cooldown > 0 {
			skillCooldowns[skill.ID] = skill.Cooldown
		}
	}

	p.WriteInt16(int16(len(skillCooldowns)))

	for id, cooldown := range skillCooldowns {
		p.WriteInt32(id)
		p.WriteInt16(cooldown)
	}

	// Quests
	writeActiveQuests(p, plr.quests.inProgressList())
	writeCompletedQuests(p, plr.quests.completedList())

	p.WriteInt16(0) // MiniGames
	/*
	   - uint16 count
	   - repeat count times:
	       - int32 a
	       - int32 b
	       - int32 c
	       - int32 d
	       - int32 e
	*/
	p.WriteInt16(0) // Rings
	/*
	   - uint16 count
	   - repeat count times:
	       - decode ring object
	*/

	// Teleport rocks (5 normal, 10 VIP) INT32 = Saved MapID
	for i := 0; i < 5; i++ {
		p.WriteInt32(999999999) // Reg Tele rocks
	}
	for i := 0; i < 10; i++ {
		p.WriteInt32(999999999) // VIP Tele rocks
	}
}

func (p *Player) canReceiveItems(items []Item) bool {
	invCounts := map[byte]int{}
	for _, item := range items {
		invType := byte(item.ID / 1000000)
		invCounts[invType]++
	}

	for invType, needed := range invCounts {
		var cur, max byte
		switch invType {
		case 1:
			cur = byte(len(p.equip))
			max = p.equipSlotSize
		case 2:
			cur = byte(len(p.use))
			max = p.useSlotSize
		case 3:
			cur = byte(len(p.setUp))
			max = p.setupSlotSize
		case 4:
			cur = byte(len(p.etc))
			max = p.etcSlotSize
		case 5:
			cur = byte(len(p.cash))
			max = p.cashSlotSize
		default:
			continue
		}
		if cur+byte(needed) > max {
			return false
		}
	}
	return true
}

func packetPlayerPetUpdate(sn int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelStatChange)
	p.WriteBool(false)
	p.WriteInt32(constant.PetID)
	p.WriteUint64(uint64(sn))
	p.WriteByte(0)

	return p
}

func packetPlayerGiveForeignBuff(charID int32, mask []byte, values []byte, delay int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerGiveForeignBuff)
	p.WriteInt32(charID)

	// Normalize to 8 bytes (low dword, high dword) like self path
	if len(mask) < 8 {
		tmp := make([]byte, 8)
		copy(tmp[8-len(mask):], mask)
		mask = tmp
	} else if len(mask) > 8 {
		mask = mask[len(mask)-8:]
	}
	p.WriteBytes(mask)

	// Subset payload in reference order
	p.WriteBytes(values)

	// Delay (usually 0)
	p.WriteInt16(delay)
	return p
}

func packetPlayerShowChair(plrID, chairID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerSit)
	p.WriteInt32(plrID)
	p.WriteInt32(chairID)
	return p
}

func packetPlayerChairResult(chairID int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerSitResult)
	p.WriteBool(chairID != -1)
	if chairID != -1 {
		p.WriteInt16(chairID)
	}
	return p
}

func packetPlayerChairUpdate() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelStatChange)
	p.WriteInt16(1)
	p.WriteInt32(0)
	return p
}
