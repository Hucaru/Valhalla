package connection

import (
	"log"

	"github.com/Hucaru/Valhalla/crypt"
	"github.com/Hucaru/Valhalla/maplepacket"
)

// Connection -
type connection interface {
	Read(p maplepacket.Packet) error
	Close()
	String() string
}

// HandleNewConnection -
func HandleNewConnection(conn connection, handler func(p maplepacket.Reader), sizeOfRead int) {
	sizeToRead := sizeOfRead
	isHeader := true

	for {
		buffer := make([]byte, sizeToRead)

		err := conn.Read(buffer)

		if err != nil {
			log.Println("Error in reading from", conn, ", closing the connection", err)
			conn.Close()
			return
		}

		if isHeader {
			sizeToRead = crypt.GetPacketLength(buffer)
		} else {
			p := maplepacket.NewPacket()
			p.Append(buffer)
			handler(maplepacket.NewReader(&p))
			sizeToRead = sizeOfRead
		}

		isHeader = !isHeader
	}
}
