package client

import (
	"crypto/sha512"
	"encoding/hex"
	"log"
	"net"
	"os"
	"strings"

	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/loginServer/handlers"
	"github.com/Hucaru/gopacket"
)

func Handle() {
	log.Println("LoginServer")

	listener, err := net.Listen("tcp", "0.0.0.0:8484")

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer connection.Db.Close()
	connection.ConnectToDb()

	log.Println("Client listener ready")

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error in accepting client", err)
		}

		defer conn.Close()
		clientConnection := NewConnection(conn)

		log.Println("New client connection from", clientConnection)

		go connection.HandleNewConnection(clientConnection, func(p gopacket.Reader) {
			handlePacket(clientConnection, p)
		}, constants.CLIENT_HEADER_SIZE, true)
	}
}

func handlePacket(conn *Connection, reader gopacket.Reader) {
	switch reader.ReadByte() {
	case constants.RECV_RETURN_TO_LOGIN_SCREEN:
		handleReturnToLoginScreen(reader, conn)
	case constants.RECV_LOGIN_REQUEST:
		handleLoginRequest(reader, conn)
	case constants.RECV_LOGIN_CHECK_LOGIN:
		handleGoodLogin(reader, conn)
	case constants.RECV_LOGIN_WORLD_SELECT:
		handleWorldSelect(reader, conn)
	case constants.RECV_LOGIN_CHANNEL_SELECT:
		handleChannelSelect(reader, conn)
	case constants.RECV_LOGIN_NAME_CHECK:
		handleNameCheck(reader, conn)
	case constants.RECV_LOGIN_NEW_CHARACTER:
		handleNewCharacter(reader, conn)
	case constants.RECV_LOGIN_DELETE_CHAR:
		handleDeleteCharacter(reader, conn)
	case constants.RECV_LOGIN_SELECT_CHARACTER:
		handleSelectCharacter(reader, conn)
	default:
		log.Println("UNKNOWN LOGIN PACKET:", reader)
	}

}

func handleReturnToLoginScreen(reader gopacket.Reader, conn *Connection) {
	conn.Write(channelToLogin())
}

func handleLoginRequest(reader gopacket.Reader, conn *Connection) {
	usernameLength := reader.ReadInt16()
	username := reader.ReadString(int(usernameLength))

	passwordLength := reader.ReadInt16()
	password := reader.ReadString(int(passwordLength))

	// hash the password, cba to salt atm
	hasher := sha512.New()
	hasher.Write([]byte(password))
	hashedPassword := hex.EncodeToString(hasher.Sum(nil))

	var userID uint32
	var user string
	var databasePassword string
	var gender byte
	var isLogedIn bool
	var isBanned int
	var isAdmin byte

	err := connection.Db.QueryRow("SELECT userID, username, password, gender, isLogedIn, isBanned, isAdmin FROM users WHERE username=?", username).
		Scan(&userID, &user, &databasePassword, &gender, &isLogedIn, &isBanned, &isAdmin)

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

	// Banned = 2, Deleted or Blocked = 3, Invalid Password = 4, Not Registered = 5, Sys Error = 6,
	// Already online = 7, System error = 9, Too many requests = 10, Older than 20 = 11, Master cannot login on this IP = 13

	if result <= 0x01 {
		conn.SetGender(gender)
		conn.SetAdmin(byte(0x01) == isAdmin)
		conn.SetUserID(userID)
		conn.SetIsLogedIn(true)
		records, err := connection.Db.Query("UPDATE users set isLogedIn=1 WHERE userID=?", userID)

		defer records.Close()

		if err != nil {
			log.Println("Database error with approving login of userID", userID, err)
		} else {
			log.Println("User", userID, "has logged in from", conn)
		}
	}

	conn.Write(loginResponce(result, userID, gender, isAdmin, username, isBanned))
}

func handleGoodLogin(reader gopacket.Reader, conn *Connection) {
	var username, password string

	userID := conn.GetUserID()

	err := connection.Db.QueryRow("SELECT username, password FROM users WHERE userID=?", userID).
		Scan(&username, &password)

	if err != nil {
		log.Println("handleCheckLogin database retrieval issue for userID:", userID, err)
	}

	hasher := sha512.New()
	hasher.Write([]byte(username + password)) // should be unique
	hash := hex.EncodeToString(hasher.Sum(nil))
	conn.SetSessionHash(hash)

	returnChan := make(chan gopacket.Packet)
	handlers.LoginServer <- connection.NewMessage(handlers.RequestWorlds(conn.IsAdmin()), returnChan)
	worldsPacket := <-returnChan
	worlds := gopacket.NewReader(&worldsPacket)
	nWorlds := int(worlds.ReadByte())

	for i := 0; i < nWorlds; i++ {
		conn.Write(worlds.ReadBytes(int(worlds.ReadInt16())))
	}

	conn.Write(endWorldList())
}

func handleWorldSelect(reader gopacket.Reader, conn *Connection) {
	worldID := reader.ReadInt16()
	conn.SetWorldID(uint32(worldID))

	returnChan := make(chan gopacket.Packet)
	handlers.LoginServer <- connection.NewMessage(handlers.RequestWorldInfo(worldID), returnChan)
	worldsInfo := <-returnChan
	conn.Write(worldsInfo)
}

func handleChannelSelect(reader gopacket.Reader, conn *Connection) {
	selectedWorld := reader.ReadByte() // world
	conn.SetChanID(reader.ReadByte())  // Channel

	var characters []character.Character

	if uint32(selectedWorld) == conn.GetWorldID() {
		characters = character.GetCharacters(conn.GetUserID(), conn.GetWorldID())
	}

	conn.Write(displayCharacters(characters))
}

func handleNameCheck(reader gopacket.Reader, conn *Connection) {
	nameLength := reader.ReadInt16()
	newCharName := reader.ReadString(int(nameLength))

	var nameFound int
	err := connection.Db.QueryRow("SELECT count(*) name FROM characters WHERE name=?", newCharName).
		Scan(&nameFound)

	if err != nil {
		panic(err.Error())
	}

	conn.Write(nameCheck(newCharName, nameFound))
}

func handleNewCharacter(reader gopacket.Reader, conn *Connection) {
	nameLength := reader.ReadInt16()
	name := reader.ReadString(int(nameLength))
	face := reader.ReadInt32()
	hair := reader.ReadInt32()
	hairColour := reader.ReadInt32()
	skin := reader.ReadInt32()
	top := reader.ReadInt32()
	bottom := reader.ReadInt32()
	shoes := reader.ReadInt32()
	weapon := reader.ReadInt32()

	str := reader.ReadByte()
	dex := reader.ReadByte()
	intelligence := reader.ReadByte()
	luk := reader.ReadByte()

	// Add str, dex, int, luk validation (check to see if client generates a constant sum)

	var counter int

	err := connection.Db.QueryRow("SELECT count(*) FROM characters where name=? and worldID=?", name, conn.GetWorldID()).Scan(&counter)

	if err != nil {
		panic(err.Error())
	}

	allowedEyes := []int32{20000, 20001, 20002, 21000, 21001, 21002, 20100, 20401, 20402, 21700, 21201, 21002}
	allowedHair := []int32{30000, 30020, 30030, 31000, 31040, 31050}
	allowedHairColour := []int32{0, 7, 3, 2}
	allowedBottom := []int32{1060002, 1060006, 1061002, 1061008, 1062115}
	allowedTop := []int32{1040002, 1040006, 1040010, 1041002, 1041006, 1041010, 1041011, 1042167}
	allowedShoes := []int32{1072001, 1072005, 1072037, 1072038, 1072383}
	allowedWeapons := []int32{1302000, 1322005, 1312004, 1442079}
	allowedSkinColour := []int32{0, 1, 2, 3}

	inSlice := func(val int32, s []int32) bool {
		for _, b := range s {
			if b == val {
				return true
			}
		}
		return false
	}

	valid := inSlice(face, allowedEyes) && inSlice(hair, allowedHair) && inSlice(hairColour, allowedHairColour) &&
		inSlice(bottom, allowedBottom) && inSlice(top, allowedTop) && inSlice(shoes, allowedShoes) &&
		inSlice(weapon, allowedWeapons) && inSlice(skin, allowedSkinColour) && (counter == 0)

	var newCharacter character.Character

	if conn.IsAdmin() {
		name = "[GM]" + name
	} else if strings.ContainsAny(name, "[]") {
		valid = false // hacked client or packet editting
	}

	if valid {
		res, err := connection.Db.Exec("INSERT INTO characters (name, userID, worldID, face, hair, skin, gender, str, dex, `int`, luk) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			name, conn.GetUserID(), conn.GetWorldID(), face, hair+hairColour, skin, conn.GetGender(), str, dex, intelligence, luk)

		characterID, err := res.LastInsertId()

		if err != nil {
			panic(err.Error())
		}

		if conn.IsAdmin() {
			addCharacterItem(characterID, 1002140, -1) // Hat
			addCharacterItem(characterID, 1032006, -4) // Earrings
			addCharacterItem(characterID, 1042003, -5)
			addCharacterItem(characterID, 1062007, -6)
			addCharacterItem(characterID, 1072004, -7)
			addCharacterItem(characterID, 1082002, -8)  // Gloves
			addCharacterItem(characterID, 1102054, -9)  // Cape
			addCharacterItem(characterID, 1092008, -10) // Shield
			addCharacterItem(characterID, 1322013, -11)
		} else {
			addCharacterItem(characterID, top, -5)
			addCharacterItem(characterID, bottom, -6)
			addCharacterItem(characterID, shoes, -7)
			addCharacterItem(characterID, weapon, -11)
		}

		if err != nil {
			panic(err.Error())
		}

		characters := character.GetCharacters(conn.GetUserID(), conn.GetWorldID())
		newCharacter = characters[len(characters)-1]
	}

	conn.Write(createdCharacter(valid, newCharacter))
}

func handleDeleteCharacter(reader gopacket.Reader, conn *Connection) {
	dob := reader.ReadInt32()
	charID := reader.ReadInt32()

	var storedDob int32
	var charCount int

	err := connection.Db.QueryRow("SELECT dob FROM users where userID=?", conn.GetUserID()).Scan(&storedDob)
	err = connection.Db.QueryRow("SELECT count(*) FROM characters where userID=? AND id=?", conn.GetUserID(), charID).Scan(&charCount)

	if err != nil {
		panic(err.Error())
	}

	hacking := false
	deleted := false

	if charCount != 1 {
		log.Println(conn.GetUserID(), "attempted to delete a character they do not own:", charID)
		hacking = true
	}

	if dob == storedDob {
		records, err := connection.Db.Query("DELETE FROM items where characterID=?", charID)

		if err != nil {
			panic(err.Error())
		}

		records.Close()

		records, err = connection.Db.Query("DELETE FROM characters where id=?", charID)

		if err != nil {
			panic(err.Error())
		}

		records.Close()

		deleted = true
	}

	conn.Write(deleteCharacter(charID, deleted, hacking))
}

func handleSelectCharacter(reader gopacket.Reader, conn *Connection) {
	charID := reader.ReadInt32()

	var charCount int

	err := connection.Db.QueryRow("SELECT count(*) FROM characters where userID=? AND id=?", conn.GetUserID(), charID).Scan(&charCount)

	if err != nil {
		panic(err.Error())
	}

	if charCount == 1 {
		returnChan := make(chan gopacket.Packet)
		handlers.LoginServer <- connection.NewMessage(handlers.RequestMigrationInfo(conn.GetWorldID(), conn.GetChanID(), charID), returnChan)
		p := <-returnChan

		info := gopacket.NewReader(&p)
		ip := info.ReadBytes(4)
		port := info.ReadUint16()

		if port != 0xFFFF {
			log.Println("Migrating", charID, "to:", ip, ":", port)
			conn.Write(migrateClient(ip, port, charID))
		} else {
			log.Println("Bad migrate for char", charID)
			conn.Write(sendBadMigrate())
		}

	}
}

func addCharacterItem(characterID int64, itemID int32, slot int32) {
	_, err := connection.Db.Exec("INSERT INTO items (characterID, itemID, slotNumber) VALUES (?, ?, ?)", characterID, itemID, slot)

	if err != nil {
		panic(err.Error())
	}
}
