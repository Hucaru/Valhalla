package skills

import (
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/gopacket"
)

func skillAnimationPacket(charID uint32, skillID uint32, tByte, targets, hits, display, animation byte, damages map[uint32][]uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_USE_SKILL)
	p.WriteUint32(charID)
	// p.WriteByte(targets*0x10 + hits)
	p.WriteByte(tByte)
	if skillID != 0 {
		p.WriteByte(1)
		p.WriteUint32(skillID)
	} else {
		p.WriteByte(0)
	}
	p.WriteByte(display)
	p.WriteByte(animation)

	p.WriteByte(0)   // mastery
	p.WriteUint32(0) // starID?

	for k, v := range damages {
		p.WriteUint32(k)
		p.WriteByte(0x6)
		for _, dmg := range v {
			p.WriteUint32(dmg)
		}
	}

	return p
}
