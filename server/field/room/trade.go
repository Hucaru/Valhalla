package room

// Trade behaviours
type Trade interface{}

// Trade window
type trade struct {
	room
}

// NewTrade a trade
func NewTrade(id int32) Trade {
	return &trade{}
}
