package mnet

import (
	"github.com/Hucaru/Valhalla/common/dataController"
	"math/rand"
	"net"
	"sync"
	"time"

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

		//log.Println("readSize", readSize)

		header = !header
	}
}

type SendChannelWrapper struct {
	ch    chan mpacket.Packet
	chNew chan bool
	lock  sync.RWMutex
}

type baseConn struct {
	net.Conn
	eSend  chan mpacket.Packet
	eRecv  chan *Event
	reader func()
	closed bool

	sendChannelLock  sync.RWMutex
	sendChannelQueue dataController.LKQueue

	sendChannelWrappwer SendChannelWrapper

	cryptSend *crypt.Maple
	cryptRecv *crypt.Maple

	interServer bool

	latency int
	jitter  int
	pSend   chan func()
}

func (bc *baseConn) Reader() {
	bc.reader()
}

func (bc *baseConn) Writer() {
	for {
		p, ok := <-bc.eSend
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

		if bc.latency > 0 {
			now := time.Now().UnixNano()
			sendTime := now + int64(rand.Intn(bc.jitter)+bc.latency)*1000000
			bc.pSend <- func() {
				now := time.Now().UnixNano()
				delta := sendTime - now

				if delta > 0 {
					time.Sleep(time.Duration(delta))
				}

				bc.Conn.Write(tmp)
			}
		} else {
			bc.Conn.Write(tmp)
		}
	}
}

func (bc *baseConn) MetaWriter() {

	for {
		if bc.closed {
			return
		}

		alloced := false

		select {
		case p, ok := <-bc.sendChannelWrappwer.ch:
			{
				if !ok {
					break
				}

				bc.Write(p)
				//if err != nil {
				//	log.Println(err)
				//	if bc.closed {
				//		break
				//	}
				//}

				if len(bc.sendChannelWrappwer.ch) >= cap(bc.sendChannelWrappwer.ch) {
					close(bc.sendChannelWrappwer.ch)
					bc.sendChannelWrappwer.ch = make(chan mpacket.Packet, 4)
					alloced = true
				}
			}
		}

		if bc.closed {
			if alloced {
				close(bc.sendChannelWrappwer.ch)
			}
			return
		}
	}
}

func (bc *baseConn) Send(p mpacket.Packet) {

	if bc.closed {
		return
	}

	bc.sendChannelWrappwer.ch <- p
}

func (bc *baseConn) String() string {
	return bc.Conn.RemoteAddr().String()
}

func (bc *baseConn) Cleanup() {
	bc.closed = true
	close(bc.sendChannelWrappwer.ch)
}
