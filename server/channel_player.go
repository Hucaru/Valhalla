package server

import (
	"fmt"
	"log"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/entity"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/field"
	"github.com/Hucaru/Valhalla/server/movement"
	"github.com/Hucaru/Valhalla/server/player"
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

	plr := player.LoadFromID(server.db, charID, conn)

	server.players = append(server.players, &plr)

	conn.Send(player.PacketPlayerEnterGame(plr, int32(server.id)))
	conn.Send(packetMessageScrollingHeader(server.header))

	field, ok := server.fields[plr.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(0)

	if err != nil {
		return
	}

	newPlr, err := server.players.getFromConn(conn)

	if err != nil {
		log.Println(err)
		return
	}

	inst.AddPlayer(newPlr)
}

func (server *ChannelServer) playerChangeChannel(conn mnet.Client, reader mpacket.Reader) {
	id := reader.ReadByte()

	server.migrating = append(server.migrating, conn)
	player, err := server.players.getFromConn(conn)

	if err != nil {
		log.Println("Unable to get player from connection", conn)
		return
	}

	if int(id) < len(server.channels) {
		if server.channels[id].port == 0 {
			conn.Send(entity.PacketCannotChangeChannel())
		} else {
			_, err := server.db.Exec("UPDATE characters SET migrationID=? WHERE id=?", id, player.ID())

			if err != nil {
				log.Println(err)
				return
			}

			conn.Send(entity.PacketChangeChannel(server.channels[id].ip, server.channels[id].port))
		}
	}
}

func (server ChannelServer) playerMovement(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		log.Println("Unable to get player from connection", conn)
		return
	}

	if plr.PortalCount() != reader.ReadByte() {
		return
	}

	moveData, finalData := movement.ParseMovement(reader)

	if !moveData.ValidateChar(plr) {
		return
	}

	moveBytes := movement.GenerateMovementBytes(moveData)

	plr.UpdateMovement(finalData)

	field, ok := server.fields[plr.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(plr.InstanceID())

	if err != nil {
		return
	}

	inst.SendExcept(entity.PacketPlayerMove(plr.ID(), moveBytes), plr)
}

func (server ChannelServer) playerEmote(conn mnet.Client, reader mpacket.Reader) {
	emote := reader.ReadInt32()

	player, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[player.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(player.InstanceID())

	if err != nil {
		return
	}

	inst.SendExcept(entity.PacketPlayerEmoticon(player.ID(), emote), player)
}

func (server ChannelServer) playerUseMysticDoor(conn mnet.Client, reader mpacket.Reader) {
	player, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	fmt.Println(player.Name(), "has used the mystic door", reader)
}

func (server ChannelServer) playerAddStatPoint(conn mnet.Client, reader mpacket.Reader) {
	player, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	if player.AP() > 0 {
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
	player, err := server.players.getFromID(reader.ReadInt32())

	if err != nil {
		return
	}

	conn.Send(packetPlayerAvatarSummaryWindow(player.ID(), *player))
}

func (server ChannelServer) playerPassiveRegen(conn mnet.Client, reader mpacket.Reader) {
	reader.ReadBytes(4) //?

	hp := reader.ReadInt16()
	mp := reader.ReadInt16()

	player, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	if player.HP() == 0 || hp > 400 || mp > 1000 || (hp > 0 && mp > 0) {
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
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	if plr.SP() < 1 {
		return // hacker
	}

	skillID := reader.ReadInt32()
	skill, ok := plr.Skills()[skillID]

	if ok {
		skill, err = player.CreateSkillFromData(skillID, skill.Level+1)

		if err != nil {
			return
		}

		plr.UpdateSkill(skill)
	} else {
		// check if class can have skill
		baseSkillID := skillID / 10000

		if !validateSkillWithJob(plr.Job(), baseSkillID) {
			conn.Send(entity.PacketPlayerNoChange())
			return
		}

		skill, err = player.CreateSkillFromData(skillID, 1)

		if err != nil {
			return
		}

		plr.UpdateSkill(skill)
	}

	plr.GiveSP(-1)
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
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	if plr.PortalCount() != reader.ReadByte() {
		conn.Send(entity.PacketPlayerNoChange())
		return
	}

	entryType := reader.ReadInt32()
	field, ok := server.fields[plr.MapID()]

	if !ok {
		return
	}

	srcInst, err := field.GetInstance(plr.InstanceID())

	if err != nil {
		return
	}

	switch entryType {
	case 0:
		if plr.HP() == 0 {
			dstField, ok := server.fields[field.Data.ReturnMap]

			if !ok {
				return
			}

			portal, err := srcInst.GetRandomSpawnPortal()

			if err == nil {
				conn.Send(entity.PacketPlayerNoChange())
				return
			}

			server.warpPlayer(plr, dstField, portal)
			plr.SetHP(50)
		}
	case -1:
		portalName := reader.ReadString(reader.ReadInt16())
		srcPortal, err := srcInst.GetPortalFromName(portalName)

		if !plr.CheckPos(srcPortal.Pos(), 100, 10) { // trying to account for lag
			if conn.GetAdminLevel() > 0 {
				conn.Send(entity.PacketMessageRedText("Portal - " + srcPortal.Pos().String() + " Player - " + plr.Pos().String()))
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

		dstInst, err := dstField.GetInstance(plr.InstanceID())

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

		server.warpPlayer(plr, dstField, dstPortal)

	default:
		log.Println("Unknown portal entry type, packet:", reader)
	}
}

func (server ChannelServer) warpPlayer(plr *player.Data, dstField *field.Field, dstPortal field.Portal) error {
	srcField, ok := server.fields[plr.MapID()]

	if !ok {
		return fmt.Errorf("Error in map id %d", plr.MapID())
	}

	srcInst, err := srcField.GetInstance(plr.InstanceID())

	if err != nil {
		return err
	}

	dstInst, err := dstField.GetInstance(plr.InstanceID())

	if err != nil {
		if dstInst, err = dstField.GetInstance(0); err != nil { // Check player is not in higher level instance than available
			return err
		}

		plr.SetInstanceID(0)
	}

	srcInst.RemovePlayer(plr)

	plr.SetMapID(dstField.ID)
	plr.SetMapPosID(dstPortal.ID())
	plr.SetPos(dstPortal.Pos())
	plr.SetFoothold(0)
	plr.Send(entity.PacketMapChange(dstField.ID, int32(server.id), dstPortal.ID(), plr.HP()))

	dstInst.AddPlayer(plr)

	return nil
}

func (server ChannelServer) playerMoveInventoryItem(conn mnet.Client, reader mpacket.Reader) {
	inv := reader.ReadByte()
	pos1 := reader.ReadInt16()
	pos2 := reader.ReadInt16()
	amount := reader.ReadInt16()

	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	var maxInvSize byte

	switch inv {
	case 1:
		maxInvSize = plr.EquipSlotSize()
	case 2:
		maxInvSize = plr.UseSlotSize()
	case 3:
		maxInvSize = plr.SetupSlotSize()
	case 4:
		maxInvSize = plr.EtcSlotSize()
	case 5:
		maxInvSize = plr.CashSlotSize()
	}

	if pos2 > int16(maxInvSize) {
		return // Moving to item slot the user does not have
	}

	field, ok := server.fields[plr.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(plr.InstanceID())

	plr.MoveItem(pos1, pos2, amount, inv, inst)

	// item1, err := plr.GetItem(inv, pos1)

	// if err != nil {
	// 	return // Player moving item that doesn't exit
	// }

	// if pos2 == 0 { // drop item
	// 	fmt.Println(inv, pos1, pos2, amount)
	// } else {
	// 	item2, err := plr.GetItem(inv, pos2)

	// 	if err != nil { // Move item into empty slot
	// 		if pos2 < 0 {
	// 			if item1.TwoHanded() {
	// 				if _, err = plr.GetItem(inv, -10); err == nil { // check for shield
	// 					conn.Send(entity.PacketPlayerNoChange())
	// 					conn.Send(entity.PacketMessageRedText("Cannot equip"))
	// 					return
	// 				}
	// 			} else if item1.Shield() {
	// 				if weapon, err := plr.GetItem(inv, -11); err == nil {
	// 					if weapon.TwoHanded() {
	// 						conn.Send(entity.PacketPlayerNoChange())
	// 						conn.Send(entity.PacketMessageRedText("Cannot equip"))
	// 						return
	// 					}
	// 				}
	// 			}
	// 		}

	// 		item1.SetSlotID(pos2)
	// 		plr.UpdateItem(item1, item1)
	// 		conn.Send(entity.PacketInventoryChangeItemSlot(inv, pos1, pos2))
	// 	} else {
	// 		if item1.IsStackable() && item2.IsStackable() && (item1.Amount()+item2.Amount()) <= constant.MaxItemStack {
	// 			item2.SetAmount(item2.Amount() + item1.Amount())
	// 			plr.UpdateItem(item2, item2)
	// 			plr.RemoveItem(item1)
	// 			conn.Send(entity.PacketInventoryAddItem(item2, false))
	// 			conn.Send(entity.PacketInventoryRemoveItem(item1))
	// 		} else { // swap
	// 			if item1.TwoHanded() {
	// 				if _, err = plr.GetItem(inv, -10); err == nil {
	// 					conn.Send(entity.PacketPlayerNoChange())
	// 					conn.Send(entity.PacketMessageRedText("Cannot equip"))
	// 					return
	// 				}
	// 			} else if item1.Shield() { // This condition should not be possible....
	// 				if weapon, err := plr.GetItem(inv, -11); err == nil {
	// 					if weapon.TwoHanded() {
	// 						conn.Send(entity.PacketPlayerNoChange())
	// 						conn.Send(entity.PacketMessageRedText("Cannot equip"))
	// 						return
	// 					}
	// 				}
	// 			}

	// 			item2.SetSlotID(pos1)
	// 			plr.UpdateItem(item2, item2)
	// 			item1.SetSlotID(pos2)
	// 			plr.UpdateItem(item1, item1)
	// 			conn.Send(packetInventoryChangeItemSlot(inv, pos1, pos2))
	// 		}
	// 	}
	// }

	// if (pos1 < 0 || pos2 < 0) && inv == 1 { // Change equip
	// 	field, ok := server.fields[plr.MapID()]

	// 	if !ok {
	// 		return
	// 	}

	// 	inst, err := field.GetInstance(plr.InstanceID())

	// 	if err != nil {
	// 		return
	// 	}

	// 	inst.Send(packetInventoryChangeEquip(plr))
	// }
}
