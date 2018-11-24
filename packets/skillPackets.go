package packets

import (
	"github.com/Hucaru/Valhalla/consts/opcodes"
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func SkillMelee(char def.Character, attackData def.AttackData) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelPlayerUseMeleeSkill)
	p.WriteInt32(char.ID)
	p.WriteByte(attackData.Targets*0x10 + attackData.Hits)
	p.WriteByte(attackData.SkillLevel)

	if attackData.SkillLevel != 0 {
		p.WriteInt32(attackData.SkillID)
	}

	if attackData.FacesLeft {
		p.WriteByte(attackData.Action | (1 << 7))
	} else {
		p.WriteByte(attackData.Action | 0)
	}

	p.WriteByte(attackData.AttackType)

	p.WriteByte(char.Skills[attackData.SkillID].Mastery) // mastery
	p.WriteInt32(attackData.StarID)                      // starID

	for _, info := range attackData.AttackInfo {
		p.WriteInt32(info.SpawnID)
		p.WriteByte(info.HitAction)

		if attackData.IsMesoExplosion {
			p.WriteByte(byte(len(info.Damages)))
		}

		for _, dmg := range info.Damages {
			p.WriteInt32(dmg)
		}
	}

	return p
}

func SkillRanged(char def.Character, attackData def.AttackData) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelPlayerUseRangedSkill)
	p.WriteInt32(char.ID)
	p.WriteByte(attackData.Targets*0x10 + attackData.Hits)
	p.WriteByte(attackData.SkillLevel)

	if attackData.SkillLevel != 0 {
		p.WriteInt32(attackData.SkillID)
	}

	if attackData.FacesLeft {
		p.WriteByte(attackData.Action | (1 << 7))
	} else {
		p.WriteByte(attackData.Action | 0)
	}

	p.WriteByte(attackData.AttackType)

	p.WriteByte(char.Skills[attackData.SkillID].Mastery) // mastery
	p.WriteInt32(attackData.StarID)                      // starID

	for _, info := range attackData.AttackInfo {
		p.WriteInt32(info.SpawnID)
		p.WriteByte(info.HitAction)

		for _, dmg := range info.Damages {
			p.WriteInt32(dmg)
		}
	}

	return p
}

func SkillMagic(char def.Character, attackData def.AttackData) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelPlayerUseMagicSkill)
	p.WriteInt32(char.ID)
	p.WriteByte(attackData.Targets*0x10 + attackData.Hits)
	p.WriteByte(attackData.SkillLevel)

	if attackData.SkillLevel != 0 {
		p.WriteInt32(attackData.SkillID)
	}

	if attackData.FacesLeft {
		p.WriteByte(attackData.Action | (1 << 7))
	} else {
		p.WriteByte(attackData.Action | 0)
	}

	p.WriteByte(attackData.AttackType)

	p.WriteByte(char.Skills[attackData.SkillID].Mastery) // mastery
	p.WriteInt32(attackData.StarID)                      // starID

	for _, info := range attackData.AttackInfo {
		p.WriteInt32(info.SpawnID)
		p.WriteByte(info.HitAction)

		for _, dmg := range info.Damages {
			p.WriteInt32(dmg)
		}
	}

	return p
}

func SkillAnimation(charID int32, skillID int32, level byte) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelPlayerAnimation)
	p.WriteInt32(charID)
	p.WriteByte(0x01)
	p.WriteInt32(skillID)
	p.WriteByte(level)

	return p
}

func SkillGmHide(isHidden bool) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(opcodes.SendChannelEmployee)
	p.WriteByte(0x0F)
	p.WriteBool(isHidden)

	return p
}
