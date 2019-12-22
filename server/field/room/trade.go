package room

// Trade interface to the omok struct
type Trade interface {
}

type trade struct {
	room
}

// NewTrade returns an interface of Trade
func NewTrade(id int32) Trade {
	return &trade{}
}
