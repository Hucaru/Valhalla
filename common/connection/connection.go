package connection

import (
	"fmt"

	"github.com/Hucaru/Valhalla/common/packet"
)

type Connection interface {
	Write(p packet.Packet) error
	Read(p packet.Packet) error
	IsOpen() bool
	Close()
}

type PacketHandler func(p packet.Packet)

func HandleNewConnection(conn Connection, handler PacketHandler, headerSize int) {
	sizeToRead := 2

	for {
		buffer := packet.NewPacket(sizeToRead)
		err := conn.Read(buffer)

		if err != nil {
			fmt.Println("Error in reading from connection")
			return
		}

		if buffer.Size() == headerSize {
			sizeToRead = int(buffer.ReadShort(0))
		} else {
			sizeToRead = headerSize
		}

		handler(buffer)
	}
}
