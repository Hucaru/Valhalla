package connection

// Player struct for duration of connection
type Player struct {
	userID    uint32
	hash      string
	isLogedin bool
}

func (p *Player) SetUserID(userID uint32) {
	p.userID = userID
}

func (p *Player) GetUserID() uint32 {
	return p.userID
}

func (p *Player) SetSessionHash(hash string) {
	p.hash = hash
}

func (p *Player) GetSessionHash() string {
	return p.hash
}

func (p *Player) SetIsLogedIn(status bool) {
	p.isLogedin = status
}

func (p *Player) GetIsLogedIn() bool {
	return p.isLogedin
}
