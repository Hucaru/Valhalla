package character

type Skill struct {
	skillID uint32
	level   byte
}

func (s *Skill) GetID() uint32 {
	return s.skillID
}
func (s *Skill) SetID(val uint32) {
	s.skillID = val
}
func (s *Skill) GetLevel() byte {
	return s.level
}
func (s *Skill) SetLevel(val byte) {
	s.level = val
}
