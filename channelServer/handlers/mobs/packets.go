package mobs

import (
	"github.com/Hucaru/gopacket"
)

func SpawnMob() gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(0x86)
	p.WriteUint32(0)      // mob spawn id
	p.WriteByte(0x01)     // ?
	p.WriteUint32(100101) // mob id

	p.WriteUint32(0) //?

	p.WriteInt16(-187) // mob x
	p.WriteInt16(170)  // mob y
	p.WriteByte(0x02)  // a mob type byte / stance direction - (2 fr, 3 fl, s still fr)
	p.WriteUint16(114) // foothold
	p.WriteUint16(0)
	p.WriteByte(0xFE) // 0xFF (show) or 0xFE (spawn), spawn or display
	p.WriteInt32(0)

	return p
}

func ShowMob() gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(0x86)
	p.WriteUint32(0)      // mob spawn id
	p.WriteByte(0x01)     // ?
	p.WriteUint32(100101) // mob id

	p.WriteUint32(0) //?

	p.WriteInt16(-187) // mob x
	p.WriteInt16(170)  // mob y
	p.WriteByte(0x02)  // a mob type byte / stance direction - (2 fr, 3 fl, s still fr)
	p.WriteUint16(114) // foothold
	p.WriteUint16(0)   // ?
	p.WriteByte(0xFF)  // 0xFF (show) or 0xFE (spawn), spawn or display
	p.WriteInt32(0)

	return p
}

func ControlMob() gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(0x88)
	p.WriteByte(0x01)     // ?
	p.WriteUint32(0)      // mob spawn id
	p.WriteByte(0x01)     // ?
	p.WriteUint32(100101) // mob id

	p.WriteUint32(0) // ?

	p.WriteInt16(-187) // mob x
	p.WriteInt16(170)  // mob y
	p.WriteByte(0x02)  // a mob type byte / stance direction - (2 fr, 3 fl, s still fr)
	p.WriteUint16(114) // foothold
	p.WriteUint16(0)   // ?
	p.WriteByte(0xFF)  // controls spawn or already there
	p.WriteUint32(0)

	return p
}

func ValidateControlMoveMob() gopacket.Packet {
	p := gopacket.NewPacket()

	p.WriteByte(0x88)
	p.WriteUint32(0)
	p.WriteBool(false)
	p.WriteUint32(0)
	p.WriteByte(0)
	// add parts of fragmentation packet
	return p
}
