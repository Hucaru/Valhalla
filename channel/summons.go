package channel

import (
	"time"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mpacket"
)

// Summon represents a player-owned summon/puppet instance.
type summon struct {
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
	puppet *summon
	summon *summon
}

func packetShowSummon(ownerID int32, su *summon) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSpecialMapObjectSpawn)
	p.WriteInt32(ownerID)
	p.WriteInt32(su.SkillID)
	p.WriteByte(su.Level)
	p.WriteInt16(su.Pos.x)
	p.WriteInt16(su.Pos.y)
	p.WriteByte(su.Stance)
	p.WriteInt16(su.Foothold)
	p.WriteBool(!su.IsPuppet)
	p.WriteBool(false)
	return p
}

func packetRemoveSummon(ownerID int32, summonID int32, reason byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSpecialMapObjectRemove)
	p.WriteInt32(ownerID)
	p.WriteInt32(summonID)
	p.WriteByte(reason)
	p.WriteInt64(0)
	p.WriteInt64(0)
	p.WriteInt64(0)
	return p
}

func packetSummonMove(ownerID int32, summonID int32, moveBytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSummonMove)
	p.WriteInt32(ownerID)
	p.WriteInt32(summonID)
	p.WriteBytes(moveBytes)
	p.WriteInt64(0)
	p.WriteInt64(0)
	p.WriteInt64(0)
	return p
}

func packetSummonAttack(ownerID int32, summonID int32, anim byte, targets byte, mobDamages map[int32][]int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSummonAttack)
	p.WriteInt32(ownerID)
	p.WriteInt32(summonID)
	p.WriteByte(anim)
	p.WriteByte(targets)
	for mobID, dList := range mobDamages {
		p.WriteInt32(mobID)
		p.WriteByte(constant.SummonAttackMob)
		for _, d := range dList {
			p.WriteInt32(d)
		}
	}
	p.WriteInt64(0)
	p.WriteInt64(0)
	p.WriteInt64(0)
	return p
}

func packetSummonDamage(ownerID int32, summonID int32, damage int32, mobID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSummonDamage)
	p.WriteInt32(ownerID)
	p.WriteInt32(summonID)
	p.WriteByte(constant.SummonTakeDamage)
	p.WriteInt32(damage)
	p.WriteInt32(mobID)
	p.WriteByte(0)
	p.WriteInt64(0)
	p.WriteInt64(0)
	p.WriteInt64(0)
	return p
}
