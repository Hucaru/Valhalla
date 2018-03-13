package skills

import (
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/gopacket"
)

func standardSkillPacket(charID uint32, skillID uint32, targets, hits, display, animation byte, damages map[uint32][]uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_USE_STANDARD_SKILL)
	p.WriteUint32(charID)
	p.WriteByte(byte(targets*0x10) + hits)
	p.WriteBool(bool(skillID != 0))
	if skillID != 0 {
		p.WriteUint32(skillID)
	}
	p.WriteByte(display)
	p.WriteByte(animation)

	p.WriteByte(0)   // mastery
	p.WriteUint32(0) // starID?

	for k, v := range damages {
		p.WriteUint32(k)
		p.WriteByte(0x6)
		// if meos explosion add, another byte for something
		for _, dmg := range v {
			p.WriteUint32(dmg)
		}
	}

	return p
}

func rangedSkillPacket(charID, skillID, objID uint32, targets, hits, display, animation byte, damages map[uint32][]uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_USE_RANGED_SKILL)
	p.WriteUint32(charID)
	p.WriteByte(targets*0x10 + hits)
	p.WriteBool(bool(skillID != 0))
	if skillID != 0 {
		p.WriteUint32(skillID)
	}
	p.WriteByte(display)
	p.WriteByte(animation)

	p.WriteByte(0)       // mastery
	p.WriteUint32(objID) // starID?

	for k, v := range damages {
		p.WriteUint32(k)
		p.WriteByte(0x6)
		for _, dmg := range v {
			p.WriteUint32(dmg)
		}
	}

	return p
}

func magicSkillPacket(charID uint32, skillID uint32, targets, hits, display, animation byte, damages map[uint32][]uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_USE_MAGIC_SKILL)
	p.WriteUint32(charID)
	p.WriteByte(targets*0x10 + hits)
	p.WriteBool(bool(skillID != 0))
	if skillID != 0 {
		p.WriteUint32(skillID)
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

func skillAnimationPacket(charID uint32, skillID uint32, level byte) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_ANIMATION)
	p.WriteUint32(charID)
	p.WriteByte(0x01)
	p.WriteUint32(skillID)
	p.WriteByte(level)

	return p
}
