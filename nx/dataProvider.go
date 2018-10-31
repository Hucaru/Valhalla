package nx

import (
	"strconv"
)

func GetMobSummonType(mobID int32) (summonType byte) {
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
	ID, Level byte
	MPCon, HP int32
	Cooldown  int64
}

func GetMobSkills(mobID int32) []MobSkill {
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
					newSkill.ID = option.Data[0]
				default:
				}
			}

			mobSkills = append(mobSkills, newSkill)
		}
	})

	for index, skill := range mobSkills {
		searchNode("Skill/MobSkill.img/"+strconv.Itoa(int(skill.ID))+"/level/"+strconv.Itoa(int(skill.Level)), func(cursor *node) {
			for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
				option := nodes[cursor.ChildID+i]
				switch strLookup[option.NameID] {
				case "mpCon":
					mobSkills[index].MPCon = dataToInt32(option.Data)
				case "interval":
					mobSkills[index].Cooldown = dataToInt64(option.Data)
				case "hp":
					mobSkills[index].HP = dataToInt32(option.Data)
				default:
				}
			}
		})
	}
	return mobSkills
}

type mobAttack struct {
	MPCon     int32
	Coooldown int64
}

func GetMobAttack(mobID int32, attackID byte) (attack mobAttack, valid bool) {
	valid = true

	result := searchNode("Mob/"+strconv.Itoa(int(mobID))+".img/attack"+strconv.Itoa(int(attackID))+"/info", func(cursor *node) {
		for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
			option := nodes[cursor.ChildID+i]
			switch strLookup[option.NameID] {
			case "conMP":
				attack.MPCon = dataToInt32(option.Data)
			case "attackAfter":
				attack.Coooldown = dataToInt64(option.Data)
			default:
			}
		}
	})

	if !result {
		valid = false
	}

	return attack, valid
}
