package server

import (
	"fmt"
	"log"
	"strconv"

	"github.com/Hucaru/Valhalla/server/db"
	"github.com/Hucaru/Valhalla/server/player"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/field"
	"github.com/Hucaru/Valhalla/server/message"
	"github.com/Hucaru/Valhalla/server/metrics"
	"github.com/Hucaru/Valhalla/server/movement"
	"github.com/prometheus/client_golang/prometheus"
)

func (server *ChannelServer) playerConnect(conn mnet.Client, reader mpacket.Reader) {
	charID := reader.ReadInt32()

	var migrationID byte
	err := db.DB.QueryRow("SELECT migrationID FROM characters WHERE id=?", charID).Scan(&migrationID)

	if err != nil {
		log.Println(err)
		return
	}

	if migrationID != server.id {
		return
	}

	var accountID int32
	err = db.DB.QueryRow("SELECT accountID FROM characters WHERE id=?", charID).Scan(&accountID)

	if err != nil {
		log.Println(err)
		return
	}

	conn.SetAccountID(accountID)

	var adminLevel int
	err = db.DB.QueryRow("SELECT adminLevel FROM accounts WHERE accountID=?", conn.GetAccountID()).Scan(&adminLevel)

	if err != nil {
		log.Println(err)
		return
	}

	conn.SetAdminLevel(adminLevel)

	_, err = db.DB.Exec("UPDATE characters SET migrationID=? WHERE id=?", -1, charID)

	if err != nil {
		log.Println(err)
		return
	}

	_, err = db.DB.Exec("UPDATE characters SET channelID=? WHERE id=?", server.id, charID)

	if err != nil {
		log.Println(err)
		return
	}

	plr := player.LoadFromID(charID, conn)

	server.players = append(server.players, &plr)

	conn.Send(player.PacketPlayerEnterGame(plr, int32(server.id)))
	conn.Send(message.PacketMessageScrollingHeader(server.header))

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
	newPlr.UpdateGuildInfo()
	newPlr.UpdateBuddyInfo()

	metrics.Gauges["player_count"].With(prometheus.Labels{"channel": strconv.Itoa(int(server.id)), "world": server.worldName}).Inc()

	server.world.Send(channelPopUpdate(server.id, int16(len(server.players))))
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
			conn.Send(message.PacketCannotChangeChannel())
		} else {
			_, err := db.DB.Exec("UPDATE characters SET migrationID=? WHERE id=?", id, player.ID())

			if err != nil {
				log.Println(err)
				return
			}

			packetChangeChannel := func(ip []byte, port int16) mpacket.Packet {
				p := mpacket.CreateWithOpcode(opcode.SendChannelChange)
				p.WriteBool(true)
				p.WriteBytes(ip)
				p.WriteInt16(port)

				return p
			}

			conn.Send(packetChangeChannel(server.channels[id].ip, server.channels[id].port))
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

	inst.MovePlayer(plr.ID(), moveBytes, plr)
}

func (server ChannelServer) playerEmote(conn mnet.Client, reader mpacket.Reader) {
	emote := reader.ReadInt32()

	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[plr.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(plr.InstanceID())

	if err != nil {
		return
	}

	packetPlayerEmoticon := func(charID int32, emotion int32) mpacket.Packet {
		p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerEmoticon)
		p.WriteInt32(charID)
		p.WriteInt32(emotion)

		return p
	}

	inst.SendExcept(packetPlayerEmoticon(plr.ID(), emote), plr.Conn())
}

func (server ChannelServer) playerUseMysticDoor(conn mnet.Client, reader mpacket.Reader) {
	// doorID := reader.ReadInt32()
	// fromTown := reader.ReadBool()
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
	plr, err := server.players.getFromID(reader.ReadInt32())

	if err != nil {
		return
	}

	packetPlayerAvatarSummaryWindow := func(charID int32, plr player.Data) mpacket.Packet {
		p := mpacket.CreateWithOpcode(opcode.SendChannelAvatarInfoWindow)
		p.WriteInt32(plr.ID())
		p.WriteByte(plr.Level())
		p.WriteInt16(plr.Job())
		p.WriteInt16(plr.Fame())

		p.WriteString(plr.Guild())

		p.WriteBool(false) // if has pet
		p.WriteByte(0)     // wishlist count

		return p
	}

	conn.Send(packetPlayerAvatarSummaryWindow(plr.ID(), *plr))
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

// TODO find better place for this
func packetPlayerNoChange() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInventoryOperation)
	p.WriteByte(0x01)
	p.WriteByte(0x00)
	p.WriteByte(0x00)

	return p
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
			conn.Send(packetPlayerNoChange())
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
	if baseSkillID == 0 { // Beginner skills
		return true
	}

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
		conn.Send(packetPlayerNoChange())
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

			dstInst, err := dstField.GetInstance(plr.InstanceID())

			if err != nil {
				dstInst, err = dstField.GetInstance(0)

				if err != nil {
					return
				}
			}

			portal, err := dstInst.GetRandomSpawnPortal()

			if err != nil {
				conn.Send(packetPlayerNoChange())
				return
			}

			server.warpPlayer(plr, dstField, portal)
			plr.SetHP(50)
			// TODO: reduce exp
		}
	case -1:
		portalName := reader.ReadString(reader.ReadInt16())
		srcPortal, err := srcInst.GetPortalFromName(portalName)

		if !plr.CheckPos(srcPortal.Pos(), 100, 100) { // trying to account for lag whilst preventing teleporting
			if conn.GetAdminLevel() > 0 {
				conn.Send(message.PacketMessageRedText("Portal - " + srcPortal.Pos().String() + " Player - " + plr.Pos().String()))
			}

			conn.Send(packetPlayerNoChange())
			return
		}

		if err != nil {
			conn.Send(packetPlayerNoChange())
			return
		}

		dstField, ok := server.fields[srcPortal.DestFieldID()]

		if !ok {
			conn.Send(packetPlayerNoChange())
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
			conn.Send(packetPlayerNoChange())
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
	}

	srcInst.RemovePlayer(plr)

	plr.SetMapID(dstField.ID)
	plr.SetMapPosID(dstPortal.ID())
	plr.SetPos(dstPortal.Pos())
	// plr.SetFoothold(0)

	packetMapChange := func(mapID int32, channelID int32, mapPos byte, hp int16) mpacket.Packet {
		p := mpacket.CreateWithOpcode(opcode.SendChannelWarpToMap)
		p.WriteInt32(channelID)
		p.WriteByte(0) // character portal counter
		p.WriteByte(0) // Is connecting
		p.WriteInt32(mapID)
		p.WriteByte(mapPos)
		p.WriteInt16(hp)
		p.WriteByte(0) // flag for more reading

		return p
	}

	plr.Send(packetMapChange(dstField.ID, int32(server.id), dstPortal.ID(), plr.HP())) // plr.ChangeMap(dstField.ID, dstPortal.ID(), dstPortal.Pos(), foothold)
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

	err = plr.MoveItem(pos1, pos2, amount, inv, inst)

	if err != nil {
		log.Println(err)
	}
}

func (server ChannelServer) playerUseInventoryItem(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	slot := reader.ReadInt16()
	itemid := reader.ReadInt32()

	item, err := plr.TakeItem(itemid, slot, 1, 2)
	if err != nil {
		log.Println(err)
	}
	item.Use(plr)

}

func (server ChannelServer) playerTakeDamage(conn mnet.Client, reader mpacket.Reader) {
	// 21 FF  or -1 is mob
	// 21 FE  or -2 is bump
	// Anything bigger than -1 is magic

	dmgType := int8(reader.ReadByte())

	if dmgType >= -1 {
		server.mobDamagePlayer(conn, reader, dmgType)
	} else if dmgType == -2 {
		server.playerBumpDamage(conn, reader)
	} else {
		log.Printf("\nUNKNOWN DAMAGE PACKET: %v", reader.String())
	}
}

func (server ChannelServer) playerBumpDamage(conn mnet.Client, reader mpacket.Reader) {
	damage := reader.ReadInt32() // Damage amount

	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	plr.DamagePlayer(int16(damage))

}

func (server ChannelServer) getPlayerInstance(conn mnet.Client, reader mpacket.Reader) (*field.Instance, error) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return nil, err
	}

	field, ok := server.fields[plr.MapID()]

	if !ok {
		return nil, err
	}

	inst, err := field.GetInstance(plr.InstanceID())

	if err != nil {
		return nil, err
	}

	return inst, nil
}

func (server *ChannelServer) playerBuddyOperation(conn mnet.Client, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 1: // Add
		plr, err := server.players.getFromConn(conn)

		if err != nil {
			log.Fatal(err)
			return
		}

		if plr.BuddyListFull() {
			conn.Send(message.PacketBuddyPlayerFullList())
			return
		}

		name := reader.ReadString(reader.ReadInt16())

		var charID int32
		var accountID int32
		var buddyListSize int32

		err = db.DB.QueryRow("SELECT id,accountID,buddyListSize FROM characters WHERE name=? and worldID=?", name, conn.GetWorldID()).Scan(&charID, &accountID, &buddyListSize)

		if err != nil || accountID == conn.GetAccountID() {
			conn.Send(message.PacketBuddyNameNotRegistered())
			return
		}

		var recepientBuddyCount int32
		err = db.DB.QueryRow("SELECT COUNT(*) FROM buddy WHERE characterID=1 and accepted=1").Scan(&recepientBuddyCount)

		if err != nil {
			log.Fatal(err)
			return
		}

		if recepientBuddyCount >= buddyListSize {
			conn.Send(message.PacketBuddyOtherFullList())
			return
		}

		if conn.GetAdminLevel() == 0 {
			var gm bool
			err = db.DB.QueryRow("SELECT adminLevel from accounts where accountID=?", accountID).Scan(&gm)

			if err != nil {
				log.Fatal(err)
				return
			}

			if gm {
				conn.Send(message.PacketBuddyIsGM())
				return
			}
		}

		query := "INSERT INTO buddy(characterID,friendID) VALUES(?,?)"

		if _, err = db.DB.Exec(query, charID, plr.ID()); err != nil {
			log.Fatal(err)
			return
		}

		if recepient, err := server.players.getFromID(charID); err != nil {
			// emit a friend request event to all channels
		} else {
			recepient.Send(message.PacketBuddyReceiveRequest(plr.ID(), plr.Name(), int32(server.id)))
		}
	case 2: // Accept request
		plr, err := server.players.getFromConn(conn)

		if err != nil {
			log.Fatal(err)
			return
		}

		friendID := reader.ReadInt32()

		var friendName string
		var friendChannel int32
		var cashShop bool

		err = db.DB.QueryRow("SELECT name,channelID,inCashShop FROM characters WHERE id=?", friendID).Scan(&friendName, &friendChannel, &cashShop)

		if err != nil {
			log.Fatal(err)
			return
		}

		query := "UPDATE buddy set accepted=1 WHERE characterID=? and friendID=?"

		if _, err := db.DB.Exec(query, plr.ID(), friendID); err != nil {
			log.Fatal(err)
			return
		}

		query = "INSERT INTO buddy(characterID,friendID,accepted) VALUES(?,?,?)"

		if _, err := db.DB.Exec(query, friendID, plr.ID(), 1); err != nil {
			log.Fatal(err)
			return
		}

		if friendChannel == -1 {
			plr.AddOfflineBuddy(friendID, friendName)
		} else {
			plr.AddOnlineBuddy(friendID, friendName, friendChannel)
		}

		if recepient, err := server.players.getFromID(friendID); err != nil {
			// emit friend request accepted, along with channel id
		} else {
			// Need to set the buddy to be offline for the logged in message to appear before setting online
			recepient.AddOfflineBuddy(plr.ID(), plr.Name())
			recepient.Send(message.PacketBuddyOnlineStatus(plr.ID(), int32(server.id)))
			recepient.AddOnlineBuddy(plr.ID(), plr.Name(), int32(server.id))
		}
	case 3: // Delete/reject friend
		fmt.Println("Delete", reader)

		// Delete both sides of the friends list from the database, retreive the friendID and use in subsequent event broadcast
		// Check if on current channel otherwise emit a friend delete event to the channels
	default:
		log.Println("Unknown buddy operation:", op)
	}
}
