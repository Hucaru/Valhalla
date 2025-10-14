package login

import (
	"crypto/sha512"
	"encoding/hex"
	"log"
	"strings"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// HandleClientPacket data
func (server *Server) HandleClientPacket(conn mnet.Client, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.RecvLoginRequest:
		server.handleLoginRequest(conn, reader)
	case opcode.RecvLoginEULA:
		server.handleEULA(conn, reader)
	case opcode.RecvLoginCheckLogin:
		server.handleGoodLogin(conn, reader)
	case opcode.RecvLoginRegisterPin:
		server.handlePinRegistration(conn, reader)
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

func (server *Server) handleLoginRequest(conn mnet.Client, reader mpacket.Reader) {
	username := reader.ReadString(reader.ReadInt16())
	password := reader.ReadString(reader.ReadInt16())
	// hash the password
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
	var eula byte

	err := common.DB.QueryRow("SELECT accountID, username, password, gender, isLogedIn, isBanned, adminLevel, eula FROM accounts WHERE username=?", username).
		Scan(&accountID, &user, &databasePassword, &gender, &isLogedIn, &isBanned, &adminLevel, &eula)

	result := constant.LoginResultSuccess

	if err != nil {
		// Account does not exist
		if server.autoRegister {
			// Create new account with default values
			res, insertErr := common.DB.Exec("INSERT INTO accounts (username, password, pin, isLogedIn, adminLevel, isBanned, gender, dob, eula, nx, maplepoints) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				username, hashedPassword, constant.AutoRegisterDefaultPIN, constant.AutoRegisterDefaultIsLoggedIn, 
				constant.AutoRegisterDefaultAdminLevel, constant.AutoRegisterDefaultIsBanned, constant.AutoRegisterDefaultGender, 
				constant.AutoRegisterDefaultDOB, constant.AutoRegisterDefaultEULA, constant.AutoRegisterDefaultNX, 
				constant.AutoRegisterDefaultMaplePoints)
			
			if insertErr != nil {
				log.Println("Failed to auto-register account:", insertErr)
				result = constant.LoginResultNotRegistered
			} else {
				// Get the newly created account ID
				newAccountID, lastInsertErr := res.LastInsertId()
				if lastInsertErr != nil {
					log.Println("Failed to get new account ID:", lastInsertErr)
					result = constant.LoginResultNotRegistered
				} else {
					accountID = int32(newAccountID)
					gender = constant.AutoRegisterDefaultGender
					adminLevel = constant.AutoRegisterDefaultAdminLevel
					eula = constant.AutoRegisterDefaultEULA
					log.Println("Auto-registered new account:", username, "with ID:", accountID)
					// Account created successfully, proceed with login
					result = constant.LoginResultSuccess
				}
			}
		} else {
			result = constant.LoginResultNotRegistered
		}
	} else if hashedPassword != databasePassword {
		result = constant.LoginResultInvalidPassword
	} else if isLogedIn {
		result = constant.LoginResultAlreadyOnline
	} else if isBanned > 0 {
		result = constant.LoginResultBanned
	} else if eula == 0 {
		result = constant.LoginResultEULA
	}

	// Banned = 2, Deleted or Blocked = 3, Invalid Password = 4, Not Registered = 5, Sys Error = 6,
	// Already online = 7, System error = 9, Too many requests = 10, Older than 20 = 11, valid login = 12, Master cannot login on this IP = 13,
	// wrong gateway korean text = 14, still processing request korean text = 15, verify email = 16, gateway english text = 17,
	// verify email = 21, eula = 23

	if result <= constant.LoginResultSuccess {
		conn.SetGender(gender)
		conn.SetAdminLevel(adminLevel)
		conn.SetAccountID(accountID)
	} else if result == constant.LoginResultEULA {
		conn.SetAccountID(accountID)
	}

	conn.Send(packetLoginResponse(result, accountID, gender, adminLevel > 0, username, isBanned))
}

func (server *Server) handleEULA(conn mnet.Client, reader mpacket.Reader) {
	accept := reader.ReadBool()

	if accept {
		_, err := common.DB.Exec("UPDATE accounts SET eula=? WHERE accountID=?", 1, conn.GetAccountID())

		if err != nil {
			log.Println("Could not set EULA signed", err)
		}

		conn.Send(packetLoginReturnFromChannel())
	}
}

func (server *Server) handlePinRegistration(conn mnet.Client, reader mpacket.Reader) {
	b1 := reader.ReadByte()

	if b1 == 0 { // Client canceled pin change request
		conn.Send(packetCancelPin())
		return
	}
	reader.Skip(2)

	accountID := conn.GetAccountID()
	pin := string(reader.GetRestAsBytes())

	_, err := common.DB.Exec("UPDATE accounts SET pin=? WHERE accountID=?", pin, accountID)
	if err != nil {
		log.Println("handlePinRegistration database pin update issue for accountID:", accountID, err)
	}

	conn.Send(packetRequestPin())

}

func (server *Server) handleGoodLogin(conn mnet.Client, reader mpacket.Reader) {
	server.migrating[conn] = false
	accountID := conn.GetAccountID()

	if server.withPin {
		var pinDB string
		var authDone bool

		err := common.DB.QueryRow("SELECT pin FROM accounts WHERE accountID=?", accountID).
			Scan(&pinDB)

		if err != nil {
			log.Println("handleCheckLogin database retrieval issue for accountID:", accountID, err)
		}

		b1 := reader.ReadByte()
		b2 := reader.ReadByte()

		if b1 == 1 && b2 == 1 { // First attempt, request for pin
			if len(pinDB) == 0 {
				conn.Send(packetRegisterPin())
			} else {
				conn.Send(packetRequestPin())
			}

		} else if b1 == 1 || b1 == 2 { // Client assigned pin
			reader.Skip(6) // space padding?
			pin := string(reader.GetRestAsBytes())

			if pin != pinDB {
				conn.Send(packetRequestPinAfterFailure())

			} else if b1 == 2 { // Changing pin request
				conn.Send(packetRegisterPin())

			} else { // Authenticated successfully
				authDone = true
			}

		} else if b1 == 0 { // Client cancels pin request
			conn.Send(packetCancelPin())
		}

		if !authDone {
			return
		}
	}

	conn.SetLogedIn(true)
	_, err := common.DB.Exec("UPDATE accounts set isLogedIn=1 WHERE accountID=?", accountID)

	if err != nil {
		log.Println("Database error with approving login of accountID", accountID, err)
	} else {
		log.Println("User", accountID, "has logged in from", conn)
	}

	const maxNumberOfWorlds = 14

	for i := len(server.worlds) - 1; i > -1; i-- {
		conn.Send(packetLoginWorldListing(byte(i), server.worlds[i]))
	}

	conn.Send(packetLoginEndWorldList())
}

func (server *Server) handleWorldSelect(conn mnet.Client, reader mpacket.Reader) {
	conn.SetWorldID(reader.ReadByte())
	reader.ReadByte() // ?

	var warning, population byte = 0, 0

	if conn.GetAdminLevel() < 1 { // gms are not restricted in any capacity
		var currentPlayers int16
		var maxPlayers int16

		for _, v := range server.worlds[conn.GetWorldID()].Channels {
			currentPlayers += v.Pop
			maxPlayers += v.MaxPop
		}

		if currentPlayers >= maxPlayers {
			warning = 2
		} else if float64(currentPlayers)/float64(maxPlayers) > 0.90 { // I'm not sure if this warning is even worth it
			warning = 1
		}
	}

	conn.Send(packetLoginWorldInfo(warning, population)) // hard coded for now
}

func (server *Server) handleChannelSelect(conn mnet.Client, reader mpacket.Reader) {
	selectedWorld := reader.ReadByte()   // world
	conn.SetChannelID(reader.ReadByte()) // Channel

	if server.worlds[selectedWorld].Channels[conn.GetChannelID()].MaxPop == 0 {
		conn.Send(packetLoginReturnFromChannel())
		return
	}

	if selectedWorld == conn.GetWorldID() {
		characters := getCharactersFromAccountWorldID(conn.GetAccountID(), conn.GetWorldID())
		conn.Send(packetLoginDisplayCharacters(characters))
	}
}

func (server *Server) handleNameCheck(conn mnet.Client, reader mpacket.Reader) {
	newCharName := reader.ReadString(reader.ReadInt16())

	var nameFound int
	err := common.DB.QueryRow("SELECT count(*) name FROM characters WHERE BINARY name=?", newCharName).
		Scan(&nameFound)

	if err != nil {
		panic(err)
	}

	conn.Send(packetLoginNameCheck(newCharName, nameFound))
}

func (server *Server) handleNewCharacter(conn mnet.Client, reader mpacket.Reader) {
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

	err := common.DB.QueryRow("SELECT count(*) FROM characters where name=? and worldID=?", name, conn.GetWorldID()).Scan(&counter)

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

	newCharacter := player{}

	if conn.GetAdminLevel() > 0 {
		name = "[GM]" + name
	} else if strings.ContainsAny(name, "[]") {
		valid = false // hacked client or packet editting
	}

	if valid {
		res, err := common.DB.Exec("INSERT INTO characters (name, accountID, worldID, face, hair, skin, gender, str, dex, intt, luk) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			name, conn.GetAccountID(), conn.GetWorldID(), face, hair+hairColour, skin, conn.GetGender(), str, dex, intelligence, luk)

		if err != nil {
			log.Println(err)
		}

		characterID, err := res.LastInsertId()

		if err != nil {
			log.Println(err)
		}

		char := loadPlayerFromID(int32(characterID))

		if conn.GetAdminLevel() > 0 {
			items := map[int32]int16{
				1002140: -1,  // Hat
				1032006: -4,  // Earrings
				1042003: -5,  // top
				1062007: -6,  // bottom
				1072004: -7,  // shoes
				1082002: -8,  // Gloves
				1102054: -9,  // Cape
				1092008: -10, // Shield
				1322013: -11, // weapon
			}

			for id, pos := range items {
				item := newAdminItem(id, char.name)

				if err != nil {
					log.Println(err)
					return
				}

				item.slotID = pos
				item.creatorName = name
				item.save(int32(characterID))
				char.equip = append(char.equip, item)
			}

			// TODO: Give GM all skils maxed
		} else {
			items := map[int32]int16{
				top:    -5,
				bottom: -6,
				shoes:  -7,
				weapon: -11,
			}

			for id, pos := range items {
				item := newBeginnerItem(id)

				if err != nil {
					log.Println(err)
					return
				}

				item.slotID = pos
				item.save(int32(characterID))
				char.equip = append(char.equip, item)
			}
		}

		char.save()
		newCharacter = char
	}

	conn.Send(packetLoginCreatedCharacter(valid, newCharacter))
}

func (server *Server) handleDeleteCharacter(conn mnet.Client, reader mpacket.Reader) {
	dob := reader.ReadInt32()
	charID := reader.ReadInt32()

	var storedDob int32
	var charCount int

	err := common.DB.QueryRow("SELECT dob FROM accounts where accountID=?", conn.GetAccountID()).Scan(&storedDob)
	err = common.DB.QueryRow("SELECT count(*) FROM characters where accountID=? AND id=?", conn.GetAccountID(), charID).Scan(&charCount)

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
		records, err := common.DB.Query("DELETE FROM characters where id=?", charID)

		defer records.Close()

		if err != nil {
			log.Println(err)
			return
		}

		deleted = true
	}

	for _, v := range server.worlds {
		v.Conn.Send(internal.PacketLoginDeletedCharacter(charID))
	}

	conn.Send(packetLoginDeleteCharacter(charID, deleted, hacking))
}

func (server *Server) handleSelectCharacter(conn mnet.Client, reader mpacket.Reader) {
	charID := reader.ReadInt32()

	var charCount int

	err := common.DB.QueryRow("SELECT count(*) FROM characters where accountID=? AND id=?", conn.GetAccountID(), charID).Scan(&charCount)

	if err != nil {
		panic(err)
	}

	if charCount == 1 {
		channel := server.worlds[conn.GetWorldID()].Channels[conn.GetChannelID()]
		_, err := common.DB.Exec("UPDATE characters SET migrationID=? WHERE id=?", conn.GetChannelID(), charID)

		if err != nil {
			panic(err)
		}

		server.migrating[conn] = true

		conn.Send(packetLoginMigrateClient(channel.IP, channel.Port, charID))
	}
}

func (server *Server) addCharacterItem(characterID int64, itemID int32, slot int32, creatorName string) {
	_, err := common.DB.Exec("INSERT INTO items (characterID, itemID, slotNumber, creatorName) VALUES (?, ?, ?, ?)", characterID, itemID, slot, creatorName)

	if err != nil {
		panic(err)
	}
}

func (server *Server) handleReturnToLoginScreen(conn mnet.Client, reader mpacket.Reader) {
	conn.Send(packetLoginReturnFromChannel())
}

// HandleServerPacket from world
func (server *Server) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.WorldNew:
		server.handleNewWorld(conn, reader)
	case opcode.WorldInfo:
		server.handleWorldInfo(conn, reader)
	default:
		log.Println("UNKNOWN WORLD PACKET:", reader)
	}
}

// The following logic could do with being cleaned up
func (server *Server) handleNewWorld(conn mnet.Server, reader mpacket.Reader) {
	log.Println("Server register request from", conn)
	if len(server.worlds) > 14 {
		log.Println("Rejected")
		conn.Send(mpacket.CreateInternal(opcode.WorldRequestBad))
	} else {
		name := reader.ReadString(reader.ReadInt16())

		if name == "" {
			name = constant.WORLD_NAMES[len(server.worlds)]

			registered := false
			for i, v := range server.worlds {
				if v.Conn == nil {
					server.worlds[i].Conn = conn
					name = server.worlds[i].Name

					registered = true
					break
				}
			}

			if !registered {
				server.worlds = append(server.worlds, internal.World{Conn: conn, Name: name})
			}

			p := mpacket.CreateInternal(opcode.WorldRequestOk)
			p.WriteString(name)
			conn.Send(p)

			log.Println("Registered", name)
		} else {
			registered := false
			for i, w := range server.worlds {
				if w.Name == name {
					server.worlds[i].Conn = conn
					server.worlds[i].Name = name

					p := mpacket.CreateInternal(opcode.WorldRequestOk)
					p.WriteString(name)
					conn.Send(p)

					registered = true

					break
				}
			}

			if !registered {
				server.worlds = append(server.worlds, internal.World{Conn: conn, Name: name})

				p := mpacket.CreateInternal(opcode.WorldRequestOk)
				p.WriteString(server.worlds[len(server.worlds)-1].Name)
				conn.Send(p)
			}

			log.Println("Re-registered", name)
		}
	}
}

func (server *Server) handleWorldInfo(conn mnet.Server, reader mpacket.Reader) {
	for i, v := range server.worlds {
		if v.Conn != conn {
			continue
		}

		server.worlds[i].SerialisePacket(reader)

		if v.Name == "" {
			log.Println("Registerd new world", server.worlds[i].Name)
		} else {
			log.Println("Updated world info for", v.Name)
		}
	}
}
