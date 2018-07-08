package interop

import (
	"github.com/Hucaru/Valhalla/maplepacket"
)

type ClientConn interface {
	Close()
	Write(maplepacket.Packet) error
	String() string
	SetUserID(int32)
	GetUserID() int32
	SetAdmin(val bool)
	IsAdmin() bool
	SetIsLoggedIn(bool)
	GetIsLoggedIn() bool
	SetChanID(int32)
	GetChanID() int32
	AddCloseCallback(func())
	// Below here might not be needed
	SetWorldID(int32)
	GetWorldID() int32
}
