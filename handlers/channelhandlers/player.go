package channelhandlers

import (
	"log"

	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/packets"
	"github.com/Hucaru/Valhalla/types"
)

func playerConnect(conn mnet.MConnChannel, reader maplepacket.Reader) {
	charID := reader.ReadInt32()

	// check that the account this id is associated with has channel id of -1
	// check that the world this characters belongs to is the same as the world this channel is part of
	conn.SetLogedIn(true)

	char := types.GetCharacterFromID(charID)

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

func playerSendAllChat(conn mnet.MConnChannel, reader maplepacket.Reader) {

}

func playerRequestAvatarInfoWindow(conn mnet.MConnChannel, reader maplepacket.Reader) {
}
