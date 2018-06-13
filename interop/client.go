package interop

import (
	"github.com/Hucaru/Valhalla/maplepacket"
)

type ClientConn interface {
	Close()
	Write(maplepacket.Packet) error
	String() string
	SetUserID(uint32)
	GetUserID() uint32
	SetAdmin(val bool)
	IsAdmin() bool
	SetIsLoggedIn(bool)
	GetIsLoggedIn() bool
	SetChanID(uint32)
	GetChanID() uint32
	AddCloseCallback(func())
	// Below here might not be needed
	SetWorldID(uint32)
	GetWorldID() uint32
}
