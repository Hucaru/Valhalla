package data

import (
	"github.com/Hucaru/Valhalla/interfaces"
)

type mapleNpc struct {
	id, spawnID                                 uint32
	x, y, sx, sy, rx0, rx1, foothold, sfoothold int16
	face, state                                 byte
	controller                                  interfaces.ClientConn
	isAlive                                     bool
}

func (n *mapleNpc) SetID(id uint32)                                { n.id = id }
func (n *mapleNpc) GetID() uint32                                  { return n.id }
func (n *mapleNpc) SetSpawnID(spawnID uint32)                      { n.spawnID = spawnID }
func (n *mapleNpc) GetSpawnID() uint32                             { return n.spawnID }
func (n *mapleNpc) SetX(x int16)                                   { n.x = x }
func (n *mapleNpc) GetX() int16                                    { return n.x }
func (n *mapleNpc) SetY(y int16)                                   { n.y = y }
func (n *mapleNpc) GetY() int16                                    { return n.y }
func (n *mapleNpc) SetSX(x int16)                                  { n.sx = x }
func (n *mapleNpc) GetSX() int16                                   { return n.sx }
func (n *mapleNpc) SetSY(y int16)                                  { n.sy = y }
func (n *mapleNpc) GetSY() int16                                   { return n.sy }
func (n *mapleNpc) SetRx0(rx0 int16)                               { n.rx0 = rx0 }
func (n *mapleNpc) GetRx0() int16                                  { return n.rx0 }
func (n *mapleNpc) SetRx1(rx1 int16)                               { n.rx1 = rx1 }
func (n *mapleNpc) GetRx1() int16                                  { return n.rx1 }
func (n *mapleNpc) SetFoothold(y int16)                            { n.foothold = y }
func (n *mapleNpc) GetFoothold() int16                             { return n.foothold }
func (n *mapleNpc) SetSFoothold(y int16)                           { n.sfoothold = y }
func (n *mapleNpc) GetSFoothold() int16                            { return n.sfoothold }
func (n *mapleNpc) SetFace(face byte)                              { n.face = face }
func (n *mapleNpc) GetFace() byte                                  { return n.face }
func (n *mapleNpc) GetState() byte                                 { return n.state }
func (n *mapleNpc) SetState(state byte)                            { n.state = state }
func (n *mapleNpc) GetController() interfaces.ClientConn           { return n.controller }
func (n *mapleNpc) SetController(controller interfaces.ClientConn) { n.controller = controller }
func (n *mapleNpc) SetIsAlive(alive bool)                          { n.isAlive = alive }
func (n *mapleNpc) GetIsAlive() bool                               { return n.isAlive }
