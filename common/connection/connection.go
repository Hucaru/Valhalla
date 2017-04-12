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

	buffer := packet.NewPacket(sizeToRead)
	packetIt := packet.NewPacketIterator()

	for {
		err := conn.Read(buffer)

		if buffer.Size() == headerSize {
			sizeToRead = int(buffer.ReadShort(packetIt))
		} else {
			sizeToRead = headerSize
		}

		if err != nil {
			fmt.Println("Error in reading from connection")
			return
		}

		handler(buffer)
		packetIt.Clear()
	}
}
