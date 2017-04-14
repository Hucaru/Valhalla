package connection

import (
	"fmt"

	"github.com/Hucaru/Valhalla/common/packet"
)

// Connection -
type Connection interface {
	Write(p packet.Packet) error
	Read(p packet.Packet) error
	IsOpen() bool
	Close()
}

// PacketHandler -
type PacketHandler func(conn Connection, p packet.Packet) int

// HandleNewConnection -
func HandleNewConnection(conn Connection, handler PacketHandler, sizeOfRead int) {
	sizeToRead := sizeOfRead
	//var pos int

	for {
		buffer := packet.NewPacket(sizeToRead)
		err := conn.Read(buffer)

		if err != nil {
			fmt.Println("Error in reading from connection")
			return
		}

		sizeToRead = handler(conn, buffer)
	}
}
