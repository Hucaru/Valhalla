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
func HandleNewConnection(conn connection, handler func(p maplepacket.Reader), sizeOfRead int, isMapleCrypt bool) {
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
			if isMapleCrypt {
				sizeToRead = crypt.GetPacketLength(buffer)
			} else {
				p := maplepacket.NewPacket()
				p.Append(buffer)
				r := maplepacket.NewReader(&p)
				sizeToRead = int(r.ReadInt32())
			}
		} else {
			p := maplepacket.NewPacket()
			p.Append(buffer)
			handler(maplepacket.NewReader(&p))
			sizeToRead = sizeOfRead
		}

		isHeader = !isHeader
	}
}
