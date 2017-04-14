package connection

import (
	"fmt"

	"github.com/Hucaru/Valhalla/common/crypt"
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
type PacketHandler func(p packet.Packet)

// HandleNewConnection -
func HandleNewConnection(conn Connection, handler PacketHandler, headerSize int) {
	sizeToRead := headerSize
	//var pos int

	for {
		buffer := packet.NewPacket(sizeToRead)
		err := conn.Read(buffer)

		if err != nil {
			fmt.Println("Error in reading from connection")
			return
		}

		if buffer.Size() == headerSize {
			sizeToRead = crypt.GetPacketLength(buffer)
		} else {
			sizeToRead = headerSize
			handler(buffer)
		}
	}
}
