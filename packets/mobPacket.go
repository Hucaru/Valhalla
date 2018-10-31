package packets

import (
	"github.com/Hucaru/Valhalla/consts/opcodes"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/types"
)

func MobShow(mob types.Mob) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.Send.ChannelShowMob)
	p.Append(addMob(mob))

	return p
}

func MobControl(mob types.Mob) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.Send.ChannelControlMob)
	p.WriteByte(0x01) // flag for end control or not

	p.Append(addMob(mob))

	return p
}

func addMob(mob types.Mob) maplepacket.Packet {
	p := maplepacket.NewPacket()

	p.WriteInt32(mob.SpawnID)
	p.WriteByte(0x01) // control status?
	p.WriteInt32(mob.ID)

	p.WriteInt32(0) // some kind of status?

	p.WriteInt16(mob.X)
	p.WriteInt16(mob.Y)

	var bitfield byte

	if mob.Summoner != nil {
		bitfield = 0x08
	} else {
		bitfield = 0x02
	}

	if mob.Stance%2 == 1 {
		bitfield |= 0x01
	} else {
		bitfield |= 0
	}

	if mob.FlySpeed > 0 {
		bitfield |= 0x04
	}

	p.WriteByte(bitfield)      // 0x08 - a summon, 0x04 - flying, 0x02 - ???, 0x01 - faces left
	p.WriteInt16(mob.Foothold) // foothold to oscillate around
	p.WriteInt16(mob.Foothold) // spawn foothold
	p.WriteInt8(mob.SummonType)

	if mob.SummonType == -3 || mob.SummonType >= 0 {
		p.WriteInt32(0) // some sort of summoning options, not sure what this is
	}

	p.WriteInt32(0) // encode mob status

	return p
}

func MobControlAcknowledge(mobID int32, moveID int16, allowedToUseSkill bool, mp int16, skill byte, level byte) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.Send.ChannelControlMobAck)
	p.WriteInt32(mobID)
	p.WriteInt16(moveID)
	p.WriteBool(allowedToUseSkill)
	p.WriteInt16(mp)
	p.WriteByte(skill)
	p.WriteByte(level)

	return p
}

func MobMove(mobID int32, allowedToUseSkill bool, action byte, unknownData uint32, buf []byte) maplepacket.Packet {
	// func MobMove(mobID int32, allowedToUseSkill bool, action int8, skill, level byte, option int16, buf []byte) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.Send.ChannelMoveMob)
	p.WriteInt32(mobID)
	p.WriteBool(allowedToUseSkill)
	p.WriteByte(action)
	p.WriteUint32(unknownData)
	p.WriteBytes(buf)

	return p

}

func MobEndControl(mob types.Mob) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.Send.ChannelControlMob)
	p.WriteByte(0)
	p.WriteInt32(mob.SpawnID)

	return p
}

func MobRemove(mob mobInter, deathType byte) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.Send.ChannelRemoveMob)
	p.WriteInt32(mob.GetSpawnID())
	p.WriteByte(deathType)

	return p
}

func MobShowHpChange(spawnID int32, dmg int32) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.Send.ChannelMobChangeHP)
	p.WriteInt32(spawnID)
	p.WriteByte(0)
	p.WriteInt32(dmg)

	return p
}
