package room

// Memory interface to the omok struct
type Memory interface {
}

type memory struct {
	room
}

// NewMemory returns an interface of Match
func NewMemory(id int32, name, password string, boardType byte) Memory {
	return &memory{}
}
