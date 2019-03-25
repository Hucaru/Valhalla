package channel

import (
	"fmt"
	"log"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/database"
	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

func playerConnect(conn mnet.Client, reader mpacket.Reader) {
	charID := reader.ReadInt32()

	var accountID int32
	err := database.Handle.QueryRow("SELECT accountID FROM characters WHERE id=?", charID).Scan(&accountID)

	if err != nil {
		log.Println(err)
	}

	// check migration, channel status

	conn.SetAccountID(accountID)

	// check that the world this characters belongs to is the same as the world this channel is part of
	conn.SetLogedIn(true) // this seems redundant

	char := game.GetCharacterFromID(charID)

	var adminLevel int
	err = database.Handle.QueryRow("SELECT adminLevel FROM accounts WHERE accountID=?", conn.GetAccountID()).Scan(&adminLevel)

	if err != nil {
		log.Println(err)
	}

	conn.SetAdminLevel(adminLevel)

	conn.Send(game.PacketPlayerEnterGame(char, 0))
	conn.Send(game.PacketMessageScrollingHeader("Valhalla Archival Project"))

	game.Players[conn] = game.NewPlayer(conn, char)
	err = game.Maps[char.MapID].AddPlayer(conn, 0)

	if err != nil {
		log.Println(err)
	}
}

func playerUsePortal(conn mnet.Client, reader mpacket.Reader) {
	player, ok := game.Players[conn]

	if !ok {
		return
	}

	char := player.Char()

	if char.PortalCount != reader.ReadByte() {
		conn.Send(game.PacketPlayerNoChange())
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

func playerEnterCashShop(conn mnet.Client, reader mpacket.Reader) {

}

func playerMovement(conn mnet.Client, reader mpacket.Reader) {
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

	game.Maps[char.MapID].SendExcept(game.PacketPlayerMove(char.ID, moveBytes), conn, player.InstanceID)
}

func playerTakeDamage(conn mnet.Client, reader mpacket.Reader) {
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

	var mob *game.Mob
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

		game.Maps[char.MapID].Send(game.PacketPlayerReceivedDmg(char.ID, mobAttack, damage, reducedDamange, spawnID, mobID,
			healSkillID, stance, reflectAction, reflected, reflectX, reflectY), player.InstanceID)

		if player.Char().HP == 0 && mob.Controller == player.Client {
			mob.ResetAggro()
		}
	}

	if mobSkillID != 0 && mobSkillLevel != 0 {
		// new skill
	} else if mob != nil {
		// residual skill
	}
}

func playerRequestAvatarInfoWindow(conn mnet.Client, reader mpacket.Reader) {
	player, err := game.Players.GetFromID(reader.ReadInt32())

	if err != nil {
		return
	}

	char := player.Char()

	conn.Send(game.PacketPlayerAvatarSummaryWindow(char.ID, char, char.Guild))
}

func playerEmote(conn mnet.Client, reader mpacket.Reader) {
	emote := reader.ReadInt32()

	player, ok := game.Players[conn]

	if !ok {
		return
	}

	char := player.Char()

	game.Maps[char.MapID].SendExcept(game.PacketPlayerEmoticon(char.ID, emote), conn, player.InstanceID)
}

func playerPassiveRegen(conn mnet.Client, reader mpacket.Reader) {
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

func playerAddStatPoint(conn mnet.Client, reader mpacket.Reader) {
	player, ok := game.Players[conn]

	if !ok {
		return
	}

	if player.Char().AP > 0 {
		player.GiveAP(-1)
	}

	statID := reader.ReadInt32()

	switch statID {
	case constant.StrID:
		player.GiveStr(1)
	case constant.DexID:
		player.GiveDex(1)
	case constant.IntID:
		player.GiveInt(1)
	case constant.LukID:
		player.GiveLuk(1)
	default:
		fmt.Println("unknown stat id:", statID)
	}
}

func playerAddSkillPoint(conn mnet.Client, reader mpacket.Reader) {
	player, ok := game.Players[conn]

	if !ok {
		return
	}

	skillID := reader.ReadInt32()

	nxSkills, err := nx.GetPlayerSkill(skillID)

	if err != nil {
		fmt.Println("error")
		return
	}

	char := player.Char()

	skill, ok := char.Skills[skillID]

	if ok {
		// check that increasing skill level won't go above max
		if int(skill.Level) >= len(nxSkills) {
			return
		}

		player.UpdateSkill(game.CreateSkillFromData(skillID, skill.Level+1, nxSkills[skill.Level]))
	} else {
		// check if class can have skill
		baseSkillID := skillID / 10000

		if !validateSkillWithJob(char.Job, baseSkillID) {
			conn.Send(game.PacketPlayerNoChange())
			fmt.Println("Unknown skill learn:", char.Job, baseSkillID)
			return
		}

		// give new skill
		player.UpdateSkill(game.CreateSkillFromData(skillID, 1, nxSkills[0]))
	}

	if char.SP > 0 {
		player.GiveSP(-1)
	}

}

func playerGiveFame(conn mnet.Client, reader mpacket.Reader) {

}

func playerMoveInventoryItem(conn mnet.Client, reader mpacket.Reader) {
	invTabID := reader.ReadByte()
	origPos := reader.ReadInt16()
	newPos := reader.ReadInt16()

	// amount := reader.ReadInt16() // amount?

	if invTabID > 5 || origPos == 0 {
		conn.Send(game.PacketPlayerNoChange()) // bad packet, hacker?
	}

	player, ok := game.Players[conn]

	if !ok {
		return
	}

	var items []game.Item

	switch invTabID {
	case 1:
		items = player.Char().Inventory.Equip
	case 2:
		items = player.Char().Inventory.Use
	case 3:
		items = player.Char().Inventory.SetUp
	case 4:
		items = player.Char().Inventory.Etc
	case 5:
		items = player.Char().Inventory.Cash
	}

	if newPos == 0 { // drop

	} else { // move
		var foundItems []game.Item

		for _, item := range items {
			if item.SlotID == origPos {
				if len(foundItems) == 0 {
					foundItems = append(foundItems, item)
				} else {
					foundItems[0] = item
				}
			} else if item.SlotID == newPos {
				if len(foundItems) == 0 {
					foundItems = make([]game.Item, 2)
					foundItems[1] = item
				} else {
					foundItems = append(foundItems, item)
				}
			}
		}

		if len(foundItems) == 1 && foundItems[0].SlotID == origPos {
			player.MoveItem(foundItems[0], newPos)
		} else if len(foundItems) == 2 {
			player.SwapItems(foundItems[0], foundItems[1])
		}
	}
}

func playerUseChair(conn mnet.Client, reader mpacket.Reader) {
	// chairID := reader.ReadInt32()
}

func playerStand(conn mnet.Client, reader mpacket.Reader) {
	if reader.ReadInt16() == -1 {

	} else {
		fmt.Println("player stand", reader)
	}
}

func validateSkillWithJob(jobID int16, baseSkillID int32) bool {
	switch jobID {
	case constant.WarriorJobID:
		if baseSkillID != constant.WarriorJobID {
			return false
		}
	case constant.FighterJobID:
		if baseSkillID != constant.WarriorJobID && baseSkillID != constant.FighterJobID {
			return false
		}
	case constant.CrusaderJobID:
		if baseSkillID != constant.WarriorJobID && baseSkillID != constant.FighterJobID && baseSkillID != constant.CrusaderJobID {
			return false
		}
	case constant.PageJobID:
		if baseSkillID != constant.WarriorJobID && baseSkillID != constant.PageJobID {
			return false
		}
	case constant.WhiteKnightJobID:
		if baseSkillID != constant.WarriorJobID && baseSkillID != constant.PageJobID && baseSkillID != constant.WhiteKnightJobID {
			return false
		}
	case constant.SpearmanJobID:
		if baseSkillID != constant.WarriorJobID && baseSkillID != constant.SpearmanJobID {
			return false
		}
	case constant.DragonKnightJobID:
		if baseSkillID != constant.WarriorJobID && baseSkillID != constant.SpearmanJobID && baseSkillID != constant.DragonKnightJobID {
			return false
		}
	case constant.MagicianJobID:
		if baseSkillID != constant.MagicianJobID {
			return false
		}
	case constant.FirePoisonWizardJobID:
		if baseSkillID != constant.MagicianJobID && baseSkillID != constant.FirePoisonWizardJobID {
			return false
		}
	case constant.FirePoisonMageJobID:
		if baseSkillID != constant.MagicianJobID && baseSkillID != constant.FirePoisonWizardJobID && baseSkillID != constant.FirePoisonMageJobID {
			return false
		}
	case constant.IceLightWizardJobID:
		if baseSkillID != constant.MagicianJobID && baseSkillID != constant.IceLightWizardJobID {
			return false
		}
	case constant.IceLightMageJobID:
		if baseSkillID != constant.MagicianJobID && baseSkillID != constant.IceLightWizardJobID && baseSkillID != constant.IceLightMageJobID {
			return false
		}
	case constant.ClericJobID:
		if baseSkillID != constant.MagicianJobID && baseSkillID != constant.ClericJobID {
			return false
		}
	case constant.PriestJobID:
		if baseSkillID != constant.MagicianJobID && baseSkillID != constant.ClericJobID && baseSkillID != constant.PriestJobID {
			return false
		}
	case constant.BowmanJobID:
		if baseSkillID != constant.BowmanJobID {
			return false
		}
	case constant.HunterJobID:
		if baseSkillID != constant.BowmanJobID && baseSkillID != constant.HunterJobID {
			return false
		}
	case constant.RangerJobID:
		if baseSkillID != constant.BowmanJobID && baseSkillID != constant.HunterJobID && baseSkillID != constant.RangerJobID {
			return false
		}
	case constant.CrossbowmanJobID:
		if baseSkillID != constant.BowmanJobID && baseSkillID != constant.CrossbowmanJobID {
			return false
		}
	case constant.SniperJobID:
		if baseSkillID != constant.BowmanJobID && baseSkillID != constant.CrossbowmanJobID && baseSkillID != constant.SniperJobID {
			return false
		}
	case constant.ThiefJobID:
		if baseSkillID != constant.ThiefJobID {
			return false
		}
	case constant.AssassinJobID:
		if baseSkillID != constant.ThiefJobID && baseSkillID != constant.AssassinJobID {
			return false
		}
	case constant.HermitJobID:
		if baseSkillID != constant.ThiefJobID && baseSkillID != constant.AssassinJobID && baseSkillID != constant.HermitJobID {
			return false
		}
	case constant.BanditJobID:
		if baseSkillID != constant.ThiefJobID && baseSkillID != constant.BanditJobID {
			return false
		}
	case constant.ChiefBanditJobID:
		if baseSkillID != constant.ThiefJobID && baseSkillID != constant.BanditJobID && baseSkillID != constant.ChiefBanditJobID {
			return false
		}
	case constant.GmJobID:
		if baseSkillID != constant.GmJobID {
			return false
		}
	case constant.SuperGmJobID:
		if baseSkillID != constant.GmJobID && baseSkillID != constant.SuperGmJobID {
			return false
		}
	default:
		return false
	}

	return true
}
