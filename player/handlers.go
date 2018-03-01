package player

import (
	"fmt"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/data"
	"github.com/Hucaru/gopacket"
)

func HandleConnect(conn clientConn, reader gopacket.Reader) {
	fmt.Println(reader)
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

	data.AddOnlineCharacter(conn, &char)

	conn.SetCloseCallback(func() {
		data.RemoveOnlineCharacter(conn)
		// maps remove client connection
		// party remove client connection
	})

	conn.Write(enterGame(char, channelID))
}
