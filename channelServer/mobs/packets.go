package mobs

import (
	"github.com/Hucaru/Valhalla/common/nx"
	"github.com/Hucaru/gopacket"
)

func ShowMob(spawnID uint32, mob nx.Life, isNewSpawn bool) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(0x86)
	p.WriteUint32(spawnID)
	p.WriteByte(0x01)
	p.WriteUint32(mob.ID)

	p.WriteUint32(0) //?

	p.WriteInt16(mob.X)
	p.WriteInt16(mob.Y)
	p.WriteByte(0x02) // 0x08 and 0x02 denote something about ownership
	p.WriteInt16(mob.Fh)
	p.WriteUint16(0)

	if isNewSpawn {
		p.WriteByte(0xFF)
	} else {
		p.WriteByte(0xFE)
	}

	p.WriteUint32(0)

	return p
}

func ControlMob(spawnID uint32, mob nx.Life, isNewSpawn bool) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(0x88)
	p.WriteByte(0x01)
	p.WriteUint32(spawnID)
	p.WriteByte(0x01)
	p.WriteUint32(mob.ID)

	p.WriteUint32(0) // ?

	p.WriteInt16(mob.X)
	p.WriteInt16(mob.Y)
	p.WriteByte(0x02)
	p.WriteInt16(mob.Fh)
	p.WriteUint16(0)

	if isNewSpawn {
		p.WriteByte(0xFF)
	} else {
		p.WriteByte(0xFE)
	}
	p.WriteUint32(0)

	return p
}

func ControlMoveMob(mobID uint32, moveID uint16, useSkill byte, mp uint16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(0x8F)
	p.WriteUint32(mobID)
	p.WriteUint16(moveID)
	p.WriteByte(useSkill)
	p.WriteUint16(mp)

	return p
}

func MoveMob(mobID uint32, useSkill byte, skill byte, buf []byte) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(0x8A)
	p.WriteUint32(mobID)
	p.WriteByte(useSkill)
	p.WriteByte(skill)
	p.WriteInt32(0)
	p.WriteBytes(buf)

	return p

}
