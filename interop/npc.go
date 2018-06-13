package interop

type Npc interface {
	Life
	SetRx0(int16)
	GetRx0() int16
	SetRx1(int16)
	GetRx1() int16
}
