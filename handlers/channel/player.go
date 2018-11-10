package channel

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
	player := game.GetPlayerFromConn(conn)

	if player.Char().PortalCount != reader.ReadByte() {
		conn.Send(packets.PlayerNoChange())
		return
	}

	entryType := reader.ReadInt32()

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
	player := game.GetPlayerFromConn(conn)
	char := player.Char()

	if char.PortalCount != reader.ReadByte() {
		return
	}

	moveData, finalData := parseMovement(reader)

	if !validateCharMovement(char, moveData) {
		return
	}

	moveBytes := generateMovementBytes(moveData)

	player.UpdateMovement(finalData)

	game.SendToMapExcept(char.CurrentMap, packets.PlayerMove(char.ID, moveBytes), conn)
}

func playerTakeDamage(conn mnet.MConnChannel, reader maplepacket.Reader) {
	mobAttack := reader.ReadInt8()
	damage := reader.ReadInt32()

	if damage < -1 {
		return
	}

	reducedDamange := damage
	healSkillID := int32(0)

	player := game.GetPlayerFromConn(conn)
	char := player.Char()

	if char.HP == 0 {
		return
	}

	var mob *types.Mob
	var mobSkillID, mobSkillLevel byte = 0, 0

	if mobAttack < -1 {
		mobSkillLevel = reader.ReadByte()
		mobSkillID = reader.ReadByte()
	} else {
		magicElement := int32(0)

		if reader.ReadBool() {
			magicElement = reader.ReadInt32()
			_ = magicElement
			// 0 = no element (Grendel the Really Old, 9001001)
			// 1 = Ice (Celion? blue, 5120003)
			// 2 = Lightning (Regular big Sentinel, 3000000)
			// 3 = Fire (Fire sentinel, 5200002)
		}

		spawnID := reader.ReadInt32()
		mobID := reader.ReadInt32()

		mob := game.GetMapFromID(char.CurrentMap).GetMobFromID(spawnID)
		// mob = game.GetMobFromMapAndSpawnID(char.CurrentMap, spawnID)

		if mob == nil || mob.ID != mobID {
			return
		}

		stance := reader.ReadByte()
		reflected := reader.ReadByte()

		reflectAction := byte(0)
		var reflectX, reflectY int16 = 0, 0

		if reflected > 0 {
			reflectAction = reader.ReadByte()
			reflectX, reflectY = reader.ReadInt16(), reader.ReadInt16()
		}

		playerDamange := -damage

		// Magic guard dmg absorption

		// Fighter / Page power guard

		// Meso guard

		player.GiveHP(playerDamange)

		game.SendToMap(char.CurrentMap, packets.PlayerReceivedDmg(char.ID, mobAttack, damage,
			reducedDamange, spawnID, mobID, healSkillID, stance, reflectAction, reflected, reflectX, reflectY))
	}

	if mobSkillID != 0 && mobSkillLevel != 0 {
		// new skill
	} else if mob != nil {
		// residual skill
	}
}

func playerRequestAvatarInfoWindow(conn mnet.MConnChannel, reader maplepacket.Reader) {
}
