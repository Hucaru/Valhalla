package player

import (
	"github.com/Hucaru/Valhalla/channelServer/playerConn"
	"github.com/Hucaru/Valhalla/channelServer/server"
	"github.com/Hucaru/Valhalla/channelServer/world"
	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/gopacket"
)

func HandlePlayerEnterGame(reader gopacket.Reader, conn *playerConn.Conn) {
	charID := reader.ReadUint32() // validate this and net address from the migration packet

	_, channelID := world.GetAssignedIDs()

	char := character.GetCharacter(charID)

	char.SetEquips(character.GetCharacterEquips(char.GetCharID()))
	char.SetSkills(character.GetCharacterSkills(char.GetCharID()))
	char.SetItems(character.GetCharacterItems(char.GetCharID()))

	_, px, py := server.GetRandomSpawnPortal(char.GetCurrentMap())
	char.SetX(px)
	char.SetY(py)
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
		server.PlayerLeaveMap(conn, char.GetCurrentMap())
	})

	conn.Write(spawnGame(char, uint32(channelID)))
	server.PlayerEnterMap(conn, char.GetCurrentMap())
}

// func HandlePlayerUsePortal(reader gopacket.Reader, conn *playerConn.Conn) {
// 	reader.ReadByte() //?

// 	entryType := reader.ReadInt32()

// 	switch entryType {
// 	case 0:
// 		if conn.GetCharacter().GetHP() == 0 {
// 			currentMap := conn.GetCharacter().GetCurrentMap()
// 			returnMap := nx.Maps[currentMap].ReturnMap
// 			portal := maps.GetRandomSpawnPortal(returnMap)

// 			conn.GetCharacter().SetX(portal.X)
// 			conn.GetCharacter().SetY(portal.Y)

// 			PlayerSetHP(conn, 50)

// 			maps.PlayerChangeMap(conn, currentMap, portal.ID, conn.GetCharacter().GetHP())
// 		}
// 	case -1:
// 		nameSize := reader.ReadUint16()
// 		portalName := reader.ReadString(int(nameSize))

// 		mapID := conn.GetCharacter().GetCurrentMap()

// 		if maps.IsValidPortal(mapID, portalName) {
// 			if !maps.IsPortalOpen(mapID, portalName) {
// 				conn.Write(message.SendPortalClosed())
// 				return
// 			}

// 			for _, v := range nx.Maps[mapID].Portals {
// 				if v.Name == portalName {
// 					portal := maps.GetPortalByName(v.Tm, v.Tn)

// 					conn.GetCharacter().SetX(portal.X)
// 					conn.GetCharacter().SetY(portal.Y)

// 					maps.PlayerChangeMap(conn, v.Tm, portal.ID, conn.GetCharacter().GetHP())
// 				}
// 			}

// 		} else {
// 			// teleport/warp hacking?
// 		}

// 	default:
// 		log.Println("Unkown portal entry type:", entryType)
// 	}
// }

// func HandlePlayerSendAllChat(reader gopacket.Reader, conn *playerConn.Conn) {
// 	msg := reader.ReadString(int(reader.ReadInt16()))
// 	ind := strings.Index(msg, "!")

// 	if ind == 0 && conn.IsAdmin() {
// 		command := strings.SplitN(msg[ind+1:], " ", -1)
// 		dealWithCommand(conn, command)

// 	} else {
// 		server.SendPacketToMap(conn.GetCharacter().GetCurrentMap(), message.SendAllChat(conn.GetCharacter().GetCharID(), conn.IsAdmin(), msg), nil)
// 	}
// }

// func HandlePlayerTakeDmg(reader gopacket.Reader, conn *playerConn.Conn) {
// 	fmt.Println(reader)

// 	dmgType := reader.ReadByte() // multiple types, need a switch statement
// 	ammount := reader.ReadUint32()

// 	mobID := uint32(0)

// 	reader.ReadUint32()

// 	hit := reader.ReadByte()
// 	stance := reader.ReadByte()

// 	if dmgType != 0xFE {
// 		mobID = reader.ReadUint32()
// 	}

// 	// Handle character buffs e.g. magic guard

// 	// Modify character hp after buffs taken into account
// 	charID := conn.GetCharacter().GetCharID()
// 	server.SendPacketToMap(conn.GetCharacter().GetCurrentMap(), playerReceivedDmg(charID, ammount, dmgType, mobID, hit, stance), nil)
// }

// func HandlePlayerEmotion(reader gopacket.Reader, conn *playerConn.Conn) {
// 	emotion := reader.ReadUint32()
// 	server.SendPacketToMap(conn.GetCharacter().GetCurrentMap(), playerEmotion(conn.GetCharacter().GetCharID(), emotion), nil)
// }

// func HandlePlayerSkillUpdate(reader gopacket.Reader, conn *playerConn.Conn) {
// 	char := conn.GetCharacter()

// 	skillID := reader.ReadUint32()

// 	newSP := char.GetSP() - 1
// 	char.SetSP(newSP)

// 	conn.Write(statChangeUint16(true, spID, newSP))

// 	// Client will warp player away and await duplicate packet for confirmation?
// 	conn.Write(playerSkillUpdate(skillID, 1))
// 	conn.Write(playerSkillUpdate(skillID, 1))
// }
