package player

import (
	"log"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/Valhalla/movement"
	"github.com/Hucaru/gopacket"
)

func HandleConnect(conn interfaces.ClientConn, reader gopacket.Reader) uint32 {
	charID := reader.ReadUint32()

	char := character.GetCharacter(charID)
	char.SetEquips(character.GetCharacterEquips(char.GetCharID()))
	char.SetSkills(character.GetCharacterSkills(char.GetCharID()))
	char.SetItems(character.GetCharacterItems(char.GetCharID()))

	var isAdmin bool

	err := connection.Db.QueryRow("SELECT isAdmin from users where userID=?", char.GetUserID()).Scan(&isAdmin)

	if err != nil {
		panic(err)
	}

	channelID := uint32(0) // Either get from world server or have it be part of config file

	conn.SetAdmin(isAdmin)
	conn.SetIsLogedIn(true)
	conn.SetChanID(channelID)

	charsPtr.AddOnlineCharacter(conn, &char)

	conn.Write(enterGame(char, channelID))

	log.Println(char.GetName(), "has loged in from", conn)

	return char.GetCurrentMap()
}

func HandleMovement(conn interfaces.ClientConn, reader gopacket.Reader) (uint32, gopacket.Packet) {
	reader.ReadBytes(5) // used in movement validation
	char := charsPtr.GetOnlineCharacterHandle(conn)

	nFrags := reader.ReadByte()

	movement.ParseFragments(nFrags, char, reader)

	return char.GetCurrentMap(), playerMovePacket(char.GetCharID(), reader.GetBuffer()[2:])
}
