package interop

type Portal interface {
	GetToMap() int32
	SetToMap(int32)
	GetToPortal() string
	SetToPortal(string)
	GetName() string
	SetName(string)
	Pos
	GetIsSpawn() bool
	SetIsSpawn(bool)
}
