package handlers

import (
	"fmt"

	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/crypt"
	"github.com/Hucaru/Valhalla/common/packet"
)

// HandlePacket -
func HandlePacket(conn connection.Connection, buffer packet.Packet, isHeader bool) int {
	var size int

	if isHeader {
		// Reading encrypted header
		size = crypt.GetPacketLength(buffer)
	} else {
		// Handle data packet
		size = constants.CLIENT_HEADER_SIZE
		pos := 0

		opcode := buffer.ReadByte(&pos)

		switch opcode {
		case 0x1:
			handleLoginRequest(buffer, conn)
		}

	}

	return size
}

func handleLoginRequest(p packet.Packet, conn connection.Connection) {
	fmt.Println("Login packet received")
}
