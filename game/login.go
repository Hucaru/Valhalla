package game

import (
	"crypto/sha512"
	"encoding/hex"
	"log"
	"strings"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/database"
	"github.com/Hucaru/Valhalla/game/entity"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// Login server state
type Login struct {
}

// ClientConnected to server
func (server *Login) ClientConnected(conn mnet.Client, keyRecv, keySend []byte) {
	conn.Send(entity.PacketClientHandshake(constant.MapleVersion, keyRecv, keySend))
}

// HandleClientPacket from client
func (server *Login) HandleClientPacket(conn mnet.Client, reader mpacket.Reader) {
	switch mpacket.Opcode(reader.ReadByte()) {
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
		log.Println("UNKNOWN LOGIN PACKET:", reader)
	}
}
func (server *Login) handleLoginRequest(conn mnet.Client, reader mpacket.Reader) {
	usernameLength := reader.ReadInt16()
	username := reader.ReadString(int(usernameLength))

	passwordLength := reader.ReadInt16()
	password := reader.ReadString(int(passwordLength))

	// hash the password, cba to salt atm
	hasher := sha512.New()
	hasher.Write([]byte(password))
	hashedPassword := hex.EncodeToString(hasher.Sum(nil))

	var accountID int32
	var user string
	var databasePassword string
	var gender byte
	var isLogedIn bool
	var isInChannel int8
	var isBanned int
	var adminLevel int

	err := database.Handle.QueryRow("SELECT accountID, username, password, gender, isLogedIn, isBanned, adminLevel, isInChannel FROM accounts WHERE username=?", username).
		Scan(&accountID, &user, &databasePassword, &gender, &isLogedIn, &isBanned, &adminLevel, &isInChannel)

	result := byte(0x00)

	if err != nil {
		result = 0x05
	} else if hashedPassword != databasePassword {
		result = 0x04
	} else if isLogedIn || isInChannel > -1 {
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

		_, err := database.Handle.Exec("UPDATE accounts set isLogedIn=1 WHERE accountID=?", accountID)

		if err != nil {
			log.Println("Database error with approving login of accountID", accountID, err)
		} else {
			log.Println("User", accountID, "has logged in from", conn)
		}
	}

	conn.Send(entity.PacketLoginResponce(result, accountID, gender, adminLevel > 0, username, isBanned))
}

func (server *Login) handleGoodLogin(conn mnet.Client, reader mpacket.Reader) {
	var username, password string

	accountID := conn.GetAccountID()

	err := database.Handle.QueryRow("SELECT username, password FROM accounts WHERE accountID=?", accountID).
		Scan(&username, &password)

	if err != nil {
		log.Println("handleCheckLogin database retrieval issue for accountID:", accountID, err)
	}

	const maxNumberOfWorlds = 14

	for i := maxNumberOfWorlds; i > -1; i-- {
		conn.Send(entity.PacketLoginWorldListing(byte(i))) // hard coded for now
	}
	conn.Send(entity.PacketLoginEndWorldList())
}

func (server *Login) handleWorldSelect(conn mnet.Client, reader mpacket.Reader) {
	conn.SetWorldID(reader.ReadByte())
	reader.ReadByte() // ?

	conn.Send(entity.PacketLoginWorldInfo(0, 0)) // hard coded for now
}

func (server *Login) handleChannelSelect(conn mnet.Client, reader mpacket.Reader) {
	selectedWorld := reader.ReadByte()   // world
	conn.SetChannelID(reader.ReadByte()) // Channel

	if selectedWorld == conn.GetWorldID() {
		characters := entity.GetCharactersFromAccountWorldID(conn.GetAccountID(), conn.GetWorldID())
		conn.Send(entity.PacketLoginDisplayCharacters(characters))
	}
}

func (server *Login) handleNameCheck(conn mnet.Client, reader mpacket.Reader) {
	nameLength := reader.ReadInt16()
	newCharName := reader.ReadString(int(nameLength))

	var nameFound int
	err := database.Handle.QueryRow("SELECT count(*) name FROM characters WHERE name=?", newCharName).
		Scan(&nameFound)

	if err != nil {
		panic(err)
	}

	conn.Send(entity.PacketLoginNameCheck(newCharName, nameFound))
}

func (server *Login) handleNewCharacter(conn mnet.Client, reader mpacket.Reader) {
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

	err := database.Handle.QueryRow("SELECT count(*) FROM characters where name=? and worldID=?", name, conn.GetWorldID()).Scan(&counter)

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

	newCharacter := entity.Character{}

	if conn.GetAdminLevel() > 0 {
		name = "[GM]" + name
	} else if strings.ContainsAny(name, "[]") {
		valid = false // hacked client or packet editting
	}

	if valid {
		res, err := database.Handle.Exec("INSERT INTO characters (name, accountID, worldID, face, hair, skin, gender, str, dex, intt, luk) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
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

		characters := entity.GetCharactersFromAccountWorldID(conn.GetAccountID(), conn.GetWorldID())
		newCharacter = characters[len(characters)-1]
	}

	conn.Send(entity.PacketLoginCreatedCharacter(valid, newCharacter))
}

func (server *Login) handleDeleteCharacter(conn mnet.Client, reader mpacket.Reader) {
	dob := reader.ReadInt32()
	charID := reader.ReadInt32()

	var storedDob int32
	var charCount int

	err := database.Handle.QueryRow("SELECT dob FROM accounts where accountID=?", conn.GetAccountID()).Scan(&storedDob)
	err = database.Handle.QueryRow("SELECT count(*) FROM characters where accountID=? AND id=?", conn.GetAccountID(), charID).Scan(&charCount)

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
		records, err := database.Handle.Query("DELETE FROM characters where id=?", charID)

		if err != nil {
			panic(err)
		}

		records.Close()

		deleted = true
	}

	conn.Send(entity.PacketLoginDeleteCharacter(charID, deleted, hacking))
}

func (server *Login) handleSelectCharacter(conn mnet.Client, reader mpacket.Reader) {
	charID := reader.ReadInt32()

	var charCount int

	err := database.Handle.QueryRow("SELECT count(*) FROM characters where accountID=? AND id=?", conn.GetAccountID(), charID).Scan(&charCount)

	if err != nil {
		panic(err)
	}

	if charCount == 1 {
		ip := []byte{192, 168, 1, 240}
		port := int16(8684)
		conn.Send(entity.PacketLoginMigrateClient(ip, port, charID))
	}
}

func (server *Login) addCharacterItem(characterID int64, itemID int32, slot int32, creatorName string) {
	_, err := database.Handle.Exec("INSERT INTO items (characterID, itemID, slotNumber, creatorName) VALUES (?, ?, ?, ?)", characterID, itemID, slot, creatorName)

	if err != nil {
		panic(err)
	}
}

func (server *Login) handleReturnToLoginScreen(conn mnet.Client, reader mpacket.Reader) {
	_, err := database.Handle.Exec("UPDATE accounts SET isLogedIn=0 WHERE accountID=?", conn.GetAccountID())

	if err != nil {
		panic(err)
	}

	conn.Send(entity.PacketLoginReturnFromChannel())
}

// HandleServerPacket from client
func (server *Login) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
}
