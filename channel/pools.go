package channel

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

type lifePoolRectangle struct {
	ax, ay int16
	bx, by int16
}

func (r lifePoolRectangle) pointInRect(x, y int16) bool {
	// Since rectangle will always be orientated as follows the check is simple
	/*
		 ----------A
		 |   P    |
		 |        |
		B----------
	*/

	if r.ax < x {
		return false
	} else if r.ay < y {
		return false
	} else if r.bx > x {
		return false
	} else if r.by > y {
		return false
	}

	return true
}

type lifePool struct {
	instance *fieldInstance

	npcs map[int32]*npc
	mobs map[int32]*monster

	spawnableMobs []monster

	mobID, npcID         int32
	lastMobSpawnTime     time.Time
	mobCapMin, mobCapMax int

	activeMobCtrl map[*player]bool

	dropPool *dropPool

	rNumber *rand.Rand
}

func creatNewLifePool(inst *fieldInstance, npcData, mobData []nx.Life, mobCapMin, mobCapMax int) lifePool {
	pool := lifePool{instance: inst, activeMobCtrl: make(map[*player]bool)}

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
	for i := 0; i < 100; i++ { // Try 99 times to generate an id if first time fails
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

	return 0, fmt.Errorf("No space to generate id in drop pool")
}

func (pool *lifePool) nextNpcID() (int32, error) {
	for i := 0; i < 100; i++ { // Try 99 times to generate an id if first time fails
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

	return 0, fmt.Errorf("No space to generate id in drop pool")
}

func (pool lifePool) canClose() bool {
	return false
}

func (pool lifePool) getNPCFromSpawnID(id int32) (npc, error) {
	for _, v := range pool.npcs {
		if v.spawnID == id {
			return *v, nil
		}
	}

	return npc{}, fmt.Errorf("Could not find npc with id %d", id)
}

func (pool *lifePool) addPlayer(plr *player) {
	for i, npc := range pool.npcs {
		plr.send(packetNpcShow(npc))

		if npc.controller == nil {
			pool.npcs[i].setController(plr)
		}
	}

	for i, m := range pool.mobs {
		plr.send(packetMobShow(m))

		if m.controller == nil {
			pool.mobs[i].setController(plr, false)
		}

		pool.showMobBossHPBar(m, plr)
	}
}

func (pool *lifePool) removePlayer(plr *player) {
	for i, v := range pool.npcs {
		if v.controller.conn == plr.conn {
			pool.npcs[i].removeController()

			// find new controller
			if plr := pool.instance.findController(); plr != nil {
				if cont, ok := plr.(*player); ok {
					pool.npcs[i].setController(cont)
				}
			}
		}
	}

	for i, v := range pool.mobs {
		if v.controller != nil && v.controller.conn == plr.conn {
			pool.mobs[i].removeController()

			// find new controller
			if plr := pool.instance.findController(); plr != nil {
				if cont, ok := plr.(*player); ok {
					pool.mobs[i].setController(cont, false)
				}
			}
		}

		plr.send(packetMobRemove(v.spawnID, 0x0)) // need to tell client to remove mobs for instance swapping
	}

	delete(pool.activeMobCtrl, plr)
}

func (pool *lifePool) npcAcknowledge(poolID int32, plr *player, data []byte) {
	for i := range pool.npcs {
		if poolID == pool.npcs[i].spawnID {
			pool.npcs[i].acknowledgeController(plr, pool.instance, data)
			return
		}
	}

}

func (pool *lifePool) mobAcknowledge(poolID int32, plr *player, moveID int16, skillPossible bool, action byte, skillData uint32, moveData movement, finalData movementFrag, moveBytes []byte) {
	for i, v := range pool.mobs {
		mob := pool.mobs[i]

		if poolID == v.spawnID && v.controller.conn == plr.conn {
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
			pool.instance.sendExcept(packetMobMove(poolID, skillPossible, action, skillData, moveBytes), v.controller.conn)
		}
	}
}

func (pool *lifePool) mobDamaged(poolID int32, damager *player, dmg ...int32) {
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

					var partyExp int32 = 0

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
						plr.party.giveExp(plr.id, partyExp, true)
					}
				}

				// quest mob logic

				// on die logic
				for _, id := range v.revives {
					spawnID, err := pool.nextMobID()

					if err != nil {
						continue
					}

					newMob, err := createMonsterFromID(spawnID, int32(id), v.pos, nil, true, true)

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

				if dropEntry, ok := dropTable[v.id]; ok {
					var mesos int32
					drops := make([]item, 0, len(dropEntry))

					for _, entry := range dropEntry {
						chance := pool.rNumber.Int63n(100000)
						if int64(pool.dropPool.rates.drop*float32(entry.Chance)) < chance {
							continue
						}

						if entry.IsMesos {
							mesos = pool.rNumber.Int31n(entry.Max-entry.Min) + entry.Min
						} else {
							var amount int16 = 1

							if entry.Max != 1 {
								val := pool.rNumber.Int31n(entry.Max-entry.Min) + entry.Min

								if val > math.MaxInt16 {
									amount = math.MaxInt16
								} else {
									amount = int16(val)
								}
							}

							newItem, err := createItemFromID(entry.ItemID, amount)

							if err != nil {
								log.Println("Failed to create drop for mobID:", v.id, "with error:", err)
								continue
							}

							drops = append(drops, newItem)
						}
					}

					// TODO: droppool type determination between DropTimeoutNonOwner and DropTimeoutNonOwnerParty
					pool.dropPool.createDrop(dropSpawnNormal, dropFreeForAll, mesos, v.pos, true, 0, 0, drops...)

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

func (pool *lifePool) killMobs(deathType byte) {
	// Need to collect keys first as when iterating over the map and killing we will kill any subsequent spawns depending on map iteration order
	keys := make([]int32, 0, len(pool.mobs))

	for key := range pool.mobs {
		keys = append(keys, key)
	}

	for _, key := range keys {
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
		delete(pool.mobs, key)
	}
}

func (pool *lifePool) spawnMob(m *monster, hasAgro bool) bool {
	pool.mobs[m.spawnID] = m
	pool.instance.send(packetMobShow(m))

	if plr := pool.instance.findController(); plr != nil {
		if cont, ok := plr.(*player); ok {
			for _, v := range pool.mobs {
				v.setController(cont, hasAgro)
			}
		}
	}

	pool.showMobBossHPBar(m, nil)
	m.summonType = -1

	return false
}

func (pool *lifePool) spawnMobFromID(mobID int32, location pos, hasAgro, items, mesos bool) error {
	id, err := pool.nextMobID()

	if err != nil {
		return err
	}

	m, err := createMonsterFromID(id, mobID, location, nil, items, mesos)

	if err != nil {
		return err
	}

	pool.spawnMob(&m, hasAgro)

	return nil
}

func (pool *lifePool) spawnReviveMob(m *monster, cont *player) {
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

func (pool lifePool) showMobBossHPBar(mob *monster, plr *player) {
	if plr != nil {
		if show, mobID, hp, maxHP, hpFgColour, hpBgColour := mob.hasHPBar(); show {
			plr.send(packetMobShowBossHP(mobID, hp, maxHP, hpFgColour, hpBgColour))
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

		activePos := make([]pos, len(pool.mobs))
		mobsToSpawn := []monster{}
		boundaryCheck := false
		count := 0

		index := 0
		for _, v := range pool.mobs {
			activePos[index] = v.pos
			index++
		}

		for _, spwnMob := range pool.spawnableMobs {
			if spwnMob.spawnInterval == 0 { // normal mobs
				boundaryCheck = true
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

			if boundaryCheck {
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
					id, err := pool.nextMobID()

					if err == nil {
						spwnMob.spawnID = id
						mobsToSpawn = append(mobsToSpawn, spwnMob)
					}
				}

				boundaryCheck = false
			}

			count++
			if count >= len(pool.spawnableMobs) {
				break
			}
		}

		for mobCount > 0 && len(mobsToSpawn) > 0 {
			ind := rand.Intn(len(mobsToSpawn))
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

	return monster{}, fmt.Errorf("Could not find mob with id %d", mobID)
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
	for i := 0; i < 100; i++ { // Try 99 times to generate an id if first time fails
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

	return 0, fmt.Errorf("No space to generate id in drop pool")
}

func (pool *roomPool) playerShowRooms(plr *player) {
	for _, r := range pool.rooms {
		if game, valid := r.(gameRoomer); valid {
			plr.send(packetMapShowGameBox(game.displayBytes()))
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
		return fmt.Errorf("Could not delete room as id was not found")
	}

	if _, valid := pool.rooms[id].(gameRoomer); valid {
		pool.instance.send(packetMapRemoveGameBox(pool.rooms[id].ownerID()))
	}

	delete(pool.rooms, id)

	return nil
}

func (pool roomPool) getRoom(id int32) (roomer, error) {
	if _, ok := pool.rooms[id]; !ok {
		return nil, fmt.Errorf("Could not retrieve room as id was not found")
	}

	return pool.rooms[id], nil
}

func (pool roomPool) getPlayerRoom(id int32) (roomer, error) {
	for _, r := range pool.rooms {
		if r.present(id) {
			return r, nil
		}
	}

	return nil, fmt.Errorf("no room with id")
}

func (pool roomPool) updateGameBox(r roomer) {
	if game, valid := r.(gameRoomer); valid {
		pool.instance.send(packetMapShowGameBox(game.displayBytes()))
	}
}

func (pool *roomPool) removePlayer(plr *player) {
	r, err := pool.getPlayerRoom(plr.id)

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

type dropSet byte

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
	item  item

	expireTime  int64
	timeoutTime int64
	neverExpire bool

	originPos pos
	finalPos  pos

	dropType byte
}

const (
	dropSpawnDisappears      = 0
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
	for i := 0; i < 100; i++ { // Try 99 times to generate an id if first time fails
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

	return 0, fmt.Errorf("No space to generate id in drop pool")
}

func (pool dropPool) canClose() bool {
	return false
}

func (pool dropPool) playerShowDrops(plr *player) {
	for _, drop := range pool.drops {
		plr.send(packetShowDrop(dropSpawnShow, drop))
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

func (pool *dropPool) playerAttemptPickup(drop fieldDrop, player *player) {
	var amount int16

	pool.instance.send(packetRemoveDrop(2, drop.ID, player.id))

	if drop.mesos > 0 {
		amount = int16(pool.rates.mesos * float32(drop.mesos))
	} else {
		amount = drop.item.amount
	}

	player.send(packetPickupNotice(drop.item.id, amount, drop.mesos > 0, drop.item.invID == 1.0))
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

func (pool *dropPool) createDrop(spawnType byte, dropType byte, mesos int32, dropFrom pos, expire bool, ownerID, partyID int32, items ...item) {
	iCount := len(items)
	var offset int16 = 0

	if mesos > 0 {
		iCount++
	}

	if iCount > 0 {
		offset = int16(itemDistance * (iCount / 2))
	}

	currentTime := time.Now()
	expireTime := currentTime.Add(itemDisppearTimeout).Unix()
	var timeoutTime int64 = 0

	if dropType == dropTimeoutNonOwner || dropType == dropTimeoutNonOwnerParty {
		timeoutTime = currentTime.Add(itemLootableByAllTimeout).Unix()
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

				pool.instance.send(packetShowDrop(spawnType, drop))
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

			pool.instance.send(packetShowDrop(spawnType, drop))
		}
	}
}

func (pool dropPool) HideDrops(plr *player) {
	for id := range pool.drops {
		plr.send(packetRemoveDrop(1, id, 0))
	}
}

func (pool *dropPool) update(t time.Time) {
	id := make([]int32, 0, len(pool.drops))

	currentTime := time.Now().Unix()

	for _, v := range pool.drops {
		if v.expireTime <= currentTime {
			id = append(id, v.ID)
		}
	}

	if len(id) > 0 {
		pool.removeDrop(0, id...)
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

func packetMobMove(mobID int32, allowedToUseSkill bool, action byte, skillData uint32, moveBytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMoveMob)
	p.WriteInt32(mobID)
	p.WriteBool(allowedToUseSkill)
	p.WriteByte(action)
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

func packetShowDrop(spawnType byte, drop fieldDrop) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelDrobEnterMap)
	p.WriteByte(spawnType) // 0 = disappears on land, 1 = normal drop, 2 = show drop, 3 = fade at top of drop
	p.WriteInt32(drop.ID)

	if drop.mesos > 0 {
		p.WriteByte(1)
		p.WriteInt32(drop.mesos)
	} else {
		p.WriteByte(0)
		p.WriteInt32(drop.item.id)
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
		p.WriteByte(0x80) // constants to indicate it's for item
		p.WriteByte(0x05)

		if drop.neverExpire {
			p.WriteInt32(400967355)
			p.WriteByte(2)
		} else {
			p.WriteInt32(int32((drop.expireTime - 946681229830) / 1000 / 60)) // TODO: figure out what time this is for
			p.WriteByte(1)
		}
	}

	p.WriteByte(0) // Did player drop it, used by pet with equip?

	return p
}

func packetRemoveDrop(dropType int8, dropID int32, lootedBy int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelDropExitMap)
	p.WriteInt8(dropType) // 0 - fade away, 1 - instant, 2 - loot by user, 3 - loot by mob
	p.WriteInt32(dropID)
	p.WriteInt32(lootedBy)

	return p
}

func packetPickupNotice(itemID int32, amount int16, isMesos bool, isEquip bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteInt8(0) //??

	p.WriteBool(isMesos)

	if isMesos {
		p.WriteInt32(int32(amount))
		p.WriteInt16(0)
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
