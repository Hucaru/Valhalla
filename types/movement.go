package types

type MovementFrag struct {
	X, Y, Vx, Vy, Foothold, Duration int16
	Stance, MType                    byte
}

type MovementData struct {
	OrigX, OrigY int16
	Frags        []MovementFrag
}
