package nx

// import (
// 	"strconv"
// )

// type JobSkill struct {
// 	Level    int32
// 	MobCount int32
// 	Damage   int32
// 	Range    int32
// 	Speed    int32
// 	Jump     int32
// 	Time     int64
// }

// var JobSkills = make(map[int32][]JobSkill)

// type MobSkill struct {
// 	Level               byte
// 	MpCon, Interval, HP int32
// }

// var MobSkills = make(map[int32][]MobSkill)

// func getJobSkills() {
// 	searchNode("Skill", func(cursor *node) {
// 		for i := uint32(0); i < uint32(cursor.ChildCount); i++ {
// 			jobID := nodes[cursor.ChildID+i]
// 			if strLookup[jobID.NameID] == "MobSkill.img" {
// 				continue
// 			}

// 			for j := uint32(0); j < uint32(jobID.ChildCount); j++ {
// 				item := nodes[jobID.ChildID+j]
// 				if strLookup[item.NameID] != "skill" {
// 					continue
// 				}

// 				for k := uint32(0); k < uint32(item.ChildCount); k++ {
// 					skillID := nodes[item.ChildID+j]
// 					id, err := strconv.Atoi(strLookup[skillID.NameID])

// 					if err != nil {
// 						continue
// 					}

// 					for l := uint32(0); l < uint32(skillID.ChildCount); l++ {
// 						if strLookup[item.NameID] != "level" {
// 							continue
// 						}

// 					}

// 					newSkill := JobSkill{}

// 					JobSkills[int32(id)] = append(JobSkills[int32(id)], newSkill)
// 				}
// 			}
// 		}
// 	})
// }

// func getMobSkills() {

// }
