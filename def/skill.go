package def

import "github.com/Hucaru/Valhalla/database"

type Skill struct {
	ID             int32
	Level, Mastery byte
	Cooldown       int16
}

func GetSkillsFromCharID(id int32) []Skill {
	skills := []Skill{}

	filter := "skillID, level, mastery, cooldown"

	row, err := database.Handle.Query("SELECT "+filter+" FROM skills where characterID=?", id)

	if err != nil {
		panic(err)
	}

	defer row.Close()

	for row.Next() {
		skill := Skill{}

		row.Scan(&skill.ID, &skill.Level, &skill.Mastery, &skill.Cooldown)

		skills = append(skills, skill)
	}

	return skills
}
