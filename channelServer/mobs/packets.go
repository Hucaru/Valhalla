package mobs

import (
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/nx"
	"github.com/Hucaru/gopacket"
)

func showMob(spawnID uint32, mob nx.Life, isNewSpawn bool) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_SHOW_MOB)
	p.WriteUint32(spawnID)
	p.WriteByte(0x01)
	p.WriteUint32(mob.ID)

	p.WriteUint32(0) //?

	p.WriteInt16(mob.X)
	p.WriteInt16(mob.Y)
	p.WriteByte(0x02) // direction / state
	p.WriteInt16(mob.Fh)
	p.WriteInt16(mob.Fh)

	if isNewSpawn {
		p.WriteByte(0xFE)
	} else {
		p.WriteByte(0xFF)
	}

	p.WriteUint32(0)

	return p
}

func controlMob(spawnID uint32, mob nx.Life, isNewSpawn bool) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_CONTROL_MOB)
	p.WriteByte(0x01)
	p.WriteUint32(spawnID)
	p.WriteByte(0x01)
	p.WriteUint32(mob.ID)

	p.WriteUint32(0) // ?

	p.WriteInt16(mob.X)
	p.WriteInt16(mob.Y)
	p.WriteByte(0x02)    // which direction it faces?
	p.WriteInt16(mob.Fh) // foothold to oscillate around
	p.WriteInt16(mob.Fh) // spawn foothold

	if isNewSpawn {
		p.WriteByte(0xFE)
	} else {
		p.WriteByte(0xFF)
	}
	p.WriteUint32(0)

	return p
}

func controlAck(mobID uint32, moveID uint16, useSkill bool, skill byte, level byte, mp uint16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_CONTROL_MOB_ACK)
	p.WriteUint32(mobID)
	p.WriteUint16(moveID)
	p.WriteBool(useSkill)
	p.WriteUint16(mp)
	p.WriteByte(skill)
	p.WriteByte(level)
	p.WriteUint16(0)

	return p
}

func moveMob(mobID uint32, skillUsed bool, skill byte, x int16, y int16, buf []byte) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNE_MOVE_MOB)
	p.WriteUint32(mobID)
	p.WriteBool(skillUsed)
	p.WriteByte(skill)
	p.WriteInt16(x) // a position thing? This is not the mob position info. That is stored in the buf
	p.WriteInt16(y) // a position thing? This is not the mob position info. That is stored in the buf
	p.WriteBytes(buf)

	return p

}

func endControl(mobID uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_CONTROL_MOB)
	p.WriteByte(0)
	p.WriteUint32(mobID)

	return p
}
