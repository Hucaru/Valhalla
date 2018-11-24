package packet

import (
	"github.com/Hucaru/Valhalla/consts/opcodes"
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func MobShow(mob def.Mob) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelShowMob)
	p.Append(addMob(mob))

	return p
}

func MobControl(mob def.Mob) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelControlMob)
	p.WriteByte(0x01) // flag for end control or not

	p.Append(addMob(mob))

	return p
}

func addMob(mob def.Mob) maplepacket.Packet {
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

	if mob.FacesLeft {
		bitfield |= 0x01
	} else {
		bitfield |= 0x04
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
		p.WriteInt32(mob.SummonOption) // some sort of summoning options, not sure what this is
	}

	p.WriteInt32(0) // encode mob status

	return p
}

func MobControlAcknowledge(mobID int32, moveID int16, allowedToUseSkill bool, mp int16, skill byte, level byte) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelControlMobAck)
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
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelMoveMob)
	p.WriteInt32(mobID)
	p.WriteBool(allowedToUseSkill)
	p.WriteByte(action)
	p.WriteUint32(unknownData)
	p.WriteBytes(buf)

	return p

}

func MobEndControl(mob def.Mob) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelControlMob)
	p.WriteByte(0)
	p.WriteInt32(mob.SpawnID)

	return p
}

func MobRemove(mob def.Mob, deathType byte) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelRemoveMob)
	p.WriteInt32(mob.SpawnID)
	p.WriteByte(deathType)

	return p
}

func MobShowHpChange(spawnID int32, dmg int32) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelMobChangeHP)
	p.WriteInt32(spawnID)
	p.WriteByte(0)
	p.WriteInt32(dmg)

	return p
}
