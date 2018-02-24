package player

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Hucaru/Valhalla/channelServer/maps"
	"github.com/Hucaru/Valhalla/channelServer/message"
	"github.com/Hucaru/Valhalla/channelServer/playerConn"
	"github.com/Hucaru/Valhalla/channelServer/server"
	"github.com/Hucaru/Valhalla/channelServer/world"
	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/nx"
	"github.com/Hucaru/gopacket"
)

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
		server.RemovePlayerFromList(conn)
	})

	server.AddPlayerToList(conn)

	conn.Write(spawnGame(char, uint32(channelID)))

	maps.RegisterNewPlayer(conn, char.GetCurrentMap())
}

func HandlePlayerUsePortal(reader gopacket.Reader, conn *playerConn.Conn) {
	reader.ReadByte() //?

	entryType := reader.ReadInt32()

	switch entryType {
	case 0:
		if conn.GetCharacter().GetHP() == 0 {
			currentMap := conn.GetCharacter().GetCurrentMap()
			returnMap := nx.Maps[currentMap].ReturnMap
			portal := maps.GetRandomSpawnPortal(returnMap)

			conn.GetCharacter().SetX(portal.X)
			conn.GetCharacter().SetY(portal.Y)

			PlayerSetHP(conn, 50)

			maps.PlayerChangeMap(conn, currentMap, portal.ID, conn.GetCharacter().GetHP())
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
					portal := maps.GetPortalByName(v.Tm, v.Tn)

					fmt.Println("using pn:", portalName, "on map:", mapID, "going to map:", v.Tm, "portal name:", portal.Name, "id:", portal.ID)

					conn.GetCharacter().SetX(portal.X)
					conn.GetCharacter().SetY(portal.Y)

					maps.PlayerChangeMap(conn, v.Tm, portal.ID, conn.GetCharacter().GetHP())
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
				return
			}

			id := uint32(val)

			if _, exist := nx.Maps[id]; exist {
				portal := maps.GetRandomSpawnPortal(id)

				if len(command) > 2 {
					pos, err := strconv.Atoi(command[2])

					if err == nil {
						portal = maps.GetPortalByID(uint32(id), byte(pos))
					}
				}

				maps.PlayerChangeMap(conn, uint32(id), portal.ID, conn.GetCharacter().GetHP())
			} else {
				// check if player id in else if
			}
		case "job":
			val, err := strconv.Atoi(command[1])

			if err != nil {
				return
			}

			PlayerChangeJob(conn, uint16(val))
		case "level":
			val, err := strconv.Atoi(command[1])

			if err != nil {
				return
			}

			PlayerSetLevel(conn, byte(val))
		case "exp":
			val, err := strconv.Atoi(command[1])

			if err != nil {
				return
			}

			PlayerAddExp(conn, uint32(val))
		case "hp":
			val, err := strconv.Atoi(command[1])

			if err != nil {
				return
			}

			PlayerSetHP(conn, uint16(val))
		case "mp":
			val, err := strconv.Atoi(command[1])

			if err != nil {
				return
			}

			PlayerSetMP(conn, uint16(val))
		default:
			log.Println("Unkown GM command", command)
		}

	} else {
		server.SendPacketToMap(conn.GetCharacter().GetCurrentMap(), message.SendAllChat(conn.GetCharacter().GetCharID(), conn.IsAdmin(), msg))
	}
}

func HandlePlayerEmotion(reader gopacket.Reader, conn *playerConn.Conn) {
	emotion := reader.ReadUint32()
	server.SendPacketToMap(conn.GetCharacter().GetCurrentMap(), playerEmotion(conn.GetCharacter().GetCharID(), emotion))
}

func HandlePlayerSkillUpdate(reader gopacket.Reader, conn *playerConn.Conn) {
	char := conn.GetCharacter()

	skillID := reader.ReadUint32()

	newSP := char.GetSP() - 1
	char.SetSP(newSP)

	conn.Write(statChangeUint16(true, spID, newSP))

	// Client will warp player away and await duplicate packet for confirmation?
	conn.Write(playerSkillUpdate(skillID, 1))
	conn.Write(playerSkillUpdate(skillID, 1))
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
