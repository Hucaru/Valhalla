package channel

import (
	"time"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

// Summon represents a player-owned summon/puppet instance.
type Summon struct {
	OwnerID int32
	SkillID int32
	Level   byte
	HP      int

	Pos       pos
	Stance    byte
	Foothold  int16
	ExpiresAt time.Time

	// Flags
	IsPuppet   bool
	SummonType int32
}

type summonState struct {
	puppet  *Summon
	summon  *Summon
	puTimer *time.Timer
	smTimer *time.Timer
}

func (p *player) ensureSummonState() {
	if p.summons == nil {
		p.summons = &summonState{}
	}
}

// addSummon registers and broadcasts a summon; replaces any existing same-type summon.
func (s *Server) addSummon(p *player, su *Summon, durationSec int) {
	p.ensureSummonState()

	su.ExpiresAt = time.Now().Add(time.Duration(durationSec) * time.Second)

	if su.IsPuppet {
		// Replace current puppet
		if p.summons.puppet != nil {
			s.removeSummon(p, true, 0x04) // replaced
		}
		p.summons.puppet = su

		if p.summons.puTimer != nil {
			p.summons.puTimer.Stop()
		}
		p.summons.puTimer = time.AfterFunc(time.Duration(durationSec)*time.Second, func() {
			s.removeSummon(p, true, 0x02) // expired
		})
	} else {
		// Replace current non-puppet summon
		if p.summons.summon != nil {
			s.removeSummon(p, false, 0x04) // replaced
		}
		p.summons.summon = su

		if p.summons.smTimer != nil {
			p.summons.smTimer.Stop()
		}
		p.summons.smTimer = time.AfterFunc(time.Duration(durationSec)*time.Second, func() {
			s.removeSummon(p, false, 0x02) // expired
		})
	}

	s.broadcastShowSummon(p, su)
}

func (s *Server) removeSummon(p *player, puppet bool, reason byte) {
	p.ensureSummonState()

	if puppet {
		if p.summons.puppet == nil {
			return
		}
		s.broadcastRemoveSummon(p, p.summons.puppet.SkillID, reason)
		p.summons.puppet = nil
		if p.summons.puTimer != nil {
			p.summons.puTimer.Stop()
		}
	} else {
		if p.summons.summon == nil {
			return
		}
		s.broadcastRemoveSummon(p, p.summons.summon.SkillID, reason)
		p.summons.summon = nil
		if p.summons.smTimer != nil {
			p.summons.smTimer.Stop()
		}
	}
}

func (p *player) getSummon(skillID int32) *Summon {
	p.ensureSummonState()
	if p.summons.summon != nil && p.summons.summon.SkillID == skillID {
		return p.summons.summon
	}
	if p.summons.puppet != nil && p.summons.puppet.SkillID == skillID {
		return p.summons.puppet
	}
	return nil
}

func (s *Server) broadcastShowSummon(p *player, su *Summon) {
	field, ok := s.fields[p.mapID]
	if !ok {
		return
	}
	inst, err := field.getInstance(p.inst.id)
	if err != nil {
		return
	}
	inst.send(packetShowSummon(p.id, su))
}

func (s *Server) broadcastRemoveSummon(p *player, summonSkillID int32, reason byte) {
	field, ok := s.fields[p.mapID]
	if !ok {
		return
	}
	inst, err := field.getInstance(p.inst.id)
	if err != nil {
		return
	}
	inst.send(packetRemoveSummon(p.id, summonSkillID, reason))
}

func packetShowSummon(ownerID int32, su *Summon) mpacket.Packet {
	p := mpacket.CreateWithOpcode(0x5F)
	p.WriteByte(0x4A)
	p.WriteInt32(ownerID)
	p.WriteInt32(su.SkillID)
	p.WriteByte(su.Level)
	p.WriteInt16(su.Pos.x)
	p.WriteInt16(su.Pos.y)
	p.WriteByte(su.Stance)
	p.WriteInt16(su.Foothold)
	p.WriteBool(!su.IsPuppet) // true if aggressive/attacking summon, false for puppet
	p.WriteBool(false)        // animated spawn by default (C# sends !animated -> we send false to animate)
	return p
}

func packetRemoveSummon(ownerID int32, summonID int32, reason byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSummonOperation)
	p.WriteByte(0x4B)
	p.WriteInt32(ownerID)
	p.WriteInt32(summonID)
	p.WriteByte(reason)
	p.WriteInt64(0)
	p.WriteInt64(0)
	p.WriteInt64(0)
	return p
}

func packetSummonMove(ownerID int32, summonID int32, moveBytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSummonOperation)
	p.WriteByte(0x4C)
	p.WriteInt32(ownerID)
	p.WriteInt32(summonID)
	p.WriteBytes(moveBytes)
	p.WriteInt64(0)
	p.WriteInt64(0)
	p.WriteInt64(0)
	return p
}

func packetSummonAttack(ownerID int32, summonID int32, anim byte, targets byte, mobDamages map[int32][]int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSummonOperation)
	p.WriteByte(0x4D)
	p.WriteInt32(ownerID)
	p.WriteInt32(summonID)
	p.WriteByte(anim)
	p.WriteByte(targets)
	for mobID, dList := range mobDamages {
		p.WriteInt32(mobID)
		p.WriteByte(0x06)
		for _, d := range dList {
			p.WriteInt32(d)
		}
	}
	p.WriteInt64(0)
	p.WriteInt64(0)
	p.WriteInt64(0)
	return p
}

func packetSummonDamage(ownerID int32, summonID int32, unk int8, damage int32, mobID int32, unk2 byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSummonOperation)
	p.WriteByte(0x4E)
	p.WriteInt32(ownerID)
	p.WriteInt32(summonID)
	p.WriteByte(byte(unk))
	p.WriteInt32(damage)
	p.WriteInt32(mobID)
	p.WriteByte(unk2)
	p.WriteInt64(0)
	p.WriteInt64(0)
	p.WriteInt64(0)
	return p
}
