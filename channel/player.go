package channel

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/nx"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
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

type player struct {
	conn mnet.Client
	inst *fieldInstance

	id        int32 // Unique identifier of the character
	accountID int32
	worldID   byte

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

	name    string
	gender  byte
	skin    byte
	face    int32
	hair    int32
	chairID int32
	stance  byte
	pos     pos

	equipSlotSize byte
	useSlotSize   byte
	setupSlotSize byte
	etcSlotSize   byte
	cashSlotSize  byte

	equip []item
	use   []item
	setUp []item
	etc   []item
	cash  []item

	mesos int32

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

	// Per-player RNG for deterministic randomness
	rng *rand.Rand

	// write-behind persistence
	dirty DirtyBits
}

// Helper: mark dirty and schedule debounced save.
func (d *player) markDirty(bits DirtyBits, debounce time.Duration) {
	d.dirty |= bits
	scheduleSave(d, debounce)
}

// Helper: clear dirty bits after successful flush (kept for future; saver currently doesn't feed back)
func (d *player) clearDirty(bits DirtyBits) {
	d.dirty &^= bits
}

func (d *player) FlushNow() {
	flushNow(d)
}

// SeedRNGDeterministic seeds the per-player RNG using stable identifiers so
// gain sequences are reproducible across restarts and processes.
func (d *player) SeedRNGDeterministic() {
	// Compose as uint64 to avoid int64 constant overflow, then cast at runtime.
	const gamma uint64 = 0x9e3779b97f4a7c15
	seed64 := gamma ^
		(uint64(uint32(d.id)) << 1) ^
		(uint64(uint32(d.accountID)) << 33) ^
		(uint64(d.worldID) << 52)

	seed := int64(seed64) // two's complement wrapping is fine for rand.Source
	d.rng = rand.New(rand.NewSource(seed))
}

// ensureRNG guarantees d.rng is initialized. If a deterministic seed has not
// been set yet, it will use a time-based seed (non-deterministic).
func (d *player) ensureRNG() {
	if d.rng == nil {
		// Default to deterministic seeding for stability unless you want variability:
		d.SeedRNGDeterministic()
	}
}

func (d *player) randIntn(n int) int {
	d.ensureRNG()
	return d.rng.Intn(n)
}

// levelUpGains returns (hpGain, mpGain) using per-player RNG and job family.
// The random component uses a small range similar to legacy behavior.
// Tweak the constants to match your balance targets if needed.
func (d *player) levelUpGains() (int16, int16) {
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
func (d *player) send(packet mpacket.Packet) {
	if d == nil || d.conn == nil {
		return
	}
	d.conn.Send(packet)
}

func (d *player) setJob(id int16) {
	d.job = id
	d.conn.Send(packetPlayerStatChange(true, constant.JobID, int32(id)))
	d.markDirty(DirtyJob, 300*time.Millisecond)

	if d.party != nil {
		d.UpdatePartyInfo(d.party.ID, d.id, int32(d.job), int32(d.level), d.mapID, d.name)
	}
}

func (d *player) levelUp() {
	d.giveAP(5)
	d.giveSP(3)

	// Use per-player RNG and job-based helper for deterministic gains.
	hpGain, mpGain := d.levelUpGains()

	// Apply gains; clamp to avoid overflow/underflow
	newMaxHP := d.maxHP + hpGain
	newMaxMP := d.maxMP + mpGain

	// Basic sanity caps; adjust to your server balance
	if newMaxHP < 1 {
		newMaxHP = 1
	}
	if newMaxMP < 0 {
		newMaxMP = 0
	}
	// Update current HP/MP minimally: keep at least previous current + gains
	d.maxHP = newMaxHP
	d.maxMP = newMaxMP
	d.hp = int16(math.Min(float64(d.maxHP), float64(d.hp+hpGain)))
	d.mp = int16(math.Min(float64(d.maxMP), float64(d.mp+mpGain)))

	d.setHP(d.hp)
	d.setMaxHP(d.hp)

	d.setMP(d.mp)
	d.setMaxMP(d.mp)

	d.giveLevel(1)
}

func (d *player) setEXP(amount int32) {
	if d.level > 199 {
		d.exp = amount
		d.send(packetPlayerStatChange(false, constant.ExpID, int32(amount)))
		d.markDirty(DirtyEXP, 800*time.Millisecond)
		return
	}
	remainder := amount - constant.ExpTable[d.level-1]
	if remainder >= 0 {
		d.levelUp()
		d.setEXP(remainder)
	} else {
		d.exp = amount
		d.send(packetPlayerStatChange(false, constant.ExpID, int32(amount)))
		d.markDirty(DirtyEXP, 800*time.Millisecond)
	}
}

func (d *player) giveEXP(amount int32, fromMob, fromParty bool) {
	amount = int32(d.rates.exp * float32(amount))
	if fromMob {
		d.send(packetMessageExpGained(true, false, amount))
	} else if fromParty {
		d.send(packetMessageExpGained(false, false, amount))
	} else {
		d.send(packetMessageExpGained(false, true, amount))
	}

	d.setEXP(d.exp + amount)
}

func (d *player) setLevel(amount byte) {
	d.level = amount
	d.send(packetPlayerStatChange(false, constant.LevelID, int32(amount)))
	d.inst.send(packetPlayerLevelUpAnimation(d.id))
	d.markDirty(DirtyLevel, 300*time.Millisecond)

	if d.party != nil {
		d.UpdatePartyInfo(d.party.ID, d.id, int32(d.job), int32(d.level), d.mapID, d.name)
	}
}

func (d *player) giveLevel(amount byte) {
	d.setLevel(d.level + amount)
}

func (d *player) setAP(amount int16) {
	d.ap = amount
	d.send(packetPlayerStatChange(false, constant.ApID, int32(amount)))
	d.markDirty(DirtyAP, 300*time.Millisecond)
}

func (d *player) giveAP(amount int16) {
	d.setAP(d.ap + amount)
}

func (d *player) setSP(amount int16) {
	d.sp = amount
	d.send(packetPlayerStatChange(false, constant.SpID, int32(amount)))
	d.markDirty(DirtySP, 300*time.Millisecond)
}

func (d *player) giveSP(amount int16) {
	d.setSP(d.sp + amount)
}

func (d *player) setStr(amount int16) {
	d.str = amount
	d.send(packetPlayerStatChange(true, constant.StrID, int32(amount)))
	d.markDirty(DirtyStr, 500*time.Millisecond)
}

func (d *player) giveStr(amount int16) {
	d.setStr(d.str + amount)
}

func (d *player) setDex(amount int16) {
	d.dex = amount
	d.send(packetPlayerStatChange(true, constant.DexID, int32(amount)))
	d.markDirty(DirtyDex, 500*time.Millisecond)
}

func (d *player) giveDex(amount int16) {
	d.setDex(d.dex + amount)
}

func (d *player) setInt(amount int16) {
	d.intt = amount
	d.send(packetPlayerStatChange(true, constant.IntID, int32(amount)))
	d.markDirty(DirtyInt, 500*time.Millisecond)
}

func (d *player) giveInt(amount int16) {
	d.setInt(d.intt + amount)
}

func (d *player) setLuk(amount int16) {
	d.luk = amount
	d.send(packetPlayerStatChange(true, constant.LukID, int32(amount)))
	d.markDirty(DirtyLuk, 500*time.Millisecond)
}

func (d *player) giveLuk(amount int16) {
	d.setLuk(d.luk + amount)
}

func (d *player) setHP(amount int16) {
	if amount > constant.MaxHpValue {
		amount = constant.MaxHpValue
	}
	d.hp = amount
	d.send(packetPlayerStatChange(true, constant.HpID, int32(amount)))
	d.markDirty(DirtyHP, 500*time.Millisecond)
}

func (d *player) giveHP(amount int16) {
	newHP := d.hp + amount
	if newHP < 0 {
		d.setHP(0)
		return
	}
	if newHP > d.maxHP {
		d.setHP(d.maxHP)
		return
	}
	d.setHP(newHP)
}

func (d *player) giveMP(amount int16) {
	newMP := d.mp + amount
	if newMP < 0 {
		d.setMP(0)
		return
	}
	if newMP > d.maxMP {
		d.setMP(d.maxMP)
		return
	}
	d.setMP(newMP)
}

func (d *player) setMaxHP(amount int16) {
	if amount > constant.MaxHpValue {
		amount = constant.MaxHpValue
	}
	d.maxHP = amount
	d.send(packetPlayerStatChange(true, constant.MaxHpID, int32(amount)))
	d.markDirty(DirtyMaxHP, 500*time.Millisecond)
}

func (d *player) setMP(amount int16) {
	if amount > constant.MaxMpValue {
		amount = constant.MaxMpValue
	}
	d.mp = amount
	d.send(packetPlayerStatChange(true, constant.MpID, int32(amount)))
	d.markDirty(DirtyMP, 500*time.Millisecond)
}

func (d *player) setMaxMP(amount int16) {
	if amount > constant.MaxMpValue {
		amount = constant.MaxMpValue
	}
	d.maxMP = amount
	d.send(packetPlayerStatChange(true, constant.MaxMpID, int32(amount)))
	d.markDirty(DirtyMaxMP, 500*time.Millisecond)
}

func (d *player) setFame(amount int16) {
	d.fame = amount
	d.send(packetPlayerStatChange(true, constant.FameID, int32(amount)))

	_, err := common.DB.Exec("UPDATE characters SET fame=? WHERE id=?", d.fame, d.id)
	if err != nil {
		log.Printf("setFame: failed to save fame for character %d: %v", d.id, err)
	}
}

func (d *player) setMesos(amount int32) {
	d.mesos = amount
	d.send(packetPlayerStatChange(true, constant.MesosID, amount))
	// write-behind instead of immediate DB write
	d.markDirty(DirtyMesos, 200*time.Millisecond)
}

func (d *player) giveMesos(amount int32) {
	d.setMesos(d.mesos + int32(d.rates.mesos*float32(amount)))
}

func (d *player) takeMesos(amount int32) {
	d.setMesos(d.mesos - amount)
}

func (d *player) saveMesos() error {
	query := "UPDATE characters SET mesos=? WHERE accountID=? and name=?"
	_, err := common.DB.Exec(query, d.mesos, d.accountID, d.name)
	return err
}

// UpdateMovement - update Data from position data
func (d *player) UpdateMovement(frag movementFrag) {
	d.pos.x = frag.x
	d.pos.y = frag.y
	d.pos.foothold = frag.foothold
	d.stance = frag.stance
}

// SetPos of Data
func (d *player) SetPos(pos pos) {
	d.pos = pos
}

// checks Data is within a certain range of a position
func (d player) checkPos(pos pos, xRange, yRange int16) bool {
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

func (d *player) setMapID(id int32) {
	oldMapID := d.mapID
	d.mapID = id

	if d.party != nil {
		d.UpdatePartyInfo(d.party.ID, d.id, int32(d.job), int32(d.level), d.mapID, d.name)
	}

	// write-behind for mapID/pos (mapPos updated on save())
	d.markDirty(DirtyMap, 500*time.Millisecond)
	if err := d.saveMapID(id, oldMapID); err != nil {
		log.Println(err)
	}
}

func (d *player) saveMapID(newMapId, oldMapId int32) error {
	query := "UPDATE characters SET mapID=?,previousMapID=? WHERE accountID=? and name=?"

	_, err := common.DB.Exec(query,
		newMapId,
		oldMapId,
		d.accountID,
		d.name)

	if err != nil {
		return err
	}

	return nil
}

func (d player) noChange() {
	d.send(packetInventoryNoChange())
}

func (d *player) giveItem(newItem item) error { // TODO: Refactor
	findFirstEmptySlot := func(items []item, size byte) (int16, error) {
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
			return 0, fmt.Errorf("No empty item slot left")
		}

		return int16(slot), nil
	}

	switch newItem.invID {
	case 1: // Equip
		slotID, err := findFirstEmptySlot(d.equip, d.equipSlotSize)

		if err != nil {
			return err
		}

		newItem.slotID = slotID
		newItem.amount = 1 // just in case
		newItem.save(d.id)
		d.equip = append(d.equip, newItem)
		d.send(packetInventoryAddItem(newItem, true))
	case 2: // Use
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
				if v.id == newItem.id && v.amount < constant.MaxItemStack {
					slotID = v.slotID
					index = i
					break
				}
			}

			if slotID == 0 {
				slotID, err := findFirstEmptySlot(d.use, d.useSlotSize)

				if err != nil {
					return err
				}

				newItem.slotID = slotID
				newItem.save(d.id)
				d.use = append(d.use, newItem)
				d.send(packetInventoryAddItem(newItem, true))
			} else {
				remainder := newItem.amount - (constant.MaxItemStack - d.use[index].amount)

				if remainder > 0 { //partial merge
					slotID, err := findFirstEmptySlot(d.use, d.useSlotSize)

					if err != nil {
						return err
					}

					newItem.amount = value
					newItem.slotID = slotID
					newItem.save(d.id)

					d.use = append(d.use, newItem)
					d.send(packetInventoryAddItems([]item{d.use[index], newItem}, []bool{false, true}))
				} else { // full merge
					d.use[index].amount = d.use[index].amount + newItem.amount
					d.send(packetInventoryAddItem(d.use[index], false))
					d.use[index].save(d.id)
				}
			}

		}
	case 3: // Set-up
		slotID, err := findFirstEmptySlot(d.setUp, d.setupSlotSize)

		if err != nil {
			return err
		}

		newItem.slotID = slotID
		newItem.save(d.id)
		d.setUp = append(d.setUp, newItem)
		d.send(packetInventoryAddItem(newItem, true))
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
				if v.id == newItem.id && v.amount < constant.MaxItemStack {
					slotID = v.slotID
					index = i
					break
				}
			}

			if slotID == 0 {
				slotID, err := findFirstEmptySlot(d.etc, d.etcSlotSize)

				if err != nil {
					return err
				}

				newItem.slotID = slotID
				newItem.save(d.id)
				d.etc = append(d.etc, newItem)
				d.send(packetInventoryAddItem(newItem, true))
			} else {
				remainder := newItem.amount - (constant.MaxItemStack - d.etc[index].amount)

				if remainder > 0 { //partial merge
					slotID, err := findFirstEmptySlot(d.etc, d.etcSlotSize)

					if err != nil {
						return err
					}

					newItem.amount = value
					newItem.slotID = slotID
					newItem.save(d.id)

					d.etc = append(d.etc, newItem)
					d.send(packetInventoryAddItems([]item{d.etc[index], newItem}, []bool{false, true}))
				} else { // full merge
					d.etc[index].amount = d.etc[index].amount + newItem.amount
					d.send(packetInventoryAddItem(d.etc[index], false))
					d.etc[index].save(d.id)
				}
			}

		}
	case 5: // Cash
		// some are stackable, how to tell?
		slotID, err := findFirstEmptySlot(d.cash, d.cashSlotSize)

		if err != nil {
			return err
		}

		newItem.slotID = slotID
		newItem.save(d.id)
		d.cash = append(d.cash, newItem)
		d.send(packetInventoryAddItem(newItem, true))
	default:
		return fmt.Errorf("Unkown inventory id: %d", newItem.invID)
	}

	return nil
}

func (d *player) takeItem(id int32, slot int16, amount int16, invID byte) (item, error) {
	item, err := d.getItem(invID, slot)
	if err != nil {
		return item, err
	}

	if item.id != id {
		return item, fmt.Errorf("item.ID(%d) does not match ID(%d) provided", item.id, id)
	}

	maxRemove := math.Min(float64(item.amount), float64(amount))
	item.amount = item.amount - int16(maxRemove)
	if item.amount == 0 {
		// Delete item
		d.removeItem(item)
	} else {
		// Update item with new stack size
		d.updateItemStack(item)

	}

	return item, nil

}

func (d player) updateItemStack(item item) {
	item.save(d.id)
	d.updateItem(item)
	d.send(packetInventoryModifyItemAmount(item))
}

func (d *player) updateItem(new item) {
	var items = d.findItemInventory(new)

	for i, v := range items {
		if v.dbID == new.dbID {
			items[i] = new
			break
		}
	}
	d.updateItemInventory(new.invID, items)
}

func (d *player) updateItemInventory(invID byte, inventory []item) {
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

func (d *player) findItemInventory(item item) []item {
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

func (d player) getItem(invID byte, slotID int16) (item, error) {
	var items []item

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

	return item{}, fmt.Errorf("Could not find item")
}

func (d *player) swapItems(item1, item2 item, start, end int16) {
	item1.slotID = end
	item1.save(d.id)
	d.updateItem(item1)

	item2.slotID = start
	item2.save(d.id)
	d.updateItem(item2)

	d.send(packetInventoryChangeItemSlot(item1.invID, start, end))
}

func (d *player) removeItem(item item) {
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
	d.send(packetInventoryRemoveItem(item))
}

func (d *player) dropMesos(amount int32) error {
	if d.mesos < amount {
		return errors.New("not enough mesos")
	}

	d.takeMesos(amount)
	d.inst.dropPool.createDrop(dropSpawnNormal, dropFreeForAll, amount, d.pos, true, d.id, d.id)

	return nil
}

func (d *player) moveItem(start, end, amount int16, invID byte) error {
	if end == 0 { //drop item
		item, err := d.getItem(invID, start)

		if err != nil {
			return fmt.Errorf("Item to move doesn't exist")
		}

		dropItem := item
		dropItem.amount = amount
		dropItem.dbID = 0

		_, err = d.takeItem(item.id, item.slotID, amount, item.invID)
		if err != nil {
			return fmt.Errorf("unable to take item")
		}

		d.inst.dropPool.createDrop(dropSpawnNormal, dropFreeForAll, 0, d.pos, true, d.id, 0, dropItem)
	} else if end < 0 { // Move to equip slot
		item1, err := d.getItem(invID, start)

		if err != nil {
			return fmt.Errorf("Item to move doesn't exist")
		}

		if item1.twoHanded {
			if _, err := d.getItem(invID, -10); err == nil {
				d.send(packetInventoryNoChange()) // Should this do switching if space is available?
				return nil
			}
		} else if item1.shield() {
			if weapon, err := d.getItem(invID, -11); err == nil && weapon.twoHanded {
				d.send(packetInventoryNoChange()) // should this move weapon into into item 1 slot?
				return nil
			}
		}

		item2, err := d.getItem(invID, end)

		if err == nil {
			item2.slotID = start
			item2.save(d.id)
			d.updateItem(item2)
		}

		item1.slotID = end
		item1.save(d.id)
		d.updateItem(item1)

		d.send(packetInventoryChangeItemSlot(invID, start, end))
		d.inst.send(packetInventoryChangeEquip(*d))
	} else { // move within inventory
		item1, err := d.getItem(invID, start)

		if err != nil {
			return fmt.Errorf("Item to move doesn't exist")
		}

		item2, err := d.getItem(invID, end)

		if err != nil { // empty slot
			item1.slotID = end
			item1.save(d.id)
			d.updateItem(item1)

			d.send(packetInventoryChangeItemSlot(invID, start, end))
		} else { // moved onto item
			if (item1.isStackable() && item2.isStackable()) && (item1.id == item2.id) {
				if item1.amount == constant.MaxItemStack || item2.amount == constant.MaxItemStack { // swap items
					d.swapItems(item1, item2, start, end)
				} else if item2.amount < constant.MaxItemStack { // full merge
					if item2.amount+item1.amount <= constant.MaxItemStack {
						item2.amount = item2.amount + item1.amount
						item2.save(d.id)
						d.updateItem(item2)
						d.send(packetInventoryAddItem(item2, false))

						d.removeItem(item1)
					} else { // partial merge is just a swap
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
	}

	return nil
}

func (d *player) openStorage(npcID int32) bool {
	// TODO: implement storage UI and wiring
	return false
}

// SetSkinColor updates the player's skin tone and broadcasts an avatar refresh.
func (d *player) setSkinColor(skin byte) bool {
	d.skin = skin
	if d.inst != nil {
		d.inst.send(packetInventoryChangeEquip(*d))
	}
	return true
}

func (d *player) takeItemAnySlot(id int32, amount int16) (item, error) {
	if amount <= 0 {
		return item{}, fmt.Errorf("amount must be > 0")
	}

	var invID byte
	switch id / 1000000 {
	case 1:
		invID = 1 // Equip
	case 2:
		invID = 2 // Use
	case 3:
		invID = 3 // Setup
	case 4:
		invID = 4 // Etc
	case 5:
		invID = 5 // Cash
	default:
		return item{}, fmt.Errorf("unable to infer inventory for item id %d", id)
	}

	var (
		removed   int16
		lastFound item
	)

	const maxSlots int16 = 128

	for slot := int16(1); slot <= maxSlots && removed < amount; slot++ {
		it, err := d.getItem(invID, slot)
		if err != nil {
			continue
		}
		if it.id != id {
			continue
		}

		lastFound = it

		remaining := amount - removed
		if remaining >= it.amount {
			removed += it.amount
			d.removeItem(it)
		} else {
			it.amount -= remaining
			removed = amount
			d.updateItemStack(it)
		}
	}

	if removed < amount {
		return item{}, fmt.Errorf("not enough of item %d to remove: wanted %d, removed %d", id, amount, removed)
	}

	return lastFound, nil
}

func (d *player) itemQuantity(id int32) int32 {
	var total int32
	// Equip (count each matching equip as 1)
	for i := range d.equip {
		if d.equip[i].id == id {
			total += int32(d.equip[i].amount)
		}
	}
	// Use (stackable)
	for i := range d.use {
		if d.use[i].id == id {
			total += int32(d.use[i].amount)
		}
	}
	// Setup
	for i := range d.setUp {
		if d.setUp[i].id == id {
			total += int32(d.setUp[i].amount)
		}
	}
	// Etc (stackable)
	for i := range d.etc {
		if d.etc[i].id == id {
			total += int32(d.etc[i].amount)
		}
	}
	// Cash
	for i := range d.cash {
		if d.cash[i].id == id {
			total += int32(d.cash[i].amount)
		}
	}
	return total
}

func (d *player) hasEquipped(itemID int32) bool {
	for i := range d.equip {
		if d.equip[i].slotID < 0 && d.equip[i].id == itemID && d.equip[i].amount != 0 {
			return true
		}
	}
	return false
}

func (d *player) updateSkill(updatedSkill playerSkill) {
	if d.skills == nil {
		d.skills = make(map[int32]playerSkill)
	}
	d.skills[updatedSkill.ID] = updatedSkill
	d.send(packetPlayerSkillBookUpdate(updatedSkill.ID, int32(updatedSkill.Level)))
	d.markDirty(DirtySkills, 800*time.Millisecond)
}

func (d *player) useSkill(id int32, level byte) error {
	skillInfo, _ := nx.GetPlayerSkill(id)

	for lvl, skill := range skillInfo {
		if lvl == int(level) {

			d.giveMP(-int16(skill.MpCon))

			// Use item
			// d.consumeItem(skill.itemCon, skill.itemConNo)

		}

		// If haste, etc
	}

	return nil
}

func (d player) admin() bool { return d.conn.GetAdminLevel() > 0 }

func (d player) displayBytes() []byte {
	pkt := mpacket.NewPacket()
	pkt.WriteByte(d.gender)
	pkt.WriteByte(d.skin)
	pkt.WriteInt32(d.face)
	pkt.WriteByte(0x00) // ?
	pkt.WriteInt32(d.hair)

	cashWeapon := int32(0)

	for _, b := range d.equip {
		if b.slotID < 0 && b.slotID > -20 {
			pkt.WriteByte(byte(math.Abs(float64(b.slotID))))
			pkt.WriteInt32(b.id)
		}
	}

	for _, b := range d.equip {
		if b.slotID < -100 {
			if b.slotID == -111 {
				cashWeapon = b.id
			} else {
				pkt.WriteByte(byte(math.Abs(float64(b.slotID + 100))))
				pkt.WriteInt32(b.id)
			}
		}
	}

	pkt.WriteByte(0xFF)
	pkt.WriteByte(0xFF)
	pkt.WriteInt32(cashWeapon)

	return pkt
}

// Logout flushes coalesced state and does a full checkpoint save.
func (d player) Logout() {
	if d.inst != nil {
		if pos, err := d.inst.calculateNearestSpawnPortalID(d.pos); err == nil {
			d.mapPos = pos
		}
	}

	flushNow(&d)

	if err := d.save(); err != nil {
		log.Printf("player(%d) logout save failed: %v", d.id, err)
	}

}

// Save data - this needs to be split to occur at relevant points in time
func (d player) save() error {
	query := `UPDATE characters set skin=?, hair=?, face=?, level=?,
	job=?, str=?, dex=?, intt=?, luk=?, hp=?, maxHP=?, mp=?, maxMP=?,
	ap=?, sp=?, exp=?, fame=?, mapID=?, mapPos=?, mesos=?, miniGameWins=?,
	miniGameDraw=?, miniGameLoss=?, miniGamePoints=?, buddyListSize=? WHERE id=?`

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
		d.miniGameDraw, d.miniGameLoss, d.miniGamePoints, d.buddyListSize, d.id)
	if err != nil {
		return err
	}

	query = `INSERT INTO skills(characterID,skillID,level,cooldown)
	         VALUES(?,?,?,?)
	         ON DUPLICATE KEY UPDATE level=VALUES(level), cooldown=VALUES(cooldown)`
	for skillID, skill := range d.skills {
		if _, err := common.DB.Exec(query, d.id, skillID, skill.Level, skill.Cooldown); err != nil {
			return err
		}
	}

	return nil
}

func (d *player) damagePlayer(damage int16) {
	if damage < -1 {
		return
	}
	newHP := d.hp - damage

	if newHP <= -1 {
		d.hp = 0
	} else {
		d.hp = newHP
	}

	d.send(packetPlayerStatChange(true, constant.HpID, int32(d.hp)))
}

func (d *player) setInventorySlotSizes(equip, use, setup, etc, cash byte) {
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
	d.markDirty(DirtyInvSlotSizes, 2*time.Second)
}

func (d *player) setBuddyListSize(size byte) {
	if d.buddyListSize == size {
		return
	}
	d.buddyListSize = size
	d.markDirty(DirtyBuddySize, 1*time.Second)
}

func (d *player) addMiniGameWin() {
	d.miniGameWins++
	d.markDirty(DirtyMiniGame, 1*time.Second)
}

func (d *player) addMiniGameDraw() {
	d.miniGameDraw++
	d.markDirty(DirtyMiniGame, 1*time.Second)
}

func (d *player) addMiniGameLoss() {
	d.miniGameLoss++
	d.markDirty(DirtyMiniGame, 1*time.Second)
}

func (d *player) addMiniGamePoints(delta int32) {
	d.miniGamePoints += delta
	d.markDirty(DirtyMiniGame, 1*time.Second)
}

func (d *player) sendBuddyList() {
	d.send(packetBuddyListSizeUpdate(d.buddyListSize))
	d.send(packetBuddyInfo(d.buddyList))
}

func (d player) buddyListFull() bool {
	count := 0
	for _, v := range d.buddyList {
		if v.status != 1 {
			count++
		}
	}

	return count >= int(d.buddyListSize)
}

func (d *player) addOnlineBuddy(id int32, name string, channel int32) {
	if d.buddyListFull() {
		return
	}

	for i, v := range d.buddyList {
		if v.id == id {
			d.buddyList[i].status = 0
			d.buddyList[i].channelID = channel
			d.send(packetBuddyUpdate(id, name, d.buddyList[i].status, channel, false))
			return
		}
	}

	newBuddy := buddy{id: id, name: name, status: 0, channelID: channel}

	d.buddyList = append(d.buddyList, newBuddy)
	d.send(packetBuddyInfo(d.buddyList))
}

func (d *player) addOfflineBuddy(id int32, name string) {
	if d.buddyListFull() {
		return
	}

	for i, v := range d.buddyList {
		if v.id == id {
			d.buddyList[i].status = 2
			d.buddyList[i].channelID = -1
			d.send(packetBuddyUpdate(id, name, d.buddyList[i].status, -1, false))
			return
		}
	}

	newBuddy := buddy{id: id, name: name, status: 2, channelID: -1}

	d.buddyList = append(d.buddyList, newBuddy)
	d.send(packetBuddyInfo(d.buddyList))
}

func (d player) hasBuddy(id int32) bool {
	for _, v := range d.buddyList {
		if v.id == id {
			return true
		}
	}

	return false
}

func (d *player) removeBuddy(id int32) {
	for i, v := range d.buddyList {
		if v.id == id {
			d.buddyList[i] = d.buddyList[len(d.buddyList)-1]
			d.buddyList = d.buddyList[:len(d.buddyList)-1]
			d.send(packetBuddyInfo(d.buddyList))
			return
		}
	}
}

// removeEquipAtSlot removes the equip from the given slot (equipped negative or inventory positive).
func (d *player) removeEquipAtSlot(slot int16) bool {
	if slot < 0 {
		// Equipped item; find and clear
		for i := range d.equip {
			if d.equip[i].slotID == slot {
				// Remove equipped item
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

// findUseItemBySlot returns the use item (scroll) at the given slot from USE inventory.
func (d *player) findUseItemBySlot(slot int16) *item {
	for i := range d.use {
		if d.use[i].slotID == slot {
			return &d.use[i]
		}
	}
	return nil
}

// findEquipBySlot returns the equip by slot (negative = equipped, positive = inventory slot).
func (d *player) findEquipBySlot(slot int16) *item {
	for i := range d.equip {
		if d.equip[i].slotID == slot {
			return &d.equip[i]
		}
	}
	return nil
}

func loadPlayerFromID(id int32, conn mnet.Client) player {
	c := player{}
	filter := "id,accountID,worldID,name,gender,skin,hair,face,level,job,str,dex,intt," +
		"luk,hp,maxHP,mp,maxMP,ap,sp, exp,fame,mapID,mapPos,previousMapID,mesos," +
		"equipSlotSize,useSlotSize,setupSlotSize,etcSlotSize,cashSlotSize,miniGameWins," +
		"miniGameDraw,miniGameLoss,miniGamePoints,buddyListSize"

	err := common.DB.QueryRow("SELECT "+filter+" FROM characters where id=?", id).Scan(&c.id,
		&c.accountID, &c.worldID, &c.name, &c.gender, &c.skin, &c.hair, &c.face,
		&c.level, &c.job, &c.str, &c.dex, &c.intt, &c.luk, &c.hp, &c.maxHP, &c.mp,
		&c.maxMP, &c.ap, &c.sp, &c.exp, &c.fame, &c.mapID, &c.mapPos,
		&c.previousMap, &c.mesos, &c.equipSlotSize, &c.useSlotSize, &c.setupSlotSize,
		&c.etcSlotSize, &c.cashSlotSize, &c.miniGameWins, &c.miniGameDraw, &c.miniGameLoss,
		&c.miniGamePoints, &c.buddyListSize)

	if err != nil {
		log.Println(err)
		return c
	}

	c.skills = make(map[int32]playerSkill)

	for _, s := range getSkillsFromCharID(c.id) {
		c.skills[s.ID] = s
	}

	nxMap, err := nx.GetMap(c.mapID)

	if err != nil {
		log.Println(err)
		return c
	}

	c.pos.x = nxMap.Portals[c.mapPos].X
	c.pos.y = nxMap.Portals[c.mapPos].Y

	c.equip, c.use, c.setUp, c.etc, c.cash = loadInventoryFromDb(c.id)

	c.buddyList = getBuddyList(c.id, c.buddyListSize)

	c.quests = loadQuestsFromDB(c.id)
	c.quests.init()
	c.quests.mobKills = loadQuestMobKillsFromDB(c.id)

	// Initialize the per-player buff manager so handlers can call plr.addBuff(...)
	c.buffs = NewCharacterBuffs(&c)

	c.conn = conn

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

		filter := "channelID,name,inCashShop"
		err := common.DB.QueryRow("SELECT "+filter+" FROM characters where id=?", newBuddy.id).Scan(&newBuddy.channelID, &newBuddy.name, &newBuddy.cashShop)

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
func (d *player) addBuff(skillID int32, level byte, delay int16) {
	if d == nil {
		return
	}
	if d.buffs == nil {
		d.buffs = NewCharacterBuffs(d)
	}
	// You can pass any extra “sinc” values you need later; 0/0 is fine for standard buffs.
	d.buffs.AddBuff(d.id, skillID, level, false, delay)
}

func (d *player) addForeignBuff(charId, skillID int32, level byte, delay int16) {
	d.buffs.AddBuff(charId, skillID, level, true, delay)
}

func (d *player) removeAllCooldowns() {
	if d == nil || d.skills == nil {
		return
	}
	for _, ps := range d.skills {
		ps.Cooldown = 0
		ps.TimeLastUsed = 0
		d.updateSkill(ps)
	}
}

func (d *player) saveBuffSnapshot() {
	if d.buffs == nil {
		return
	}

	// Ensure we don't snapshot already-stale buffs
	d.buffs.AuditAndExpireStaleBuffs()

	snaps := d.buffs.Snapshot()
	if len(snaps) == 0 {
		_, _ = common.DB.Exec("DELETE FROM character_buffs WHERE characterID=?", d.id)
		return
	}

	tx, err := common.DB.Begin()
	if err != nil {
		log.Println("saveBuffSnapshot: begin tx:", err)
		return
	}
	defer func() { _ = tx.Commit() }()

	if _, err := tx.Exec("DELETE FROM character_buffs WHERE characterID=?", d.id); err != nil {
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
		if _, err := stmt.Exec(d.id, s.SourceID, s.Level, s.ExpiresAtMs); err != nil {
			log.Println("saveBuffSnapshot: insert:", err)
			return
		}
	}
}

func (d *player) loadAndApplyBuffSnapshot() {
	rows, err := common.DB.Query("SELECT sourceID, level, expiresAtMs FROM character_buffs WHERE characterID=?", d.id)
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
		args = append(args, d.id)
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
		d.buffs.RestoreFromSnapshot(snaps)
	}
}

// countItem returns total count across USE/SETUP/ETC for an item id.
func (d *player) countItem(itemID int32) int32 {
	var total int32
	for i := range d.use {
		if d.use[i].id == itemID {
			total += int32(d.use[i].amount)
		}
	}
	for i := range d.setUp {
		if d.setUp[i].id == itemID {
			total += int32(d.setUp[i].amount)
		}
	}
	for i := range d.etc {
		if d.etc[i].id == itemID {
			total += int32(d.etc[i].amount)
		}
	}
	return total
}

// removeItemsByID removes up to reqCount across USE/SETUP/ETC. Returns true if fully removed.
func (d *player) removeItemsByID(itemID int32, reqCount int32) bool {
	if reqCount <= 0 {
		return true
	}
	remaining := reqCount

	drain := func(invID byte, items []item) {
		for i := range items {
			if remaining == 0 {
				return
			}
			it := items[i]
			if it.id != itemID || it.amount <= 0 {
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

	return remaining == 0
}

func (d *player) meetsPrevQuestState(req nx.QuestStateReq) bool {
	switch req.State {
	case 2: // completed
		return d.quests.hasCompleted(req.ID)
	case 1: // in progress
		return d.quests.hasInProgress(req.ID)
	default:
		return true
	}
}

// meetsQuestBlock validates prereqs/item counts.
func (d *player) meetsQuestBlock(blk nx.CheckBlock) bool {
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

// applyQuestAct grants EXP/Mesos and applies item +/- from NX Act block.
func (d *player) applyQuestAct(act nx.ActBlock) {
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
			if it, err := createItemFromID(ai.ID, int16(ai.Count)); err == nil {
				_ = d.giveItem(it)
			}
		case ai.Count < 0:
			_ = d.removeItemsByID(ai.ID, -ai.Count)
		}
	}
}

// tryStartQuest validates NX Start requirements, starts quest, applies Act(0).
func (d *player) tryStartQuest(questID int16) bool {
	q, err := nx.GetQuest(questID)
	if err != nil {
		log.Printf("[Quest] start fail nx lookup: char=%s id=%d err=%v", d.name, questID, err)
		return false
	}

	if !d.meetsQuestBlock(q.Start) {
		return false
	}

	d.quests.add(questID, "")
	upsertQuestRecord(d.id, questID, "")
	d.send(packetQuestUpdate(questID, ""))

	d.applyQuestAct(q.ActOnStart)
	return true
}

func (d *player) tryCompleteQuest(questID int16) bool {
	q, err := nx.GetQuest(questID)
	if err != nil {
		log.Printf("[Quest] complete fail nx lookup: char=%s id=%d err=%v", d.name, questID, err)
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
	setQuestCompleted(d.id, questID, nowMs)
	clearQuestMobKills(d.id, q.ID)

	d.send(packetQuestUpdate(questID, ""))
	d.send(packetQuestComplete(questID))

	d.applyQuestAct(q.ActOnComplete)

	if q.ActOnComplete.NextQuest != 0 {
		_ = d.tryStartQuest(q.ActOnComplete.NextQuest)
	}
	return true
}

func (d *player) buildQuestKillString(q nx.Quest) string {
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

func (d *player) onMobKilled(mobID int32) {
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
			d.quests.mobKills = make(map[int16]map[int32]int32, 16)
		}
		if d.quests.mobKills[qid] == nil {
			d.quests.mobKills[qid] = make(map[int32]int32, 4)
		}

		cur := d.quests.mobKills[qid][mobID]
		if cur < needed {
			d.quests.mobKills[qid][mobID] = cur + 1
			upsertQuestMobKill(d.id, qid, mobID, 1)
		}

		d.send(packetQuestUpdateMobKills(qid, d.buildQuestKillString(q)))
	}
}

func (d *player) meetsMobKills(questID int16, reqs []nx.ReqMob) bool {
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

func (d *player) allowsQuestDrop(qid int32) bool {
	if qid == 0 {
		return true
	}
	if d == nil {
		return false
	}
	return d.quests.hasInProgress(int16(qid))
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
	p.WriteByte(0x00)

	return p
}

func packetPlayerSkillAnimSelf(charID int32, skillID int32, level byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerAnimation)
	p.WriteInt32(charID)
	p.WriteByte(0x01)
	p.WriteInt32(skillID)
	p.WriteByte(level)
	return p
}

func packetPlayerSkillAnimThirdParty(charID int32, party bool, self bool, skillID int32, level byte) mpacket.Packet {
	var p mpacket.Packet
	if party && self {
		p = mpacket.CreateWithOpcode(opcode.SendChannelSkillAnimation)
	} else {
		p = mpacket.CreateWithOpcode(opcode.SendChannelPlayerAnimation)
		p.WriteInt32(charID)
	}

	if party {
		p.WriteByte(0x02)
	} else {
		p.WriteByte(0x01)
	}
	p.WriteInt32(skillID)
	// Basis uses WriteInt for level; encode as int32 to match
	p.WriteInt32(int32(level))
	p.WriteUint64(0)
	p.WriteUint64(0)
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

func packetPlayerGiveForeignBuff(charID int32, mask []byte, values []byte, extra byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerGiveForeignBuff)
	p.WriteInt32(charID)

	// Normalize to 8 bytes (low dword, high dword)
	if len(mask) < 8 {
		tmp := make([]byte, 8)
		copy(tmp[8-len(mask):], mask)
		mask = tmp
	} else if len(mask) > 8 {
		mask = mask[len(mask)-8:]
	}

	p.WriteBytes(mask)

	// Per-stat value triples
	p.WriteBytes(values)

	// Foreign path: no delay, but still optional extra byte
	if buffMaskNeedsExtraByte(mask) {
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

func packetPlayerGiveDebuff(mask []byte, values []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelTempStatChange)
	p.WriteInt64(0)
	p.WriteBytes(mask)
	p.WriteBytes(values)
	p.WriteInt16(900)
	p.WriteByte(1)

	return p
}

func packetPlayerGiveForeignDebuff(charID int32, mask []byte, skillID, skillLevel int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerGiveForeignBuff)
	p.WriteInt32(charID)
	p.WriteInt64(0)
	p.WriteBytes(mask)
	p.WriteInt16(skillID)
	p.WriteInt16(skillLevel)
	p.WriteInt16(900)

	return p
}

func packetPlayerCancelForeignBuff(charID int32, mask []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerResetForeignBuff)
	p.WriteInt32(charID)
	p.WriteUint64(0)
	p.WriteBytes(mask)
	return p
}

func packetPlayerCancelDebuff(mask []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRemoveTempStat)
	p.WriteInt64(0)
	p.WriteBytes(mask)
	p.WriteInt32(0)

	return p
}

func packetPlayerCancelForeignDebuff(charID int32, mask []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerResetForeignBuff)
	p.WriteInt32(charID)
	p.WriteUint64(0)
	p.WriteBytes(mask)

	return p
}

func packetPlayerShowForeignEffect(charID int32, effectID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerShowForeignEffect)
	p.WriteInt32(charID)
	p.WriteInt32(effectID)

	return p
}

func packetPlayerShowBuffEffect(charID, skillID, effectID int32, direction byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerShowItemGainChat)
	p.WriteInt32(charID)
	p.WriteInt32(effectID)
	p.WriteInt32(skillID)
	p.WriteByte(1)
	if direction != 0x03 {
		p.WriteByte(direction)
	}
	return p
}

func packetPlayerShowOwnBuffEffect(skillId, effectId int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerShowForeignEffect)
	p.WriteInt32(0)
	p.WriteInt32(effectId)
	p.WriteInt32(skillId)
	p.WriteByte(0)

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

func packetPlayerEnterGame(plr player, channelID int32) mpacket.Packet {
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

	// Are active buffs name encoded in here?
	p.WriteByte(0xFF)
	p.WriteByte(0xFF)

	p.WriteInt32(plr.id)
	p.WritePaddedString(plr.name, 13)
	p.WriteByte(plr.gender)
	p.WriteByte(plr.skin)
	p.WriteInt32(plr.face)
	p.WriteInt32(plr.hair)

	p.WriteInt64(0) // Pet Cash ID

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
			p.WriteBytes(v.inventoryBytes())
		}
	}

	p.WriteByte(0)

	// Equips
	for _, v := range plr.equip {
		if v.slotID < 0 && v.cash {
			p.WriteBytes(v.inventoryBytes())
		}
	}

	p.WriteByte(0)

	// Inventory windows starts
	for _, v := range plr.equip {
		if v.slotID > -1 {
			p.WriteBytes(v.inventoryBytes())
		}
	}

	p.WriteByte(0)

	for _, v := range plr.use {
		p.WriteBytes(v.inventoryBytes())
	}

	p.WriteByte(0)

	for _, v := range plr.setUp {
		p.WriteBytes(v.inventoryBytes())
	}

	p.WriteByte(0)

	for _, v := range plr.etc {
		p.WriteBytes(v.inventoryBytes())
	}

	p.WriteByte(0)

	for _, v := range plr.cash {
		p.WriteBytes(v.inventoryBytes())
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

func packetInventoryAddItem(item item, newItem bool) mpacket.Packet {
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

func packetInventoryModifyItemAmount(item item) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(item.invID)
	p.WriteInt16(item.slotID)
	p.WriteInt16(item.amount)

	return p
}

func packetInventoryAddItems(items []item, newItem []bool) mpacket.Packet {
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

func packetInventoryRemoveItem(item item) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x01)
	p.WriteByte(0x03)
	p.WriteByte(item.invID)
	p.WriteInt16(item.slotID)
	p.WriteUint64(0) //?

	return p
}

func packetInventoryChangeEquip(chr player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerChangeAvatar)
	p.WriteInt32(chr.id)
	p.WriteByte(1)
	p.WriteBytes(chr.displayBytes())

	// Pet ID (spawned pet item id).
	p.WriteInt32(0)

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

// It is possible to change id's using this packet, however if the id is a request it will crash the users
// client when selecting an option in notification, therefore the id has not been allowed to change
func packetBuddyUpdate(id int32, name string, status byte, channelID int32, cashShop bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x08)
	p.WriteInt32(id) // original id
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

func packetPlayerAvatarSummaryWindow(charID int32, plr player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelAvatarInfoWindow)
	p.WriteInt32(plr.id)
	p.WriteByte(plr.level)
	p.WriteInt16(plr.job)
	p.WriteInt16(plr.fame)

	if plr.guild != nil {
		p.WriteString(plr.guild.name)
	} else {
		p.WriteString("")
	}

	p.WriteBool(false) // if has pet
	p.WriteByte(0)     // wishlist count

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
