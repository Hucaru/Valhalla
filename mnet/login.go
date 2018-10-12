package mnet

import (
	"crypto/rand"
	"net"

	"github.com/Hucaru/Valhalla/consts"
	"github.com/Hucaru/Valhalla/crypt"

	"github.com/Hucaru/Valhalla/maplepacket"
)

type MConnLogin interface {
	MConn

	Cleanup()

	GetAccountID() int32
	SetAccountID(int32)

	GetGender() byte
	SetGender(byte)
}

type login struct {
	baseConn

	accountID int32
	gender    byte
}

func NewLogin(conn net.Conn, eRecv chan *Event, queueSize int) *login {
	l := &login{}
	l.Conn = conn

	l.eSend = make(chan maplepacket.Packet, queueSize)
	l.eRecv = eRecv

	key := [4]byte{}
	rand.Read(key[:])
	l.cryptSend = crypt.New(key, consts.MapleVersion)
	rand.Read(key[:])
	l.cryptRecv = crypt.New(key, consts.MapleVersion)

	l.reader = func() {
		clientReader(l, l.eRecv, consts.MapleVersion, consts.ClientHeaderSize, l.cryptRecv, l.cryptSend)
	}

	return l
}

func (l *login) Cleanup() {
	l.baseConn.Cleanup()
}

func (l *login) GetAccountID() int32 {
	return l.accountID
}

func (l *login) SetAccountID(accountID int32) {
	l.accountID = accountID
}

func (l *login) GetGender() byte {
	return l.gender
}

func (l *login) SetGender(gender byte) {
	l.gender = gender
}
