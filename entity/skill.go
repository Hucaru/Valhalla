package entity

import (
	"database/sql"

	"github.com/Hucaru/Valhalla/nx"
)

type Skill struct {
	ID             int32
	Level, Mastery byte
	Cooldown       int16
	TimeLastUsed   int64
}

func CreateSkillFromData(ID int32, level byte, skill nx.PlayerSkill) Skill {
	return Skill{ID: ID,
		Level:        level,
		Mastery:      byte(skill.Mastery),
		Cooldown:     int16(skill.Time),
		TimeLastUsed: 0}
}

func GetSkillsFromCharID(db *sql.DB, id int32) []Skill {
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
