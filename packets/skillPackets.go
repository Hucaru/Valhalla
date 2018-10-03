package packets

import (
	"github.com/Hucaru/Valhalla/consts/opcodes"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func SkillMelee(charID int32, skillID int32, targets, hits, display, animation byte, damages map[int32][]int32) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.Send.ChannelPlayerUseStandardSkill)
	p.WriteInt32(charID)
	p.WriteByte(byte(targets*0x10) + hits)
	p.WriteBool(bool(skillID != 0))
	if skillID != 0 {
		p.WriteInt32(skillID)
	}
	p.WriteByte(display)
	p.WriteByte(animation)

	p.WriteByte(0)  // mastery
	p.WriteInt32(0) // starID?

	for k, v := range damages {
		p.WriteInt32(k)
		p.WriteByte(0x6)
		// if meos explosion add, another byte for something
		for _, dmg := range v {
			p.WriteInt32(dmg)
		}
	}

	return p
}

func SkillRanged(charID, skillID, objID int32, targets, hits, display, animation byte, damages map[int32][]int32) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.Send.ChannelPlayerUseRangedSkill)
	p.WriteInt32(charID)
	p.WriteByte(targets*0x10 + hits)
	p.WriteBool(bool(skillID != 0))
	if skillID != 0 {
		p.WriteInt32(skillID)
	}
	p.WriteByte(display)
	p.WriteByte(animation)

	p.WriteByte(0)      // mastery
	p.WriteInt32(objID) // starID?

	for k, v := range damages {
		p.WriteInt32(k)
		p.WriteByte(0x6)
		for _, dmg := range v {
			p.WriteInt32(dmg)
		}
	}

	return p
}

func SkillMagic(charID int32, skillID int32, targets, hits, display, animation byte, damages map[int32][]int32) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.Send.ChannelPlayerUseMagicSkill)
	p.WriteInt32(charID)
	p.WriteByte(targets*0x10 + hits)
	p.WriteBool(bool(skillID != 0))
	if skillID != 0 {
		p.WriteInt32(skillID)
	}
	p.WriteByte(display)
	p.WriteByte(animation)

	p.WriteByte(0)  // mastery
	p.WriteInt32(0) // starID?

	for k, v := range damages {
		p.WriteInt32(k)
		p.WriteByte(0x6)
		for _, dmg := range v {
			p.WriteInt32(dmg)
		}
	}

	return p
}

func SkillAnimation(charID int32, skillID int32, level byte) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.Send.ChannelPlayerAnimation)
	p.WriteInt32(charID)
	p.WriteByte(0x01)
	p.WriteInt32(skillID)
	p.WriteByte(level)

	return p
}

func SkillGmHide(isHidden bool) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.Send.ChannelEmployee)
	p.WriteByte(0x0F)
	p.WriteBool(isHidden)

	return p
}
