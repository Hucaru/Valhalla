package interop

type Portal interface {
	GetToMap() uint32
	SetToMap(uint32)
	GetToPortal() string
	SetToPortal(string)
	GetName() string
	SetName(string)
	Pos
	GetIsSpawn() bool
	SetIsSpawn(bool)
}
