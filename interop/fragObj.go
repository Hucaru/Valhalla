package interop

type FragObj interface {
	Pos
	GetState() byte
	SetState(byte)
	SetFoothold(int16)
	GetFoothold() int16
}
