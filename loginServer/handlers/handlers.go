package handlers

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/crypt"
	"github.com/Hucaru/Valhalla/common/packet"
)

// HandlePacket -
func HandlePacket(conn connection.Connection, buffer packet.Packet, isHeader bool) int {
	size := constants.CLIENT_HEADER_SIZE

	if isHeader {
		// Reading encrypted header
		size = crypt.GetPacketLength(buffer)
	} else {
		// Handle data packet
		pos := 0

		opcode := buffer.ReadByte(&pos)

		switch opcode {
		case constants.LOGIN_OP:
			handleLoginRequest(buffer, &pos, conn)
		}

	}

	return size
}

func handleLoginRequest(p packet.Packet, pos *int, conn connection.Connection) {
	fmt.Println("Login packet received")
	usernameLength := p.ReadShort(pos)
	username := p.ReadString(pos, usernameLength)

	passwordLength := p.ReadShort(pos)
	password := p.ReadString(pos, passwordLength)

	// hash and salt the password#
	hasher := sha512.New()
	hasher.Write([]byte(password))
	hashedPassword := hex.EncodeToString(hasher.Sum(nil))

	fmt.Println("Attempted login from user:", username, "- password (hashed):", hashedPassword)

	// Check username and passwd against db
	validLogin := false

	if validLogin {

	} else {

	}

}
