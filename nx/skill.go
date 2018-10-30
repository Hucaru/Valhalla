package nx

type JobSkill struct {
}

var JobSkills = make(map[int32]JobSkill)

type MobSkill struct {
	Level               byte
	MpCon, Interval, HP int32
}

var MobSkills = make(map[int32]MobSkill)

func getJobSkills() {

}

func getMobSkills() {

}
