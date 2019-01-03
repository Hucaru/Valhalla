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
		log.Println(err)
	}

	// check migration, channel status

	conn.SetAccountID(accountID)

	// check that the world this characters belongs to is the same as the world this channel is part of
	conn.SetLogedIn(true)

	char := def.GetCharacterFromID(charID)

	var adminLevel int
	err = database.Handle.QueryRow("SELECT adminLevel FROM accounts WHERE accountID=?", conn.GetAccountID()).Scan(&adminLevel)

	if err != nil {
		log.Println(err)
	}

	conn.SetAdminLevel(adminLevel)

	conn.Send(packet.PlayerEnterGame(char, 0))
	conn.Send(packet.MessageScrollingHeader("Valhalla Archival Project"))

	game.Players[conn] = game.NewPlayer(conn, char)
	err = game.Maps[char.MapID].AddPlayer(conn, 0)

	if err != nil {
		log.Println(err)
	}
}

func playerUsePortal(conn mnet.MConnChannel, reader mpacket.Reader) {
	player, ok := game.Players[conn]

	if !ok {
		return
	}

	char := player.Char()

	if char.PortalCount != reader.ReadByte() {
		conn.Send(packet.PlayerNoChange())
		return
	}

	entryType := reader.ReadInt32()

	currentMap, err := nx.GetMap(char.MapID)

	if err != nil {
		return
	}

	switch entryType {
	case 0:
		if char.HP == 0 {
			returnMapID := currentMap.ReturnMap
			portal, id, _ := game.Maps[returnMapID].GetRandomSpawnPortal()
			player.ChangeMap(returnMapID, portal, id)
			player.GiveHP(50)
		}
	case -1:
		portalName := reader.ReadString(int(reader.ReadInt16()))

		for _, src := range currentMap.Portals {
			if src.Pn == portalName {
				destMap, err := nx.GetMap(src.Tm)

				if err != nil {
					return
				}

				for i, dest := range destMap.Portals {
					if dest.Pn == src.Tn {
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
	player, ok := game.Players[conn]

	if !ok {
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

	game.Maps[char.MapID].SendExcept(packet.PlayerMove(char.ID, moveBytes), conn, player.InstanceID)
}

func playerTakeDamage(conn mnet.MConnChannel, reader mpacket.Reader) {
	mobAttack := reader.ReadInt8()
	damage := reader.ReadInt32()

	if damage < -1 {
		return
	}

	reducedDamange := damage
	healSkillID := int32(0)

	player, ok := game.Players[conn]

	if !ok {
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

		mob, err := game.Maps[char.MapID].GetMobFromSpawnID(spawnID, player.InstanceID)

		if err != nil {
			return
		}

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

		game.Maps[char.MapID].Send(packet.PlayerReceivedDmg(char.ID, mobAttack, damage, reducedDamange, spawnID, mobID,
			healSkillID, stance, reflectAction, reflected, reflectX, reflectY), player.InstanceID)
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

	player, ok := game.Players[conn]

	if !ok {
		return
	}

	char := player.Char()

	game.Maps[char.MapID].SendExcept(packet.PlayerEmoticon(char.ID, emote), conn, player.InstanceID)
}

func playerPassiveRegen(conn mnet.MConnChannel, reader mpacket.Reader) {
	reader.ReadBytes(4) //?

	hp := reader.ReadInt16()
	mp := reader.ReadInt16()

	// validate the values

	player, ok := game.Players[conn]

	if !ok {
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

func playerAddStatPoint(conn mnet.MConnChannel, reader mpacket.Reader) {
	// statID := reader.ReadInt32()
}

func playerAddSkillPoint(conn mnet.MConnChannel, reader mpacket.Reader) {
	// skillID := reader.ReadInt32()
}

func playerGiveFame(conn mnet.MConnChannel, reader mpacket.Reader) {

}

func playerMoveInventoryItem(conn mnet.MConnChannel, reader mpacket.Reader) {
	// invTabID := reader.ReadByte()
	// origPos := reader.ReadInt16()
	// newPos := reader.ReadInt16()

	// amount := reader.ReadInt16() // amount?
}
