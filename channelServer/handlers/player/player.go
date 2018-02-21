package player

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/Hucaru/Valhalla/channelServer/handlers/maps"
	"github.com/Hucaru/Valhalla/channelServer/handlers/message"
	"github.com/Hucaru/Valhalla/channelServer/handlers/playerConn"
	"github.com/Hucaru/Valhalla/channelServer/handlers/world"
	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/nx"
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

// TODO: Add cheat detection - use an audit thread maybe? to not block current socket
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

	nFragaments := reader.ReadByte()

	for i := byte(0); i < nFragaments; i++ {
		movementType := reader.ReadByte()
		switch movementType { // Movement type
		// Absolute movement
		case 0x00: // normal move
			fallthrough
		case 0x05: // normal move
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
		case 0x01: // jump
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
		default:
			log.Println("Unkown movement type received", movementType, reader.GetRestAsBytes())

		}
	}

	reader.GetRestAsBytes() // used in movement validation

	maps.PlayerMove(conn, reader.GetBuffer()[2:])
}

func HandlePlayerUsePortal(reader gopacket.Reader, conn *playerConn.Conn) {
	reader.ReadByte() //?

	entryType := reader.ReadInt32()

	switch entryType {
	case 0:
		if conn.GetCharacter().GetHP() == 0 {
			fmt.Println("Implement death handling through portal")
		}
	case -1:
		nameSize := reader.ReadUint16()
		portalName := reader.ReadString(int(nameSize))

		mapID := conn.GetCharacter().GetCurrentMap()

		if maps.IsValidPortal(mapID, portalName) {
			if !maps.IsPortalOpen(mapID, portalName) {
				conn.Write(message.SendPortalClosed())
				return
			}

			for _, v := range nx.Maps[mapID].Portals {
				if v.Name == portalName {
					ChangeMap(conn, v.Tm, conn.GetChannelID(), maps.GetPortalByName(v.Tm, portalName), conn.GetCharacter().GetHP())
				}
			}

		} else {
			// teleport/warp hacking?
		}

	default:
		log.Println("Unkown portal entry type:", entryType)
	}
}

func HandlePlayerSendAllChat(reader gopacket.Reader, conn *playerConn.Conn) {
	msg := reader.ReadString(int(reader.ReadInt16()))
	ind := strings.Index(msg, "!")

	if ind == 0 && conn.IsAdmin() {
		command := strings.SplitN(msg[ind+1:], " ", -1)
		switch command[0] {
		case "packet":
			packet := string(command[1])
			data, err := hex.DecodeString(packet)

			if err != nil {
				log.Println("Eror in decoding string for gm command packet:", packet)
				break
			}
			log.Println("Sent packet:", hex.EncodeToString(data))
			conn.Write(data)
		case "warp":
			val, err := strconv.Atoi(command[1])

			if err != nil {
				panic(err)
			}

			id := uint32(val)

			if _, ok := nx.Maps[id]; ok {
				ChangeMap(conn, uint32(id), 0, maps.GetRandomSpawnPortal(id), conn.GetCharacter().GetHP())
			} else {
				// check if player id in else if
			}
		default:
			log.Println("Unkown GM command", command)
		}

	} else {
		maps.SendPacketToMap(conn.GetCharacter().GetCurrentMap(), message.SendAllChat(conn.GetCharacter().GetCharID(), conn.IsAdmin(), msg))
	}
}

func SendPlayerPacket(name string, p gopacket.Packet) {
	playerListMutex.Lock()

	if val, ok := playerList[name]; ok {
		val.Write(p)
	}

	playerListMutex.Unlock()
}

func ChangeMap(conn *playerConn.Conn, newMapID uint32, channelID uint32, portal nx.Portal, hp uint16) {
	conn.GetCharacter().SetX(portal.X)
	conn.GetCharacter().SetY(portal.Y)

	conn.Write(changeMap(newMapID, channelID, portal.ID, hp))
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
