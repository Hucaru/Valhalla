package data

import "github.com/Hucaru/Valhalla/interfaces"

type mapleNpc struct {
	id             uint32
	x, y, foothold int16
	face           bool
	controller     interfaces.ClientConn
}

func (n *mapleNpc) SetID(id uint32)                                { n.id = id }
func (n *mapleNpc) GetID() uint32                                  { return n.id }
func (n *mapleNpc) SetX(x int16)                                   { n.x = x }
func (n *mapleNpc) GetX() int16                                    { return n.x }
func (n *mapleNpc) SetY(y int16)                                   { n.y = y }
func (n *mapleNpc) GetY() int16                                    { return n.y }
func (n *mapleNpc) SetFoothold(y int16)                            { n.foothold = y }
func (n *mapleNpc) GetFoothold() int16                             { return n.foothold }
func (n *mapleNpc) SetFace(face bool)                              { n.face = face }
func (n *mapleNpc) GetFace() bool                                  { return n.face }
func (n *mapleNpc) GetController() interfaces.ClientConn           { return n.controller }
func (n *mapleNpc) SetController(controller interfaces.ClientConn) { n.controller = controller }
