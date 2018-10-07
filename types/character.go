package types

type character struct {
	ID     int32
	UserID int32

	CurrentMap    int32
	CurrentMapPos byte
	PreviousMap   int32

	Mesos int32

	Job int16

	Level byte
	Str   int16
	Dex   int16
	Intt  int16
	Luk   int16
	Hp    int16
	MaxHP int16
	Mp    int16
	MaxMP int16
	Ap    int16
	Sp    int16
	Exp   int32
	Fame  int16

	Skills map[int32]int32

	MiniGameWins, MiniGameTies, MiniGameLosses int32
}
