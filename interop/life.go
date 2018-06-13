package interop

type Life interface {
	SetID(uint32)
	GetID() uint32
	SetSpawnID(uint32)
	GetSpawnID() uint32
	Pos

	SetFoothold(int16)
	GetFoothold() int16
	SetFace(byte)
	GetFace() byte
	GetState() byte
	SetState(byte)
}
