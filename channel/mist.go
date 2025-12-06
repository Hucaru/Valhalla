package channel

import (
	"log"
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
	lastTick time.Time
}

func createNewMistPool(inst *fieldInstance) mistPool {
	return mistPool{
		instance: inst,
		mists:    make(map[int32]*fieldMist),
	}
}

func (pool *mistPool) nextID() int32 {
	pool.poolID++
	if pool.poolID == 0 {
		pool.poolID = 1
	}
	return pool.poolID
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

func (pool *mistPool) update(t time.Time) {
	if pool == nil || pool.instance == nil {
		return
	}

	for id, mist := range pool.mists {
		if mist == nil || mist.duration <= 0 {
			continue
		}

		if t.After(mist.createdAt.Add(time.Duration(mist.duration) * time.Second)) {
			pool.removeMist(id)
		}
	}

	if pool.lastTick.IsZero() || t.Sub(pool.lastTick) >= time.Second {
		pool.lastTick = t

		for _, mist := range pool.mists {
			if mist == nil || !mist.isPoisonMist {
				continue
			}

			remain := mist.createdAt.Add(time.Duration(mist.duration) * time.Second).Sub(t)
			remainSec := int16(remain / time.Second)
			if remainSec < 1 {
				remainSec = 1
			}

			for _, mob := range pool.instance.lifePool.mobs {
				if mob == nil || mob.hp <= 0 {
					continue
				}
				if !mist.isInMist(mob.pos) {
					continue
				}

				if (mob.statBuff & skill.MobStat.Poison) == 0 {
					mob.applyBuff(mist.skillID, mist.skillLevel, skill.MobStat.Poison, pool.instance)
				}

				if mob.buffs != nil {
					if b, ok := mob.buffs[skill.MobStat.Poison]; ok && b != nil {
						b.ownerID = mist.ownerID
						b.duration = remainSec
						b.expiresAt = t.Add(time.Duration(remainSec) * time.Second).UnixMilli()
						pool.instance.send(packetMobStatSet(mob.spawnID, skill.MobStat.Poison, b.value, b.skillID, b.duration, 0))
					}
				}
			}
		}
	}
}

func (pool *mistPool) removeMistsByOwner(ownerID int32) {
	if pool == nil {
		return
	}
	for id, mist := range pool.mists {
		if mist != nil && mist.ownerID == ownerID {
			pool.removeMist(id)
		}
	}
}
