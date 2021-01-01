package player

import (
	"fmt"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/nx"
)

// Skill data
type Skill struct {
	ID             int32
	Level, Mastery byte
	Cooldown       int16
	CooldownTime   int16
	TimeLastUsed   int64
}

// CreateSkillFromData - creates a player skill for a given id and level
func CreateSkillFromData(ID int32, level byte) (Skill, error) {
	skill, err := nx.GetPlayerSkill(ID)

	if err != nil {
		return Skill{}, fmt.Errorf("Not a valid skill ID %v level %v", ID, level)
	}

	if int(level) > len(skill) {
		return Skill{}, fmt.Errorf("Invalid skill level")
	}

	return Skill{ID: ID,
		Level:        level,
		Mastery:      byte(skill[level-1].Mastery),
		Cooldown:     0,
		CooldownTime: int16(skill[level-1].Time),
		TimeLastUsed: 0}, nil
}

func getSkillsFromCharID(id int32) []Skill {
	skills := []Skill{}

	filter := "skillID, level, cooldown"

	row, err := common.DB.Query("SELECT "+filter+" FROM skills where characterID=?", id)

	if err != nil {
		panic(err)
	}

	defer row.Close()

	for row.Next() {
		skill := Skill{}

		row.Scan(&skill.ID, &skill.Level, &skill.Cooldown)

		skillData, err := nx.GetPlayerSkill(skill.ID)

		if err != nil {
			return skills
		}

		skill.CooldownTime = int16(skillData[skill.Level-1].Time)

		skills = append(skills, skill)
	}

	return skills
}
