package data

type mapleNpc struct {
	x, y int16
	face byte
}

func (n *mapleNpc) SetX(x int16) {
	n.x = x
}

func (n *mapleNpc) GetX() int16 {
	return n.x
}

func (n *mapleNpc) SetY(y int16) {
	n.y = y
}

func (n *mapleNpc) GetY() int16 {
	return n.y
}

func (n *mapleNpc) SetFace(face byte) {
	n.face = face
}

func (n *mapleNpc) GetFace() byte {
	return n.face
}
