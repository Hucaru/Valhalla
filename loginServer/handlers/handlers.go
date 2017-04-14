package handlers

import (
	"fmt"

	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/crypt"
	"github.com/Hucaru/Valhalla/common/packet"
)

// HandlePacket -
func HandlePacket(conn connection.Connection, buffer packet.Packet) int {
	var size int
	if buffer.Size() == constants.CLIENT_HEADER_SIZE {
		// Reading encrypted header
		size = crypt.GetPacketLength(buffer)
	} else {
		// Handle data packet
		size = constants.CLIENT_HEADER_SIZE
		pos := 0

		opcode := buffer.ReadByte(&pos)

		switch opcode {
		case 0x1:
			fmt.Println("Login packet received")
		}

	}

	return size
}
