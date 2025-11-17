package channel

import (
	"log"
	"math"
	"time"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant/skill"
	"github.com/Hucaru/Valhalla/mpacket"
)

// fieldMist represents a poison mist on the field
type fieldMist struct {
	ID           int32
	ownerID      int32
	skillID      int32
	skillLevel   byte
	box          mistBox
	createdAt    time.Time
	duration     int64
	isPoisonMist bool
	magicAttack  int16
}

// mistBox defines the rectangular area of a mist
type mistBox struct {
	x1, y1 int16
	x2, y2 int16
}

// mistPool manages all mists in a field instance
type mistPool struct {
	instance *fieldInstance
	poolID   int32
	mists    map[int32]*fieldMist
}

func createNewMistPool(inst *fieldInstance) mistPool {
	return mistPool{
		instance: inst,
		mists:    make(map[int32]*fieldMist),
	}
}

func (pool *mistPool) nextID() int32 {
	for i := 0; i < 100; i++ {
		pool.poolID++
		if pool.poolID == math.MaxInt32-1 {
			pool.poolID = math.MaxInt32 / 2
		} else if pool.poolID == 0 {
			pool.poolID = 1
		}

		if _, ok := pool.mists[pool.poolID]; !ok {
			return pool.poolID
		}
	}
	return 0
}

// createMist spawns a new mist on the field
func (pool *mistPool) createMist(ownerID, skillID int32, skillLevel byte, pos pos, duration int64, isPoisonMist bool, magicAttack int16) *fieldMist {
	mistID := pool.nextID()
	if mistID == 0 {
		log.Println("Mist: Failed to generate mist ID")
		return nil
	}

	const mistWidth int16 = 150
	const mistHeight int16 = 100

	mist := &fieldMist{
		ID:         mistID,
		ownerID:    ownerID,
		skillID:    skillID,
		skillLevel: skillLevel,
		box: mistBox{
			x1: pos.x - mistWidth,
			y1: pos.y - mistHeight,
			x2: pos.x + mistWidth,
			y2: pos.y + mistHeight,
		},
		createdAt:    time.Now(),
		duration:     duration,
		isPoisonMist: isPoisonMist,
		magicAttack:  magicAttack,
	}

	pool.mists[mistID] = mist
	pool.instance.send(packetMistSpawn(mist))

	if duration > 0 {
		go func() {
			time.Sleep(time.Duration(duration) * time.Second)
			pool.instance.dispatch <- func() {
				pool.removeMist(mistID)
			}
		}()
	}

	return mist
}

// removeMist removes a mist from the field
func (pool *mistPool) removeMist(mistID int32) {
	if mist, ok := pool.mists[mistID]; ok {
		pool.instance.send(packetMistRemove(mistID, mist.isPoisonMist))
		delete(pool.mists, mistID)
	}
}

// playerShowMists shows all active mists to a player joining the map
func (pool mistPool) playerShowMists(plr *Player) {
	for _, mist := range pool.mists {
		plr.Send(packetMistSpawn(mist))
	}
}

// isInMist checks if a position is within a mist's area
func (m *fieldMist) isInMist(p pos) bool {
	return p.x >= m.box.x1 && p.x <= m.box.x2 && p.y >= m.box.y1 && p.y <= m.box.y2
}

func packetMistSpawn(mist *fieldMist) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelAffectedAreaCreate)
	p.WriteInt32(mist.ID)
	p.WriteBool(false)
	p.WriteInt32(mist.skillID)
	p.WriteByte(mist.skillLevel)
	p.WriteInt16(0) // delay
	p.WriteInt32(int32(mist.box.x1))
	p.WriteInt32(int32(mist.box.y1))
	p.WriteInt32(int32(mist.box.x2))
	p.WriteInt32(int32(mist.box.y2))

	return p
}

func packetMistRemove(mistID int32, isPoisonMist bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelAffectedAreaRemove)
	p.WriteInt32(mistID)

	return p
}

// startPoisonMistTicker applies poison buff to mobs entering the mist area
func (server *Server) startPoisonMistTicker(inst *fieldInstance, mist *fieldMist) {
	if !mist.isPoisonMist || inst == nil {
		return
	}

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		defer ticker.Stop()

		endTime := mist.createdAt.Add(time.Duration(mist.duration) * time.Second)

		for range ticker.C {
			if time.Now().After(endTime) {
				return
			}

			if _, exists := inst.mistPool.mists[mist.ID]; !exists {
				return
			}

			inst.dispatch <- func() {
				now := time.Now()
				remain := endTime.Sub(now)
				remainSec := int16(remain / time.Second)
				if remainSec < 1 {
					remainSec = 1
				}

				for _, mob := range inst.lifePool.mobs {
					if mob == nil || mob.hp <= 0 {
						continue
					}
					if mist.isInMist(mob.pos) {
						if (mob.statBuff & skill.MobStat.Poison) == 0 {
							mob.applyBuff(mist.skillID, mist.skillLevel, skill.MobStat.Poison, inst)
						}

						if mob.buffs != nil {
							if b, ok := mob.buffs[skill.MobStat.Poison]; ok && b != nil {
								b.ownerID = mist.ownerID
								b.duration = remainSec
								b.expiresAt = now.Add(time.Duration(remainSec) * time.Second).UnixMilli()

								inst.send(packetMobStatSet(mob.spawnID, skill.MobStat.Poison, b.value, b.skillID, b.duration, 0))
							}
						}
					}
				}
			}
		}
	}()
}
