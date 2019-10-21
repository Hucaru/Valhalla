package entity

import (
	"database/sql"
	"fmt"

	"github.com/Hucaru/Valhalla/nx"
)

type Skill struct {
	ID             int32
	Level, Mastery byte
	Cooldown       int16
	TimeLastUsed   int64
}

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
		Cooldown:     int16(skill[level-1].Time),
		TimeLastUsed: 0}, nil
}

func getSkillsFromCharID(db *sql.DB, id int32) []Skill {
	skills := []Skill{}

	filter := "skillID, level, cooldown"

	row, err := db.Query("SELECT "+filter+" FROM skills where characterID=?", id)

	if err != nil {
		panic(err)
	}

	defer row.Close()

	for row.Next() {
		skill := Skill{}

		row.Scan(&skill.ID, &skill.Level, &skill.Cooldown)

		skills = append(skills, skill)
	}

	return skills
}
