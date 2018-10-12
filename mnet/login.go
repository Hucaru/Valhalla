package mnet

import (
	"net"

	"github.com/Hucaru/Valhalla/consts"
	"github.com/Hucaru/Valhalla/crypt"

	"github.com/Hucaru/Valhalla/maplepacket"
)

type MConnLogin interface {
	MConn

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

func NewLogin(conn net.Conn, eRecv chan *Event, queueSize int, keySend, keyRecv [4]byte) *login {
	l := &login{}
	l.Conn = conn

	l.eSend = make(chan maplepacket.Packet, queueSize)
	l.eRecv = eRecv

	l.cryptSend = crypt.New(keySend, consts.MapleVersion)
	l.cryptRecv = crypt.New(keyRecv, consts.MapleVersion)

	l.reader = func() {
		clientReader(l, l.eRecv, consts.MapleVersion, consts.ClientHeaderSize, l.cryptRecv, l.cryptSend)
	}

	return l
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
