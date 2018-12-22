package nx

import (
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Hucaru/gonx"
)

// PlayerSkill data from nx
type PlayerSkill struct {
	Mastery               int64
	Mad, Mdd, Pad, Pdd    int64
	Hp, Mp, HpCon, MpCon  int64
	BulletConsume         int64
	MoneyConsume          int64
	ItemCon               int64
	ItemConNo             int64
	Time                  int64
	Eva, Acc, Jump, Speed int64
	Range                 int64
	MobCount              int64
	AttackCount           int64
	Damage                int64
	Fixdamage             int64
	Rb, Lt                gonx.Vector
	Hs                    string
	X, Y, Z               int64
	Prop                  int64
	BulletCount           int64
	Action                string
}

// MobSkill data from nx
type MobSkill struct {
	HP              int64
	Limit, Interval int64
	MobID           []int64
}

func extractSkills(nodes []gonx.Node, textLookup []string) (map[int32][]PlayerSkill, map[int32][]MobSkill) {
	playerSkills := make(map[int32][]PlayerSkill)
	mobSkills := make(map[int32][]MobSkill)

	search := "/Skill"

	valid := gonx.FindNode(search, nodes, textLookup, func(node *gonx.Node) {
		for i := uint32(0); i < uint32(node.ChildCount); i++ {
			skillSectionNode := nodes[node.ChildID+i]
			name := textLookup[skillSectionNode.NameID]

			if _, err := strconv.Atoi(strings.TrimSuffix(name, filepath.Ext(name))); err != nil {
				mobSkillSearch := search + "/" + name
				skillIDs := []string{}

				valid := gonx.FindNode(mobSkillSearch, nodes, textLookup, func(node *gonx.Node) {
					for j := uint32(0); j < uint32(node.ChildCount); j++ {
						skillNode := nodes[node.ChildID+j]
						skillIDs = append(skillIDs, textLookup[skillNode.NameID])
					}
				})

				for _, s := range skillIDs {
					valid = gonx.FindNode(mobSkillSearch+"/"+s+"/level", nodes, textLookup, func(node *gonx.Node) {
						skillID, err := strconv.Atoi(s)

						if err != nil {
							return
						}

						mobSkills[int32(skillID)] = make([]MobSkill, node.ChildCount)

						for j := uint32(0); j < uint32(node.ChildCount); j++ {
							skillNode := nodes[node.ChildID+j]
							skillLevel := textLookup[skillNode.NameID]
							level, err := strconv.Atoi(skillLevel)

							if err == nil {
								mobSkills[int32(skillID)][level-1] = getMobSkill(&skillNode, nodes, textLookup)
							}
						}
					})
				}

				if !valid {
					log.Println("Invalid node search:", mobSkillSearch)
				}
			} else {
				playerSkillSearch := search + "/" + name + "/skill"
				skillIDs := []string{}

				valid := gonx.FindNode(playerSkillSearch, nodes, textLookup, func(node *gonx.Node) {
					for j := uint32(0); j < uint32(node.ChildCount); j++ {
						skillNode := nodes[node.ChildID+j]
						skillIDs = append(skillIDs, textLookup[skillNode.NameID])
					}
				})

				for _, s := range skillIDs {
					valid = gonx.FindNode(playerSkillSearch+"/"+s+"/level", nodes, textLookup, func(node *gonx.Node) {
						skillID, err := strconv.Atoi(s)

						if err != nil {
							return
						}

						playerSkills[int32(skillID)] = make([]PlayerSkill, node.ChildCount)

						for j := uint32(0); j < uint32(node.ChildCount); j++ {
							skillNode := nodes[node.ChildID+j]
							skillLevel := textLookup[skillNode.NameID]
							level, err := strconv.Atoi(skillLevel)

							if err == nil {
								playerSkills[int32(skillID)][level-1] = getPlayerSkill(&skillNode, nodes, textLookup)
							}
						}
					})
				}

				if !valid {
					log.Println("Invalid node search:", playerSkillSearch)
				}
			}

		}
	})

	if !valid {
		log.Println("Invalid node search:", search)
	}

	return playerSkills, mobSkills
}

func getPlayerSkill(node *gonx.Node, nodes []gonx.Node, textLookup []string) PlayerSkill {
	skill := PlayerSkill{}

	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		option := nodes[node.ChildID+i]
		optionName := textLookup[option.NameID]

		switch optionName {
		case "mad":
			skill.Mad = gonx.DataToInt64(option.Data)
		case "mdd":
			skill.Mdd = gonx.DataToInt64(option.Data)
		case "pad":
			skill.Pad = gonx.DataToInt64(option.Data)
		case "pdd":
			skill.Pdd = gonx.DataToInt64(option.Data)
		case "hp":
			skill.Hp = gonx.DataToInt64(option.Data)
		case "mp":
			skill.Mp = gonx.DataToInt64(option.Data)
		case "hpCon":
			skill.HpCon = gonx.DataToInt64(option.Data)
		case "mpCon":
			skill.MpCon = gonx.DataToInt64(option.Data)
		case "bulletConsume":
			skill.BulletConsume = gonx.DataToInt64(option.Data)
		case "moneyCon":
			skill.MoneyConsume = gonx.DataToInt64(option.Data)
		case "itemCon":
			skill.ItemCon = gonx.DataToInt64(option.Data)
		case "itemConNo":
			skill.ItemConNo = gonx.DataToInt64(option.Data)
		case "mastery":
			skill.Mastery = gonx.DataToInt64(option.Data)
		case "time":
			skill.Time = gonx.DataToInt64(option.Data)
		case "eva":
			skill.Eva = gonx.DataToInt64(option.Data)
		case "acc":
			skill.Acc = gonx.DataToInt64(option.Data)
		case "jump":
			skill.Jump = gonx.DataToInt64(option.Data)
		case "speed":
			skill.Speed = gonx.DataToInt64(option.Data)
		case "range":
			skill.Range = gonx.DataToInt64(option.Data)
		case "mobCount":
			skill.MobCount = gonx.DataToInt64(option.Data)
		case "attackCount":
			skill.AttackCount = gonx.DataToInt64(option.Data)
		case "damage":
			skill.Damage = gonx.DataToInt64(option.Data)
		case "fixdamage":
			skill.Fixdamage = gonx.DataToInt64(option.Data)
		case "rb":
			skill.Rb = gonx.DataToVector(option.Data)
		case "hs":
			skill.Hs = textLookup[gonx.DataToUint32(option.Data)]
		case "lt":
			skill.Lt = gonx.DataToVector(option.Data)
		case "x":
			skill.X = gonx.DataToInt64(option.Data)
		case "y":
			skill.Y = gonx.DataToInt64(option.Data)
		case "z":
			skill.Z = gonx.DataToInt64(option.Data)
		case "prop":
			skill.Prop = gonx.DataToInt64(option.Data)
		case "ball":
		case "hit":
		case "bulletCount":
			skill.BulletCount = gonx.DataToInt64(option.Data)
		case "action":
			skill.Action = textLookup[gonx.DataToUint32(option.Data)]
		case "58": //?
		default:
			log.Println("Unsupported NX player skill option:", optionName, "->", option.Data)
		}
	}

	return skill
}

func getMobSkill(node *gonx.Node, nodes []gonx.Node, textLookup []string) MobSkill {
	skill := MobSkill{}

	for i := uint32(0); i < uint32(node.ChildCount); i++ {
		option := nodes[node.ChildID+i]
		optionName := textLookup[option.NameID]

		switch optionName {
		case "hp":
		case "interval":
		case "limit":
		case "summonEffect":
		case "time":
		case "mpCon":

		// ?
		case "0":
		case "1":
		case "2":
		case "3":
		case "4":
		case "5":

		// not sure what these are used for
		case "lt":
		case "rb":
		case "effect":
		case "x":
		case "y":
		case "tile":
		case "prop":
		case "affected":
		case "mob":
		case "mob0":
		default:
			log.Println("Unsupported NX mob skill option:", optionName, "->", option.Data)
		}
	}

	return skill
}
