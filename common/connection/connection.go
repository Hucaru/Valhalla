package connection

import (
	"log"
	"net"
	"strconv"

	"github.com/Hucaru/Valhalla/common/crypt"
	"github.com/Hucaru/gopacket"
)

// Connection -
type connection interface {
	Read(p gopacket.Packet) error
	Close()
	String() string
}

// CreateServerListener -
func CreateServerListener(protocol string, address string, startingPort int) (net.Listener, error, uint16) {
	listener, err := net.Listen(protocol, address+":"+strconv.Itoa(startingPort))
	port := startingPort

	if err != nil {
		for port = startingPort + 1; port < startingPort+100; port++ {
			listener, err = net.Listen(protocol, address+":"+strconv.Itoa(port))

			if err == nil {
				break
			}

		}
	}

	return listener, err, uint16(port)
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
