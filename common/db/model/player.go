package model

type Player struct {
	AccountID   int64
	UId         string
	CharacterID int64
	RegionID    int64
	Character   *Character
	Interaction *Interaction
	IsBot       int32
	ModifiedAt  int64
}
