package handlers

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/packet"
	"github.com/Hucaru/Valhalla/loginServer/loginConn"
	"github.com/Hucaru/Valhalla/loginServer/worlds"
)

// HandlePacket -
func HandlePacket(conn *loginConn.Connection, buffer packet.Packet) {
	pos := 0
	opcode := buffer.ReadByte(&pos)

	switch opcode {
	case constants.LOGIN_REQUEST:
		handleLoginRequest(buffer, &pos, conn)
	case constants.LOGIN_CHECK_LOGIN:
		handleGoodLogin(buffer, &pos, conn)
	default:
		fmt.Println("UNKNOWN LOGIN PACKET:", buffer)
	}

}

func handleLoginRequest(p packet.Packet, pos *int, conn *loginConn.Connection) {
	usernameLength := p.ReadInt16(pos)
	username := p.ReadString(pos, int(usernameLength))

	passwordLength := p.ReadInt16(pos)
	password := p.ReadString(pos, int(passwordLength))

	// hash the password, cba to salt atm
	hasher := sha512.New()
	hasher.Write([]byte(password))
	hashedPassword := hex.EncodeToString(hasher.Sum(nil))

	var userID uint32
	var user string
	var databasePassword string
	var isLogedIn bool
	var isBanned int
	var isAdmin bool

	err := connection.Db.QueryRow("SELECT userID, username, password, isLogedIn, isBanned, isAdmin FROM users WHERE username=?", username).
		Scan(&userID, &user, &databasePassword, &isLogedIn, &isBanned, &isAdmin)

	result := byte(0x00)

	if err != nil {
		result = 0x05
	} else if hashedPassword != databasePassword {
		result = 0x04
	} else if isLogedIn {
		result = 0x07
	} else if isBanned > 0 {
		result = 0x02
	}

	// -Banned- = 2
	// Deleted or Blocked = 3
	// Invalid Password = 4
	// Not Registered = 5
	// Sys Error = 6
	// Already online = 7
	// System error = 9
	// Too many requests = 10
	// Older than 20 = 11
	// Master cannot login on this IP = 13

	pac := packet.NewPacket()
	pac.WriteByte(constants.LOGIN_RESPONCE)
	pac.WriteByte(result)
	pac.WriteByte(0x00)
	pac.WriteInt32(0)

	if result <= 0x01 {
		pac.WriteUint32(userID)
		pac.WriteByte(0x00)

		if isAdmin {
			pac.WriteByte(0x01)
		} else {
			pac.WriteByte(0x00)
		}

		pac.WriteByte(0x01)
		pac.WriteString(username)

		conn.SetUserID(userID)
		conn.SetIsLogedIn(true)
		_, err = connection.Db.Query("UPDATE users set isLogedIn=1 WHERE userID=?", userID)

		if err != nil {
			fmt.Println("Database error with approving login of userID", userID, err)
		} else {
			fmt.Println("User", userID, "has logged in from", conn)
		}
	} else if result == 0x02 {
		pac.WriteByte(byte(isBanned))
		pac.WriteInt64(0) // Expire time, for now let set this to epoch
	}

	pac.WriteInt64(0)
	pac.WriteInt64(0)
	pac.WriteInt64(0)
	conn.Write(pac)

	if err != nil {
		fmt.Println(err)
	}
}

func handleGoodLogin(p packet.Packet, pos *int, conn *loginConn.Connection) {
	// What the fuck is going on here? Is it the pin system?
	pac := packet.NewPacket()
	pac.WriteByte(0x03)
	pac.WriteByte(0x04)
	pac.WriteByte(0x00)
	conn.Write(pac)

	var username, password string

	userID := conn.GetUserID()

	err := connection.Db.QueryRow("SELECT username, password FROM users WHERE userID=?", userID).
		Scan(&username, &password)

	if err != nil {
		fmt.Println("handleCheckLogin database retrieval issue for userID:", userID, err)
	}

	hasher := sha512.New()
	hasher.Write([]byte(username + password)) // should be unique
	hash := hex.EncodeToString(hasher.Sum(nil))
	conn.SetSessionHash(hash)

	conn.WorldMngr = make(chan worlds.Message, 1)
	worlds.NewClient(conn.WorldMngr, conn.GetSessionHash())

	result := make(chan [][]byte)
	conn.WorldMngr <- worlds.Message{Opcode: worlds.WORLD_LIST, Message: result}
	worlds := <-result
	for world := range worlds {
		conn.Write(worlds[world])
	}

	pac = packet.NewPacket()
	pac.WriteByte(constants.LOGIN_SEND_WORLD_LIST)
	pac.WriteByte(0xFF) // Probs indicates end of list
	conn.Write(pac)
}
