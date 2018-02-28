package player

type clientConn interface {
	Close()
	String() string
	SetUserID(val uint32)
	GetUserID() uint32
	SetAdmin(val bool)
	IsAdmin() bool
	SetIsLogedIn(val bool)
	GetIsLogedIn() bool
	// Below here might not be needed
	SetWorldID(val uint32)
	GetWorldID() uint32
	SetChanID(val byte)
	GetChanID() byte
}
