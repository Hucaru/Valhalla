package room

// Memory behaviours
type Memory interface{}

type memory struct {
	room
}

// NewMemory a new memory
func NewMemory(id int32, name, password string, boardType byte) Memory {
	return &memory{}
}
