package interop

type Life interface {
	SetID(int32)
	GetID() int32
	SetSpawnID(int32)
	GetSpawnID() int32
	Pos

	SetFoothold(int16)
	GetFoothold() int16
	SetFace(byte)
	GetFace() byte
	GetState() byte
	SetState(byte)
}
