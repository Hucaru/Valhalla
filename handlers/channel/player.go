package channel

import (
	"log"

	"github.com/Hucaru/Valhalla/database"
	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/game/packet"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

func playerConnect(conn mnet.MConnChannel, reader mpacket.Reader) {
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

	char := def.GetCharacterFromID(charID)

	var adminLevel int
	err = database.Handle.QueryRow("SELECT adminLevel FROM accounts WHERE accountID=?", conn.GetAccountID()).Scan(&adminLevel)

	if err != nil {
		panic(err)
	}

	conn.SetAdminLevel(adminLevel)

	conn.Send(packet.PlayerEnterGame(char, 0))
	conn.Send(packet.MessageScrollingHeader("dummy header"))

	game.AddPlayer(game.NewPlayer(conn, char))
}

func playerUsePortal(conn mnet.MConnChannel, reader mpacket.Reader) {
	player, err := game.GetPlayerFromConn(conn)

	if err != nil {
		return
	}

	char := player.Char()

	if char.PortalCount != reader.ReadByte() {
		conn.Send(packet.PlayerNoChange())
		return
	}

	entryType := reader.ReadInt32()

	switch entryType {
	case 0:
		if char.HP == 0 {
			returnMapID := nx.Maps[char.MapID].ReturnMap
			portal, id := game.GetRandomSpawnPortal(returnMapID)
			player.ChangeMap(returnMapID, portal, id)
		}
	case -1:
		portalName := reader.ReadString(int(reader.ReadInt16()))

		for _, src := range nx.Maps[char.MapID].Portals {
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

func playerEnterCashShop(conn mnet.MConnChannel, reader mpacket.Reader) {

}

func playerMovement(conn mnet.MConnChannel, reader mpacket.Reader) {
	player, err := game.GetPlayerFromConn(conn)

	if err != nil {
		return
	}

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

	game.SendToMapExcept(char.MapID, packet.PlayerMove(char.ID, moveBytes), conn)
}

func playerTakeDamage(conn mnet.MConnChannel, reader mpacket.Reader) {
	mobAttack := reader.ReadInt8()
	damage := reader.ReadInt32()

	if damage < -1 {
		return
	}

	reducedDamange := damage
	healSkillID := int32(0)

	player, err := game.GetPlayerFromConn(conn)

	if err != nil {
		return
	}

	char := player.Char()

	if char.HP == 0 {
		return
	}

	var mob *def.Mob
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

		mob := game.GetMapFromID(char.MapID).GetMobFromID(spawnID)
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

		game.SendToMap(char.MapID, packet.PlayerReceivedDmg(char.ID, mobAttack, damage,
			reducedDamange, spawnID, mobID, healSkillID, stance, reflectAction, reflected, reflectX, reflectY))
	}

	if mobSkillID != 0 && mobSkillLevel != 0 {
		// new skill
	} else if mob != nil {
		// residual skill
	}
}

func playerRequestAvatarInfoWindow(conn mnet.MConnChannel, reader mpacket.Reader) {
	player, err := game.GetPlayerFromID(reader.ReadInt32())

	if err != nil {
		return
	}

	char := player.Char()

	conn.Send(packet.PlayerAvatarSummaryWindow(char.ID, char, char.Guild))
}

func playerEmote(conn mnet.MConnChannel, reader mpacket.Reader) {
	emote := reader.ReadInt32()

	player, err := game.GetPlayerFromConn(conn)

	if err != nil {
		return
	}

	char := player.Char()

	mapID := char.MapID

	game.SendToMapExcept(mapID, packet.PlayerEmoticon(char.ID, emote), conn)
}

func playerPassiveRegen(conn mnet.MConnChannel, reader mpacket.Reader) {
	reader.ReadBytes(4) //?

	hp := reader.ReadInt16()
	mp := reader.ReadInt16()

	// validate the values

	player, err := game.GetPlayerFromConn(conn)

	if err != nil {
		return
	}

	char := player.Char()

	if char.HP == 0 || hp > 400 || mp > 1000 || (hp > 0 && mp > 0) {
		return
	}

	if hp > 0 {
		player.GiveHP(int32(hp))
	} else if mp > 0 {
		player.GiveMP(int32(mp))
	}
}
