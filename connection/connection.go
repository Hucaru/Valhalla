package connection

import (
	"log"

	"github.com/Hucaru/Valhalla/crypt"
	"github.com/Hucaru/gopacket"
)

// Connection -
type connection interface {
	Read(p gopacket.Packet) error
	Close()
	String() string
}

// HandleNewConnection -
func HandleNewConnection(conn connection, handler func(p gopacket.Reader), sizeOfRead int, isMapleCrypt bool) {
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
				p := gopacket.NewPacket()
				p.Append(buffer)
				r := gopacket.NewReader(&p)
				sizeToRead = int(r.ReadInt32())
			}
		} else {
			p := gopacket.NewPacket()
			p.Append(buffer)
			handler(gopacket.NewReader(&p))
			sizeToRead = sizeOfRead
		}

		isHeader = !isHeader
	}
}
