package mnet

type MConnChannel interface {
	MConn

	GetAdmin() bool
	SetAdmin(bool)
}

type channel struct {
	baseConn
}
