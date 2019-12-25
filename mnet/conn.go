package mnet

import (
	"net"

	"github.com/Hucaru/Valhalla/mnet/crypt"

	"github.com/Hucaru/Valhalla/mpacket"
)

type MConn interface {
	String() string
	Send(mpacket.Packet)
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
			readSize = headerSize

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
			readSize = int(buffer[0])
		} else {
			readSize = headerSize
			eRecv <- &Event{Type: MEServerPacket, Conn: conn, Packet: buffer}
		}

		header = !header
	}
}

type baseConn struct {
	net.Conn
	eSend   chan mpacket.Packet
	eRecv   chan *Event
	endSend chan bool
	reader  func()

	cryptSend *crypt.Maple
	cryptRecv *crypt.Maple

	interServer bool
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

			tmp := make(mpacket.Packet, len(p))
			copy(tmp, p)

			if bc.cryptSend != nil {
				bc.cryptSend.Encrypt(tmp, true, false)
			}

			if bc.interServer {
				tmp[0] = byte(len(tmp) - 1)
			}

			bc.Conn.Write(tmp)
		}
	}
}

func (bc *baseConn) Send(p mpacket.Packet) {
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
