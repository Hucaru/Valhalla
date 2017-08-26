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
	case constants.LOGIN_WORLD_SELECT:
		handleWorldSelect(buffer, &pos, conn)
	case constants.LOGIN_CHANNEL_SELECT:
		handleChannelSelect(buffer, &pos, conn)
	case constants.LOGIN_NAME_CHECK:
		handleNameCheck(buffer, &pos, conn)
	case constants.LOGIN_NEW_CHARACTER:
		handleNewCharacter(buffer, &pos, conn)
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

	pac := packet.NewPacket()
	pac.WriteByte(constants.LOGIN_SEND_WORLD_LIST)
	pac.WriteByte(0xFF) // Probs indicates end of list
	conn.Write(pac)
}

func handleWorldSelect(p packet.Packet, pos *int, conn *loginConn.Connection) {
	p.ReadInt16(pos) // World ID
	pac := packet.NewPacket()
	pac.WriteByte(constants.LOGIN_WORLD_META)
	pac.WriteByte(0x00) // Warning - 0 = no warning, 1 - high amount of concurent users, 2 = max uesrs in world
	pac.WriteByte(0x00) // Population marker - 0 = No maker, 1 = Highly populated, 2 = over populated
	conn.Write(pac)
}

func handleChannelSelect(p packet.Packet, pos *int, conn *loginConn.Connection) {
	p.ReadByte(pos) // world
	p.ReadByte(pos) // Channel

	pac := packet.NewPacket()
	pac.WriteByte(constants.LOGIN_CHARACTER_DATA)
	pac.WriteByte(0) // ?
	pac.WriteByte(0) // Character count
	conn.Write(pac)
}

func handleNameCheck(p packet.Packet, pos *int, conn *loginConn.Connection) {
	nameLength := p.ReadInt16(pos)
	newCharName := p.ReadString(pos, int(nameLength))

	pac := packet.NewPacket()
	pac.WriteByte(constants.LOGIN_NAME_CHECK_RESULT)
	pac.WriteString(newCharName)
	pac.WriteByte(0x0) // 0 = good name, 1 = bad name
	conn.Write(pac)
}

func handleNewCharacter(p packet.Packet, pos *int, conn *loginConn.Connection) {
	nameLength := p.ReadInt16(pos)
	name := p.ReadString(pos, int(nameLength))

	face := p.ReadInt32(pos)
	hair := p.ReadInt32(pos)
	hairColour := p.ReadInt32(pos)
	skin := p.ReadInt32(pos)
	top := p.ReadInt32(pos)
	bottom := p.ReadInt32(pos)
	shoes := p.ReadInt32(pos)
	weapon := p.ReadInt32(pos)

	str := p.ReadByte(pos)
	dex := p.ReadByte(pos)
	intelligence := p.ReadByte(pos)
	luk := p.ReadByte(pos)

	// Validate, name, equipment, stats
	// Insert into database if valid

	pac := packet.NewPacket()
	pac.WriteByte(0x0D)
	pac.WriteByte(0x0) // if creation was sucessfull - 0 = good, 1 = bad

	pac.WriteString(name)
	pac.WriteByte(0x0) //gender
	pac.WriteByte(byte(skin))
	pac.WriteByte(byte(face))
	pac.WriteByte(byte(hair))

	pac.WriteInt64(0x0) // Pet cash ID

	pac.WriteByte(200) // level
	pac.WriteInt16(0)  // Job
	pac.WriteInt16(int16(str))
	pac.WriteInt16(int16(dex))
	pac.WriteInt16(int16(intelligence))
	pac.WriteInt16(int16(luk))
	pac.WriteInt16(100) // hp
	pac.WriteInt16(100) // max hp
	pac.WriteInt16(100) // max mp
	pac.WriteInt16(100) // mp
	pac.WriteInt16(100) // ap
	pac.WriteInt16(100) // sp
	pac.WriteInt32(100) // exp
	pac.WriteInt16(100) // fame

	pac.WriteInt32(0) // map id
	pac.WriteByte(0)  // map pos

	pac.WriteByte(0x0) //gender
	pac.WriteByte(byte(skin))
	pac.WriteInt32(face)
	pac.WriteByte(0x0) // ?
	pac.WriteInt32(hair)

	// hidden equip - byte for type id , int for value
	// shown equip - byte for type id , int for value

	pac.WriteByte(0xFF)
	pac.WriteByte(0xFF)

	pac.WriteByte(0)  // Rankings
	pac.WriteInt32(0) // ?
	pac.WriteInt32(0) // world old pos

	conn.Write(pac)
}
