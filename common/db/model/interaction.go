package model

type Interaction struct {
	AccountID       int64
	UId             string
	CharacterID     int64
	objectIndex     int32
	animMontageName string
	destinationX    float32
	destinationY    float32
	destinationZ    float32
}
