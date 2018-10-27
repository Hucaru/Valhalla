package channelhandlers

import (
	"log"

	"github.com/Hucaru/Valhalla/database"
	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/packets"
	"github.com/Hucaru/Valhalla/types"
)

func playerConnect(conn mnet.MConnChannel, reader maplepacket.Reader) {
	charID := reader.ReadInt32()

	var accountID int32
	err := database.Handle.QueryRow("SELECT accountID FROM characters WHERE id=?", charID).Scan(&accountID)

	if err != nil {
		panic(err)
	}

	// check migration, channel status

	conn.SetAccountID(accountID)

	// check that the world this characters belongs to is the same as the world this channel is part of
	conn.SetLogedIn(true)

	char := types.GetCharacterFromID(charID)

	var adminLevel int
	err = database.Handle.QueryRow("SELECT adminLevel FROM accounts WHERE accountID=?", conn.GetAccountID()).Scan(&adminLevel)

	if err != nil {
		panic(err)
	}

	conn.SetAdminLevel(adminLevel)

	conn.Send(packets.PlayerEnterGame(char, 0))
	conn.Send(packets.MessageScrollingHeader("dummy header"))

	game.AddPlayer(game.NewPlayer(conn, char))
}

func playerUsePortal(conn mnet.MConnChannel, reader maplepacket.Reader) {
	reader.ReadByte()
	entryType := reader.ReadInt32()

	player := game.GetPlayerFromConn(conn)

	switch entryType {
	case 0:
		if player.Char().HP == 0 {
			returnMapID := nx.Maps[player.Char().CurrentMap].ReturnMap
			portal, id := game.GetRandomSpawnPortal(returnMapID)
			player.ChangeMap(returnMapID, portal, id)
		}
	case -1:
		portalName := reader.ReadString(int(reader.ReadInt16()))

		for _, src := range nx.Maps[player.Char().CurrentMap].Portals {
			if src.Name == portalName {
				for i, dest := range nx.Maps[src.Tm].Portals {
					if dest.Name == src.Tn {
						player.ChangeMap(src.Tm, dest, byte(i))
					}
				}
			}
		}

	default:
		log.Println("Unknown portal entry type, packet:", reader)
	}
}

func playerEnterCashShop(conn mnet.MConnChannel, reader maplepacket.Reader) {

}

func playerMovement(conn mnet.MConnChannel, reader maplepacket.Reader) {

}

func playerStandardSkill(conn mnet.MConnChannel, reader maplepacket.Reader) {

}

func playerRangedSkill(conn mnet.MConnChannel, reader maplepacket.Reader) {

}

func playerMagicSkill(conn mnet.MConnChannel, reader maplepacket.Reader) {

}

func playerTakeDamage(conn mnet.MConnChannel, reader maplepacket.Reader) {

}

func playerRequestAvatarInfoWindow(conn mnet.MConnChannel, reader maplepacket.Reader) {
}
