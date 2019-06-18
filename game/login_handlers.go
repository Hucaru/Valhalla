package game

import (
	"crypto/sha512"
	"encoding/hex"
	"log"
	"strings"

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// HandleClientPacket data
func (server *LoginServer) HandleClientPacket(conn mnet.Client, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.RecvLoginRequest:
		server.handleLoginRequest(conn, reader)
	case opcode.RecvLoginCheckLogin:
		server.handleGoodLogin(conn, reader)
	case opcode.RecvLoginWorldSelect:
		server.handleWorldSelect(conn, reader)
	case opcode.RecvLoginChannelSelect:
		server.handleChannelSelect(conn, reader)
	case opcode.RecvLoginNameCheck:
		server.handleNameCheck(conn, reader)
	case opcode.RecvLoginNewCharacter:
		server.handleNewCharacter(conn, reader)
	case opcode.RecvLoginDeleteChar:
		server.handleDeleteCharacter(conn, reader)
	case opcode.RecvLoginSelectCharacter:
		server.handleSelectCharacter(conn, reader)
	case opcode.RecvReturnToLoginScreen:
		server.handleReturnToLoginScreen(conn, reader)
	default:
		log.Println("UNKNOWN CLIENT PACKET:", reader)
	}
}
func (server *LoginServer) handleLoginRequest(conn mnet.Client, reader mpacket.Reader) {
	username := reader.ReadString(reader.ReadInt16())
	password := reader.ReadString(reader.ReadInt16())

	// hash the password, cba to salt atm
	hasher := sha512.New()
	hasher.Write([]byte(password))
	hashedPassword := hex.EncodeToString(hasher.Sum(nil))

	var accountID int32
	var user string
	var databasePassword string
	var gender byte
	var isLogedIn bool
	var isBanned int
	var adminLevel int

	err := server.db.QueryRow("SELECT accountID, username, password, gender, isLogedIn, isBanned, adminLevel FROM accounts WHERE username=?", username).
		Scan(&accountID, &user, &databasePassword, &gender, &isLogedIn, &isBanned, &adminLevel)

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
		conn.SetLogedIn(true)
		conn.SetGender(gender)
		conn.SetAdminLevel(adminLevel)
		conn.SetAccountID(accountID)

		_, err := server.db.Exec("UPDATE accounts set isLogedIn=1 WHERE accountID=?", accountID)

		if err != nil {
			log.Println("Database error with approving login of accountID", accountID, err)
		} else {
			log.Println("User", accountID, "has logged in from", conn)
		}
	}

	conn.Send(packetLoginResponce(result, accountID, gender, adminLevel > 0, username, isBanned))
}

func (server *LoginServer) handleGoodLogin(conn mnet.Client, reader mpacket.Reader) {
	server.migrating[conn] = false
	var username, password string

	accountID := conn.GetAccountID()

	err := server.db.QueryRow("SELECT username, password FROM accounts WHERE accountID=?", accountID).
		Scan(&username, &password)

	if err != nil {
		log.Println("handleCheckLogin database retrieval issue for accountID:", accountID, err)
	}

	const maxNumberOfWorlds = 14

	for i := len(server.worlds) - 1; i > -1; i-- {
		conn.Send(packetLoginWorldListing(byte(i), server.worlds[i]))
	}

	conn.Send(packetLoginEndWorldList())
}

func (server *LoginServer) handleWorldSelect(conn mnet.Client, reader mpacket.Reader) {
	conn.SetWorldID(reader.ReadByte())
	reader.ReadByte() // ?

	var warning, population byte = 0, 0

	if conn.GetAdminLevel() < 1 { // gms are not restricted in any capacity
		var currentPlayers int16
		var maxPlayers int16

		for _, v := range server.worlds[conn.GetWorldID()].channels {
			currentPlayers += v.pop
			maxPlayers += v.maxPop
		}

		if currentPlayers >= maxPlayers {
			warning = 2
		} else if float64(currentPlayers)/float64(maxPlayers) > 0.95 { // I'm not sure if this warning is even worth it
			warning = 1
		}

		// implement server total registered characters lookup for population field
	}

	conn.Send(packetLoginWorldInfo(warning, population)) // hard coded for now
}

func (server *LoginServer) handleChannelSelect(conn mnet.Client, reader mpacket.Reader) {
	selectedWorld := reader.ReadByte()   // world
	conn.SetChannelID(reader.ReadByte()) // Channel

	if server.worlds[selectedWorld].channels[conn.GetChannelID()].maxPop == 0 {
		conn.Send(packetMessageDialogueBox("Channel currently unavailable"))
		return
	}

	if selectedWorld == conn.GetWorldID() {
		characters := GetCharactersFromAccountWorldID(server.db, conn.GetAccountID(), conn.GetWorldID())
		conn.Send(packetLoginDisplayCharacters(characters))
	}
}

func (server *LoginServer) handleNameCheck(conn mnet.Client, reader mpacket.Reader) {
	newCharName := reader.ReadString(reader.ReadInt16())

	var nameFound int
	err := server.db.QueryRow("SELECT count(*) name FROM characters WHERE name=?", newCharName).
		Scan(&nameFound)

	if err != nil {
		panic(err)
	}

	conn.Send(packetLoginNameCheck(newCharName, nameFound))
}

func (server *LoginServer) handleNewCharacter(conn mnet.Client, reader mpacket.Reader) {
	name := reader.ReadString(reader.ReadInt16())
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

	err := server.db.QueryRow("SELECT count(*) FROM characters where name=? and worldID=?", name, conn.GetWorldID()).Scan(&counter)

	if err != nil {
		panic(err)
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

	newCharacter := character{}

	if conn.GetAdminLevel() > 0 {
		name = "[GM]" + name
	} else if strings.ContainsAny(name, "[]") {
		valid = false // hacked client or packet editting
	}

	if valid {
		res, err := server.db.Exec("INSERT INTO characters (name, accountID, worldID, face, hair, skin, gender, str, dex, intt, luk) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			name, conn.GetAccountID(), conn.GetWorldID(), face, hair+hairColour, skin, conn.GetGender(), str, dex, intelligence, luk)

		characterID, err := res.LastInsertId()

		if err != nil {
			panic(err)
		}

		if conn.GetAdminLevel() > 0 {
			server.addCharacterItem(characterID, 1002140, -1, name) // Hat
			server.addCharacterItem(characterID, 1032006, -4, name) // Earrings
			server.addCharacterItem(characterID, 1042003, -5, name)
			server.addCharacterItem(characterID, 1062007, -6, name)
			server.addCharacterItem(characterID, 1072004, -7, name)
			server.addCharacterItem(characterID, 1082002, -8, name)  // Gloves
			server.addCharacterItem(characterID, 1102054, -9, name)  // Cape
			server.addCharacterItem(characterID, 1092008, -10, name) // Shield
			server.addCharacterItem(characterID, 1322013, -11, name)
		} else {
			server.addCharacterItem(characterID, top, -5, "")
			server.addCharacterItem(characterID, bottom, -6, "")
			server.addCharacterItem(characterID, shoes, -7, "")
			server.addCharacterItem(characterID, weapon, -11, "")
		}

		if err != nil {
			panic(err)
		}

		characters := GetCharactersFromAccountWorldID(server.db, conn.GetAccountID(), conn.GetWorldID())
		newCharacter = characters[len(characters)-1]
	}

	conn.Send(packetLoginCreatedCharacter(valid, newCharacter))
}

func (server *LoginServer) handleDeleteCharacter(conn mnet.Client, reader mpacket.Reader) {
	dob := reader.ReadInt32()
	charID := reader.ReadInt32()

	var storedDob int32
	var charCount int

	err := server.db.QueryRow("SELECT dob FROM accounts where accountID=?", conn.GetAccountID()).Scan(&storedDob)
	err = server.db.QueryRow("SELECT count(*) FROM characters where accountID=? AND id=?", conn.GetAccountID(), charID).Scan(&charCount)

	if err != nil {
		panic(err)
	}

	hacking := false
	deleted := false

	if charCount != 1 {
		log.Println(conn.GetAccountID(), "attempted to delete a character they do not own:", charID)
		hacking = true
	}

	if dob == storedDob {
		records, err := server.db.Query("DELETE FROM characters where id=?", charID)

		if err != nil {
			panic(err)
		}

		records.Close()

		deleted = true
	}

	conn.Send(packetLoginDeleteCharacter(charID, deleted, hacking))
}

func (server *LoginServer) handleSelectCharacter(conn mnet.Client, reader mpacket.Reader) {
	charID := reader.ReadInt32()

	var charCount int

	err := server.db.QueryRow("SELECT count(*) FROM characters where accountID=? AND id=?", conn.GetAccountID(), charID).Scan(&charCount)

	if err != nil {
		panic(err)
	}

	if charCount == 1 {
		channel := server.worlds[conn.GetWorldID()].channels[conn.GetChannelID()]
		_, err := server.db.Exec("UPDATE characters SET migrationID=? WHERE id=?", conn.GetChannelID(), charID)

		if err != nil {
			panic(err)
		}

		server.migrating[conn] = true

		conn.Send(packetLoginMigrateClient(channel.ip, channel.port, charID))
	}
}

func (server *LoginServer) addCharacterItem(characterID int64, itemID int32, slot int32, creatorName string) {
	_, err := server.db.Exec("INSERT INTO items (characterID, itemID, slotNumber, creatorName) VALUES (?, ?, ?, ?)", characterID, itemID, slot, creatorName)

	if err != nil {
		panic(err)
	}
}

func (server *LoginServer) handleReturnToLoginScreen(conn mnet.Client, reader mpacket.Reader) {
	conn.Send(packetLoginReturnFromChannel())
}
