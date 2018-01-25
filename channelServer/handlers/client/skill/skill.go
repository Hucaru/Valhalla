package skill

import (
	"github.com/Hucaru/Valhalla/common/connection"
)

type Skill struct {
	SkillID uint32
	Level   byte
}

func GetCharacterSkills(charID uint32) []Skill {
	filter := "skillID,level"
	row, err := connection.Db.Query("SELECT "+filter+" FROM skills WHERE characterID=?", charID)

	if err != nil {
		panic(err.Error())
	}

	defer row.Close()

	var skills []Skill

	for row.Next() {
		var newSkill Skill

		row.Scan(&newSkill.SkillID, &newSkill.Level)

		skills = append(skills, newSkill)
	}

	return skills
}
