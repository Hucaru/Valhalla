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

	var databaseUser string
	var databasePassword string
	var databaseIsLogedIn bool
	var databaseBanned int
	var databaseIsAdmin bool

	err := connection.Db.QueryRow("SELECT username, password, isLogedIn, isBanned, isAdmin FROM users WHERE username=?", username).
		Scan(&databaseUser, &databasePassword, &databaseIsLogedIn, &databaseBanned, &databaseIsAdmin)

	result := byte(0x00)

	if err != nil {
		result = 0x05
	} else if hashedPassword != databasePassword {
		result = 0x04
	} else if databaseIsLogedIn {
		result = 0x07
	} else if databaseBanned > 0 {
		result = 0x03
	}

	fmt.Println(username, "has logged in", result)

	packet := packet.NewPacket()
	packet.WriteByte(0x01)
	packet.WriteByte(0x05)
	packet.WriteByte(0x00)
	packet.WriteInt(0)

	packet.WriteLong(0)
	packet.WriteLong(0)
	packet.WriteLong(0)

	conn.Write(packet)
}
