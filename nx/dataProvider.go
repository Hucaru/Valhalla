package nx

import (
	"strconv"
)

func GetMobSummonType(mobID uint32) (summonType byte) {
	result := searchNode("Mob/"+strconv.Itoa(int(mobID))+".img/info", func(cursor *node) {
		for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
			option := nodes[cursor.ChildID+i]
			switch strLookup[option.NameID] {
			case "summonType":
				summonType = option.Data[0]
			default:
			}
		}
	})

	if !result {
		return 0
	}

	return summonType
}

type MobSkill struct {
	SkillID, Level      byte
	MpCon, Interval, HP uint32
}

func GetMobSkills(mobID uint32) []MobSkill {
	mobSkills := make([]MobSkill, 0)

	searchNode("Mob/"+strconv.Itoa(int(mobID))+".img/info/skill", func(cursor *node) {
		for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
			skill := nodes[cursor.ChildID+i]
			newSkill := MobSkill{}
			for j := uint32(0); j < uint32(skill.ChildCount); j++ {
				option := nodes[skill.ChildID+j]
				switch strLookup[option.NameID] {
				case "level":
					newSkill.Level = option.Data[0]
				case "skill":
					newSkill.SkillID = option.Data[0]
				default:
				}
			}

			mobSkills = append(mobSkills, newSkill)
		}
	})

	for index, skill := range mobSkills {
		searchNode("Skill/MobSkill.img/"+strconv.Itoa(int(skill.SkillID))+"/level/"+strconv.Itoa(int(skill.Level)), func(cursor *node) {
			for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
				option := nodes[cursor.ChildID+i]
				switch strLookup[option.NameID] {
				case "mpCon":
					mobSkills[index].MpCon = dataToUint32(option.Data)
				case "interval":
					mobSkills[index].Interval = dataToUint32(option.Data)
				case "hp":
					mobSkills[index].HP = dataToUint32(option.Data)
				default:
				}
			}
		})
	}
	return mobSkills
}

type mobAttack struct {
	ConMP uint32
}

func GetMobAttack(mobID uint32, attackID int8) (attack mobAttack, valid bool) {
	valid = true

	result := searchNode("Mob/"+strconv.Itoa(int(mobID))+".img/attack"+strconv.Itoa(int(attackID))+"/info", func(cursor *node) {
		for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
			option := nodes[cursor.ChildID+i]
			switch strLookup[option.NameID] {
			case "conMP":
				attack.ConMP = dataToUint32(option.Data)
			default:
			}
		}
	})

	if !result {
		valid = false
	}

	return attack, valid
}
