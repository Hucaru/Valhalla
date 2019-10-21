package server

import (
	"fmt"
	"log"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/entity"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (server *ChannelServer) playerConnect(conn mnet.Client, reader mpacket.Reader) {
	charID := reader.ReadInt32()

	var migrationID byte
	err := server.db.QueryRow("SELECT migrationID FROM characters WHERE id=?", charID).Scan(&migrationID)

	if err != nil {
		log.Println(err)
		return
	}

	if migrationID != server.id {
		return
	}

	var accountID int32
	err = server.db.QueryRow("SELECT accountID FROM characters WHERE id=?", charID).Scan(&accountID)

	if err != nil {
		log.Println(err)
		return
	}

	conn.SetAccountID(accountID)
	char := entity.Character{}
	char.LoadFromID(server.db, charID)

	var adminLevel int
	err = server.db.QueryRow("SELECT adminLevel FROM accounts WHERE accountID=?", conn.GetAccountID()).Scan(&adminLevel)

	if err != nil {
		log.Println(err)
		return
	}

	conn.SetAdminLevel(adminLevel)

	_, err = server.db.Exec("UPDATE characters SET migrationID=? WHERE id=?", -1, charID)

	if err != nil {
		log.Println(err)
		return
	}

	_, err = server.db.Exec("UPDATE characters SET channelID=? WHERE id=?", server.id, charID)

	if err != nil {
		log.Println(err)
		return
	}

	server.players = append(server.players, entity.NewPlayer(conn, char))

	conn.Send(entity.PacketPlayerEnterGame(char, int32(server.id)))
	conn.Send(entity.PacketMessageScrollingHeader(server.header))

	field, ok := server.fields[char.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(0)

	if err != nil {
		return
	}

	inst.AddPlayer(server.players[len(server.players)-1])
}

func (server *ChannelServer) playerChangeChannel(conn mnet.Client, reader mpacket.Reader) {
	id := reader.ReadByte()

	server.migrating = append(server.migrating, conn)
	player, err := server.players.GetFromConn(conn)

	if err != nil {
		log.Println("Unable to get player from connection", conn)
		return
	}

	char := player.Char()

	if int(id) < len(server.channels) {
		if server.channels[id].port == 0 {
			conn.Send(entity.PacketCannotChangeChannel())
		} else {
			_, err := server.db.Exec("UPDATE characters SET migrationID=? WHERE id=?", id, char.ID())

			if err != nil {
				log.Println(err)
				return
			}

			conn.Send(entity.PacketChangeChannel(server.channels[id].ip, server.channels[id].port))
		}
	}
}

func (server ChannelServer) playerMovement(conn mnet.Client, reader mpacket.Reader) {
	player, err := server.players.GetFromConn(conn)

	if err != nil {
		log.Println("Unable to get player from connection", conn)
		return
	}

	char := player.Char()

	if char.PortalCount() != reader.ReadByte() {
		return
	}

	moveData, finalData := entity.ParseMovement(reader)

	if !moveData.ValidateChar(char) {
		return
	}

	moveBytes := entity.GenerateMovementBytes(moveData)

	player.UpdateMovement(finalData)

	field, ok := server.fields[char.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(player.InstanceID())

	if err != nil {
		return
	}

	inst.SendExcept(entity.PacketPlayerMove(char.ID(), moveBytes), conn)
}

func (server ChannelServer) playerEmote(conn mnet.Client, reader mpacket.Reader) {
	emote := reader.ReadInt32()

	player, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	char := player.Char()

	field, ok := server.fields[char.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(player.InstanceID())

	if err != nil {
		return
	}

	inst.SendExcept(entity.PacketPlayerEmoticon(char.ID(), emote), conn)
}

func (server ChannelServer) playerUseMysticDoor(conn mnet.Client, reader mpacket.Reader) {
	player, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	fmt.Println(player.Char().Name(), "has used the mystic door", reader)
}

func (server ChannelServer) playerAddStatPoint(conn mnet.Client, reader mpacket.Reader) {
	player, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	if player.Char().AP() > 0 {
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

func (server ChannelServer) playerRequestAvatarInfoWindow(conn mnet.Client, reader mpacket.Reader) {
	player, err := server.players.GetFromID(reader.ReadInt32())

	if err != nil {
		return
	}

	char := player.Char()

	conn.Send(entity.PacketPlayerAvatarSummaryWindow(char.ID(), char, char.Guild()))
}

func (server ChannelServer) playerPassiveRegen(conn mnet.Client, reader mpacket.Reader) {
	reader.ReadBytes(4) //?

	hp := reader.ReadInt16()
	mp := reader.ReadInt16()

	player, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	char := player.Char()

	if char.HP() == 0 || hp > 400 || mp > 1000 || (hp > 0 && mp > 0) {
		return
	}

	if hp > 0 {
		player.GiveHP(int16(hp))
	} else if mp > 0 {
		player.GiveMP(int16(mp))
	}
}

func (server ChannelServer) playerUseChair(conn mnet.Client, reader mpacket.Reader) {
	fmt.Println("use chair:", reader)
	// chairID := reader.ReadInt32()
}

func (server ChannelServer) playerStand(conn mnet.Client, reader mpacket.Reader) {
	fmt.Println(reader)
	if reader.ReadInt16() == -1 {

	} else {
	}
}

func (server ChannelServer) playerAddSkillPoint(conn mnet.Client, reader mpacket.Reader) {
	player, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	char := player.Char()

	if char.SP() < 1 {
		return // hacker
	}

	skillID := reader.ReadInt32()
	skill, ok := char.Skills()[skillID]

	if ok {
		skill, err = entity.CreateSkillFromData(skillID, skill.Level+1)

		if err != nil {
			return
		}

		player.UpdateSkill(skill)
	} else {
		// check if class can have skill
		baseSkillID := skillID / 10000

		if !validateSkillWithJob(char.Job(), baseSkillID) {
			conn.Send(entity.PacketPlayerNoChange())
			return
		}

		skill, err = entity.CreateSkillFromData(skillID, 1)

		if err != nil {
			return
		}

		player.UpdateSkill(skill)
	}

	player.GiveSP(-1)
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

func (server ChannelServer) playerUsePortal(conn mnet.Client, reader mpacket.Reader) {
	player, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	char := player.Char()

	if char.PortalCount() != reader.ReadByte() {
		conn.Send(entity.PacketPlayerNoChange())
		return
	}

	entryType := reader.ReadInt32()
	field, ok := server.fields[char.MapID()]

	if !ok {
		return
	}

	srcInst, err := field.GetInstance(player.InstanceID())

	if err != nil {
		return
	}

	switch entryType {
	case 0:
		if char.HP() == 0 {
			dstField, ok := server.fields[field.Data.ReturnMap]

			if !ok {
				return
			}

			portal, err := srcInst.GetRandomSpawnPortal()

			if err == nil {
				conn.Send(entity.PacketPlayerNoChange())
				return
			}

			server.warpPlayer(player, dstField, portal)
			player.SetHP(50)
		}
	case -1:
		portalName := reader.ReadString(reader.ReadInt16())
		srcPortal, err := srcInst.GetPortalFromName(portalName)

		if !player.CheckPos(srcPortal.Pos(), 100, 10) { // I'm guessing what the portal hit box is
			if conn.GetAdminLevel() > 0 {
				conn.Send(entity.PacketMessageRedText("Portal - " + srcPortal.Pos().String() + " Player - " + player.Char().Pos().String()))
			}

			conn.Send(entity.PacketPlayerNoChange())
			return
		}

		if err != nil {
			conn.Send(entity.PacketPlayerNoChange())
			return
		}

		dstField, ok := server.fields[srcPortal.DestFieldID()]

		if !ok {
			conn.Send(entity.PacketPlayerNoChange())
			return
		}

		dstInst, err := dstField.GetInstance(player.InstanceID())

		if err != nil {
			if dstInst, err = dstField.GetInstance(0); err != nil {
				return
			}
		}

		dstPortal, err := dstInst.GetPortalFromName(srcPortal.DestName())

		if err != nil {
			conn.Send(entity.PacketPlayerNoChange())
			return
		}

		server.warpPlayer(player, dstField, dstPortal)

	default:
		log.Println("Unknown portal entry type, packet:", reader)
	}
}

func (server ChannelServer) warpPlayer(player *entity.Player, dstField *entity.Field, dstPortal entity.Portal) error {
	srcField, ok := server.fields[player.Char().MapID()]

	if !ok {
		return fmt.Errorf("Error in map id %d", player.Char().MapID())
	}

	srcInst, err := srcField.GetInstance(player.InstanceID())

	if err != nil {
		return err
	}

	dstInst, err := dstField.GetInstance(player.InstanceID())

	if err != nil {
		if dstInst, err = dstField.GetInstance(0); err != nil { // Check player is not in higher level instance than available
			return err
		}

		player.SetInstance(0)
	}

	srcInst.RemovePlayer(player)

	player.SetMapID(dstField.ID)
	player.SetMapPosID(dstPortal.ID())
	player.SetPos(dstPortal.Pos())
	player.SetFoothold(0)
	player.Send(entity.PacketMapChange(dstField.ID, int32(server.id), dstPortal.ID(), player.Char().HP()))

	dstInst.AddPlayer(player)

	return nil
}

// TODO: Refactor/Change logic
func (server ChannelServer) playerMoveInventoryItem(conn mnet.Client, reader mpacket.Reader) {
	inv := reader.ReadByte()
	pos1 := reader.ReadInt16()
	pos2 := reader.ReadInt16()
	amount := reader.ReadInt16()

	player, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	var maxInvSize byte

	switch inv {
	case 1:
		maxInvSize = player.Char().EquipSlotSize()
	case 2:
		maxInvSize = player.Char().UseSlotSize()
	case 3:
		maxInvSize = player.Char().SetupSlotSize()
	case 4:
		maxInvSize = player.Char().EtcSlotSize()
	case 5:
		maxInvSize = player.Char().CashSlotSize()
	}

	if pos2 > int16(maxInvSize) {
		return // Moving to item slot the user does not have
	}

	item1, err := player.GetItem(inv, pos1)

	if err != nil {
		return // Player moving item that doesn't exit
	}

	if pos2 == 0 { // drop item
		fmt.Println(inv, pos1, pos2, amount)
	} else {
		item2, err := player.GetItem(inv, pos2)

		if err != nil { // Move item into empty slot
			if pos2 < 0 {
				if item1.TwoHanded() {
					if _, err = player.GetItem(inv, -10); err == nil { // check for shield
						conn.Send(entity.PacketPlayerNoChange())
						conn.Send(entity.PacketMessageRedText("Cannot equip"))
						return
					}
				} else if item1.Shield() {
					if weapon, err := player.GetItem(inv, -11); err == nil {
						if weapon.TwoHanded() {
							conn.Send(entity.PacketPlayerNoChange())
							conn.Send(entity.PacketMessageRedText("Cannot equip"))
							return
						}
					}
				}
			}

			item1.SetSlotID(pos2)
			player.UpdateItem(item1, item1)
			conn.Send(entity.PacketInventoryChangeItemSlot(inv, pos1, pos2))
		} else {
			if item1.IsStackable() && item2.IsStackable() && (item1.Amount()+item2.Amount()) <= constant.MaxItemStack {
				item2.SetAmount(item2.Amount() + item1.Amount())
				player.UpdateItem(item2, item2)
				player.RemoveItem(item1)
				conn.Send(entity.PacketInventoryAddItem(item2, false))
				conn.Send(entity.PacketInventoryRemoveItem(item1))
			} else { // swap
				if item1.TwoHanded() {
					if _, err = player.GetItem(inv, -10); err == nil {
						conn.Send(entity.PacketPlayerNoChange())
						conn.Send(entity.PacketMessageRedText("Cannot equip"))
						return
					}
				} else if item1.Shield() { // This condition should not be possible....
					if weapon, err := player.GetItem(inv, -11); err == nil {
						if weapon.TwoHanded() {
							conn.Send(entity.PacketPlayerNoChange())
							conn.Send(entity.PacketMessageRedText("Cannot equip"))
							return
						}
					}
				}

				item2.SetSlotID(pos1)
				player.UpdateItem(item2, item2)
				item1.SetSlotID(pos2)
				player.UpdateItem(item1, item1)
				conn.Send(entity.PacketInventoryChangeItemSlot(inv, pos1, pos2))
			}
		}
	}

	if (pos1 < 0 || pos2 < 0) && inv == 1 { // Change equip
		field, ok := server.fields[player.Char().MapID()]

		if !ok {
			return
		}

		inst, err := field.GetInstance(player.InstanceID())

		if err != nil {
			return
		}

		inst.Send(entity.PacketInventoryChangeEquip(player.Char()))
	}

}
