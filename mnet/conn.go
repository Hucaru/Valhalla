package mnet

import (
	"net"

	"github.com/Hucaru/Valhalla/consts"
	"github.com/Hucaru/Valhalla/mnet/crypt"

	"github.com/Hucaru/Valhalla/maplepacket"
)

type MConn interface {
	String() string
	Send(maplepacket.Packet)
	Cleanup()
}

func clientReader(conn net.Conn, eRecv chan *Event, mapleVersion int16, headerSize int, cryptRecv *crypt.Maple) {
	eRecv <- &Event{Type: MEClientConnected, Conn: conn}

	header := true
	readSize := headerSize

	for {
		buffer := make([]byte, readSize)

		if _, err := conn.Read(buffer); err != nil {
			eRecv <- &Event{Type: MEClientDisconnect, Conn: conn}
			break
		}

		if header {
			readSize = crypt.GetPacketLength(buffer)
		} else {
			readSize = consts.ClientHeaderSize

			if cryptRecv != nil {
				cryptRecv.Decrypt(buffer, true, false)
			}

			eRecv <- &Event{Type: MEClientPacket, Conn: conn, Packet: buffer}
		}

		header = !header
	}
}

func serverReader(conn net.Conn, eRecv chan *Event, headerSize int) {
	eRecv <- &Event{Type: MEServerConnected, Conn: conn}

	header := true
	readSize := headerSize

	for {
		buffer := make([]byte, readSize)

		if _, err := conn.Read(buffer); err != nil {
			eRecv <- &Event{Type: MEServerDisconnect, Conn: conn}
			break
		}

		if header {
			readSize = crypt.GetPacketLength(buffer)
		} else {
			readSize = consts.ClientHeaderSize
			eRecv <- &Event{Type: MEClientPacket, Conn: conn, Packet: buffer}
		}

		header = !header
	}
}

type baseConn struct {
	net.Conn
	eSend   chan maplepacket.Packet
	eRecv   chan *Event
	endSend chan bool
	reader  func()

	cryptSend *crypt.Maple
	cryptRecv *crypt.Maple
}

func (bc *baseConn) Reader() {
	bc.reader()
}

func (bc *baseConn) Writer() {
	for {
		select {
		case p, ok := <-bc.eSend:
			if !ok {
				return
			}

			if bc.cryptSend != nil {
				bc.cryptSend.Encrypt(p, true, false)
			}

			bc.Conn.Write(p)
		}
	}
}

func (bc *baseConn) Send(p maplepacket.Packet) {
	select {
	case bc.eSend <- p:
	case <-bc.endSend:
		close(bc.eSend)
	}

}

func (bc *baseConn) String() string {
	return bc.Conn.RemoteAddr().String()
}

func (bc *baseConn) Cleanup() {
	bc.endSend <- true
}
