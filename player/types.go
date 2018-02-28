package player

type clientConn interface {
	Close()
	String() string
	SetUserID(uint32)
	GetUserID() uint32
	SetAdmin(val bool)
	IsAdmin() bool
	SetIsLogedIn(bool)
	GetIsLogedIn() bool
	SetChanID(uint32)
	GetChanID() uint32
	SetCloseCallback(func())
	// Below here might not be needed
	SetWorldID(uint32)
	GetWorldID() uint32
}
