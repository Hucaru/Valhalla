package connection

import (
	"fmt"

	"github.com/Hucaru/Valhalla/common/crypt"
	"github.com/Hucaru/Valhalla/common/packet"
)

// Connection -
type connection interface {
	Read(p packet.Packet) error
	Close()
	String() string
}

// HandleNewConnection -
func HandleNewConnection(conn connection, handler func(p packet.Packet), sizeOfRead int) {
	sizeToRead := sizeOfRead
	isHeader := true
	fmt.Println("New connection from", conn)
	for {
		buffer := make([]byte, sizeToRead)

		err := conn.Read(buffer)

		if err != nil {
			fmt.Println("Error in reading from", conn, ", closing the connection", err)
			conn.Close()
			return
		}

		if isHeader {
			sizeToRead = crypt.GetPacketLength(buffer)
		} else {
			p := packet.NewPacket()
			p.Append(buffer)
			handler(p)
			sizeToRead = sizeOfRead
		}

		isHeader = !isHeader
	}
}
