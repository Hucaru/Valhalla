package mnet

import (
	"github.com/Hucaru/Valhalla/common/dataController"
	"log"
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
	ch       chan mpacket.Packet
	chFinish bool
}

type baseConn struct {
	net.Conn
	eSend  chan mpacket.Packet
	eRecv  chan *Event
	reader func()
	closed bool

	sendChannelLock  sync.RWMutex
	sendChannelQueue *dataController.MapQueue

	sendChannel chan mpacket.Packet

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
	//for {
	//	if bc.closed {
	//		return
	//	}
	//
	//	for {
	//		v := bc.sendChannelQueue.Dequeue()
	//		if v != nil {
	//			bc.Write(v)
	//		} else {
	//			break
	//		}
	//	}
	//
	//	runtime.Gosched()
	//}

	//for {
	//
	//}
}

func (bc *baseConn) Send(p mpacket.Packet) {
	if bc.closed {
		return
	}

	if len(bc.sendChannel) == cap(bc.sendChannel) {
		//if cap(bc.sendChannel) > 0 {
		//	close(bc.sendChannel)
		//}
		bc.sendChannel = make(chan mpacket.Packet, 1024*1024)
		//log.Println("full realloc len : ", len(bc.sendChannel))

		go func(c <-chan mpacket.Packet) {
			for _v := range c {
				if _v == nil {
					break
				}
				bc.Write(_v)
			}

			log.Println("finish channel : ", len(c), cap(c))
		}(bc.sendChannel)
		if bc.closed {
			return
		}
		bc.sendChannel <- p
	} else {
		if bc.closed {
			return
		}
		bc.sendChannel <- p
	}

}

func (bc *baseConn) String() string {
	return bc.Conn.RemoteAddr().String()
}

func (bc *baseConn) Cleanup() {
	bc.closed = true
	//close(bc.sendChannel)
}
