package game

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

func PacketMobShow(mob Mob) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelShowMob)
	p.Append(addMob(mob))

	return p
}

func PacketMobControl(mob Mob, chase bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelControlMob)
	if chase {
		p.WriteByte(0x02) // 2 chase, 1 no chase, 0 no control
	} else {
		p.WriteByte(0x01)
	}

	p.Append(addMob(mob))

	return p
}

func addMob(mob Mob) mpacket.Packet {
	p := mpacket.NewPacket()

	p.WriteInt32(mob.SpawnID)
	p.WriteByte(0x00) // control status?
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

	if mob.FaceLeft {
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

func PacketMobControlAcknowledge(mobID int32, moveID int16, allowedToUseSkill bool, mp int16, skill byte, level byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelControlMobAck)
	p.WriteInt32(mobID)
	p.WriteInt16(moveID)
	p.WriteBool(allowedToUseSkill)
	p.WriteInt16(mp)
	p.WriteByte(skill)
	p.WriteByte(level)

	return p
}

func PacketMobMove(mobID int32, allowedToUseSkill bool, action byte, unknownData uint32, buf []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMoveMob)
	p.WriteInt32(mobID)
	p.WriteBool(allowedToUseSkill)
	p.WriteByte(action)
	p.WriteUint32(unknownData)
	p.WriteBytes(buf)

	return p

}

func PacketMobEndControl(mob Mob) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelControlMob)
	p.WriteByte(0)
	p.WriteInt32(mob.SpawnID)

	return p
}

func PacketMobRemove(mob Mob, deathType byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRemoveMob)
	p.WriteInt32(mob.SpawnID)
	p.WriteByte(deathType)

	return p
}

func PacketMobShowHpChange(spawnID int32, dmg int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMobChangeHP)
	p.WriteInt32(spawnID)
	p.WriteByte(0)
	p.WriteInt32(dmg)

	return p
}
