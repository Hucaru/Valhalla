package data

type mapleMob struct {
	hp, maxHp, mp, maxMp int16
	x, y, fh             int16
	face                 byte
}

func (m *mapleMob) SetX(x int16) {
	m.x = x
}

func (m *mapleMob) GetX() int16 {
	return m.x
}

func (m *mapleMob) SetY(y int16) {
	m.y = y
}

func (m *mapleMob) GetY() int16 {
	return m.y
}

func (m *mapleMob) SetFace(face byte) {
	m.face = face
}

func (m *mapleMob) GetFace() byte {
	return m.face
}
