package player

import (
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
	char.SetStance(0) // Not sure how to populate this

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

func HandlePlayerDisconnect() {

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
