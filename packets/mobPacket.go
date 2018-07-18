package packets

import (
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/nx"
)

func MobShow(mob mobInter, isNewSpawn bool) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_SHOW_MOB)
	p.Append(addMob(mob, isNewSpawn))

	return p
}

func MobControl(mob mobInter, isNewSpawn bool) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_CONTROL_MOB)
	p.WriteByte(0x01) // flag for end control or not

	p.Append(addMob(mob, isNewSpawn))

	return p
}

func addMob(mob mobInter, isNewSpawn bool) maplepacket.Packet {
	p := maplepacket.NewPacket()

	p.WriteInt32(mob.GetSpawnID())
	p.WriteByte(0x01) // control status?
	p.WriteInt32(mob.GetID())

	p.WriteInt32(0) // some kind of status?

	p.WriteInt16(mob.GetX())
	p.WriteInt16(mob.GetY())

	var bitfield byte

	if mob.GetSummoner() != nil {
		bitfield = 0x08
	} else {
		bitfield = 0x02
	}

	if mob.GetState()%2 == 1 {
		bitfield |= 0x01
	} else {
		bitfield |= 0
	}

	if mob.GetFlySpeed() > 0 {
		bitfield |= 0x04
	}

	p.WriteByte(bitfield)           // 0x08 - a summon, 0x04 - flying, 0x02 - ???, 0x01 - faces left
	p.WriteInt16(mob.GetFoothold()) // foothold to oscillate around
	p.WriteInt16(mob.GetFoothold()) // spawn foothold

	if mob.GetSummoner() != nil {
		p.WriteByte(nx.GetMobSummonType(mob.GetID()))
	} else {
		if isNewSpawn {
			p.WriteByte(0xFE)
		} else {
			p.WriteByte(0xFF)
		}
	}

	p.WriteInt32(0)

	return p
}

func MobAck(mobID int32, moveID int16, allowedToUseSkill bool, mp int16, skill byte, level byte) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_CONTROL_MOB_ACK)
	p.WriteInt32(mobID)
	p.WriteInt16(moveID)
	p.WriteBool(allowedToUseSkill)
	p.WriteInt16(mp)
	p.WriteByte(skill)
	p.WriteByte(level)

	return p
}

func MobMove(mobID int32, allowedToUseSkill bool, activity, skill, level byte, option int16, buf []byte) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_MOVE_MOB)
	p.WriteInt32(mobID)
	p.WriteBool(allowedToUseSkill)
	p.WriteByte(activity)
	p.WriteByte(skill)
	p.WriteByte(level)
	p.WriteInt16(option)
	p.WriteBytes(buf)

	return p

}

func MobEndControl(mob mobInter) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_CONTROL_MOB)
	p.WriteByte(0)
	p.WriteInt32(mob.GetSpawnID())

	return p
}

func MobRemove(mob mobInter, deathType byte) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_REMOVE_MOB)
	p.WriteInt32(mob.GetSpawnID())
	p.WriteByte(deathType)

	return p
}

func MobShowHpChange(spawnID int32, dmg int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_MOB_CHANGE_HP)
	p.WriteInt32(spawnID)
	p.WriteByte(0)
	p.WriteInt32(dmg)

	return p
}
