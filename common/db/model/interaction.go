package model

type Interaction struct {
	IsInteraction   bool
	ObjectIndex     int32
	AttachEnabled   int32
	AnimMontageName string
	DestinationX    float32
	DestinationY    float32
	DestinationZ    float32
}

func NewInteraction() Interaction {
	result := Interaction{}
	result.IsInteraction = false

	return result
}
