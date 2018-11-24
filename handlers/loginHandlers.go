package handlers

import (
	"crypto/sha512"
	"encoding/hex"
	"log"
	"strings"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/database"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/game/packet"
)

func handleReturnToLoginScreen(conn mnet.MConnLogin, reader mpacket.Reader) {
	conn.Send(packet.LoginReturnFromChannel())
}

func handleLoginRequest(conn mnet.MConnLogin, reader mpacket.Reader) {
	usernameLength := reader.ReadInt16()
	username := reader.ReadString(int(usernameLength))

	passwordLength := reader.ReadInt16()
	password := reader.ReadString(int(passwordLength))

	// hash the password, cba to salt atm
	hasher := sha512.New()
	hasher.Write([]byte(password))
	hashedPassword := hex.EncodeToString(hasher.Sum(nil))

	var userID int32
	var user string
	var databasePassword string
	var gender byte
	var isLogedIn bool
	var isBanned int
	var isAdmin byte

	err := database.Db.QueryRow("SELECT userID, username, password, gender, isLogedIn, isBanned, isAdmin FROM users WHERE username=?", username).
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
		conn.SetAccountID(userID)

		records, err := database.Db.Query("UPDATE users set isLogedIn=1 WHERE userID=?", userID)

		defer records.Close()

		if err != nil {
			log.Println("Database error with approving login of userID", userID, err)
		} else {
			log.Println("User", userID, "has logged in from", conn)
		}
	}

	conn.Send(packet.LoginResponce(result, userID, gender, isAdmin, username, isBanned))
}

func handleGoodLogin(conn mnet.MConnLogin, reader mpacket.Reader) {
	var username, password string

	userID := conn.GetUserID()

	err := database.Db.QueryRow("SELECT username, password FROM users WHERE userID=?", userID).
		Scan(&username, &password)

	if err != nil {
		log.Println("handleCheckLogin database retrieval issue for userID:", userID, err)
	}

	hasher := sha512.New()
	hasher.Write([]byte(username + password)) // should be unique
	hash := hex.EncodeToString(hasher.Sum(nil))
	conn.SetSessionHash(hash)

	const maxNumberOfWorlds = 14

	for i := maxNumberOfWorlds; i > -1; i-- {
		conn.Send(packet.LoginWorldListing(byte(i))) // hard coded for now
	}
	conn.Send(packet.LoginEndWorldList())
}

func handleWorldSelect(conn mnet.MConnLogin, reader mpacket.Reader) {
	worldID := reader.ReadInt16()
	conn.SetWorldID(int32(worldID))

	conn.Send(packet.LoginWorldInfo(0, 0)) // hard coded for now
}

func handleChannelSelect(conn mnet.MConnLogin, reader mpacket.Reader) {
	selectedWorld := reader.ReadByte() // world
	conn.SetChanID(reader.ReadByte())  // Channel

	var characters []character.Character

	if int32(selectedWorld) == conn.GetWorldID() {
		characters = character.GetCharacters(conn.GetUserID(), conn.GetWorldID())
	}

	conn.Send(packet.LoginDisplayCharacters(characters))
}

func handleNameCheck(conn mnet.MConnLogin, reader mpacket.Reader) {
	nameLength := reader.ReadInt16()
	newCharName := reader.ReadString(int(nameLength))

	var nameFound int
	err := database.Db.QueryRow("SELECT count(*) name FROM characters WHERE name=?", newCharName).
		Scan(&nameFound)

	if err != nil {
		panic(err.Error())
	}

	conn.Send(packet.LoginNameCheck(newCharName, nameFound))
}

func handleNewCharacter(conn mnet.MConnLogin, reader mpacket.Reader) {
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

	err := database.Db.QueryRow("SELECT count(*) FROM characters where name=? and worldID=?", name, conn.GetWorldID()).Scan(&counter)

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
		res, err := database.Db.Exec("INSERT INTO characters (name, userID, worldID, face, hair, skin, gender, str, dex, intt, luk) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
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

	conn.Send(packet.LoginCreatedCharacter(valid, newCharacter))
}

func handleDeleteCharacter(conn mnet.MConnLogin, reader mpacket.Reader) {
	dob := reader.ReadInt32()
	charID := reader.ReadInt32()

	var storedDob int32
	var charCount int

	err := database.Db.QueryRow("SELECT dob FROM users where userID=?", conn.GetUserID()).Scan(&storedDob)
	err = database.Db.QueryRow("SELECT count(*) FROM characters where userID=? AND id=?", conn.GetUserID(), charID).Scan(&charCount)

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
		records, err := database.Db.Query("DELETE FROM characters where id=?", charID)

		if err != nil {
			panic(err.Error())
		}

		records.Close()

		deleted = true
	}

	conn.Send(packet.LoginDeleteCharacter(charID, deleted, hacking))
}

func handleSelectCharacter(conn mnet.MConnLogin, reader mpacket.Reader) {
	charID := reader.ReadInt32()

	var charCount int

	err := database.Db.QueryRow("SELECT count(*) FROM characters where userID=? AND id=?", conn.GetUserID(), charID).Scan(&charCount)

	if err != nil {
		panic(err.Error())
	}

	if charCount == 1 {
		ip := []byte{192, 168, 1, 240}
		port := int16(8686)
		conn.Send(packet.LoginMigrateClient(ip, port, charID))
	}
}

func addCharacterItem(characterID int64, itemID int32, slot int32) {
	_, err := database.Db.Exec("INSERT INTO items (characterID, itemID, slotNumber, creatorName) VALUES (?, ?, ?, ?)", characterID, itemID, slot, "")

	if err != nil {
		panic(err.Error())
	}
}
