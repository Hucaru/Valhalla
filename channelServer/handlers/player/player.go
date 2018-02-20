package player

import (
	"fmt"
	"log"
	"sync"

	"github.com/Hucaru/Valhalla/channelServer/handlers/maps"
	"github.com/Hucaru/Valhalla/channelServer/handlers/playerConn"
	"github.com/Hucaru/Valhalla/channelServer/handlers/world"
	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

var playerList = make(map[string]*playerConn.Conn)
var playerListMutex = &sync.Mutex{}

func HandlePlayerEnterGame(reader gopacket.Reader, conn *playerConn.Conn) {
	charID := reader.ReadUint32() // validate this and net address from the migration packet

	if !validateNewConnection(charID) {
		conn.Close()
	}

	_, channelID := world.GetAssignedIDs()

	char := character.GetCharacter(charID)

	char.SetEquips(character.GetCharacterEquips(char.GetCharID()))
	char.SetSkills(character.GetCharacterSkills(char.GetCharID()))
	char.SetItems(character.GetCharacterItems(char.GetCharID()))

	portal := maps.GetSpawnPortal(char.GetCurrentMap(), char.GetCurrentMapPos())
	char.SetX(portal.X)
	char.SetY(portal.Y)
	char.SetState(0) // Not sure how to populate this

	conn.SetCharacter(char)
	conn.SetChanneldID(uint32(channelID))

	var isAdmin bool

	err := connection.Db.QueryRow("SELECT isAdmin from users where userID=?", char.GetUserID()).Scan(&isAdmin)

	if err != nil {
		panic(err)
	}

	conn.SetIsLogedIn(true)
	conn.SetAdmin(isAdmin)

	conn.SetCloseCallback(func() {
		maps.PlayerLeftGame(conn)
	})

	conn.Write(spawnGame(char, uint32(channelID)))

	maps.RegisterNewPlayer(conn, char.GetCurrentMap())
}

func HandlePlayerMovement(reader gopacket.Reader, conn *playerConn.Conn) {
	// http://mapleref.wikia.com/wiki/Movement
	/*
		State enum:
			left / right: Action
			3 / 2: Walk
			5 / 4: Standing
			7 / 6: Jumping & Falling
			9 / 8: Normal attack
			11 / 10: Prone
			13 / 12: Rope
			15 / 14: Ladder
	*/
	reader.ReadBytes(5) // used in movement validation
	char := conn.GetCharacter()

	// Used to validate movement:
	for len(reader.GetRestAsBytes()) > 18 {
		movementType := reader.ReadByte()
		switch movementType { // Movement type
		// Absolute movement
		case 0x00:
			fallthrough
		case 0x05:
			fallthrough
		case 0x17:
			posX := reader.ReadInt16()
			posY := reader.ReadInt16()
			velX := reader.ReadInt16()
			velY := reader.ReadInt16()

			reader.ReadUint16()

			state := reader.ReadByte()
			duration := reader.ReadUint16()

			char.SetX(posX + velX*int16(duration))
			char.SetY(posY + velY*int16(duration))
			char.SetState(state)

		// Relative movement
		case 0x01:
			fallthrough
		case 0x02:
			fallthrough
		case 0x06:
			fallthrough
		case 0x12:
			fallthrough
		case 0x13:
			fallthrough
		case 0x16:
			reader.ReadInt16() // velX
			reader.ReadInt16() // velY

			state := reader.ReadByte()
			reader.ReadUint16() // duration

			char.SetState(state)

		// Instant movement
		case 0x03:
			fallthrough
		case 0x04:
			fallthrough
		case 0x07:
			fallthrough
		case 0x08:
			fallthrough
		case 0x09:
			fallthrough
		case 0x014:
			posX := reader.ReadInt16()
			posY := reader.ReadInt16()
			reader.ReadInt16() // velX
			reader.ReadInt16() // velY

			state := reader.ReadByte()

			char.SetX(posX)
			char.SetY(posY)
			char.SetState(state)

		// Equip movement
		case 0x10:
			reader.ReadByte() // ?

		// Jump down movement
		case 0x11:
			posX := reader.ReadInt16()
			posY := reader.ReadInt16()
			velX := reader.ReadInt16()
			velY := reader.ReadInt16()

			reader.ReadUint16()

			foothold := reader.ReadUint16()
			duration := reader.ReadUint16()

			char.SetX(posX + velX*int16(duration))
			char.SetY(posY + velY*int16(duration))
			char.SetFh(foothold)
			fmt.Println("foothold found:", foothold)
		default:
			log.Println("Unkown movement type received", movementType, reader.GetRestAsBytes())

		}
	}

	reader.ReadBytes(18) // used in movement validation

	maps.PlayerMove(conn, reader.GetBuffer()[2:])
}

func SendPlayerPacket(name string, p gopacket.Packet) {
	playerListMutex.Lock()

	if val, ok := playerList[name]; ok {
		val.Write(p)
	}

	playerListMutex.Unlock()
}

func ChangeMap(conn *playerConn.Conn, newMapID uint32, channelID uint32, mapPos byte, hp uint16) {
	portal := maps.GetSpawnPortal(newMapID, mapPos)
	conn.GetCharacter().SetX(portal.X)
	conn.GetCharacter().SetY(portal.Y)

	conn.Write(changeMap(newMapID, channelID, mapPos, hp))
	maps.PlayerChangeMap(conn, newMapID)
}

func validateNewConnection(charID uint32) bool {
	var migratingWorldID, migratingChannelID int8
	err := connection.Db.QueryRow("SELECT isMigratingWorld,isMigratingChannel FROM characters where id=?", charID).Scan(&migratingWorldID, &migratingChannelID)

	if err != nil {
		panic(err.Error())
	}

	if migratingWorldID < 0 || migratingChannelID < 0 {

		return false
	}

	msg := make(chan gopacket.Packet)
	world.InterServer <- connection.NewMessage([]byte{constants.CHANNEL_GET_INTERNAL_IDS}, msg)
	result := <-msg
	r := gopacket.NewReader(&result)

	if r.ReadByte() != byte(migratingWorldID) && r.ReadByte() != byte(migratingChannelID) {
		log.Println("Received invalid migration info for character", charID, "remote hacking")
		records, err := connection.Db.Query("UPDATE characters set migratingWorldID=?, migratingChannelID=? WHERE id=?", -1, -1, charID)

		defer records.Close()

		if err != nil {
			panic(err.Error())
		}

		return false
	}

	return true
}
