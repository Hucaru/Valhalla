package lifepool

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/server/field/droppool"
	"github.com/Hucaru/Valhalla/server/field/lifepool/mob"
	"github.com/Hucaru/Valhalla/server/field/lifepool/npc"
	"github.com/Hucaru/Valhalla/server/item"
	"github.com/Hucaru/Valhalla/server/movement"
	"github.com/Hucaru/Valhalla/server/pos"
)

type field interface {
	Send(mpacket.Packet) error
	SendExcept(mpacket.Packet, mnet.Client) error
	FindController() interface{}
}

type controller interface {
	Send(mpacket.Packet)
	Conn() mnet.Client
}

type player interface {
	controller
	GiveEXP(int32, bool, bool)
	MapID() int32
}

type dropPool interface {
	CreateDrop(byte, byte, int32, pos.Data, bool, int32, int32, ...item.Data)
}

type party interface {
}

type rectangle struct {
	ax, ay int16
	bx, by int16
}

func (r rectangle) pointInRect(x, y int16) bool {
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

// Data structure for pool
type Data struct {
	instance field

	npcs []npc.Data
	mobs []mob.Data

	spawnableMobs []mob.Data

	poolID               int32
	lastMobSpawnTime     time.Time
	mobCapMin, mobCapMax int

	activeMobCtrl map[controller]bool

	dropPool dropPool

	rNumber *rand.Rand
}

// CreatNewPool for life
func CreatNewPool(inst field, npcData, mobData []nx.Life, mobCapMin, mobCapMax int) Data {
	pool := Data{instance: inst, activeMobCtrl: make(map[controller]bool)}

	pool.npcs = make([]npc.Data, len(npcData))

	for i, l := range npcData {
		pool.npcs[i] = npc.CreateFromData(pool.nextID(), l)
	}

	pool.mobs = make([]mob.Data, len(mobData))
	pool.spawnableMobs = make([]mob.Data, len(mobData))

	for i, v := range mobData {
		m, err := nx.GetMob(v.ID)

		if err != nil {
			continue
		}

		pool.mobs[i] = mob.CreateFromData(pool.nextID(), v, m, true, true)
		pool.mobs[i].SetSummonType(-1)

		pool.spawnableMobs[i] = mob.CreateFromData(pool.nextID(), v, m, true, true)
	}

	pool.mobCapMin = mobCapMin
	pool.mobCapMax = mobCapMax

	randomSource := rand.NewSource(time.Now().UTC().UnixNano())
	pool.rNumber = rand.New(randomSource)

	return pool
}

// SetDropPool to use
func (pool *Data) SetDropPool(drop *droppool.Data) {
	pool.dropPool = drop
}

func (pool *Data) nextID() int32 {
	pool.poolID++

	if pool.poolID == math.MaxInt32-1 {
		pool.poolID = math.MaxInt32 / 2
	} else if pool.poolID == 0 {
		pool.poolID = 1
	}

	return pool.poolID
}

// CanClose the pool down
func (pool Data) CanClose() bool {
	return false
}

// GetNPCFromSpawnID - get npc data from spawn id
func (pool Data) GetNPCFromSpawnID(id int32) (npc.Data, error) {
	for _, v := range pool.npcs {
		if v.SpawnID() == id {
			return v, nil
		}
	}

	return npc.Data{}, fmt.Errorf("Could not find npc with id %d", id)
}

// AddPlayer to be added to the pool
func (pool *Data) AddPlayer(plr controller) {
	for i, npc := range pool.npcs {
		plr.Send(packetNpcShow(npc))

		if npc.Controller() == nil {
			pool.npcs[i].SetController(plr)
		}
	}

	for i, m := range pool.mobs {
		plr.Send(packetMobShow(m))

		if m.Controller() == nil {
			pool.mobs[i].SetController(plr, false)
		}

		pool.showMobBossHPBar(m)
	}
}

// RemovePlayer from pool
func (pool *Data) RemovePlayer(plr controller) {
	for i, v := range pool.npcs {
		if v.Controller().Conn() == plr.Conn() {
			pool.npcs[i].RemoveController()

			// find new controller
			if plr := pool.instance.FindController(); plr != nil {
				if cont, ok := plr.(controller); ok {
					pool.npcs[i].SetController(cont)
				}
			}
		}
	}

	for i, v := range pool.mobs {
		if v.Controller() != nil && v.Controller().Conn() == plr.Conn() {
			pool.mobs[i].RemoveController()

			// find new controller
			if plr := pool.instance.FindController(); plr != nil {
				if cont, ok := plr.(controller); ok {
					pool.mobs[i].SetController(cont, false)
				}
			}
		}

		plr.Send(packetMobRemove(v.SpawnID(), 0x0)) // need to tell client to remove mobs for instance swapping
	}

	delete(pool.activeMobCtrl, plr)
}

// NpcAcknowledge bytes to be applied to the pool
func (pool *Data) NpcAcknowledge(poolID int32, plr controller, data []byte) {
	for i := range pool.npcs {
		if poolID == pool.npcs[i].SpawnID() {
			pool.npcs[i].AcknowledgeController(plr, pool.instance, data)
			return
		}
	}

}

// MobAcknowledge bytes to be applied to the pool
func (pool *Data) MobAcknowledge(poolID int32, plr controller, moveID int16, skillPossible bool, action byte, skillData uint32, moveData movement.Data, finalData movement.Frag, moveBytes []byte) {
	for i, v := range pool.mobs {
		if poolID == v.SpawnID() && v.Controller().Conn() == plr.Conn() {
			skillID := byte(skillData)
			skillLevel := byte(skillData >> 8)
			skillDelay := int16(skillData >> 16)

			actualAction := int(byte(action >> 1))

			if action < 0 {
				actualAction = -1
			}

			if actualAction >= 21 && actualAction <= 25 {
				pool.mobs[i].PerformSkill(skillDelay, skillLevel, skillID)
			} else if actualAction > 12 && actualAction < 20 {
				pool.mobs[i].PerformAttack(byte(actualAction - 12))
			}

			if !moveData.ValidateMob(v) {
				return
			}

			pool.mobs[i].AcknowledgeController(moveID, finalData, skillPossible, skillID, skillLevel)
			pool.instance.SendExcept(packetMobMove(poolID, skillPossible, action, skillData, moveBytes), v.Controller().Conn())
		}
	}
}

// MobDamaged handling
func (pool *Data) MobDamaged(poolID int32, damager player, prty party, dmg ...int32) {
	for i, v := range pool.mobs {
		if v.SpawnID() == poolID {
			pool.mobs[i].RemoveController()

			if damager != nil {
				pool.mobs[i].SetController(damager, true)
				pool.activeMobCtrl[damager] = true
				pool.mobs[i].GiveDamage(damager, dmg...)
			} else {
				pool.mobs[i].GiveDamage(nil, dmg...)
			}

			pool.showMobBossHPBar(v)

			if pool.mobs[i].HP() < 1 {
				for cont, dmg := range pool.mobs[i].GetDamage() {
					plr, ok := cont.(player)

					if !ok {
						continue
					}

					if damager != nil && damager.MapID() != plr.MapID() {
						continue
					}

					if dmg == v.MaxHP() {
						plr.GiveEXP(v.Exp(), true, false)
					} else if float64(dmg)/float64(v.MaxHP()) > 0.60 {
						plr.GiveEXP(v.Exp(), true, false)
					} else {
						newExp := int32(float64(v.Exp()) * 0.25)

						if newExp == 0 {
							newExp = 1
						}

						plr.GiveEXP(newExp, true, false)
					}

					if prty != nil {
						// party exp share logic
					}
				}

				// quest mob logic

				// on die logic
				for _, id := range v.Revives() {
					newMob, err := mob.CreateFromID(pool.nextID(), int32(id), v.Pos(), nil, true, true)

					if err != nil {
						log.Println(err)
						continue
					}

					newMob.SetFaceLeft(v.FaceLeft())
					newMob.SetSummonType(-3)
					newMob.SetSummonOption(v.SpawnID())
					pool.spawnReviveMob(newMob, damager)
				}

				if dropEntry, ok := item.DropTable[v.ID()]; ok {
					chance := pool.rNumber.Int31n(100000)

					var mesos int32
					drops := make([]item.Data, 0, len(dropEntry))

					for _, entry := range dropEntry {
						if entry.Chance < chance {
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

							newItem, err := item.CreateFromID(entry.ItemID, amount)

							if err != nil {
								log.Println("Failed to create drop for mobID:", v.ID(), "with error:", err)
								continue
							}

							drops = append(drops, newItem)
						}
					}

					// TODO: droppool type determination between DropTimeoutNonOwner and DropTimeoutNonOwnerParty
					pool.dropPool.CreateDrop(droppool.SpawnNormal, droppool.DropFreeForAll, mesos, v.Pos(), true, 0, 0, drops...)
				}

				pool.removeMob(v.SpawnID(), 0x1)

				if v.SpawnInterval() > 0 {
					for i, k := range pool.spawnableMobs {
						if k.ID() == v.ID() { // if this needs strengthening then add a spawn pos check
							pool.spawnableMobs[i].SetTimeToSpawn(time.Now().Add(time.Millisecond * time.Duration(v.SpawnInterval())))
							break
						}
					}
				}
			}
			break
		}
	}
}

// KillMobs in the pool
func (pool *Data) KillMobs(deathType byte) {
	for _, v := range pool.mobs {
		pool.MobDamaged(v.SpawnID(), nil, nil, v.HP())
	}
}

func (pool *Data) spawnMob(m mob.Data, hasAgro bool) bool {
	pool.mobs = append(pool.mobs, m)
	pool.instance.Send(packetMobShow(m))

	if plr := pool.instance.FindController(); plr != nil {
		if cont, ok := plr.(controller); ok {
			pool.mobs[len(pool.mobs)-1].SetController(cont, hasAgro)
		}
	}

	pool.showMobBossHPBar(m)

	return false
}

// SpawnMobFromID into the pool
func (pool *Data) SpawnMobFromID(mobID int32, location pos.Data, hasAgro, items, mesos bool) error {
	m, err := mob.CreateFromID(pool.nextID(), mobID, location, nil, items, mesos)

	if err != nil {
		return err
	}

	pool.spawnMob(m, hasAgro)

	return nil
}

func (pool *Data) spawnReviveMob(m mob.Data, cont controller) {
	pool.spawnMob(m, true)

	pool.mobs[len(pool.mobs)-1].SetSummonType(-2)
	pool.mobs[len(pool.mobs)-1].SetSummonOption(0)

	if cont != nil {
		pool.mobs[len(pool.mobs)-1].SetController(cont, true)
	}
}

func (pool *Data) removeMob(poolID int32, deathType byte) {
	for i, v := range pool.mobs {
		if v.SpawnID() == poolID {
			pool.mobs = append(pool.mobs[:i], pool.mobs[i+1:]...)
			pool.instance.Send(packetMobRemove(poolID, deathType))
			return
		}
	}
}

// ShowMobBossHPBar to instance if possible
func (pool Data) showMobBossHPBar(mob mob.Data) {
	if show, mobID, hp, maxHP, hpFgColour, hpBgColour := mob.HasHPBar(); show {
		pool.instance.Send(packetMobShowBossHP(mobID, hp, maxHP, hpFgColour, hpBgColour))
	}
}

// Update logic for the pool e.g. mob spawning
func (pool *Data) Update(t time.Time) {
	for i := range pool.mobs {
		pool.mobs[i].Update(t)
	}

	pool.attemptMobSpawn(false)
}

func (pool *Data) attemptMobSpawn(poolReset bool) {
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

		activePos := make([]pos.Data, len(pool.mobs))
		mobsToSpawn := []mob.Data{}
		boundaryCheck := false
		count := 0

		for i, v := range pool.mobs {
			activePos[i] = v.Pos()
		}

		for _, spwnMob := range pool.spawnableMobs {
			if spwnMob.SpawnInterval() == 0 { // normal mobs
				boundaryCheck = true
			} else if spwnMob.SpawnInterval() > 0 || poolReset { // boss mobs or reset
				active := false

				for _, k := range pool.mobs {
					if k.ID() == spwnMob.ID() {
						active = true
						break
					}
				}

				if !active && currentTime.After(spwnMob.TimeToSpawn()) {
					mobsToSpawn = append(mobsToSpawn, spwnMob)
				}
			}

			if boundaryCheck {
				rct := rectangle{
					ax: spwnMob.Pos().X() - 100,
					ay: spwnMob.Pos().Y() + 100,
					bx: spwnMob.Pos().X() + 100,
					by: spwnMob.Pos().Y() - 100,
				}

				add := true
				for _, pos := range activePos {
					if rct.pointInRect(pos.X(), pos.Y()) {
						add = false
						break
					}
				}

				if add {
					spwnMob.SetSpawnID(pool.nextID())
					mobsToSpawn = append(mobsToSpawn, spwnMob)
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
			pool.spawnMob(mobsToSpawn[ind], false)

			mobsToSpawn[ind] = mobsToSpawn[len(mobsToSpawn)-1]
			mobsToSpawn = mobsToSpawn[:len(mobsToSpawn)-1]

			mobCount--
		}

		pool.lastMobSpawnTime = currentTime
	}
}

// GetMobFromID returns the mob data from mobID
func (pool *Data) GetMobFromID(mobID int32) (mob.Data, error) {
	for _, m := range pool.mobs {
		if m.SpawnID() == mobID {
			return m, nil
		}
	}

	return mob.Data{}, fmt.Errorf("Could not find mob with id %d", mobID)
}
