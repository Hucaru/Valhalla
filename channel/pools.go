package channel

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

type lifePoolRectangle struct {
	ax, ay int16
	bx, by int16
}

func (r lifePoolRectangle) pointInRect(x, y int16) bool {
	// Rectangle is defined with ax <= bx and by <= ay in typical usage:
	// r := { ax: center.x-100, ay: center.y+100, bx: center.x+100, by: center.y-100 }
	minX, maxX := r.ax, r.bx
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	minY, maxY := r.by, r.ay
	if minY > maxY {
		minY, maxY = maxY, minY
	}
	return x >= minX && x <= maxX && y >= minY && y <= maxY
}

type lifePool struct {
	instance *fieldInstance

	npcs map[int32]*npc
	mobs map[int32]*monster

	spawnableMobs []monster

	mobID, npcID         int32
	lastMobSpawnTime     time.Time
	mobCapMin, mobCapMax int

	activeMobCtrl map[*Player]bool

	dropPool *dropPool

	rNumber *rand.Rand
}

func creatNewLifePool(inst *fieldInstance, npcData, mobData []nx.Life, mobCapMin, mobCapMax int) lifePool {
	pool := lifePool{instance: inst, activeMobCtrl: make(map[*Player]bool)}

	pool.npcs = make(map[int32]*npc)

	for _, l := range npcData {
		id, err := pool.nextNpcID()

		if err != nil {
			continue
		}

		val := createNpcFromData(id, l)
		pool.npcs[id] = &val
	}

	pool.mobs = make(map[int32]*monster)
	pool.spawnableMobs = make([]monster, len(mobData))

	for i, v := range mobData {
		m, err := nx.GetMob(v.ID)

		if err != nil {
			continue
		}

		id, err := pool.nextMobID()

		if err != nil {
			continue
		}

		val := createMonsterFromData(id, v, m, true, true)
		pool.mobs[id] = &val
		pool.mobs[id].summonType = -1

		pool.spawnableMobs[i] = createMonsterFromData(id, v, m, true, true)
	}

	pool.mobCapMin = mobCapMin
	pool.mobCapMax = mobCapMax

	randomSource := rand.NewSource(time.Now().UTC().UnixNano())
	pool.rNumber = rand.New(randomSource)

	return pool
}

func (pool *lifePool) setDropPool(drop *dropPool) {
	pool.dropPool = drop
}

func (pool lifePool) mobCount() int {
	return len(pool.mobs)
}

func (pool *lifePool) nextMobID() (int32, error) {
	for i := 0; i < 100; i++ { // Try 100 times to generate an ID if first time fails
		pool.mobID++

		if pool.mobID == math.MaxInt32-1 {
			pool.mobID = math.MaxInt32 / 2
		} else if pool.mobID == 0 {
			pool.mobID = 1
		}

		if _, ok := pool.mobs[pool.mobID]; !ok {
			return pool.mobID, nil
		}
	}

	return 0, fmt.Errorf("no space to generate ID in life pool")
}

func (pool *lifePool) nextNpcID() (int32, error) {
	for i := 0; i < 100; i++ { // Try 100 times to generate an ID if first time fails
		pool.npcID++

		if pool.npcID == math.MaxInt32-1 {
			pool.npcID = math.MaxInt32 / 2
		} else if pool.npcID == 0 {
			pool.npcID = 1
		}

		if _, ok := pool.npcs[pool.npcID]; !ok {
			return pool.npcID, nil
		}
	}

	return 0, fmt.Errorf("no space to generate ID in life pool")
}

func (pool lifePool) canPause() bool {
	// TODO: Need to check if any status effects are on monsters, if none are present then this pool can pause
	return false
}

func (pool lifePool) getNPCFromSpawnID(id int32) (npc, error) {
	for _, v := range pool.npcs {
		if v.spawnID == id {
			return *v, nil
		}
	}

	return npc{}, fmt.Errorf("Could not find npc with ID %d", id)
}

func (pool *lifePool) addPlayer(plr *Player) {
	for i, npc := range pool.npcs {
		plr.Send(packetNpcShow(npc))

		if npc.controller == nil {
			pool.npcs[i].setController(plr)
		}
	}

	for i, m := range pool.mobs {
		plr.Send(packetMobShow(m))

		if m.controller == nil {
			pool.mobs[i].setController(plr, false)
		}

		pool.showMobBossHPBar(m, plr)
	}
}

func (pool *lifePool) removePlayer(plr *Player) {
	for i, v := range pool.npcs {
		if v.controller != nil && v.controller.Conn == plr.Conn {
			pool.npcs[i].removeController()

			// find new controller
			if plr := pool.instance.findController(); plr != nil {
				if cont, ok := plr.(*Player); ok {
					pool.npcs[i].setController(cont)
				}
			}
		}
	}

	for i, v := range pool.mobs {
		if v.controller != nil && v.controller.Conn == plr.Conn {
			pool.mobs[i].removeController()

			// find new controller
			if plr := pool.instance.findController(); plr != nil {
				if cont, ok := plr.(*Player); ok {
					pool.mobs[i].setController(cont, false)
				}
			}
		}

		plr.Send(packetMobRemove(v.spawnID, 0x0)) // need to tell client to remove mobs for instance swapping
	}

	delete(pool.activeMobCtrl, plr)
}

func (pool *lifePool) npcAcknowledge(poolID int32, plr *Player, data []byte) {
	for i := range pool.npcs {
		if poolID == pool.npcs[i].spawnID {
			pool.npcs[i].acknowledgeController(plr, pool.instance, data)
			return
		}
	}

}

func (pool *lifePool) mobAcknowledge(poolID int32, plr *Player, moveID int16, skillPossible bool, action int8, skillData uint32, moveData movement, finalData movementFrag, moveBytes []byte) {
	for i, v := range pool.mobs {
		mob := pool.mobs[i]

		if poolID == v.spawnID && v.controller.Conn == plr.Conn {
			skillID := byte(skillData)
			skillLevel := byte(skillData >> 8)
			skillDelay := int16(skillData >> 16)

			var actualAction int8

			if action < 0 {
				actualAction = -1
			} else {
				actualAction = int8(action) >> 1
			}

			// Perform either skill or attack
			if actualAction >= 21 && actualAction <= 25 {
				pool.mobs[i].performSkill(skillDelay, skillLevel, skillID)
			} else if actualAction > 12 && actualAction < 20 {
				attackID := byte(actualAction - 12)

				mobSkills := mob.skills

				// check mob can use attack
				if level, valid := mobSkills[attackID]; valid {
					levels, err := nx.GetMobSkill(attackID)

					if err != nil {
						return
					}

					if int(level) < len(levels) {
						skill := levels[level]
						mob.mp = mob.mp - skill.MpCon
						if mob.mp < 0 {
							mob.mp = 0
						}
					}

				}

				pool.mobs[i].performAttack(attackID)
			}

			skillID, skillLevel = mob.canUseSkill(skillPossible)

			if !moveData.validateMob(v) {
				return
			}

			pool.mobs[i].acknowledgeController(moveID, finalData, skillPossible, skillID, skillLevel)
			pool.instance.sendExcept(packetMobMove(poolID, skillPossible, action, skillData, moveBytes), v.controller.Conn)
		}
	}
}

func (pool *lifePool) mobDamaged(poolID int32, damager *Player, dmg ...int32) {
	for i, v := range pool.mobs {
		if v.spawnID == poolID {
			pool.mobs[i].removeController()

			if damager != nil {
				pool.mobs[i].setController(damager, true)
				pool.activeMobCtrl[damager] = true
				pool.mobs[i].giveDamage(damager, dmg...)
			} else {
				pool.mobs[i].giveDamage(nil, dmg...)
			}

			pool.showMobBossHPBar(v, nil)

			if pool.mobs[i].hp < 1 {
				for plr, dmg := range pool.mobs[i].dmgTaken {
					if damager != nil && damager.mapID != plr.mapID {
						continue
					}

					var partyExp int32

					if dmg == v.maxHP {
						plr.giveEXP(v.exp, true, false)
						partyExp = int32(float64(v.exp) * 0.25) // TODO: party exp needs to be properly calculated
					} else if float64(dmg)/float64(v.maxHP) > 0.60 {
						plr.giveEXP(v.exp, true, false)
						partyExp = int32(float64(v.exp) * 0.25) // TODO: party exp needs to be properly calculated
					} else {
						newExp := int32(float64(v.exp) * 0.25)

						if newExp == 0 {
							newExp = 1
						}

						plr.giveEXP(newExp, true, false)
						partyExp = int32(float64(newExp) * 0.25) // TODO: party exp needs to be properly calculated
					}

					if plr.party != nil {
						// TODO: check level difference is appropriate
						plr.party.giveExp(plr.ID, partyExp, true)
					}
				}

				// quest mob logic

				// on die logic
				for _, id := range v.revives {
					spawnID, err := pool.nextMobID()

					if err != nil {
						continue
					}

					newMob, err := createMonsterFromID(spawnID, int32(id), v.pos, nil, true, true, 0)

					if err != nil {
						log.Println(err)
						continue
					}

					newMob.faceLeft = v.faceLeft
					newMob.summonType = -3
					newMob.summonOption = v.spawnID
					pool.spawnReviveMob(&newMob, damager)
				}

				pool.removeMob(v.spawnID, 0x1)
				damager.onMobKilled(v.id)

				if dropEntry, ok := dropTable[v.id]; ok {
					var mesos int32
					drops := make([]Item, 0, len(dropEntry))

					for _, entry := range dropEntry {
						if entry.IsMesos {
							mesos = randRangeInclusive(pool.rNumber, entry.Min, entry.Max)
							continue
						}

						// Quest-gated Item: only allow if killer has quest active
						// This should probably be hidden from instance and only viewable to Player
						if entry.QuestID != 0 && !damager.allowsQuestDrop(entry.QuestID) {
							continue
						}

						if !rollDrop(pool.rNumber, entry.Chance, pool.dropPool.rates.drop) {
							continue
						}

						var amount int16 = 1
						minAmt := entry.Min
						maxAmt := entry.Max
						if maxAmt != 1 {
							val := randRangeInclusive(pool.rNumber, minAmt, maxAmt)
							if val > math.MaxInt16 {
								amount = math.MaxInt16
							} else if val < 1 {
								amount = 1
							} else {
								amount = int16(val)
							}
						}

						newItem, err := CreateItemFromID(entry.ItemID, amount)
						if err != nil {
							log.Println("Failed to create drop for mobID:", v.id, "with error:", err)
							continue
						}
						drops = append(drops, newItem)
					}

					// TODO: droppool type determination between DropTimeoutNonOwner and DropTimeoutNonOwnerParty
					pool.dropPool.createDrop(dropSpawnNormal, dropFreeForAll, int32(damager.rates.mesos*float32(mesos)), v.pos, true, 0, 0, drops...)

					// If has hp bar: remove
				}

				if v.spawnInterval > 0 {
					for i, k := range pool.spawnableMobs {
						if k.id == v.id { // if this needs strengthening then add a spawn pos check
							pool.spawnableMobs[i].timeToSpawn = time.Now().Add(time.Millisecond * time.Duration(v.spawnInterval))
							break
						}
					}
				}
			}
			break
		}
	}
}

func randRangeInclusive(r *rand.Rand, lo, hi int32) int32 {
	if hi <= lo {
		return lo
	}
	delta := hi - lo + 1
	if delta <= 0 {
		return lo
	}
	return r.Int31n(delta) + lo
}

func rollDrop(r *rand.Rand, baseChance int64, rate float32) bool {
	const denom int64 = 1000000

	// Fast-path clamps
	if baseChance <= 0 {
		return false
	}
	if baseChance >= denom && rate >= 1 {
		return true
	}

	// Scale with rounding in float domain, then convert back to int64
	scaled := int64(math.Round(float64(baseChance) * float64(rate)))
	if scaled <= 0 {
		return false
	}
	if scaled >= denom {
		return true
	}

	// Roll
	return r.Int63n(denom) < scaled
}

func (pool *lifePool) killMobs(deathType byte) {
	// Need to collect keys first as when iterating over the map and killing we will kill any subsequent spawns depending on map iteration order
	keys := make([]int32, 0, len(pool.mobs))

	for key := range pool.mobs {
		keys = append(keys, key)
	}

	for _, key := range keys {
		// Apply the provided deathType for consistency
		pool.instance.send(packetMobRemove(pool.mobs[key].spawnID, deathType))
		pool.mobDamaged(pool.mobs[key].spawnID, nil, pool.mobs[key].hp)
	}
}

func (pool *lifePool) eraseMobs() {
	keys := make([]int32, 0, len(pool.mobs))

	for key := range pool.mobs {
		keys = append(keys, key)
	}
	for _, key := range keys {
		pool.removeMob(key, 0)
		// removeMob already deletes from pool.mobs
	}
}

func (pool *lifePool) spawnMob(m *monster, hasAgro bool) bool {
	pool.mobs[m.spawnID] = m
	pool.instance.send(packetMobShow(m))

	if plr := pool.instance.findController(); plr != nil {
		if cont, ok := plr.(*Player); ok {
			for _, v := range pool.mobs {
				v.setController(cont, hasAgro)
			}
		}
	}

	pool.showMobBossHPBar(m, nil)
	m.summonType = -1

	return false
}

func (pool *lifePool) spawnMobFromID(mobID int32, location pos, hasAgro, items, mesos bool, summoner int32) error {
	id, err := pool.nextMobID()

	if err != nil {
		return err
	}

	m, err := createMonsterFromID(id, mobID, location, nil, items, mesos, summoner)

	if err != nil {
		return err
	}

	pool.spawnMob(&m, hasAgro)

	return nil
}

func (pool *lifePool) spawnReviveMob(m *monster, cont *Player) {
	pool.spawnMob(m, true)

	pool.mobs[m.spawnID].summonType = -2
	pool.mobs[m.spawnID].summonOption = 0

	if cont != nil {
		pool.mobs[m.spawnID].setController(cont, true)
	}
}

func (pool *lifePool) removeMob(poolID int32, deathType byte) {
	if _, ok := pool.mobs[poolID]; !ok {
		return
	}

	delete(pool.mobs, poolID)
	pool.instance.send(packetMobRemove(poolID, deathType))
}

func (pool lifePool) showMobBossHPBar(mob *monster, plr *Player) {
	if plr != nil {
		if show, mobID, hp, maxHP, hpFgColour, hpBgColour := mob.hasHPBar(); show {
			plr.Send(packetMobShowBossHP(mobID, hp, maxHP, hpFgColour, hpBgColour))
		}
	} else {
		if show, mobID, hp, maxHP, hpFgColour, hpBgColour := mob.hasHPBar(); show {
			pool.instance.send(packetMobShowBossHP(mobID, hp, maxHP, hpFgColour, hpBgColour))
		}
	}
}

func (pool *lifePool) update(t time.Time) {
	for i := range pool.mobs {
		pool.mobs[i].update(t)
	}

	pool.attemptMobSpawn(false)
}

func (pool *lifePool) attemptMobSpawn(poolReset bool) {
	if len(pool.spawnableMobs) == 0 {
		return
	}

	currentTime := time.Now()

	if poolReset || currentTime.Sub(pool.lastMobSpawnTime).Milliseconds() >= 7000 {
		mobCapacity := pool.mobCapMin

		if len(pool.activeMobCtrl) > (mobCapacity / 2) {
			if len(pool.activeMobCtrl) < (2 * mobCapacity) {
				mobCapacity = pool.mobCapMin + (pool.mobCapMax-pool.mobCapMin)*(2*len(pool.activeMobCtrl)-pool.mobCapMin)/(3*pool.mobCapMax)
			} else {
				mobCapacity = pool.mobCapMax
			}
		}

		mobCount := mobCapacity - len(pool.mobs)
		if mobCount <= 0 {
			return
		}

		activePos := make([]pos, 0, len(pool.mobs))
		for _, v := range pool.mobs {
			activePos = append(activePos, v.pos)
		}

		mobsToSpawn := []monster{}

		for _, spwnMob := range pool.spawnableMobs {
			if spwnMob.spawnInterval == 0 { // normal mobs: boundary check
				rct := lifePoolRectangle{
					ax: spwnMob.pos.x - 100,
					ay: spwnMob.pos.y + 100,
					bx: spwnMob.pos.x + 100,
					by: spwnMob.pos.y - 100,
				}

				add := true
				for _, pos := range activePos {
					if rct.pointInRect(pos.x, pos.y) {
						add = false
						break
					}
				}

				if add {
					if id, err := pool.nextMobID(); err == nil {
						spwnMob.spawnID = id
						mobsToSpawn = append(mobsToSpawn, spwnMob)
					}
				}
			} else if spwnMob.spawnInterval > 0 || poolReset { // boss mobs or reset
				active := false
				for _, k := range pool.mobs {
					if k.id == spwnMob.id {
						active = true
						break
					}
				}

				if !active && currentTime.After(spwnMob.timeToSpawn) {
					mobsToSpawn = append(mobsToSpawn, spwnMob)
				}
			}
		}

		for mobCount > 0 && len(mobsToSpawn) > 0 {
			ind := pool.rNumber.Intn(len(mobsToSpawn))
			newMob := mobsToSpawn[ind]
			pool.spawnMob(&newMob, false)

			mobsToSpawn[ind] = mobsToSpawn[len(mobsToSpawn)-1]
			mobsToSpawn = mobsToSpawn[:len(mobsToSpawn)-1]

			mobCount--
		}

		pool.lastMobSpawnTime = currentTime
	}
}

func (pool *lifePool) getMobFromID(mobID int32) (monster, error) {
	if m, ok := pool.mobs[mobID]; ok {
		return *m, nil
	}

	return monster{}, fmt.Errorf("Could not find mob with ID %d", mobID)
}

type roomPool struct {
	instance *fieldInstance
	rooms    map[int32]roomer
	poolID   int32
}

func createNewRoomPool(inst *fieldInstance) roomPool {
	return roomPool{instance: inst, rooms: make(map[int32]roomer)}
}

func (pool *roomPool) nextID() (int32, error) {
	for i := 0; i < 100; i++ { // Try 99 times to generate an ID if first time fails
		pool.poolID++

		if pool.poolID == math.MaxInt32-1 {
			pool.poolID = math.MaxInt32 / 2
		} else if pool.poolID == 0 {
			pool.poolID = 1
		}

		if _, ok := pool.rooms[pool.poolID]; !ok {
			return pool.poolID, nil
		}
	}

	return 0, fmt.Errorf("No space to generate ID in drop pool")
}

func (pool *roomPool) playerShowRooms(plr *Player) {
	for _, r := range pool.rooms {
		if game, valid := r.(gameRoomer); valid {
			plr.Send(packetMapShowGameBox(game.displayBytes()))
		}
	}
}

func (pool *roomPool) addRoom(r roomer) error {
	id, err := pool.nextID()

	if err != nil {
		return err
	}

	r.setID(id)

	pool.rooms[id] = r

	pool.updateGameBox(r)

	return nil
}

func (pool *roomPool) removeRoom(id int32) error {
	if _, ok := pool.rooms[id]; !ok {
		return fmt.Errorf("Could not delete room as ID was not found")
	}

	if _, valid := pool.rooms[id].(gameRoomer); valid {
		pool.instance.send(packetMapRemoveGameBox(pool.rooms[id].ownerID()))
	}

	delete(pool.rooms, id)

	return nil
}

func (pool roomPool) getRoom(id int32) (roomer, error) {
	if _, ok := pool.rooms[id]; !ok {
		return nil, fmt.Errorf("Could not retrieve room as ID was not found")
	}

	return pool.rooms[id], nil
}

func (pool roomPool) getPlayerRoom(id int32) (roomer, error) {
	for _, r := range pool.rooms {
		if r.present(id) {
			return r, nil
		}
	}

	return nil, fmt.Errorf("no room with ID")
}

func (pool roomPool) updateGameBox(r roomer) {
	if game, valid := r.(gameRoomer); valid {
		pool.instance.send(packetMapShowGameBox(game.displayBytes()))
	}
}

func (pool *roomPool) removePlayer(plr *Player) {
	r, err := pool.getPlayerRoom(plr.ID)

	if err != nil {
		return
	}

	if game, valid := r.(gameRoomer); valid {
		game.kickPlayer(plr, 0x0)

		if r.closed() {
			pool.removeRoom(r.id())
		}
	}
}

const (
	dropTimeoutNonOwner      = 0
	dropTimeoutNonOwnerParty = 1
	dropFreeForAll           = 2
	dropExplosiveFreeForAll  = 3 // e.g. ludi pq extra stage boxes
)

type fieldDrop struct {
	ID      int32
	ownerID int32
	partyID int32

	mesos int32
	item  Item

	expireTime  int64
	timeoutTime int64
	neverExpire bool

	originPos pos
	finalPos  pos

	dropType byte
}

const (
	dropSpawnDisappears      = 0 // disappears as it is thrown in the air
	dropSpawnNormal          = 1
	dropSpawnShow            = 2
	dropSpawnFadeAtTopOfDrop = 3
)

type dropPool struct {
	instance *fieldInstance
	poolID   int32
	drops    map[int32]fieldDrop // If this struct doesn't stay static change to a ptr
	rates    *rates
}

func createNewDropPool(inst *fieldInstance, rates *rates) dropPool {
	return dropPool{instance: inst, drops: make(map[int32]fieldDrop), rates: rates}
}

func (pool *dropPool) nextID() (int32, error) {
	for i := 0; i < 100; i++ { // Try 99 times to generate an ID if first time fails
		pool.poolID++

		if pool.poolID == math.MaxInt32-1 {
			pool.poolID = math.MaxInt32 / 2
		} else if pool.poolID == 0 {
			pool.poolID = 1
		}

		if _, ok := pool.drops[pool.poolID]; !ok {
			return pool.poolID, nil
		}
	}

	return 0, fmt.Errorf("No space to generate ID in drop pool")
}

func (pool dropPool) canPause() bool {
	return len(pool.drops) == 0
}

func (pool dropPool) playerShowDrops(plr *Player) {
	for _, drop := range pool.drops {
		plr.Send(packetShowDrop(dropSpawnShow, drop, true))
	}
}

func (pool *dropPool) removeDrop(dropType int8, id ...int32) {
	for _, id := range id {
		pool.instance.send(packetRemoveDrop(dropType, id, 0))

		if _, ok := pool.drops[id]; ok {
			delete(pool.drops, id)
		}
	}
}

func (pool *dropPool) eraseDrops() {
	pool.drops = make(map[int32]fieldDrop)
}

func (pool *dropPool) clearDrops() {
	for id, _ := range pool.drops {
		pool.instance.send(packetRemoveDrop(0, id, 0))

		if _, ok := pool.drops[id]; ok {
			delete(pool.drops, id)
		}
	}
}

func (pool *dropPool) playerAttemptPickup(drop fieldDrop, player *Player, pickupType int8) {
	var amount int16

	pool.instance.send(packetRemoveDrop(pickupType, drop.ID, player.ID))

	if drop.mesos > 0 {
		amount = int16(pool.rates.mesos * float32(drop.mesos))
	} else {
		amount = drop.item.amount
	}

	player.Send(packetPickupNotice(drop.item.ID, amount, drop.mesos > 0, drop.item.invID == 1.0))
	delete(pool.drops, drop.ID)
}

func (pool *dropPool) findDropFromID(dropID int32) (error, fieldDrop) {
	drop, ok := pool.drops[dropID]

	if !ok {
		return errors.New("unavailable drop"), fieldDrop{}
	}

	return nil, drop
}

const itemDistance = 20 // Between 15 and 20?
const itemDisppearTimeout = time.Minute * 2
const itemLootableByAllTimeout = time.Minute * 1

func (pool *dropPool) createDrop(spawnType byte, dropType byte, mesos int32, dropFrom pos, expire bool, ownerID, partyID int32, items ...Item) {
	iCount := len(items)
	var offset int16 = 0

	if mesos > 0 {
		iCount++
	}

	if iCount > 0 {
		offset = int16(itemDistance * (iCount / 2))
	}

	now := time.Now()
	expireTime := now.Add(itemDisppearTimeout).UnixMilli()
	var timeoutTime int64 = 0

	if dropType == dropTimeoutNonOwner || dropType == dropTimeoutNonOwnerParty {
		timeoutTime = now.Add(itemLootableByAllTimeout).UnixMilli()
	}
	if len(items) > 0 {
		for i, item := range items {
			tmp := dropFrom
			tmp.x = dropFrom.x - offset + int16(i*itemDistance)
			finalPos := pool.instance.calculateFinalDropPos(tmp)

			if poolID, err := pool.nextID(); err == nil {
				drop := fieldDrop{
					ID:      poolID,
					ownerID: ownerID,
					partyID: partyID,
					mesos:   0,
					item:    item,

					expireTime:  expireTime,
					timeoutTime: timeoutTime,
					neverExpire: false,

					originPos: dropFrom,
					finalPos:  finalPos,

					dropType: dropType,
				}

				pool.drops[drop.ID] = drop

				pool.instance.send(packetShowDrop(spawnType, drop, true))

				d := drop
				time.AfterFunc(5*time.Second, func() {
					if _, ok := pool.drops[d.ID]; !ok {
						return
					}
					pool.instance.reactorPool.tryTriggerByDrop(d)
				})
			}
		}
	}

	if mesos > 0 {
		tmp := dropFrom

		if iCount > 1 {
			tmp.x = tmp.x - offset + int16((iCount-1)*itemDistance)
		}

		finalPos := pool.instance.calculateFinalDropPos(tmp)

		if poolID, err := pool.nextID(); err == nil {
			drop := fieldDrop{
				ID:      poolID,
				ownerID: ownerID,
				partyID: partyID,
				mesos:   mesos,

				expireTime:  expireTime,
				timeoutTime: timeoutTime,
				neverExpire: false,

				originPos: dropFrom,
				finalPos:  finalPos,

				dropType: dropType,
			}

			pool.drops[drop.ID] = drop

			pool.instance.send(packetShowDrop(spawnType, drop, true))
		}
	}
}

func (pool *dropPool) update(t time.Time) {
	id := make([]int32, 0, len(pool.drops))

	currentTime := time.Now().UnixMilli()

	for _, v := range pool.drops {
		if v.expireTime <= currentTime {
			id = append(id, v.ID)
		}
	}

	if len(id) > 0 {
		pool.removeDrop(0, id...)
	}
}

func (pool dropPool) HideDrops(plr *Player) {
	for id := range pool.drops {
		plr.Send(packetRemoveDrop(1, id, 0))
	}
}

// Reactors

type fieldReactor struct {
	spawnID    int32
	templateID int32
	state      byte
	frameDelay int16
	pos        pos
	faceLeft   bool
	name       string

	info        nx.ReactorInfo
	reactorTime int32
}

type reactorPool struct {
	instance *fieldInstance
	reactors map[int32]*fieldReactor
	nextID   int32
}

func createNewReactorPool(inst *fieldInstance, data []nx.Reactor) reactorPool {
	pool := reactorPool{
		instance: inst,
		reactors: make(map[int32]*fieldReactor),
	}

	for _, r := range data {
		if r.ID == 0 {
			log.Println("Reactor skipped: templateID=0, map:", inst.fieldID, "name:", r.Name)
			continue
		}

		info, err := nx.GetReactorInfo(int32(r.ID))
		if err != nil {
			log.Println("Reactor skipped: template not found:", r.ID, "map:", inst.fieldID, "name:", r.Name, "err:", err)
			continue
		}

		id, err := pool.nextReactorID()
		if err != nil {
			continue
		}

		p := pos{x: int16(r.X), y: int16(r.Y)}
		pool.reactors[id] = &fieldReactor{
			spawnID:    id,
			templateID: int32(r.ID),
			state:      0, // default initial state
			frameDelay: 0,
			pos:        p,
			faceLeft:   r.FaceLeft != 0,
			name:       r.Name,
			info:       info,
			reactorTime: func() int32 {
				if r.ReactorTime > math.MaxInt32 {
					return math.MaxInt32
				}
				return int32(r.ReactorTime)
			}(),
		}
	}

	return pool
}

func (pool *reactorPool) nextReactorID() (int32, error) {
	for i := 0; i < 100; i++ {
		pool.nextID++

		if pool.nextID == math.MaxInt32-1 {
			pool.nextID = math.MaxInt32 / 2
		} else if pool.nextID == 0 {
			pool.nextID = 1
		}

		if _, ok := pool.reactors[pool.nextID]; !ok {
			return pool.nextID, nil
		}
	}
	return 0, fmt.Errorf("no space to generate ID in reactor pool")
}

func (pool *reactorPool) playerShowReactors(plr *Player) {
	for _, r := range pool.reactors {
		plr.Send(packetMapReactorEnterField(r.spawnID, r.templateID, r.state, r.pos.x, r.pos.y, r.faceLeft))
	}
}

func (pool *reactorPool) Reset(send bool) {
	for _, r := range pool.reactors {
		r.state = 0
		r.frameDelay = 0
		if send {
			pool.instance.send(packetMapReactorChangeState(r.spawnID, r.state, r.pos.x, r.pos.y, r.frameDelay, r.faceLeft, 0))
		}
	}
}

type reactorTableEntry map[string]interface{}

var reactorTable map[string]map[string]reactorTableEntry

func populateReactorTable(reactorJSON string) error {
	f, err := os.Open(reactorJSON)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &reactorTable)
}

type reactorDrops struct {
	items []int
	money int
}

var reactorDropTable map[string]reactorDrops

func populateReactorDropTable(reactorJSON string) error {
	b, err := os.ReadFile(reactorJSON)
	if err != nil {
		return err
	}

	var raw map[string]map[string]map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	reactorDropTable = make(map[string]reactorDrops, len(raw))
	for rid, slots := range raw {
		rd := reactorDrops{}

		for _, slot := range slots {
			if v, ok := slot["item"].(string); ok {
				if id, _ := strconv.Atoi(v); id != 0 {
					rd.items = append(rd.items, id)
				}
			}
			if v, ok := slot["money"].(string); ok {
				if m, _ := strconv.Atoi(v); m != 0 {
					rd.money = m
				}
			}
		}

		reactorDropTable[rid] = rd
	}

	return nil
}

type rect struct{ left, top, right, bottom int16 }

func (r rect) contains(x, y int16) bool {
	return !(r.right < r.left || r.bottom < r.top) && x >= r.left && x <= r.right && y >= r.top && y <= r.bottom
}

func (r *fieldReactor) nextStateFromTemplate() (byte, bool) {
	cur := int(r.state)
	st, ok := r.info.States[cur]
	if ok && len(st.Events) > 0 {
		ns := int(st.Events[0].State)
		if _, ok2 := r.info.States[ns]; ok2 {
			return byte(ns), true
		}
	}
	if _, ok := r.info.States[cur+1]; ok {
		return byte(cur + 1), true
	}
	return r.state, false
}

func (r *fieldReactor) isTerminal() bool {
	_, ok := r.info.States[int(r.state)+1]
	return !ok
}

func (pool *reactorPool) changeState(r *fieldReactor, next byte, frameDelay int16, cause byte, server *Server, plr *Player) {
	r.state = next
	r.frameDelay = frameDelay
	pool.instance.send(packetMapReactorChangeState(r.spawnID, r.state, r.pos.x, r.pos.y, r.frameDelay, r.faceLeft, cause))
	pool.processStateSideEffects(r, server, plr)
}

func (pool *reactorPool) leaveAndMaybeRespawn(r *fieldReactor, _ int) {
	pool.instance.send(packetMapReactorLeaveField(r.spawnID, r.state, r.pos.x, r.pos.y))
	if r.reactorTime > 0 {
		time.AfterFunc(time.Duration(r.reactorTime)*time.Second, func() {
			r.state = 0
			r.frameDelay = 0
			pool.instance.send(packetMapReactorEnterField(r.spawnID, r.templateID, r.state, r.pos.x, r.pos.y, r.faceLeft))
		})
	}
}

func (pool *reactorPool) triggerHit(spawnID int32, cause byte, server *Server, plr *Player) {
	r, ok := pool.reactors[spawnID]
	if !ok {
		return
	}
	if next, ok := r.nextStateFromTemplate(); ok && next != r.state {
		pool.changeState(r, next, 0, cause, server, plr)
		if r.isTerminal() {
			pool.leaveAndMaybeRespawn(r, 0)
		}
	}
}

func (pool *reactorPool) tryTriggerByDrop(drop fieldDrop) bool {
	if drop.mesos > 0 {
		return false
	}
	for _, r := range pool.reactors {
		st, has := r.info.States[int(r.state)]
		if !has || len(st.Events) == 0 {
			continue
		}
		ev := st.Events[0]
		if ev.ReqItemID != drop.item.ID {
			continue
		}
		if ev.ReqItemCnt > 0 && int16(ev.ReqItemCnt) != drop.item.amount {
			continue
		}
		if next, okNext := r.nextStateFromTemplate(); okNext && next != r.state {
			pool.changeState(r, next, 0, 0, &Server{}, nil)
			pool.instance.dropPool.removeDrop(0, drop.ID)
			if r.isTerminal() {
				pool.leaveAndMaybeRespawn(r, 0)
			}
			return true
		}
	}
	return false
}

func getInt(e reactorTableEntry, key string, def int) int {
	if v, ok := e[key]; ok && v != nil {
		switch t := v.(type) {
		case float64:
			return int(t)
		case int:
			return t
		case int32:
			return int(t)
		case int64:
			return int(t)
		}
	}
	return def
}

func getString(e reactorTableEntry, key, def string) string {
	if v, ok := e[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func entriesForReactor(r *fieldReactor) []reactorTableEntry {
	groupName := strings.TrimSpace(r.info.Action)
	if groupName == "" {
		groupName = strings.TrimSpace(r.name)
	}
	group, ok := reactorTable[groupName]
	if !ok {
		return nil
	}
	type kv struct {
		n int
		k string
	}
	keys := make([]kv, 0, len(group))
	for k := range group {
		if n, err := strconv.Atoi(k); err == nil {
			keys = append(keys, kv{n: n, k: k})
		}
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].n < keys[j].n })

	out := make([]reactorTableEntry, 0, len(keys))
	for _, it := range keys {
		e := group[it.k]
		if getInt(e, "state", -1) == int(r.state) {
			out = append(out, e)
		}
	}
	return out
}

type reactorWarpInfo struct {
	warpAll bool
	targets []reactorWarpTarget
}
type reactorWarpTarget struct {
	mapID  int32
	portal string
}

func loadReactorWarpData(entry reactorTableEntry) *reactorWarpInfo {
	if entry["type"].(float64) != 0 {
		return nil
	}

	var out reactorWarpInfo
	out.warpAll = int(entry["0"].(float64)) == 1

	for i := 1; ; i += 2 {
		mapKey := strconv.Itoa(i)
		portalKey := strconv.Itoa(i + 1)

		_, ok1 := entry[mapKey]
		_, ok2 := entry[portalKey]
		if !ok1 || !ok2 {
			break
		}

		mapID := int32(getInt(entry, mapKey, 0))
		if mapID == 0 {
			continue
		}

		portalName := getString(entry, portalKey, "")
		if portalName == "" {
			continue
		}

		out.targets = append(out.targets, reactorWarpTarget{
			mapID:  mapID,
			portal: portalName,
		})
	}
	return &out
}

func pickRndMap(warp *reactorWarpInfo) (reactorWarpTarget, error) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := len(warp.targets)
	if n == 0 {
		return reactorWarpTarget{}, errors.New("no warp targets available")
	}
	return warp.targets[rnd.Intn(n)], nil
}

func (pool *reactorPool) processStateSideEffects(r *fieldReactor, server *Server, plr *Player) {
	entries := entriesForReactor(r)
	if len(entries) == 0 {
		return
	}

	for _, e := range entries {
		if msg := getString(e, "message", ""); msg != "" {
			pool.instance.send(packetMessageRedText(msg))
		}

		actionType := getInt(e, "type", -1)

		switch actionType {
		case constant.ReactorWarp:
			var players []*Player

			warpInfo := loadReactorWarpData(e)
			mapToWarpTo, err := pickRndMap(warpInfo)
			if err != nil {
				log.Println(err)
				return
			}

			if warpInfo.warpAll {
				players = append(players, pool.instance.lifePool.instance.players...)
			} else {
				players = append(players, plr)
			}

			for _, player := range players {
				err := server.warpPlayer(player,
					server.fields[mapToWarpTo.mapID],
					portal{name: mapToWarpTo.portal})
				if err != nil {
					log.Println(err)
				}
			}

		case constant.ReactorSpawn:
			mobID := getInt(e, "0", 0)
			if mobID <= 0 {
				continue
			}
			count := getInt(e, "2", 1)
			spawnPos := pool.instance.calculateFinalDropPos(r.pos)
			for i := 0; i < count; i++ {
				_ = pool.instance.lifePool.spawnMobFromID(int32(mobID), spawnPos, false, true, true, 0)
			}
		case constant.ReactorDrop:
			reactorID := strconv.Itoa(int(r.info.ID))
			reactorDrops := reactorDropTable[reactorID]
			var items []Item
			for _, val := range reactorDrops.items {
				newItem, err := CreateItemFromID(int32(val), 1)
				if err != nil {
					log.Println(err)
					continue
				}
				items = append(items, newItem)
			}

			pool.instance.dropPool.createDrop(dropSpawnNormal, dropFreeForAll, int32(reactorDrops.money), r.pos, true, 0, 0, items...)

		case constant.ReactorSpawnNPC:
		case constant.ReactorRunScript:

		default:
			continue
		}
	}
}

func packetNpcShow(npc *npc) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcShow)
	p.WriteInt32(npc.spawnID)
	p.WriteInt32(npc.id)
	p.WriteInt16(npc.pos.x)
	p.WriteInt16(npc.pos.y)

	p.WriteBool(!npc.faceLeft)

	p.WriteInt16(npc.pos.foothold)
	p.WriteInt16(npc.rx0)
	p.WriteInt16(npc.rx1)

	return p
}

func packetNpcRemove(npcID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcRemove)
	p.WriteInt32(npcID)

	return p
}

func packetMobShow(mob *monster) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelShowMob)
	p.Append(mob.displayBytes())

	return p
}

func packetMobMove(mobID int32, allowedToUseSkill bool, action int8, skillData uint32, moveBytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMoveMob)
	p.WriteInt32(mobID)
	p.WriteBool(allowedToUseSkill)
	p.WriteInt8(action)
	p.WriteUint32(skillData)
	p.WriteBytes(moveBytes)

	return p

}

func packetMobRemove(spawnID int32, deathType byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRemoveMob)
	p.WriteInt32(spawnID)
	p.WriteByte(deathType)

	return p
}

func packetMobShowBossHP(mobID, hp, maxHP int32, colourFg, colourBg byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMapEffect) // field effect
	p.WriteByte(5)                                             // 1, tremble effect, 3 - mapEffect (string), 4 - mapSound (string), arbitary - environemnt change int32 followed by string
	p.WriteInt32(mobID)
	p.WriteInt32(hp)
	p.WriteInt32(maxHP)
	p.WriteByte(colourFg)
	p.WriteByte(colourBg)

	return p
}

func packetMapShowGameBox(displayBytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoomBox)
	p.WriteBytes(displayBytes)

	return p
}

func packetMapRemoveGameBox(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoomBox)
	p.WriteInt32(charID)
	p.WriteInt32(0)

	return p
}

func packetShowDrop(spawnType byte, drop fieldDrop, allowPet bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelDrobEnterMap)
	p.WriteByte(spawnType) // 0 = disappears on land, 1 = normal drop, 2 = show drop, 3 = fade at top of drop
	p.WriteInt32(drop.ID)

	if drop.mesos > 0 {
		p.WriteByte(1)
		p.WriteInt32(drop.mesos)
	} else {
		p.WriteByte(0)
		p.WriteInt32(drop.item.ID)
	}

	p.WriteInt32(drop.ownerID)
	p.WriteByte(drop.dropType) // drop type 0 = timeout for non owner, 1 = timeout for non-owner party, 2 = free for all, 3 = explosive free for all
	p.WriteInt16(drop.finalPos.x)
	p.WriteInt16(drop.finalPos.y)

	if drop.dropType == dropTimeoutNonOwner {
		p.WriteInt32(drop.ownerID)
	} else {
		p.WriteInt32(0)
	}

	if spawnType != dropSpawnShow {
		p.WriteInt16(drop.originPos.x)        // drop from x
		p.WriteInt16(drop.originPos.y)        // drop from y
		p.WriteInt16(drop.originPos.foothold) // foothold
	}

	if drop.mesos == 0 {
		p.WriteByte(0)    // ?
		p.WriteByte(0x80) // constants to indicate it's for Item
		p.WriteByte(0x05)

		if drop.neverExpire {
			p.WriteInt32(400967355)
			p.WriteByte(2)
		} else {
			// drop.expireTime is in milliseconds; protocol expects minutes since a base epoch
			p.WriteInt32(int32((drop.expireTime - 946681229830) / 1000 / 60))
			p.WriteByte(1)
		}
	}

	p.WriteBool(allowPet)

	return p
}

func packetRemoveDrop(dropType int8, dropID int32, lootedBy int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelDropExitMap)
	p.WriteInt8(dropType) // 0 - fade away, 1 - instant, 2 - loot by user, 3 - loot by mob, 4 - explode, 5 - loot by pet
	p.WriteInt32(dropID)
	p.WriteInt32(lootedBy)

	return p
}

func packetPickupNotice(itemID int32, amount int16, isMesos bool, isEquip bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteInt8(0) // This is the value in switch statement in client for "onMessage" function

	p.WriteBool(isMesos)

	if isMesos {
		p.WriteInt32(int32(amount))
		p.WriteInt16(0) // Internet Cafe Bonus
	} else {
		p.WriteInt32(itemID)

		if !isEquip {
			p.WriteInt16(amount)
		}

		p.WriteInt32(0)
	}
	return p
}

func packetDropNotAvailable() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteInt8(0)
	p.WriteInt8(-2)

	return p
}

func packetInventoryFull() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteInt8(0)
	p.WriteInt8(-1)

	return p
}
