package channel

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/constant/skill"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus metrics for packet observability
var (
	packetsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "channel_packets_total",
			Help: "Total number of received packets by opcode",
		},
		[]string{"opcode"},
	)
	unknownPacketsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "channel_packets_unknown_total",
			Help: "Total number of unknown/unhandled packets",
		},
	)
)

func init() {
	prometheus.MustRegister(packetsTotal, unknownPacketsTotal)
}

func (server *Server) HandleClientPacket(conn mnet.Client, reader mpacket.Reader) {
	// Read opcode first for logging/metrics and to make panic logs useful
	op := reader.ReadByte()
	packetsTotal.WithLabelValues(fmt.Sprintf("%d", op)).Inc()

	// Panic guard per packet to avoid dropping the connection loop on handler bugs
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic in HandleClientPacket op=%d: %v", op, r)
		}
	}()

	switch op {
	case opcode.RecvPing:
	case opcode.RecvClientMigrate:
		server.playerConnect(conn, reader)
	case opcode.RecvCHannelChangeChannel:
		server.playerChangeChannel(conn, reader)
	case opcode.RecvChannelUserPortal:
		// This opcode is used for revival UI as well.
		server.playerUsePortal(conn, reader)
	case opcode.RecvChannelScriptedPortal:
		server.playerUseScriptedPortal(conn, reader)
	case opcode.RecvChannelEnterCashShop:
		server.playerEnterCashShop(conn, reader)
	case opcode.RecvChannelPlayerMovement:
		server.playerMovement(conn, reader)
	case opcode.RecvChannelPlayerStand:
		server.playerStand(conn, reader)
	case opcode.RecvChannelChairHeal:
		server.playerChairHeal(conn, reader)
	case opcode.RecvChannelInvUseCashItem:
		server.playerUseCash(conn, reader)
	case opcode.RecvChannelPlayerUseChair:
		server.playerUseChair(conn, reader)
	case opcode.RecvChannelMeleeSkill:
		server.playerMeleeSkill(conn, reader)
	case opcode.RecvChannelRangedSkill:
		server.playerRangedSkill(conn, reader)
	case opcode.RecvChannelMagicSkill:
		server.playerMagicSkill(conn, reader)
	case opcode.RecvChannelDmgRecv:
		server.playerTakeDamage(conn, reader)
	case opcode.RecvChannelPlayerSendAllChat:
		server.chatSendAll(conn, reader)
	case opcode.RecvChannelGroupChat:
		server.chatGroup(conn, reader)
	case opcode.RecvChannelSlashCommands:
		server.chatSlashCommand(conn, reader)
	case opcode.RecvChannelCharacterUIWindow:
		server.roomWindow(conn, reader)
	case opcode.RecvChannelEmote:
		server.playerEmote(conn, reader)
	case opcode.RecvChannelNpcDialogue:
		server.npcChatStart(conn, reader)
	case opcode.RecvChannelNpcDialogueContinue:
		server.npcChatContinue(conn, reader)
	case opcode.RecvChannelNpcShop:
		server.npcShop(conn, reader)
	case opcode.RecvChannelInvMoveItem:
		server.playerMoveInventoryItem(conn, reader)
	case opcode.RecvChannelPlayerDropMesos:
		server.playerDropMesos(conn, reader)
	case opcode.RecvChannelPlayerFame:
		server.playerFame(conn, reader)
	case opcode.RecvChannelInvUseItem:
		server.playerUseInventoryItem(conn, reader)
	case opcode.RecvChannelNearestTown:
		// Return Scroll / Nearest Town Scroll
		server.playerUseReturnScroll(conn, reader)
	case opcode.RecvChannelUseScroll:
		server.playerUseScroll(conn, reader)
	case opcode.RecvChannelPlayerPickup:
		server.playerPickupItem(conn, reader)
	case opcode.RecvChannelAddStatPoint:
		server.playerAddStatPoint(conn, reader)
	case opcode.RecvChannelPassiveRegen:
		server.playerPassiveRegen(conn, reader)
	case opcode.RecvChannelAddSkillPoint:
		server.playerAddSkillPoint(conn, reader)
	case opcode.RecvChannelSpecialSkill:
		server.playerSpecialSkill(conn, reader)
	case opcode.RecvChannelRequestBuffCancel:
		server.playerStopSkill(conn, reader)
	case opcode.RecvChannelCharacterInfo:
		server.playerRequestAvatarInfoWindow(conn, reader)
	case opcode.RecvChannelLieDetectorResult:
	case opcode.RecvChannelPartyInfo:
		server.playerPartyInfo(conn, reader)
	case opcode.RecvChannelGuildManagement:
		server.guildManagement(conn, reader)
	case opcode.RecvChannelGuildReject:
		server.guildInviteResult(conn, reader)
	case opcode.RecvChannelBuddyOperation:
		server.playerBuddyOperation(conn, reader)
	case opcode.RecvChannelUseMysticDoor:
		server.playerUseMysticDoor(conn, reader)
	case opcode.RecvChannelMobControl:
		server.mobControl(conn, reader)
	case opcode.RecvChannelDistance:
		server.mobDistance(conn, reader)
	case opcode.RecvChannelNpcMovement:
		server.npcMovement(conn, reader)
	case opcode.RecvChannelBoatMap:
		// [mapID int32][? byte]
	case opcode.RecvChannelAcknowledgeBuff:
		// Consume
	case opcode.RecvChannelCancelBuff:
		server.playerCancelBuff(conn, reader)
	case opcode.RecvChannelQuestOperation:
		server.playerQuestOperation(conn, reader)
	case opcode.RecvChannelSummonMove:
		server.playerSummonMove(conn, reader)
	case opcode.RecvChannelSummonDamage:
		server.playerSummonDamage(conn, reader)
	case opcode.RecvChannelSummonAttack:
		server.playerSummonAttack(conn, reader)
	case opcode.RecvChannelReactorHit:
		server.playerHitReactor(conn, reader)
	case opcode.RecvChannelNpcStorage:
		server.playerUseStorage(conn, reader)
	case opcode.RecvChannelMessenger:
		server.playerHandleMessenger(conn, reader)
	case opcode.RecvChannelPetSpawn:
		server.playerPetSpawn(conn, reader)
	case opcode.RecvChannelPetMove:
		server.playerPetMove(conn, reader)
	case opcode.RecvChannelPetAction:
		server.playerPetAction(conn, reader)
	case opcode.RecvChannelPetInteraction:
		server.playerPetInteraction(conn, reader)
	case opcode.RecvChannelPetLoot:
		server.playerPetLoot(conn, reader)
	case opcode.RecvChannelUseSack:
		server.playerUseSack(conn, reader)
	default:
		unknownPacketsTotal.Inc()
		// Let's send a no change to make sure characters aren't stuck on unknown packets
		conn.Send(packetPlayerNoChange())
		log.Println("UNKNOWN CLIENT PACKET(", op, "):", reader)
	}
}

func (server *Server) playerConnect(conn mnet.Client, reader mpacket.Reader) {
	charID := reader.ReadInt32()

	// Fetch channelID, migrationID and accountID in a single query
	var (
		migrationID byte
		channelID   int8
		accountID   int32
	)
	err := common.DB.QueryRow(
		"SELECT channelID, migrationID, accountID FROM characters WHERE ID=?",
		charID,
	).Scan(&channelID, &migrationID, &accountID)
	if err != nil {
		log.Println("playerConnect query error:", err)
		return
	}

	if migrationID != server.id {
		// Not for this server; silently ignore to avoid leaking info
		return
	}

	conn.SetAccountID(accountID)

	var adminLevel int
	err = common.DB.QueryRow("SELECT adminLevel FROM accounts WHERE accountID=?", conn.GetAccountID()).Scan(&adminLevel)

	if err != nil {
		log.Println(err)
		return
	}

	conn.SetAdminLevel(adminLevel)

	_, err = common.DB.Exec("UPDATE characters SET migrationID=? WHERE ID=?", -1, charID)

	if err != nil {
		log.Println(err)
		return
	}

	_, err = common.DB.Exec("UPDATE characters SET channelID=? WHERE ID=?", server.id, charID)

	if err != nil {
		log.Println(err)
		return
	}

	plr := LoadPlayerFromID(charID, conn)
	plr.rates = &server.rates

	server.players.Add(&plr)

	conn.Send(packetPlayerEnterGame(plr, int32(server.id)))
	conn.Send(packetMessageScrollingHeader(server.header))

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(0)

	if err != nil {
		return
	}

	newPlr, err := server.players.GetFromConn(conn)

	if err != nil {
		log.Println(err)
		return
	}

	for _, party := range server.parties {
		if party.addExistingPlayer(newPlr) {
			break
		}
	}

	newPlr.sendBuddyList()

	newPlr.UpdatePartyInfo = func(partyID, playerID, job, level, mapID int32, name string) {
		server.world.Send(internal.PacketChannelPartyUpdateInfo(partyID, playerID, job, level, mapID, name))
	}

	var guildID sql.NullInt32
	err = common.DB.QueryRow("SELECT guildID FROM characters WHERE ID=?", newPlr.ID).Scan(&guildID)

	if err != nil {
		log.Fatal(err)
	}

	if guildID.Valid {
		if guild, ok := server.guilds[guildID.Int32]; !ok {
			guild, err = loadGuildFromDb(guildID.Int32, &server.players)

			if err == nil {
				server.guilds[guildID.Int32] = guild
				newPlr.guild = guild
			}
		} else {
			newPlr.guild = server.guilds[guildID.Int32]
		}
	} else {
		var guildID int32
		var inviter string
		row, err := common.DB.Query("SELECT guildID, inviter FROM guild_invites WHERE playerID=?", newPlr.ID)

		if err != nil {
			log.Fatal(err)
		}

		defer row.Close()

		for row.Next() { // We should only ever have 1 row
			row.Scan(&guildID, &inviter)
			newPlr.Send(packetGuildInviteCard(guildID, inviter))
		}
	}

	err = inst.addPlayer(newPlr)

	if err != nil {
		log.Println(err)
		return
	}

	// Restore buffs (if any) saved during CC or previous logout, then audit for stale
	newPlr.loadAndApplyBuffSnapshot()

	if newPlr.buffs != nil {
		newPlr.buffs.plr.inst = newPlr.inst
		newPlr.buffs.AuditAndExpireStaleBuffs()
	}

	// Check for quest mob kills
	if len(newPlr.quests.inProgressList()) > 0 {
		for _, q := range newPlr.quests.inProgressList() {
			m := newPlr.quests.mobKills[q.id]
			if m == nil {
				continue
			}
			questData, err := nx.GetQuest(q.id)
			if err != nil {
				continue
			}

			newPlr.Send(packetQuestUpdateMobKills(q.id, newPlr.buildQuestKillString(questData)))
		}
	}

	common.MetricsGauges["player_count"].With(prometheus.Labels{"channel": strconv.Itoa(int(server.id)), "world": server.worldName}).Inc()

	server.world.Send(internal.PacketChannelPopUpdate(server.id, int16(server.players.count())))

	if guildID.Valid {
		server.world.Send(internal.PacketChannelPlayerConnected(plr.ID, plr.Name, server.id, channelID > -1, newPlr.mapID, guildID.Int32))
	} else {
		server.world.Send(internal.PacketChannelPlayerConnected(plr.ID, plr.Name, server.id, channelID > -1, newPlr.mapID, 0))
	}
}

func (server *Server) playerChangeChannel(conn mnet.Client, reader mpacket.Reader) {
	id := reader.ReadByte()

	server.migrating = append(server.migrating, conn)
	player, err := server.players.GetFromConn(conn)
	if err != nil {
		log.Println("Unable to get Player from connection", conn)
		return
	}

	// Expire Summon Buffs
	player.expireSummons()
	server.removeSummonsFromField(player)
	player.saveBuffSnapshot()

	if int(id) < len(server.channels) {
		if server.channels[id].Port == 0 {
			conn.Send(packetCannotChangeChannel())
		} else {
			if _, err := common.DB.Exec("UPDATE characters SET migrationID=? WHERE ID=?", id, player.ID); err != nil {
				log.Println(err)
				return
			}

			conn.Send(packetChangeChannel(server.channels[id].IP, server.channels[id].Port))
		}
	}
}

func (server Server) playerMovement(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)

	if err != nil {
		log.Println("Unable to get Player from connection", conn)
		return
	}

	if plr.portalCount != reader.ReadByte() {
		return
	}

	moveData, finalData := parseMovement(reader)

	if !moveData.validateChar(plr) {
		return
	}

	moveBytes := generateMovementBytes(moveData)

	plr.UpdateMovement(finalData)

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)

	if err != nil {
		return
	}

	inst.movePlayer(plr.ID, moveBytes, plr)
}

func (server Server) playerEmote(conn mnet.Client, reader mpacket.Reader) {
	emote := reader.ReadInt32()

	plr, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)

	if err != nil {
		return
	}

	inst.sendExcept(packetPlayerEmoticon(plr.ID, emote), plr.Conn)
}

func (server Server) playerUseMysticDoor(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	if plr.inst == nil {
		plr.Send(packetPlayerNoChange())
		return
	}

	var (
		nearestDoor *mysticDoorInfo
		ownerID     int32
		minDist     int32 = 1<<30 - 1
	)
	for oid, door := range plr.inst.mysticDoors {
		dx := int32(plr.pos.x) - int32(door.pos.x)
		dy := int32(plr.pos.y) - int32(door.pos.y)
		dist := dx*dx + dy*dy
		if dist < minDist {
			minDist = dist
			nearestDoor = door
			ownerID = oid
		}
	}

	if nearestDoor == nil {
		plr.Send(packetPlayerNoChange())
		return
	}

	authorized := plr.ID == ownerID
	if !authorized && plr.party != nil {
		for _, pid := range plr.party.PlayerID {
			if pid == ownerID {
				authorized = true
				break
			}
		}
	}
	if !authorized {
		plr.Send(packetPlayerNoChange())
		return
	}

	if nearestDoor.townPortal {
		// Town -> Source
		destMapID := nearestDoor.destMapID
		destField, ok := server.fields[destMapID]
		if !ok {
			plr.Send(packetPlayerNoChange())
			return
		}

		var (
			dstInst       *fieldInstance
			destPortalIdx int
		)
		for instID := 0; instID < len(destField.instances); instID++ {
			if inst, err := destField.getInstance(instID); err == nil {
				if doorData, exists := inst.mysticDoors[ownerID]; exists && !doorData.townPortal {
					dstInst = inst
					destPortalIdx = doorData.portalIndex
					break
				}
			}
		}
		if dstInst == nil {
			plr.Send(packetPlayerNoChange())
			return
		}
		if destPortalIdx < 0 || destPortalIdx >= len(dstInst.portals) {
			plr.Send(packetPlayerNoChange())
			return
		}
		dstPortal := dstInst.portals[destPortalIdx]
		if err := server.warpPlayer(plr, destField, dstPortal, true); err != nil {
			plr.Send(packetPlayerNoChange())
			return
		}
	} else {
		destMapID := nearestDoor.destMapID
		destField, ok := server.fields[destMapID]
		if !ok {
			plr.Send(packetPlayerNoChange())
			return
		}

		dstInst, err2 := destField.getInstance(0)
		if err2 != nil {
			plr.Send(packetPlayerNoChange())
			return
		}
		townDoor, ok := dstInst.mysticDoors[ownerID]
		if !ok || !townDoor.townPortal {
			plr.Send(packetPlayerNoChange())
			return
		}
		destPortalIdx := townDoor.portalIndex
		if destPortalIdx < 0 || destPortalIdx >= len(dstInst.portals) {
			plr.Send(packetPlayerNoChange())
			return
		}
		dstPortal := dstInst.portals[destPortalIdx]
		if err := server.warpPlayer(plr, destField, dstPortal, true); err != nil {
			plr.Send(packetPlayerNoChange())
			return
		}
	}
}

func (server Server) playerAddStatPoint(conn mnet.Client, reader mpacket.Reader) {
	player, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	if player.ap > 0 {
		player.giveAP(-1)
	}

	statID := reader.ReadInt32()

	switch statID {
	case constant.StrID:
		player.giveStr(1)
	case constant.DexID:
		player.giveDex(1)
	case constant.IntID:
		player.giveInt(1)
	case constant.LukID:
		player.giveLuk(1)
	default:
		fmt.Println("unknown stat ID:", statID)
	}
}

func (server Server) playerRequestAvatarInfoWindow(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromID(reader.ReadInt32())

	if err != nil {
		return
	}

	conn.Send(packetPlayerAvatarSummaryWindow(plr.ID, *plr))
}

func (server Server) playerPassiveRegen(conn mnet.Client, reader mpacket.Reader) {
	flag := reader.ReadInt32()

	var hp, mp int16

	if flag&0x0400 != 0 {
		hp = reader.ReadInt16()
	} else {
		hp = 0
	}

	if flag&0x1000 != 0 {
		mp = reader.ReadInt16()
	} else {
		mp = 0
	}

	player, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	if player.hp == 0 || hp > 400 || mp > 1000 || (hp > 0 && mp > 0) {
		return
	}

	if hp > 0 {
		player.giveHP(hp)
	} else if mp > 0 {
		player.giveMP(mp)
	}
}

func (server Server) playerUseChair(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		log.Println(err)
		return
	}

	chairID := reader.ReadInt32()
	plr.chairID = chairID

	plr.inst.sendExcept(packetPlayerShowChair(plr.ID, chairID), plr.Conn)
	plr.Send(packetPlayerChairUpdate())
}

func (server Server) playerStand(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		log.Println(err)
		return
	}

	chairIndex := reader.ReadInt16()

	plr.chairID = 0
	plr.inst.sendExcept(packetPlayerShowChair(plr.ID, 0), plr.Conn)
	plr.Send(packetPlayerChairResult(chairIndex))
	plr.Send(packetPlayerChairUpdate())
}

func (server Server) playerChairHeal(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		log.Println(err)
		return
	}

	healAmount := reader.ReadInt16()

	if plr.chairID == 0 {
		return
	}

	chair, err := nx.GetItem(plr.chairID)
	if err != nil {
		return
	}

	if healAmount > int16(chair.RecoveryHP) {
		return
	}

	if time.Since(plr.lastChairHeal) < 10*time.Second {
		return
	}

	plr.giveHP(healAmount)
	plr.lastChairHeal = time.Now()
}

func (server Server) playerAddSkillPoint(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	if plr.sp < 1 {
		return // hacker
	}

	skillID := reader.ReadInt32()
	skill, ok := plr.skills[skillID]

	if ok {
		skill, err = createPlayerSkillFromData(skillID, skill.Level+1)

		if err != nil {
			return
		}

		plr.updateSkill(skill)
	} else {
		// check if class can have skill
		baseSkillID := skillID / 10000
		if !validateSkillWithJob(plr.job, baseSkillID) {
			conn.Send(packetPlayerNoChange())
			return
		}

		skill, err = createPlayerSkillFromData(skillID, 1)

		if err != nil {
			return
		}

		plr.updateSkill(skill)
	}

	plr.giveSP(-1)
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

func (server *Server) playerEnterCashShop(conn mnet.Client, reader mpacket.Reader) {
	server.migrating = append(server.migrating, conn)
	player, err := server.players.GetFromConn(conn)
	if err != nil {
		log.Println("Unable to get Player from connection", conn)
		return
	}

	// Expire Summon Buffs
	player.expireSummons()
	server.removeSummonsFromField(player)
	player.saveBuffSnapshot()

	if len(server.cashShop.IP) > 0 || server.cashShop.Port == 0 {
		if _, err := common.DB.Exec("UPDATE characters SET migrationID=?, previousChannelID=?, inCashShop=1 WHERE ID=?", 50, server.id, player.ID); err != nil {
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
		conn.Send(packetChangeChannel(server.cashShop.IP, server.cashShop.Port))
	} else {
		conn.Send(packetCannotEnterCashShop())
	}
}

func (server *Server) removeSummonsFromField(player *Player) {
	if player.summons != nil {
		if field, ok := server.fields[player.mapID]; ok && field != nil {
			if inst, e := field.getInstance(player.inst.id); e == nil && inst != nil {
				if player.summons.puppet != nil {
					inst.send(packetRemoveSummon(player.ID, player.summons.puppet.SkillID, constant.SummonRemoveReasonCancel))
				}
				if player.summons.summon != nil {
					inst.send(packetRemoveSummon(player.ID, player.summons.summon.SkillID, constant.SummonRemoveReasonCancel))
				}
			}
		}
		player.summons.puppet = nil
		player.summons.summon = nil
	}
}

func (server Server) playerUsePortal(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	if plr.portalCount != reader.ReadByte() {
		conn.Send(packetPlayerNoChange())
		return
	}

	entryType := reader.ReadInt32()

	curField, ok := server.fields[plr.mapID]
	if !ok || curField == nil {
		return
	}

	instID := -1
	if plr.inst != nil {
		instID = plr.inst.id
	}

	srcInst, err := curField.getInstance(instID)
	if err != nil {
		if inst0, e2 := curField.getInstance(0); e2 == nil {
			if _, has := inst0.getPlayerFromID(plr.ID); has != nil {
				_ = inst0.addPlayer(plr)
			}
			plr.inst = inst0
			srcInst = inst0
		} else {
			return
		}
	}

	chooseDstPortal := func(dstInst *fieldInstance, backToMapID int32, srcName, preferName string) (portal, error) {
		if preferName != "" {
			if p, e := dstInst.getPortalFromName(preferName); e == nil {
				return p, nil
			}
		}
		for _, p := range dstInst.portals {
			if p.destFieldID == backToMapID && p.destName == srcName {
				return p, nil
			}
		}
		for _, p := range dstInst.portals {
			if p.destFieldID == backToMapID {
				return p, nil
			}
		}
		return dstInst.getRandomSpawnPortal()
	}

	switch entryType {
	case 0:
		// Death revive to return map
		if plr.hp == 0 {
			dstFld, ok := server.fields[curField.Data.ReturnMap]
			if !ok || dstFld == nil {
				return
			}

			dstInst, err := dstFld.getInstance(instID)
			if err != nil {
				dstInst, err = dstFld.getInstance(0)
				if err != nil {
					return
				}
			}

			dstPortal, err := chooseDstPortal(dstInst, curField.id, "", "")
			if err != nil {
				conn.Send(packetPlayerNoChange())
				return
			}

			if err := server.warpPlayer(plr, dstFld, dstPortal, true); err != nil {
				return
			}

			plr.setHP(50)
			return
		}

	case -1:
		nameLen := reader.ReadInt16()
		if nameLen <= 0 {
			conn.Send(packetPlayerNoChange())
			return
		}
		portalName := reader.ReadString(nameLen)

		srcPortal, err := srcInst.getPortalFromName(portalName)
		if err != nil {
			conn.Send(packetPlayerNoChange())
			return
		}

		if !plr.checkPos(srcPortal.pos, 100, 100) {
			if conn.GetAdminLevel() > 0 {
				conn.Send(packetMessageRedText("Portal - " + srcPortal.pos.String() + " Player - " + plr.pos.String()))
			}
			conn.Send(packetPlayerNoChange())
			return
		}

		dstFld, ok := server.fields[srcPortal.destFieldID]
		if !ok || dstFld == nil {
			conn.Send(packetPlayerNoChange())
			return
		}

		dstInst, err := dstFld.getInstance(instID)
		if err != nil {
			dstInst, err = dstFld.getInstance(0)
			if err != nil {
				conn.Send(packetPlayerNoChange())
				return
			}
		}

		dstPortal, err := chooseDstPortal(dstInst, curField.id, srcPortal.name, srcPortal.destName)
		if err != nil {
			conn.Send(packetPlayerNoChange())
			return
		}

		if err := server.warpPlayer(plr, dstFld, dstPortal, true); err != nil {
			return
		}
	}
}

func (server Server) playerUseScriptedPortal(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	nameLen := reader.ReadInt16()

	if nameLen <= 0 {
		plr.Send(packetPlayerNoChange())
		return
	}

	portalName := reader.ReadString(nameLen)

	srcPortal, err := plr.inst.getPortalFromName(portalName)

	if err != nil {
		plr.Send(packetPlayerNoChange())
		return
	}

	// Validate range to prevent teleport hacks
	if !plr.checkPos(srcPortal.pos, 100, 100) {
		if conn.GetAdminLevel() > 0 {
			plr.Send(packetMessageRedText("ScriptedPortal - " + srcPortal.pos.String() + " Player - " + plr.pos.String()))
		}
		plr.Send(packetPlayerNoChange())
		return
	}

	warp := func(plr *Player, mapID int32, portalName string, checkActive bool, maxPlayers int, minLevel byte) {
		dstField, ok := server.fields[mapID]

		if !ok {
			plr.Send(packetPlayerNoChange())
			return
		}

		inst, err := dstField.getInstance(plr.inst.id)

		if err != nil {
			log.Println(err)
			plr.Send(packetPlayerNoChange())
			return
		}

		if checkActive {
			bossActive, ok := inst.properties["eventActive"].(bool)

			if !ok {
				bossActive = false
			}

			if bossActive {
				plr.Send(packetMessageRedText("The fight is already active. Please try again later."))
				plr.Send(packetPlayerNoChange())
				return
			}
		}

		if maxPlayers > 0 && len(inst.players) >= maxPlayers {
			plr.Send(packetMessageRedText("The boss room is currently full."))
			plr.Send(packetPlayerNoChange())
			return
		}

		if plr.level < minLevel {
			plr.Send(packetMessageRedText(fmt.Sprintf("You must be level %d or higher to enter.", minLevel)))
			plr.Send(packetPlayerNoChange())
			return
		}

		portal, err := inst.getPortalFromName(portalName)

		if err != nil {
			log.Println(err)
			plr.Send(packetPlayerNoChange())
			return
		}

		server.warpPlayer(plr, dstField, portal, true)
	}

	switch portalName {
	case constant.PortalFreeMarketEnter:
		warp(plr, constant.MapFreeMarket, "out00", false, 0, 10)
	case constant.PortalFreeMarketLeave:
		previousMap := plr.previousMap
		warp(plr, previousMap, "market00", false, 0, 0)
		plr.previousMap = previousMap
	case constant.PortalPapulatus:
		warp(plr, constant.MapBossPapulatus, "st00", true, 0, 120)
	case constant.PortalPianus:
		warp(plr, constant.MapBossPianus, "out00", false, 10, 0)
	case constant.PortalZakum:
		warp(plr, constant.MapBossZakum, "st00", true, 20, 50)
	}
}

func (server Server) warpPlayer(plr *Player, dstField *field, dstPortal portal, usedPortal bool) error {
	srcField, ok := server.fields[plr.mapID]
	if !ok {
		return fmt.Errorf("Error in map ID %d", plr.mapID)
	}

	srcInst, err := srcField.getInstance(plr.inst.id)
	if err != nil {
		return err
	}

	dstInst, err := dstField.getInstance(plr.inst.id)
	if err != nil {
		if dstInst, err = dstField.getInstance(0); err != nil {
			return err
		}
	}

	var keptSummon *summon
	if plr.summons != nil && plr.summons.summon != nil {
		keptSummon = plr.summons.summon
	}

	server.removeSummonsFromField(plr)

	if err = srcInst.removePlayer(plr, usedPortal); err != nil {
		return err
	}

	plr.setMapID(dstField.id)
	plr.pos = dstPortal.pos

	plr.Send(packetMapChange(dstField.id, int32(server.id), dstPortal.id, plr.hp))

	if err = dstInst.addPlayer(plr); err != nil {
		return err
	}

	if keptSummon != nil && plr.shouldKeepSummonOnTransfer(keptSummon) {
		snapped := dstInst.fhHist.getFinalPosition(newPos(plr.pos.x, plr.pos.y, 0))
		keptSummon.Pos = snapped
		keptSummon.Foothold = snapped.foothold
		keptSummon.Stance = 0
		plr.addSummon(keptSummon)
	} else {
		// Buff is gone or summon missing, do not keep it.
		if plr.summons != nil {
			if plr.summons.summon == keptSummon {
				plr.summons.summon = nil
			}
			if plr.summons.puppet == keptSummon {
				plr.summons.puppet = nil
			}
		}
	}

	if plr.petCashID != 0 && plr.pet.spawned {
		plr.pet.pos = plr.pos
		plr.pet.pos.y -= 15
		dstInst.send(packetPetSpawn(plr.ID, plr.pet))
	}

	return nil
}

func (server Server) playerMoveInventoryItem(conn mnet.Client, reader mpacket.Reader) {
	inv := reader.ReadByte()
	pos1 := reader.ReadInt16()
	pos2 := reader.ReadInt16()
	amount := reader.ReadInt16()

	plr, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	var maxInvSize byte

	switch inv {
	case constant.InventoryEquip:
		maxInvSize = plr.equipSlotSize
	case constant.InventoryUse:
		maxInvSize = plr.useSlotSize
	case constant.InventorySetup:
		maxInvSize = plr.setupSlotSize
	case constant.InventoryEtc:
		maxInvSize = plr.etcSlotSize
	case constant.InventoryCash:
		maxInvSize = plr.cashSlotSize
	}

	if pos2 > int16(maxInvSize) {
		return // Moving to Item slot the user does not have
	}

	err = plr.moveItem(pos1, pos2, amount, inv)

	if err != nil {
		log.Println(err)
	}
}

func (server Server) playerDropMesos(conn mnet.Client, reader mpacket.Reader) {
	amount := reader.ReadInt32()
	plr, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	err = plr.dropMesos(amount)
	if err != nil {
		log.Println(err)
	}

	plr.inst.dropPool.createDrop(dropSpawnNormal, dropFreeForAll, amount, plr.pos, true, plr.ID, plr.ID)

}

func (server Server) playerUseInventoryItem(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	slot := reader.ReadInt16()
	itemid := reader.ReadInt32()

	item, err := plr.takeItem(itemid, slot, 1, 2)
	if err != nil {
		log.Println(err)
	}
	item.use(plr)

}

func (server *Server) playerUseReturnScroll(conn mnet.Client, reader mpacket.Reader) {
	slot := reader.ReadInt16()   // inventory slot in 'use' tab
	itemID := reader.ReadInt32() // Item ID

	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	// Validate: ensure the Item at the given slot in 'use' inventory matches itemID
	found := false
	for _, it := range plr.use {
		if it.slotID == slot && it.ID == itemID && it.amount > 0 {
			found = true
			break
		}
	}
	if !found {
		// Desync or invalid request
		plr.Send(packetPlayerNoChange())
		return
	}

	meta, err := nx.GetItem(itemID)
	if err != nil {
		// Missing NX data
		plr.Send(packetPlayerNoChange())
		return
	}

	// Resolve destination field ID with sensible fallbacks
	var dstFieldID int32
	// Prefer MoveTo from Item data if available
	if meta.MoveTo != 0 {
		dstFieldID = meta.MoveTo
	}

	// If MoveTo is not present, try Player's previous map
	if dstFieldID == 0 && plr.previousMap != 0 {
		dstFieldID = plr.previousMap
	}

	// If still unknown, try the current map's ReturnMap
	if dstFieldID == 0 {
		if curField, ok := server.fields[plr.mapID]; ok && curField.Data.ReturnMap != 0 {
			dstFieldID = curField.Data.ReturnMap
		}
	}

	// Final fallback: if nothing resolved, don't change state
	if dstFieldID == 0 {
		plr.Send(packetPlayerNoChange())
		return
	}

	// Resolve destination field
	dstField, ok := server.fields[dstFieldID]
	if !ok {
		if dstField == nil {
			if curField, ok2 := server.fields[plr.mapID]; ok2 && curField.Data.ReturnMap != 0 {
				if f2, ok3 := server.fields[curField.Data.ReturnMap]; ok3 {
					dstField = f2
				}
			}
		}
		// If still unresolved, abort
		if dstField == nil {
			plr.Send(packetPlayerNoChange())
			return
		}
	}

	// Resolve destination instance (use same index if possible, else 0)
	dstInst, err := dstField.getInstance(plr.inst.id)
	if err != nil {
		dstInst, err = dstField.getInstance(0)
		if err != nil {
			plr.Send(packetPlayerNoChange())
			return
		}
	}

	// Pick a spawn portal
	portal, err := dstInst.getRandomSpawnPortal()
	if err != nil {
		plr.Send(packetPlayerNoChange())
		return
	}

	// Consume one scroll from the specified slot in 'use' inventory (invID 2)
	if _, err := plr.takeItem(itemID, slot, 1, 2); err != nil {
		// If we fail to consume, do not warp
		plr.Send(packetPlayerNoChange())
		return
	}

	_ = server.warpPlayer(plr, dstField, portal, true)
}

func (server *Server) playerUseScroll(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	scrollSlot := reader.ReadInt16() // USE inv slot (scroll)
	targetSlot := reader.ReadInt16() // equip slot (negative if equipped)

	if targetSlot < -100 {
		plr.Send(packetPlayerNoChange())
		return
	}

	// Load items
	scroll := plr.findUseItemBySlot(scrollSlot)
	equip := plr.findEquipBySlot(targetSlot)
	if scroll == nil || equip == nil || scroll.amount < 1 || equip.amount != 1 {
		plr.Send(packetPlayerNoChange())
		return
	}

	// Data and basic checks
	scrollMeta, err := nx.GetItem(scroll.ID)
	if err != nil {
		plr.Send(packetPlayerNoChange())
		return
	}
	// Type compatibility and basic checks
	if !validateScrollTarget(scroll.ID, equip.ID) {
		plr.Send(packetPlayerNoChange())
		return
	}
	if int(scrollMeta.Success) == 0 || equip.getSlots() == 0 {
		plr.Send(packetPlayerNoChange())
		return
	}

	// Consume scroll
	if _, err := plr.takeItem(scroll.ID, scrollSlot, 1, 2); err != nil {
		plr.Send(packetPlayerNoChange())
		return
	}

	// consume slot
	equip.setSlots(equip.getSlots() - 1)

	// Roll success/failure
	rand.Seed(time.Now().UnixNano())
	successRoll := rand.Intn(100)

	if successRoll < int(scrollMeta.Success) {
		// Success: apply stats, decrement slot, increment scroll count
		equip.applyScrollEffects(scrollMeta)
		equip.incrementScrollCount()

		// Persist and update in-memory slice
		equip.save(plr.ID)
		plr.updateItem(*equip)

		// Send full Item update so client refreshes stats/slots
		plr.Send(packetInventoryAddItem(*equip, true))
		// Optional: refresh avatar appearance (won't change looks, but safe)
		plr.Send(packetInventoryChangeEquip(*plr))
		plr.Send(packetUseScroll(plr.ID, true, false, false))
	} else {
		curseRoll := rand.Intn(100)
		if curseRoll < int(scrollMeta.Cursed) {
			// Destroy the equip
			plr.removeItem(*equip)
			plr.Send(packetUseScroll(plr.ID, false, true, false))
		} else {
			// Normal fail (slot consumed): persist and Send full update too
			equip.save(plr.ID)
			plr.updateItem(*equip)
			plr.Send(packetInventoryAddItem(*equip, true))
			plr.Send(packetUseScroll(plr.ID, false, false, false))
		}
	}
}

func (server *Server) playerUseCash(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	slot := reader.ReadInt16()
	itemID := reader.ReadInt32()

	used := false

	switch itemID {
	case constant.ItemSafetyCharm:
		plr.hasSafetyCharm = true
		used = true

	case constant.ItemAPReset:
		statUp := reader.ReadInt32()
		statDown := reader.ReadInt32()

		if (statUp == constant.StrID || statUp == constant.DexID || statUp == constant.IntID || statUp == constant.LukID) &&
			(statDown == constant.StrID || statDown == constant.DexID || statDown == constant.IntID || statDown == constant.LukID) &&
			statUp != statDown {
			canReset := false
			switch statDown {
			case constant.StrID:
				canReset = plr.str > 4
			case constant.DexID:
				canReset = plr.dex > 4
			case constant.IntID:
				canReset = plr.intt > 4
			case constant.LukID:
				canReset = plr.luk > 4
			}

			if canReset {
				switch statDown {
				case constant.StrID:
					plr.setStr(plr.str - 1)
				case constant.DexID:
					plr.setDex(plr.dex - 1)
				case constant.IntID:
					plr.setInt(plr.intt - 1)
				case constant.LukID:
					plr.setLuk(plr.luk - 1)
				}

				switch statUp {
				case constant.StrID:
					plr.setStr(plr.str + 1)
				case constant.DexID:
					plr.setDex(plr.dex + 1)
				case constant.IntID:
					plr.setInt(plr.intt + 1)
				case constant.LukID:
					plr.setLuk(plr.luk + 1)
				}

				used = true
			}
		}

	case constant.ItemSPResetFirstJob, constant.ItemSPResetSecondJob, constant.ItemSPResetThirdJob:
		skillUp := reader.ReadInt32()
		skillDown := reader.ReadInt32()

		skillUpData, okUp := plr.skills[skillUp]
		skillDownData, okDown := plr.skills[skillDown]

		if plr.skills[skillUp].Mastery < skillUpData.Level+1 {
			plr.Send(packetPlayerNoChange())
			return
		}

		if okUp && okDown && skillDownData.Level > 0 {
			skillDownData.Level--
			if skillDownData.Level == 0 {
				delete(plr.skills, skillDown)
			} else {
				plr.updateSkill(skillDownData)
			}

			skillUpData.Level++
			plr.updateSkill(skillUpData)

			used = true
		}

	case constant.ItemVIPTeleportRock, constant.ItemRegTeleportRock:
		mode := reader.ReadByte()

		if mode == constant.TeleportToName {
			targetName := reader.ReadString(reader.ReadInt16())

			targetPlr, err := server.players.GetFromName(targetName)
			if err != nil {
				plr.Send(packetMessageRedText(fmt.Sprintf("Player %s not found in current channel.", targetName)))
				return
			}

			targetField, ok := server.fields[targetPlr.mapID]
			if !ok {
				plr.Send(packetPlayerNoChange())
				return
			}

			targetInst, err := targetField.getInstance(targetPlr.inst.id)
			if err != nil {
				plr.Send(packetPlayerNoChange())
				return
			}

			portal, err := targetInst.getRandomSpawnPortal()
			if err != nil {
				plr.Send(packetPlayerNoChange())
				return
			}

			if err := server.warpPlayer(plr, targetField, portal, false); err != nil {
				log.Println("Teleport rock warp error:", err)
				plr.Send(packetPlayerNoChange())
				return
			}

			used = true
		} else {
			mapID := reader.ReadInt32()

			targetField, ok := server.fields[mapID]
			if !ok {
				plr.Send(packetPlayerNoChange())
				return
			}

			targetInst, err := targetField.getInstance(0)
			if err != nil {
				plr.Send(packetPlayerNoChange())
				return
			}

			portal, err := targetInst.getRandomSpawnPortal()
			if err != nil {
				plr.Send(packetPlayerNoChange())
				return
			}

			if err := server.warpPlayer(plr, targetField, portal, false); err != nil {
				log.Println("Teleport rock warp error:", err)
				plr.Send(packetPlayerNoChange())
				return
			}

			used = true
		}

	case constant.ItemPetNameTag:
		newName := reader.ReadString(reader.ReadInt16())

		if plr.pet != nil && plr.pet.spawned {
			plr.pet.name = newName

			plr.MarkDirty(DirtyPet, time.Millisecond*300)

			if plr.inst != nil {
				plr.inst.send(packetPetNameChange(plr.ID, newName))
			}

			used = true
		}

	case constant.ItemWaterOfLife:
		log.Printf("Water of Life not fully implemented")

	case constant.ItemMegaphone, constant.ItemSuperMegaphone:
		msg := reader.ReadString(reader.ReadInt16())

		if itemID == constant.ItemMegaphone {
			server.players.broadcast(packetMessageBroadcastChannel(plr.Name, msg, server.id, false))
		} else {
			whisper := reader.ReadBool()
			server.world.Send(internal.PacketChatMegaphone(plr.Name, msg, whisper))
		}
		used = true
	case constant.ItemWeatherCandy, constant.ItemWeatherFlower, constant.ItemWeatherFireworks, constant.ItemWeatherSoap, constant.ItemWeatherSnow, constant.ItemWeatherSnowFlakes,
		constant.ItemWeatherPresents, constant.ItemWeatherLeaves, constant.ItemWeatherChocolate, constant.ItemWeatherFlowers:
		msg := reader.ReadString(reader.ReadInt16())

		if plr.inst != nil {
			if plr.inst.startWeatherEffect(itemID, msg) {
				used = true
			}
		}

	default:
		log.Printf("Unhandled Cash Item Use: %d", itemID)
		plr.Send(packetPlayerNoChange())
		return
	}

	if used {
		if _, err = plr.takeItem(itemID, slot, 1, 5); err != nil {
			log.Println(err)
			return
		}
	}

	plr.Send(packetPlayerNoChange())

}

func (server Server) playerPickupItem(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	posx := reader.ReadInt16()
	posy := reader.ReadInt16()
	dropID := reader.ReadInt32()

	pos := pos{
		x: posx,
		y: posy,
	}

	err, drop := plr.inst.dropPool.findDropFromID(dropID)

	if err != nil {
		plr.Send(packetDropNotAvailable())
		log.Printf("drop Unavailable: %v\nError: %s", drop, err)
		return
	}

	if plr.pos.x-pos.x > 800 || plr.pos.y-pos.y > 600 {
		// Hax
		log.Printf("Player: %s tried to pickup an Item from far away", plr.Name)
		plr.Send(packetDropNotAvailable())
		plr.Send(packetInventoryDontTake())
		return
	}

	if drop.mesos > 0 {
		amount := int32(plr.inst.dropPool.rates.mesos * float32(drop.mesos))
		plr.giveMesos(amount)
	} else {
		err, _ = plr.GiveItem(drop.item)
		if err != nil {
			plr.Send(packetInventoryFull())
			plr.Send(packetInventoryDontTake())
			return
		}

	}

	plr.inst.dropPool.playerAttemptPickup(drop, plr, 2)

}

func (server Server) playerTakeDamage(conn mnet.Client, reader mpacket.Reader) {
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

func (server Server) playerBumpDamage(conn mnet.Client, reader mpacket.Reader) {
	damage := reader.ReadInt32() // Damage amount

	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	plr.damagePlayer(int16(damage))

}

func (server Server) getPlayerInstance(conn mnet.Client, reader mpacket.Reader) (*fieldInstance, error) {
	plr, err := server.players.GetFromConn(conn)

	if err != nil {
		return nil, err
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return nil, err
	}

	inst, err := field.getInstance(plr.inst.id)

	if err != nil {
		return nil, err
	}

	return inst, nil
}

func (server *Server) playerBuddyOperation(conn mnet.Client, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 1: // Add
		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		if plr.buddyListFull() {
			conn.Send(packetBuddyPlayerFullList())
			return
		}

		name := reader.ReadString(reader.ReadInt16())

		var charID int32
		var accountID int32
		var buddyListSize int32

		err = common.DB.QueryRow("SELECT ID,accountID,buddyListSize FROM characters WHERE BINARY Name=? and worldID=?", name, conn.GetWorldID()).Scan(&charID, &accountID, &buddyListSize)

		if err != nil || accountID == conn.GetAccountID() {
			conn.Send(packetBuddyNameNotRegistered())
			return
		}

		if plr.hasBuddy(charID) {
			conn.Send(packetBuddyAlreadyAdded())
			return
		}

		var recepientBuddyCount int32
		err = common.DB.QueryRow("SELECT COUNT(*) FROM buddy WHERE characterID=1 and accepted=1").Scan(&recepientBuddyCount)

		if err != nil {
			log.Fatal(err)
			return
		}

		if recepientBuddyCount >= buddyListSize {
			conn.Send(packetBuddyOtherFullList())
			return
		}

		if conn.GetAdminLevel() == 0 {
			var gm bool
			err = common.DB.QueryRow("SELECT adminLevel from accounts where accountID=?", accountID).Scan(&gm)

			if err != nil {
				log.Fatal(err)
				return
			}

			if gm {
				conn.Send(packetBuddyIsGM())
				return
			}
		}

		query := "INSERT INTO buddy(characterID,friendID) VALUES(?,?)"

		if _, err = common.DB.Exec(query, charID, plr.ID); err != nil {
			log.Fatal(err)
			return
		}

		if recepient, err := server.players.GetFromID(charID); err != nil {
			server.world.Send(internal.PacketChannelBuddyEvent(1, charID, plr.ID, plr.Name, server.id))
		} else {
			recepient.Send(packetBuddyReceiveRequest(plr.ID, plr.Name, int32(server.id)))
		}
	case 2: // Accept request
		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		friendID := reader.ReadInt32()

		var friendName string
		var friendChannel int32
		var cashShop bool

		err = common.DB.QueryRow("SELECT Name,channelID,inCashShop FROM characters WHERE ID=?", friendID).Scan(&friendName, &friendChannel, &cashShop)

		if err != nil {
			log.Fatal(err)
			return
		}

		query := "UPDATE buddy set accepted=1 WHERE characterID=? and friendID=?"

		if _, err := common.DB.Exec(query, plr.ID, friendID); err != nil {
			log.Fatal(err)
			return
		}

		query = "INSERT INTO buddy(characterID,friendID,accepted) VALUES(?,?,?)"

		if _, err := common.DB.Exec(query, friendID, plr.ID, 1); err != nil {
			log.Fatal(err)
			return
		}

		if friendChannel == -1 {
			plr.addOfflineBuddy(friendID, friendName)
		} else {
			plr.addOnlineBuddy(friendID, friendName, friendChannel)
		}

		if recepient, err := server.players.GetFromID(friendID); err != nil {
			server.world.Send(internal.PacketChannelBuddyEvent(2, friendID, plr.ID, plr.Name, server.id))
		} else {
			// Need to set the buddy to be offline for the logged in message to appear before setting online
			recepient.addOfflineBuddy(plr.ID, plr.Name)
			recepient.Send(packetBuddyOnlineStatus(plr.ID, int32(server.id)))
			recepient.addOnlineBuddy(plr.ID, plr.Name, int32(server.id))
		}
	case 3: // Delete/reject friend
		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		id := reader.ReadInt32()

		query := "DELETE FROM buddy WHERE (characterID=? AND friendID=?) OR (characterID=? AND friendID=?)"

		if _, err = common.DB.Exec(query, id, plr.ID, plr.ID, id); err != nil {
			log.Fatal(err)
			return
		}

		plr.removeBuddy(id)

		if recepient, err := server.players.GetFromID(id); err != nil {
			server.world.Send(internal.PacketChannelBuddyEvent(3, id, plr.ID, "", server.id))
		} else {
			recepient.removeBuddy(plr.ID)
		}
	default:
		log.Println("Unknown buddy operation:", op)
	}
}

func (server *Server) playerPartyInfo(conn mnet.Client, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 1: // create party
		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		if plr.party != nil {
			plr.Send(packetPartyAlreadyJoined())
			return
		}

		server.world.Send(internal.PacketChannelPartyCreateRequest(plr.ID, server.id, plr.mapID, int32(plr.job), int32(plr.level), plr.Name))
	case 2: // leave party
		if b := reader.ReadByte(); b != 0 { // Not sure what this byte/bool does
			log.Println("Leave party byte is not zero:", b)
		}

		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		if plr.party == nil {
			return
		}

		partyID := plr.party.ID

		server.world.Send(internal.PacketChannelPartyLeave(partyID, plr.ID, false))
	case 3: // accept
		partyID := reader.ReadInt32()

		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		server.world.Send(internal.PacketChannelPartyAccept(partyID, plr.ID, int32(server.id), plr.mapID, int32(plr.job), int32(plr.level), plr.Name))
	case 4: // invite
		id := reader.ReadInt32()

		recipient, err := server.players.GetFromID(id)

		if err != nil {
			conn.Send(packetPartyUnableToFindPlayer())
			return
		}

		if recipient.party != nil {
			conn.Send(packetPartyAlreadyJoined())
			return
		}

		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		if plr.party == nil {
			plr.Send(packetPartyUnableToFindPlayer())
			return
		}

		if plr.party.full() {
			plr.Send(packetPartyToJoinIsFull())
			return
		}

		recipient.Send(packetPartyInviteNotice(plr.party.ID, plr.Name))
	case 5: // expel
		playerID := reader.ReadInt32()

		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		if plr.party == nil {
			plr.Send(packetPartyUnableToFindPlayer())
			return
		}

		server.world.Send(internal.PacketChannelPartyLeave(plr.party.ID, playerID, true))
	default:
		log.Println("Unknown party info type:", op, reader)
	}
}

func (server Server) chatGroup(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	op := reader.ReadByte()

	switch op {
	case 0: // buddy
		buffer := reader.GetRestAsBytes()
		server.world.Send(internal.PacketChannelPlayerChat(internal.OpChatBuddy, plr.Name, buffer))
	case 1: // party
		buffer := reader.GetRestAsBytes()
		server.world.Send(internal.PacketChannelPlayerChat(internal.OpChatParty, plr.Name, buffer))
	case 2: // guild
		buffer := reader.GetRestAsBytes()
		server.world.Send(internal.PacketChannelPlayerChat(internal.OpChatGuild, plr.Name, buffer))
	default:
		log.Println("Unknown group chat type:", op, reader)
	}
}

func (server Server) chatSlashCommand(conn mnet.Client, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 5: // find / map button in friend
		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}
		name := reader.ReadString(reader.ReadInt16())

		var accountID int32
		var channelID int8
		var mapID int32 = -1
		var inCashShop bool

		err = common.DB.QueryRow("SELECT accountID,channelID,mapID,inCashShop FROM characters WHERE BINARY Name=? AND worldID=?", name, conn.GetWorldID()).Scan(&accountID, &channelID, &mapID, &inCashShop)

		if err != nil || channelID == -1 {
			plr.Send(packetMessageFindResult(name, false, false, false, -1))
			return
		}

		var isGM bool

		err = common.DB.QueryRow("SELECT adminLevel from accounts where accountID=?", accountID).Scan(&isGM)

		if err != nil {
			log.Fatal(err)
			return
		}

		if isGM {
			plr.Send(packetMessageFindResult(name, false, inCashShop, false, mapID))
		} else {
			plr.Send(packetMessageFindResult(name, true, inCashShop, byte(channelID) == server.id, mapID))
		}
	case 6: // whispher
		recepientName := reader.ReadString(reader.ReadInt16())
		msg := reader.ReadString(reader.ReadInt16())

		if receiver, err := server.players.GetFromName(recepientName); err != nil {
			var online bool
			err := common.DB.QueryRow("SELECT COUNT(*) FROM characters WHERE BINARY Name=? AND worldID=? AND channelID != -1", recepientName, conn.GetWorldID()).Scan(&online)

			if err != nil || !online {
				conn.Send(packetMessageRedText("Incorrect character Name"))
				return
			}

			plr, err := server.players.GetFromConn(conn)

			if err != nil {
				return
			}

			plr.Send(packetMessageWhisper(plr.Name, msg, server.id))
			server.world.Send(internal.PacketChannelWhispherChat(recepientName, plr.Name, msg, server.id))
		} else {
			plr, err := server.players.GetFromConn(conn)

			if err != nil {
				return
			}

			plr.Send(packetMessageWhisper(plr.Name, msg, server.id))
			receiver.Send(packetMessageWhisper(plr.Name, msg, server.id))
		}
	default:
		log.Println("Unkown slash command type:", op, reader)
	}
}

func (server *Server) chatSendAll(conn mnet.Client, reader mpacket.Reader) {
	msg := reader.ReadString(reader.ReadInt16())

	if strings.Index(msg, "/") == 0 && conn.GetAdminLevel() > 0 {
		server.gmCommand(conn, msg)
	} else {
		player, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		inst, err := server.fields[player.mapID].getInstance(player.inst.id)

		if err != nil {
			return
		}

		inst.send(packetMessageAllChat(player.ID, conn.GetAdminLevel() > 0, msg))
	}
}

func (server Server) mobControl(conn mnet.Client, reader mpacket.Reader) {
	mobSpawnID := reader.ReadInt32()
	moveID := reader.ReadInt16()
	bits := reader.ReadByte()
	action := reader.ReadInt8()
	skillData := reader.ReadUint32()

	skillPossible := (bits & 0x0F) != 0

	plr, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	inst, err := server.getPlayerInstance(conn, reader)
	if err != nil {
		return
	}

	moveData, finalData := parseMovement(reader)

	moveBytes := generateMovementBytes(moveData)

	inst.lifePool.mobAcknowledge(mobSpawnID, plr, moveID, skillPossible, action, skillData, moveData, finalData, moveBytes)

}

func (server Server) mobDamagePlayer(conn mnet.Client, reader mpacket.Reader, mobAttack int8) {
	damage := reader.ReadInt32() // Damage amount
	healSkillID := int32(0)

	if damage < -1 {
		return
	}

	reducedDamage := damage

	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	field, ok := server.fields[plr.mapID]
	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)
	if err != nil {
		return
	}

	var mob monster
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

		mob, err = inst.lifePool.getMobFromID(spawnID)
		if err != nil {
			return
		}

		if mob.id != mobID {
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

		// Magic guard dmg absorption
		if _, hasMagicGuard := plr.buffs.activeSkillLevels[int32(skill.MagicGuard)]; hasMagicGuard {
			skillData, err := nx.GetPlayerSkill(int32(skill.MagicGuard))
			if err == nil {
				if ps, ok := plr.skills[int32(skill.MagicGuard)]; ok && ps.Level > 0 {
					idx := int(ps.Level) - 1
					if idx < len(skillData) {
						absorbPercent := int32(skillData[idx].X)
						if absorbPercent > 0 {
							mpDamage := (damage * absorbPercent) / 100
							hpDamage := damage - mpDamage

							if mpDamage > 0 {
								plr.giveMP(-int16(mpDamage))
							}

							damage = hpDamage
							reducedDamage = hpDamage
						}
					}
				}
			}
		}

		// Fighter / Page power guard - reflects damage back to mob
		powerGuardSkillID := int32(0)
		if _, hasPowerGuard := plr.buffs.activeSkillLevels[int32(skill.PowerGuard)]; hasPowerGuard {
			powerGuardSkillID = int32(skill.PowerGuard)
		} else if _, hasPagePowerGuard := plr.buffs.activeSkillLevels[int32(skill.PagePowerGuard)]; hasPagePowerGuard {
			powerGuardSkillID = int32(skill.PagePowerGuard)
		}

		if powerGuardSkillID > 0 {
			skillData, err := nx.GetPlayerSkill(powerGuardSkillID)
			if err == nil {
				if ps, ok := plr.skills[powerGuardSkillID]; ok && ps.Level > 0 {
					idx := int(ps.Level) - 1
					if idx < len(skillData) {
						reflectPercent := int32(skillData[idx].X)
						if reflectPercent > 0 {
							reflectedDamage := (damage * reflectPercent) / 100

							if reflectedDamage > 0 {
								inst.lifePool.mobDamaged(spawnID, plr, reflectedDamage)
							}
						}
					}
				}
			}
		}

		// Meso guard - uses mesos to reduce damage
		if _, hasMesoGuard := plr.buffs.activeSkillLevels[int32(skill.MesoGuard)]; hasMesoGuard {
			skillData, err := nx.GetPlayerSkill(int32(skill.MesoGuard))
			if err == nil {
				if ps, ok := plr.skills[int32(skill.MesoGuard)]; ok && ps.Level > 0 {
					idx := int(ps.Level) - 1
					if idx < len(skillData) {
						absorbPercent := int32(skillData[idx].X)
						mesoCost := int32(skillData[idx].Y)

						if absorbPercent > 0 && mesoCost > 0 {
							absorbedDamage := (damage * absorbPercent) / 100
							totalMesoCost := absorbedDamage * mesoCost

							if plr.mesos >= totalMesoCost {
								plr.giveMesos(-totalMesoCost)
								damage -= absorbedDamage
								reducedDamage = damage
							}
						}
					}
				}
			}
		}

		if !plr.admin() {
			plr.damagePlayer(int16(damage))
		}

		inst.send(packetPlayerReceivedDmg(plr.ID, mobAttack, damage, reducedDamage, spawnID, mobID, healSkillID, stance, reflectAction, reflected, reflectX, reflectY))
	}
	if mobSkillID != 0 && mobSkillLevel != 0 {
		// Apply mob skill debuff to the player
		levels, err := nx.GetMobSkill(mobSkillID)
		if err == nil && int(mobSkillLevel) > 0 && int(mobSkillLevel) <= len(levels) {
			skillData := levels[mobSkillLevel-1]
			durationSec := int16(0)
			if skillData.Time > 0 {
				durationSec = int16(skillData.Time) // Time is already in seconds
			}
			plr.addMobDebuff(mobSkillID, mobSkillLevel, durationSec)
		}
	}

}

func (server Server) mobDistance(conn mnet.Client, reader mpacket.Reader) {
	/*
		ID := reader.ReadInt32()
		distance := reader.ReadInt32()

		Unknown what this packet is for
	*/

}

func (server Server) playerMeleeSkill(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)

	if err != nil {
		conn.Send(packetMessageRedText(err.Error()))
		return
	}

	data, valid := getAttackInfo(reader, *plr, attackMelee)

	if !valid {
		return
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)

	if err != nil {
		conn.Send(packetMessageRedText(err.Error()))
		return
	}

	err = plr.useSkill(data.skillID, data.skillLevel, data.projectileID)
	if err != nil {
		// Send packet to stop?
		return
	}

	inst.sendExcept(packetSkillMelee(*plr, data), conn)

	if data.isMesoExplosion {
		server.handleMesoExplosion(plr, inst, data)
	} else {
		for _, attack := range data.attackInfo {
			inst.lifePool.mobDamaged(attack.spawnID, plr, attack.damages...)
		}
	}
}

func (server Server) playerRangedSkill(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)

	if err != nil {
		conn.Send(packetMessageRedText(err.Error()))
		return
	}

	data, valid := getAttackInfo(reader, *plr, attackRanged)

	if !valid {
		return
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)

	if err != nil {
		conn.Send(packetMessageRedText(err.Error()))
		return
	}

	err = plr.useSkill(data.skillID, data.skillLevel, data.projectileID)
	if err != nil {
		// Send packet to stop?
		return
	}

	// if Player in party extract

	inst.sendExcept(packetSkillRanged(*plr, data), conn)

	for _, attack := range data.attackInfo {
		inst.lifePool.mobDamaged(attack.spawnID, plr, attack.damages...)
	}
}

func (server Server) playerMagicSkill(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)

	if err != nil {
		conn.Send(packetMessageRedText(err.Error()))
		return
	}

	data, valid := getAttackInfo(reader, *plr, attackMagic)

	if !valid {
		return
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)

	if err != nil {
		conn.Send(packetMessageRedText(err.Error()))
		return
	}

	err = plr.useSkill(data.skillID, data.skillLevel, data.projectileID)
	if err != nil {
		return
	}

	if skill.Skill(data.skillID) == skill.PoisonMyst {
		skillData, err := nx.GetPlayerSkill(data.skillID)
		if err == nil && int(data.skillLevel) > 0 && int(data.skillLevel) <= len(skillData) {
			duration := skillData[data.skillLevel-1].Time
			magicAttack := int16(skillData[data.skillLevel-1].Mad)
			inst.mistPool.createMist(plr.ID, data.skillID, data.skillLevel, plr.pos, duration, true, magicAttack)
		}
	}

	inst.sendExcept(packetSkillMagic(*plr, data), conn)

	for _, attack := range data.attackInfo {
		inst.lifePool.mobDamaged(attack.spawnID, plr, attack.damages...)
	}
}

func (server *Server) handleMesoExplosion(plr *Player, inst *fieldInstance, data attackData) {
	allDropIDs := make(map[int32]struct{}, 16)
	hadMappedDrop := make([]bool, len(data.attackInfo))

	for _, at := range data.attackInfo {
		if len(at.mesoDropIDs) == 0 {
			continue
		}
		for _, did := range at.mesoDropIDs {
			allDropIDs[did] = struct{}{}
		}
	}

	existing := make(map[int32]struct{}, len(allDropIDs))
	for did := range allDropIDs {
		if drop, ok := inst.dropPool.drops[did]; ok && drop.mesos > 0 {
			existing[did] = struct{}{}
		}
	}

	for tIdx, at := range data.attackInfo {
		found := false
		for _, did := range at.mesoDropIDs {
			if _, ok := existing[did]; ok {
				found = true
				break
			}
		}
		hadMappedDrop[tIdx] = found
	}

	removed := 0
	for did := range existing {
		inst.dropPool.removeDrop(4, did)
		removed++
	}

	for tIdx, at := range data.attackInfo {
		if at.spawnID == 0 || len(at.damages) == 0 {
			continue
		}

		if hadMappedDrop[tIdx] {
			inst.lifePool.mobDamaged(at.spawnID, plr, at.damages...)
		}
	}
}

const (
	attackMelee = iota
	attackRanged
	attackMagic
	attackSummon
)

type attackInfo struct {
	spawnID                                                int32
	hitAction, foreAction, frameIndex, calcDamageStatIndex byte
	facesLeft                                              bool
	hitPosition, previousMobPosition                       pos
	hitDelay                                               int16
	hitCount                                               byte
	damages                                                []int32
	mesoDropIDs                                            []int32
}

type attackData struct {
	skillID, summonType, totalDamage, projectileID int32
	isMesoExplosion, facesLeft                     bool
	option, action, attackType                     byte
	targets, hits, skillLevel                      byte
	mesoKillDelay                                  int16

	attackInfo []attackInfo
	playerPos  pos
}

func getAttackInfo(reader mpacket.Reader, player Player, attackType int) (attackData, bool) {
	data := attackData{}

	if player.hp == 0 {
		return data, false
	}

	if reader.Time-player.lastAttackPacketTime < 350 {
		return data, false
	}
	player.lastAttackPacketTime = reader.Time

	if attackType != attackSummon {
		tByte := reader.ReadByte()
		skillID := reader.ReadInt32()

		if _, ok := player.skills[skillID]; !ok && skillID != 0 {
			return data, false
		}

		data.skillID = skillID
		if data.skillID != 0 {
			data.skillLevel = player.skills[skillID].Level
		}

		data.isMesoExplosion = (skill.Skill(data.skillID) == skill.MesoExplosion)

		data.targets = tByte / 0x10
		data.hits = tByte % 0x10

		data.option = reader.ReadByte()

		tmp := reader.ReadByte()
		data.action = tmp & 0x7F
		data.facesLeft = (tmp >> 7) == 1
		data.attackType = reader.ReadByte()

		// The client sends a 4-byte field here (checksum/random). Keep alignment by consuming it.
		_ = reader.ReadInt32()

		if attackType == attackRanged {
			projectileSlot := reader.ReadInt16()
			_ = reader.ReadInt16()
			_ = reader.ReadByte()

			if projectileSlot != 0 {
				data.projectileID = -1
				for _, item := range player.use {
					if item.slotID == projectileSlot {
						data.projectileID = item.ID
						break
					}
				}
				if data.projectileID == -1 {
					return data, false
				}
			} else {
				// No projectile slot given: must have Soul Arrow active
				hasSoulArrow := false
				if player.buffs != nil {
					if _, ok := player.buffs.activeSkillLevels[int32(skill.SoulArrow)]; ok {
						hasSoulArrow = true
					}
					if _, ok := player.buffs.activeSkillLevels[int32(skill.CBSoulArrow)]; ok {
						hasSoulArrow = true
					}
				}
				if !hasSoulArrow {
					return data, false
				}
				data.projectileID = -1
			}

		}

		// Parse per-target blocks (supports Meso Explosion per-target hit count)
		data.attackInfo = make([]attackInfo, data.targets)
		for i := byte(0); i < data.targets; i++ {
			ai := attackInfo{}
			ai.spawnID = reader.ReadInt32()
			ai.hitAction = reader.ReadByte()

			tmp := reader.ReadByte()
			ai.foreAction = tmp & 0x7F
			ai.facesLeft = (tmp >> 7) == 1
			ai.frameIndex = reader.ReadByte()

			ai.calcDamageStatIndex = reader.ReadByte()

			ai.hitPosition.x = reader.ReadInt16()
			ai.hitPosition.y = reader.ReadInt16()

			ai.previousMobPosition.x = reader.ReadInt16()
			ai.previousMobPosition.y = reader.ReadInt16()

			if data.isMesoExplosion {
				ai.hitCount = reader.ReadByte()
				ai.hitDelay = 0
			} else {
				ai.hitDelay = reader.ReadInt16()
				ai.hitCount = data.hits
			}

			if data.isMesoExplosion && ai.hitCount == 0 {
				rest := reader.GetRestAsBytes()
				peek := 24
				if len(rest) < peek {
					peek = len(rest)
				}
			}

			ai.damages = make([]int32, ai.hitCount)
			for j := byte(0); j < ai.hitCount; j++ {
				dmg := reader.ReadInt32()
				data.totalDamage += dmg
				ai.damages[j] = dmg
			}
			data.attackInfo[i] = ai
		}

		data.playerPos.x = reader.ReadInt16()
		data.playerPos.y = reader.ReadInt16()

		if data.isMesoExplosion {
			count := int(reader.ReadByte())
			for i := 0; i < count; i++ {
				objID := reader.ReadInt32()
				targetMask := reader.ReadByte()

				for tIdx := 0; tIdx < int(data.targets); tIdx++ {
					if (targetMask & (1 << uint(tIdx))) != 0 {
						data.attackInfo[tIdx].mesoDropIDs = append(data.attackInfo[tIdx].mesoDropIDs, objID)
					}
				}
			}
			data.mesoKillDelay = reader.ReadInt16()
		}

		return data, true
	}

	// Summon attack parsing (server-specific layout)
	data.summonType = reader.ReadInt32()
	stance := reader.ReadByte()
	extra := reader.ReadByte()
	data.action = stance
	data.hits = 1

	spawnID := int32(extra)

	rest := reader.GetRestAsBytes()
	delayIdx := -1
	for idx := 0; idx+5 < len(rest); idx++ {
		if rest[idx] == 0x64 && rest[idx+1] == 0x00 {
			delayIdx = idx
			break
		}
	}
	if delayIdx == -1 || delayIdx+6 > len(rest) {
		return data, false
	}

	dmg := int32(rest[delayIdx+2]) |
		int32(rest[delayIdx+3])<<8 |
		int32(rest[delayIdx+4])<<16 |
		int32(rest[delayIdx+5])<<24

	data.attackInfo = []attackInfo{{
		spawnID: spawnID,
		damages: []int32{dmg},
	}}
	data.targets = 1
	data.totalDamage = dmg

	return data, true
}

func (server *Server) npcMovement(conn mnet.Client, reader mpacket.Reader) {
	data := reader.GetRestAsBytes()
	id := reader.ReadInt32()

	plr, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)

	if err != nil {
		return
	}

	inst.lifePool.npcAcknowledge(id, plr, data)
}

func (server *Server) npcChatStart(conn mnet.Client, reader mpacket.Reader) {
	npcSpawnID := reader.ReadInt32()

	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	field, ok := server.fields[plr.mapID]
	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)
	if err != nil {
		return
	}

	npcData, err := inst.lifePool.getNPCFromSpawnID(npcSpawnID)
	if err != nil {
		return
	}

	// Start npc session
	var controller *npcChatController

	if program, ok := server.npcScriptStore.scripts[strconv.Itoa(int(npcData.id))]; ok {
		controller, err = createNpcChatController(npcData.id, conn, program, plr, server.fields, server.warpPlayer, server.world)
	} else {
		if program, ok := server.npcScriptStore.scripts["default"]; ok {
			controller, err = createNpcChatController(npcData.id, conn, program, plr, server.fields, server.warpPlayer, server.world)
		}
	}

	if controller == nil {
		log.Println("Unable to find npc script for:", npcData.id, ".... default.js not found")
		return
	}
	if err != nil {
		log.Println("script init:", err)
	}

	server.npcChat[conn] = controller
	server.updateNPCInteractionMetric(1)

	// Run the script. If it returns true, chat flow ended.
	if ended := controller.run(); ended {
		delete(server.npcChat, conn)
		server.updateNPCInteractionMetric(-1)
	}
}

func (server *Server) npcChatContinue(conn mnet.Client, reader mpacket.Reader) {
	if _, ok := server.npcChat[conn]; !ok {
		return
	}

	controller := server.npcChat[conn]
	controller.clearUserInput()

	terminate := false

	msgType := reader.ReadByte()

	switch msgType {
	case 0: // next/back
		value := reader.ReadByte()

		switch value {
		case 0: // back
			controller.stateTracker.popState()
		case 1: // next
			controller.stateTracker.addState(npcNextState)
		case 255: // 255/0xff end chat
			terminate = true
		default:
			terminate = true
			log.Println("unknown next/back:", value)
		}
	case 1: // yes/no, ok
		value := reader.ReadByte()

		switch value {
		case 0: // no
			controller.stateTracker.addState(npcNoState)
		case 1: // yes, ok
			controller.stateTracker.addState(npcYesState)
		case 255: // 255/0xff end chat
			terminate = true
		default:
			log.Println("unknown yes/no:", value)
		}
	case 2: // string input
		if reader.ReadBool() {
			controller.stateTracker.addState(npcStringInputState)
			controller.stateTracker.inputs = append(controller.stateTracker.inputs, reader.ReadString(reader.ReadInt16()))
		} else {
			terminate = true
		}
	case 3: // number input
		if reader.ReadBool() {
			controller.stateTracker.addState(npcNumberInputState)
			controller.stateTracker.numbers = append(controller.stateTracker.numbers, reader.ReadInt32())
		} else {
			terminate = true
		}
	case 4: // select option
		if reader.ReadBool() {
			controller.stateTracker.addState(npcSelectionState)
			controller.stateTracker.selections = append(controller.stateTracker.selections, reader.ReadInt32())
		} else {
			terminate = true
		}
	case 5: // style window (no way to discern between cancel button and end chat selection)
		if reader.ReadBool() {
			controller.stateTracker.addState(npcSelectionState)
			controller.stateTracker.selections = append(controller.stateTracker.selections, int32(reader.ReadByte()))
		} else {
			terminate = true
		}
	case 6:
		fmt.Println("npc pet window:", reader)
	default:
		log.Println("Unkown npc chat continue packet:", reader)
	}

	if terminate {
		delete(server.npcChat, conn)
		server.updateNPCInteractionMetric(-1)
	} else if controller.run() {
		delete(server.npcChat, conn)
		server.updateNPCInteractionMetric(-1)
	}
}

func (server *Server) npcShop(conn mnet.Client, reader mpacket.Reader) {
	getInventoryID := func(id int32) byte {
		return byte(id / 1000000)
	}
	isRechargeable := func(itemID int32) bool {
		base := itemID / 10000
		return base == 207 || base == 233
	}

	// ShopRes codes (aligned with client-side enum)
	const (
		shopBuySuccess byte = iota
		shopBuyNoStock
		shopBuyNoMoney
		shopBuyUnknown
		shopSellSuccess
		shopSellNoStock
		shopSellIncorrectRequest
		shopSellUnknown
		shopRechargeSuccess
		shopRechargeNoStock
		shopRechargeNoMoney
		shopRechargeIncorrectRequest
		shopRechargeUnknown
	)

	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	switch reader.ReadByte() {
	case 0: // Buy
		index := reader.ReadInt16()
		itemID := reader.ReadInt32()
		amount := reader.ReadInt16()

		if amount < 1 {
			plr.Send(packetNpcShopResult(shopBuyUnknown))
			return
		}
		controller, ok := server.npcChat[conn]
		if !ok || controller == nil {
			plr.Send(packetNpcShopResult(shopBuyUnknown))
			return
		}
		goods := controller.goods
		if int(index) < 0 || int(index) >= len(goods) {
			plr.Send(packetNpcShopResult(shopBuyUnknown))
			return
		}
		entry := goods[index]
		if len(entry) < 1 || entry[0] != itemID {
			plr.Send(packetNpcShopResult(shopBuyUnknown))
			return
		}

		meta, nxErr := nx.GetItem(itemID)
		if nxErr != nil {
			plr.Send(packetNpcShopResult(shopBuyUnknown))
			return
		}

		var price int32
		if len(entry) >= 2 {
			price = entry[1]
		} else {
			price = meta.Price
		}

		if isRechargeable(itemID) {
			if amount != 1 || price == 0 {
				plr.Send(packetNpcShopResult(shopBuyUnknown))
				return
			}
		}
		if meta.InvTabID == 1 && amount != 1 {
			plr.Send(packetNpcShopResult(shopBuyUnknown))
			return
		}

		totalCost := int64(price) * int64(amount)
		if totalCost < 0 || int64(plr.mesos) < totalCost {
			plr.Send(packetNpcShopResult(shopBuyNoMoney))
			return
		}

		newItem, mkErr := createAverageItemFromID(itemID, amount)
		if mkErr != nil {
			plr.Send(packetNpcShopResult(shopBuyUnknown))
			return
		}
		if err, _ := plr.GiveItem(newItem); err != nil {
			plr.Send(packetNpcShopResult(shopBuyUnknown))
			return
		}

		plr.giveMesos(int32(-totalCost))
		plr.Send(packetNpcShopResult(shopBuySuccess))

	case 1: // Sell
		slotPos := reader.ReadInt16()
		itemID := reader.ReadInt32()
		amount := reader.ReadInt16()
		if amount < 1 {
			plr.Send(packetNpcShopResult(shopSellIncorrectRequest))
			return
		}

		meta, nxErr := nx.GetItem(itemID)
		if nxErr != nil {
			plr.Send(packetNpcShopResult(shopSellUnknown))
			return
		}

		invID := getInventoryID(itemID)
		sellAmount := amount
		if isRechargeable(itemID) {
			useItem := plr.findUseItemBySlot(slotPos)
			if useItem == nil || useItem.ID != itemID || useItem.amount <= 0 {
				plr.Send(packetNpcShopResult(shopSellIncorrectRequest))
				return
			}
			sellAmount = useItem.amount
		}

		if _, remErr := plr.takeItem(itemID, slotPos, sellAmount, invID); remErr != nil {
			plr.Send(packetNpcShopResult(shopSellIncorrectRequest))
			return
		}

		var payout int64
		if meta.InvTabID == 1 {
			payout = int64(meta.Price)
		} else {
			payout = int64(meta.Price) * int64(sellAmount)
		}
		if payout < 0 {
			plr.Send(packetNpcShopResult(shopSellIncorrectRequest))
			return
		}

		plr.giveMesos(int32(payout))
		plr.Send(packetNpcShopResult(shopSellSuccess))

	case 2: // Recharge
		slotPos := reader.ReadInt16()

		it := plr.findUseItemBySlot(slotPos)
		if it == nil || !isRechargeable(it.ID) {
			plr.Send(packetNpcShopResult(shopRechargeIncorrectRequest))
			return
		}

		controller, ok := server.npcChat[conn]
		if !ok || controller == nil {
			plr.Send(packetNpcShopResult(shopRechargeUnknown))
			return
		}
		found := false
		for _, g := range controller.goods {
			if len(g) > 0 && g[0] == it.ID {
				found = true
				break
			}
		}
		if !found {
			plr.Send(packetNpcShopResult(shopRechargeIncorrectRequest))
			return
		}

		meta, nxErr := nx.GetItem(it.ID)
		if nxErr != nil {
			plr.Send(packetNpcShopResult(shopRechargeUnknown))
			return
		}

		slotMax := meta.SlotMax
		if slotMax <= 0 || it.amount < 0 || it.amount >= slotMax {
			plr.Send(packetNpcShopResult(shopRechargeIncorrectRequest))
			return
		}

		toFill := int(slotMax - it.amount)
		unitPrice := meta.UnitPrice
		if unitPrice <= 0 {
			plr.Send(packetNpcShopResult(shopRechargeIncorrectRequest))
			return
		}

		cost := int(math.Ceil(unitPrice * float64(toFill)))
		if cost < 0 || int(plr.mesos) < cost {
			plr.Send(packetNpcShopResult(shopRechargeNoMoney))
			return
		}

		it.amount = slotMax
		plr.updateItemStack(*it)
		plr.giveMesos(int32(-cost))
		plr.Send(packetNpcShopResult(shopRechargeSuccess))

	case 3: // Close
		if _, ok := server.npcChat[conn]; ok {
			delete(server.npcChat, conn)
			server.updateNPCInteractionMetric(-1)
		}
	}
}

func (server Server) roomWindow(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)
	pool := inst.roomPool

	if err != nil {
		return
	}

	operation := reader.ReadByte()

	switch operation {
	case constant.MiniRoomCreate:
		switch roomType := reader.ReadByte(); roomType {
		case constant.MiniRoomTypeOmok:
			name := reader.ReadString(reader.ReadInt16())

			var password string
			if reader.ReadBool() {
				password = reader.ReadString(reader.ReadInt16())
			}

			boardType := reader.ReadByte()

			r, valid := newOmokRoom(inst.nextID(), name, password, boardType).(roomer)

			if !valid {
				return
			}

			if r.addPlayer(plr) {
				err = pool.addRoom(r)

				if err != nil {
					log.Println(err)
				}
			}
		case constant.MiniRoomTypeMatchCards:
			name := reader.ReadString(reader.ReadInt16())

			var password string
			if reader.ReadBool() {
				password = reader.ReadString(reader.ReadInt16())
			}

			boardType := reader.ReadByte()

			r, valid := newMemoryRoom(inst.nextID(), name, password, boardType).(roomer)

			if !valid {
				return
			}

			if r.addPlayer(plr) {
				err = pool.addRoom(r)

				if err != nil {
					log.Println(err)
				}
			}
		case constant.MiniRoomTypeTrade:
			r, valid := newTradeRoom(inst.nextID()).(roomer)

			if !valid {
				return
			}

			if r.addPlayer(plr) {
				err = pool.addRoom(r)

				if err != nil {
					log.Println(err)
				}
			}
		case constant.MiniRoomTypePlayerShop:
			log.Println("Personal shop not implemented")
		default:
			log.Println("Unknown room type", roomType)
		}
	case constant.MiniRoomInvite:
		id := reader.ReadInt32()

		plr2, err := inst.getPlayerFromID(id)

		if err != nil {
			plr.Send(packetRoomTradeRequireSameMap())
			return
		}

		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if trade, valid := r.(*tradeRoom); valid {
			trade.sendInvite(plr2)
		}
	case constant.MiniRoomDeclineInvite:
		id := reader.ReadInt32()
		code := reader.ReadByte()

		r, err := pool.getRoom(id)

		if err != nil {
			return
		}

		if trade, valid := r.(*tradeRoom); valid {
			trade.reject(code, plr.Name)
		}
	case constant.MiniRoomEnter:
		id := reader.ReadInt32()

		r, err := pool.getRoom(id)

		if err != nil {
			plr.Send(packetRoomTradeRequireSameMap())
			return
		}

		if reader.ReadBool() {
			password := reader.ReadString(reader.ReadInt16())

			if game, valid := r.(gameRoomer); valid {
				if !game.checkPassword(password, plr) {
					return
				}
			}
		}

		r.addPlayer(plr)

		if _, valid := r.(gameRoomer); valid {
			pool.updateGameBox(r)
		}
	case constant.MiniRoomChat:
		msg := reader.ReadString(reader.ReadInt16())

		if len(msg) > 0 {
			r, err := pool.getPlayerRoom(plr.ID)

			if err != nil {
				return
			}

			r.chatMsg(plr, msg)
		}
	case constant.MiniRoomLeave:
		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.kickPlayer(plr, 0x0)

			if r.closed() {
				err = pool.removeRoom(r.id())

				if err != nil {
					log.Println(err)
				}
			} else {
				pool.updateGameBox(r)
			}
		} else if trade, valid := r.(*tradeRoom); valid {
			trade.removePlayer(plr)
			err = pool.removeRoom(trade.roomID)

			if err != nil {
				log.Println(err)
			}
		}
	case constant.MiniRoomTradePutItem:
		invType := reader.ReadByte()
		invSlot := reader.ReadInt16()
		amount := reader.ReadInt16()
		tradeSlot := reader.ReadByte()

		item, err := plr.getItem(invType, invSlot)
		if err != nil || item.amount < amount {
			plr.Send(packetPlayerNoChange())
			return
		}

		if item.isRechargeable() {
			amount = item.amount
		}

		_, err = plr.takeItem(item.ID, invSlot, amount, invType)
		if err != nil {
			plr.Send(packetPlayerNoChange())
			return
		}

		newItem := item
		newItem.amount = amount

		if r, err := pool.getPlayerRoom(plr.ID); err == nil {
			if tr, ok := r.(*tradeRoom); ok {
				tr.insertItem(tradeSlot, plr.ID, newItem)
			}
		}

		plr.Send(packetPlayerNoChange())

	case constant.MiniRoomTradePutMesos:
		amount := reader.ReadInt32()

		if plr.mesos < amount {
			plr.Send(packetPlayerNoChange())
			return
		}

		plr.takeMesos(amount)

		if r, err := pool.getPlayerRoom(plr.ID); err == nil {
			if tr, ok := r.(*tradeRoom); ok {
				tr.updateMesos(amount, plr.ID)
			}
		}

	case constant.MiniRoomTradeAccept:
		if r, err := pool.getPlayerRoom(plr.ID); err == nil {
			if tr, ok := r.(*tradeRoom); ok {
				if tr.acceptTrade(plr) {
					err = pool.removeRoom(tr.roomID)

					if err != nil {
						log.Println(err)
					}
				}
			}
		}

	case constant.RoomRequestTie:
		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.requestTie(plr)
		}
	case constant.RoomRequestTieResult:
		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			tie := reader.ReadBool()
			game.requestTieResult(tie, plr)

			if r.closed() {
				err = pool.removeRoom(r.id())

				if err != nil {
					log.Println(err)
				}
			} else {
				pool.updateGameBox(r)
			}
		}
	case constant.RoomForfeit:
		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.forfeit(plr)

			if r.closed() {
				err = pool.removeRoom(r.id())

				if err != nil {
					log.Println(err)
				}
			} else {
				pool.updateGameBox(r)
			}
		}
	case constant.RoomRequestUndo:
		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if game, valid := r.(*omokRoom); valid {
			game.requestUndo(plr)
		}
	case constant.RoomRequestUndoResult:
		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if game, valid := r.(*omokRoom); valid {
			undo := reader.ReadBool()
			game.requestUndoResult(undo, plr)
		}
	case constant.RoomRequestExitDuringGame:
		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.requestExit(true, plr)
		}
	case constant.RoomUndoRequestExit:
		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.requestExit(false, plr)
		}
	case constant.RoomReadyButtonPressed:
		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.ready(plr)
		}
	case constant.RoomUnready:
		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.unready(plr)
		}
	case constant.RoomOwnerExpell:
		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.expel()
			pool.updateGameBox(r)
		}
	case constant.RoomGameStart:
		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.start()
			pool.updateGameBox(r)
		}
	case constant.RoomChangeTurn:
		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if game, valid := r.(gameRoomer); valid {
			game.changeTurn()
		}
	case constant.RoomPlacePiece:
		x := reader.ReadInt32()
		y := reader.ReadInt32()
		piece := reader.ReadByte()

		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if game, valid := r.(*omokRoom); valid {
			if game.placePiece(x, y, piece, plr) {
				pool.updateGameBox(r)
			}

			if r.closed() {
				err = pool.removeRoom(game.roomID)

				if err != nil {
					log.Println(err)
				}
			}
		}
	case constant.RoomSelectCard:
		turn := reader.ReadByte()
		cardID := reader.ReadByte()

		r, err := pool.getPlayerRoom(plr.ID)

		if err != nil {
			return
		}

		if game, valid := r.(*memoryRoom); valid {
			if game.selectCard(turn, cardID, plr) {
				pool.updateGameBox(r)
			}

			if r.closed() {
				err = pool.removeRoom(game.roomID)

				if err != nil {
					log.Println(err)
				}
			}
		}
	default:
		log.Println("Unknown room operation", operation)
	}
}

func (server *Server) guildManagement(conn mnet.Client, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case constant.GuildCreateDialogue:
		guildName := reader.ReadString(reader.ReadInt16())

		if len(guildName) < 4 || len(guildName) > 12 {
			conn.Send(packetGuildProblemOccurred())
			return
		}

		guildCount := 0
		err := common.DB.QueryRow("SELECT count(*) FROM guilds where Name=? AND worldID=?", guildName, conn.GetWorldID()).Scan(&guildCount)

		if err != nil {
			log.Fatal(err)
		}

		if guildCount > 0 {
			conn.Send(packetGuildNameInUse())
			return
		}

		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		if plr.party == nil {
			return
		}

		guild := createGuildContract(guildName, int32(plr.worldID), &server.players, plr)

		if guild == nil {
			return
		}

		server.guilds[guild.id] = guild
	case constant.GuildInvite:
		invitee := reader.ReadString(reader.ReadInt16())

		var playerID int32
		var guildID sql.NullInt32
		var worldID byte
		var channelID int8

		query := "SELECT ID, guildID, worldID, channelID FROM characters WHERE Name=?"
		row, err := common.DB.Query(query, invitee)

		if err != nil {
			log.Fatal(err)
		}

		defer row.Close()

		for row.Next() {
			row.Scan(&playerID, &guildID, &worldID, &channelID)
		}

		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		if plr.guild == nil {
			return // cannot invite someone if not in guild
		}

		if guildID.Valid {
			plr.Send(packetGuildAlreadyJoined())
			return
		}

		count := 0
		query = "SELECT count(*) FROM guild_invites WHERE playerID=?"
		err = common.DB.QueryRow(query, playerID).Scan(&count)

		if err != nil {
			log.Fatal(err)
		}

		if count != 0 {
			plr.Send(packetGuildInviteeHasAnother(invitee))
			return
		}

		if worldID != plr.worldID {
			plr.Send(packetMessageRedText("Could not find Player"))
			return
		}

		query = "INSERT INTO guild_invites (playerID, guildID, inviter) VALUES (?, ?, ?)"
		_, err = common.DB.Exec(query, playerID, plr.guild.id, plr.Name)

		if err != nil {
			log.Fatal(err)
		}

		server.world.Send(internal.PacketGuildInvite(plr.guild.id, plr.Name, invitee))
	case constant.GuildAcceptInvite:
		guildID := reader.ReadInt32()
		playerID := reader.ReadInt32()

		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		if plr.ID != playerID {
			return // cannot join the guild on someone else's behalf
		}

		query := "DELETE FROM guild_invites WHERE playerID=? AND guildID=?"

		if _, err := common.DB.Exec(query, playerID, guildID); err != nil {
			log.Fatal(err)
		}

		server.world.Send(internal.PacketGuildInviteAccept(playerID, guildID, plr.Name, int32(plr.job), int32(plr.level), true, 5))
	case constant.GuildLeave:
		playerID := reader.ReadInt32()
		name := reader.ReadString(reader.ReadInt16())

		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		if plr.guild.isMaster(plr) {
			server.world.Send(internal.PacketGuildDisband(plr.guild.id))
		} else {
			server.world.Send(internal.PacketGuildRemovePlayer(plr.guild.id, playerID, name, false))
		}
	case constant.GuildExpel:
		playerID := reader.ReadInt32()
		name := reader.ReadString(reader.ReadInt16())

		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		server.world.Send(internal.PacketGuildRemovePlayer(plr.guild.id, playerID, name, true))
	case constant.GuildNoticeChange:
		notice := reader.ReadString(reader.ReadInt16())
		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		server.world.Send(internal.PacketGuildUpdateNotice(plr.guild.id, notice))
	case constant.GuildUpdateTitleNames:
		master := reader.ReadString(reader.ReadInt16())
		jrMaster := reader.ReadString(reader.ReadInt16())
		member1 := reader.ReadString(reader.ReadInt16())
		member2 := reader.ReadString(reader.ReadInt16())
		member3 := reader.ReadString(reader.ReadInt16())

		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		if plr.guild == nil {
			return
		}

		server.world.Send(internal.PacketGuildTitlesChange(plr.guild.id, master, jrMaster, member1, member2, member3))
	case constant.GuildRankChange:
		playerID := reader.ReadInt32()
		rank := reader.ReadByte()

		if rank < 1 || rank > 5 {
			return
		}

		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		if plr.guild != nil {
			server.world.Send(internal.PacketGuildRankUpdate(plr.guild.id, playerID, rank))
		}
	case constant.GuildEmblemChange:
		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		if plr.guild == nil || !plr.guild.isMaster(plr) {
			return
		}

		logoBg := reader.ReadInt16()
		logoBgColour := reader.ReadByte()
		logo := reader.ReadInt16()
		logoColour := reader.ReadByte()

		plr.giveMesos(-1e6)

		server.world.Send(internal.PacketGuildUpdateEmblem(plr.guild.id, logoBg, logo, logoBgColour, logoColour))
	case constant.GuildContractSign:
		id := reader.ReadInt32()
		accepted := reader.ReadBool()

		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			return
		}

		if plr.guild == nil {
			return
		}

		if accepted {
			success := plr.guild.signContract(id)

			if success {
				server.guilds[plr.guild.id] = plr.guild
			}
		} else {
			plr, err := plr.guild.players.GetFromID(plr.guild.playerID[0]) // master will always be the first Player when creating a guild

			if err != nil {
				return
			}

			guildID := plr.guild.id

			for _, id := range plr.guild.playerID {
				member, err := server.players.GetFromID(id)

				if err != nil {
					continue
				}

				member.guild = nil
			}

			plr.Send(packetGuildContractDisagree())
			delete(server.guilds, guildID)
		}
	default:
		log.Println("Unknown guild operation", op, reader)
	}
}

func (server *Server) playerUseSack(conn mnet.Client, reader mpacket.Reader) {
	slot := reader.ReadInt16()
	itemID := reader.ReadInt32()

	plr, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	sack, err := plr.takeItem(itemID, slot, 1, 2)

	if err != nil {
		log.Println("Could not find sack")
		return
	}

	if len(sack.spawnMobs) > 0 {
		p := int32(plr.randIntn(100))

		for mobID, prob := range sack.spawnMobs {
			if prob >= p {
				summonType := constant.MobSummonTypePoof

				switch mobID {
				case constant.MobBalrog:
					summonType = constant.MobSummonTypeJrBalrog
				}

				plr.inst.lifePool.spawnMobFromID(mobID, plr.pos, false, true, true, summonType, plr.ID)
			}
		}
	}
}

func getAffectedPartyMembers(p *party, src *Player, affected byte) []*Player {
	if p == nil || src == nil {
		return nil
	}

	var total byte
	for i := 0; i < constant.MaxPartySize; i++ {
		if p.players[i] != nil {
			total++
		}
	}

	ret := make([]*Player, 0, constant.MaxPartySize)

	for i := 0; i < constant.MaxPartySize; i++ {
		idx := i + 1
		mask := partyMemberMaskForIndex(idx, total)
		if (affected & mask) == 0 {
			continue
		}

		member := p.players[i]
		if member == nil {
			continue
		}

		// Must be same map and same instance
		if member.mapID != src.mapID {
			continue
		}
		if member.inst == nil || src.inst == nil || member.inst.id != src.inst.id {
			continue
		}

		// Exclude self
		if member.ID == src.ID {
			continue
		}

		ret = append(ret, member)
	}

	return ret
}

func (server *Server) playerSpecialSkill(conn mnet.Client, reader mpacket.Reader) {
	// Minimal, safe implementation to keep packet stream in sync and apply basic validations/costs.
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	// Dead players cannot cast
	if plr.hp == 0 {
		plr.Send(packetPlayerNoChange())
		return
	}

	// The packet layout for special skills generally starts with [skillID int32][skillLevel byte]
	skillID := reader.ReadInt32()
	skillLevel := reader.ReadByte()

	// Validate the Player owns the skill and level does not exceed learned level
	ps, ok := plr.skills[skillID]
	if !ok || skillLevel == 0 || skillLevel > ps.Level {
		// Possible hack/desync; drop request
		plr.Send(packetPlayerNoChange())
		return
	}

	partyMask := reader.ReadByte() // party flags
	delay := reader.ReadInt16()    // delay

	switch skill.Skill(skillID) {
	case skill.Haste, skill.BanditHaste, skill.Bless, skill.IronWill, skill.Rage,
		skill.Meditation, skill.ILMeditation, skill.MesoUp, skill.HolySymbol, skill.HyperBody, skill.NimbleBody:
		plr.addBuff(skillID, skillLevel, delay)
		plr.inst.send(packetPlayerSkillAnimation(plr.ID, false, skillID, skillLevel))

		// Apply to eligible party members in same map/instance per mask
		if plr.party != nil {
			affected := getAffectedPartyMembers(plr.party, plr, partyMask)
			for _, member := range affected {
				if member == nil {
					continue
				}

				member.addForeignBuff(member.ID, skillID, skillLevel, delay)
				member.Send(packetPlayerEffectSkill(true, skillID, skillLevel))
				member.inst.send(packetPlayerSkillAnimation(member.ID, true, skillID, skillLevel))
			}
		}
	case skill.SuperGMHaste, skill.SuperGMBless, skill.SuperGMHolySymbol:
		plr.addBuff(skillID, skillLevel, delay)
		plr.inst.send(packetPlayerSkillAnimation(plr.ID, false, skillID, skillLevel))

		// Apply buff to the entire map
		for _, member := range plr.inst.players {
			if member.ID == plr.ID {
				continue
			}
			member.addForeignBuff(member.ID, skillID, skillLevel, delay)
			member.Send(packetPlayerEffectSkill(true, skillID, skillLevel))
			member.inst.send(packetPlayerSkillAnimation(member.ID, true, skillID, skillLevel))
		}
	case skill.SuperGMResurrection:
		plr.inst.send(packetPlayerSkillAnimation(plr.ID, false, skillID, skillLevel))

		for _, member := range plr.inst.players {
			if member.ID == plr.ID {
				continue
			}
			if member.hp == 0 {
				member.Send(packetPlayerEffectSkill(true, skillID, skillLevel))
				member.inst.send(packetPlayerSkillAnimation(member.ID, true, skillID, skillLevel))
				member.setHP(member.maxHP)
			}
		}

	case skill.SuperGMHealDispell:
		plr.inst.send(packetPlayerSkillAnimation(plr.ID, false, skillID, skillLevel))

		for _, member := range plr.inst.players {
			if member.ID == plr.ID {
				continue
			}
			member.Send(packetPlayerEffectSkill(true, skillID, skillLevel))
			member.inst.send(packetPlayerSkillAnimation(member.ID, true, skillID, skillLevel))
			member.setHP(member.maxHP)
			member.setMP(member.maxMP)
			member.buffs.cureAllDebuffs()
		}

	case skill.Dispel:
		plr.inst.send(packetPlayerSkillAnimation(plr.ID, false, skillID, skillLevel))

		if plr.party != nil {
			affected := getAffectedPartyMembers(plr.party, plr, partyMask)
			for _, member := range affected {
				if member == nil {
					continue
				}

				member.Send(packetPlayerEffectSkill(true, skillID, skillLevel))
				member.inst.send(packetPlayerSkillAnimation(member.ID, true, skillID, skillLevel))
				member.buffs.cureAllDebuffs()
			}
		}

	case skill.Resurrection:
		plr.inst.send(packetPlayerSkillAnimation(plr.ID, false, skillID, skillLevel))

		if plr.party != nil {
			affected := getAffectedPartyMembers(plr.party, plr, partyMask)
			for _, member := range affected {
				if member == nil {
					continue
				}

				if member.hp == 0 {
					member.Send(packetPlayerEffectSkill(true, skillID, skillLevel))
					member.inst.send(packetPlayerSkillAnimation(member.ID, true, skillID, skillLevel))
					member.setHP(member.maxHP)
				}
			}
		}

	// Nimble feet and recovery beginner skills with cooldown
	case skill.NimbleFeet, skill.Recovery:
		plr.addBuff(skillID, skillLevel, delay)
		plr.inst.send(packetPlayerSkillAnimation(plr.ID, true, skillID, skillLevel))
		// Send cooldown packet for beginner skills
		if skillData, ok := plr.skills[skillID]; ok {
			plr.Send(packetPlayerSkillCooldown(skillID, skillData.CooldownTime))
			// Start timer to clear the cooldown when it expires
			go func(sid int32, cooldownTime int16, inst *fieldInstance) {
				time.Sleep(time.Duration(cooldownTime) * time.Second)
				// Use dispatch pattern for thread safety
				if inst != nil && inst.dispatch != nil && plr != nil {
					inst.dispatch <- func() {
						plr.Send(packetPlayerSkillCooldown(sid, 0))
					}
				}
			}(skillID, skillData.CooldownTime, plr.inst)
		}

	case skill.DarkSight,
		skill.MagicGuard,
		skill.Invincible,
		skill.SoulArrow, skill.CBSoulArrow,
		skill.ShadowPartner,
		skill.MesoGuard,
		skill.SwordBooster, skill.AxeBooster, skill.PageSwordBooster, skill.BwBooster,
		skill.SpearBooster, skill.PolearmBooster,
		skill.BowBooster, skill.CrossbowBooster,
		skill.ClawBooster, skill.DaggerBooster,
		skill.SpellBooster, skill.ILSpellBooster,
		// GM Hide (mapped to invincible bit)
		skill.SuperGMHide:
		plr.addBuff(skillID, skillLevel, delay)
		plr.inst.send(packetPlayerSkillAnimation(plr.ID, true, skillID, skillLevel))

	case skill.Threaten,
		skill.Slow, skill.ILSlow,
		skill.MagicCrash,
		skill.PowerCrash,
		skill.ArmorCrash,
		skill.ILSeal, skill.Seal,
		skill.ShadowWeb,
		skill.Doom:

		_ = reader.ReadInt16() // padding

		mobIDs := make([]int32, 0)
		restBytes := reader.GetRestAsBytes()
		for i := 0; i+4 <= len(restBytes); i += 4 {
			mobID := int32(restBytes[i]) | int32(restBytes[i+1])<<8 | int32(restBytes[i+2])<<16 | int32(restBytes[i+3])<<24
			if mobID == 0 || mobID < 0 {
				break
			}
			mobIDs = append(mobIDs, mobID)
		}

		plr.inst.send(packetPlayerSkillAnimation(plr.ID, false, skillID, skillLevel))

		var statMask int32
		switch skill.Skill(skillID) {
		case skill.Threaten:
			statMask = skill.MobStat.PhysicalDefense | skill.MobStat.MagicDefense
		case skill.ArmorCrash:
			statMask = skill.MobStat.PhysicalDefense
		case skill.PowerCrash:
			statMask = skill.MobStat.PhysicalDamage
		case skill.MagicCrash:
			statMask = skill.MobStat.MagicDefense
		case skill.Slow, skill.ILSlow:
			statMask = skill.MobStat.Speed
		case skill.Seal, skill.ILSeal:
			statMask = skill.MobStat.Seal
		case skill.ShadowWeb:
			statMask = skill.MobStat.Web
		case skill.Doom:
			statMask = skill.MobStat.Doom
		}

		if plr.inst != nil {
			for _, mobID := range mobIDs {
				plr.inst.lifePool.applyMobBuff(mobID, skillID, skillLevel, statMask, plr.inst)
			}
		}

	case skill.MysticDoor:
		createMysticDoor(plr, skillID, skillLevel)

	case skill.SummonDragon,
		skill.SilverHawk, skill.GoldenEagle,
		skill.Puppet, skill.SniperPuppet:

		isPuppet := (skill.Skill(skillID) == skill.Puppet || skill.Skill(skillID) == skill.SniperPuppet)

		spawn := plr.pos

		if isPuppet {
			desiredX := plr.pos.x
			if (plr.stance & 0x01) == 0 {
				desiredX += 200 // facing right
			} else {
				desiredX -= 200 // facing left
			}
			if fld, ok := server.fields[plr.mapID]; ok {
				if inst, err := fld.getInstance(plr.inst.id); err == nil {
					snapped := inst.fhHist.getFinalPosition(newPos(desiredX, plr.pos.y, 0))
					spawn = snapped
				}
			}
		}

		summ := &summon{
			OwnerID:    plr.ID,
			SkillID:    skillID,
			Level:      skillLevel,
			Pos:        spawn,
			Stance:     0,
			Foothold:   spawn.foothold,
			IsPuppet:   isPuppet,
			SummonType: 0,
		}

		if isPuppet {
			if data, err := nx.GetPlayerSkill(skillID); err == nil {
				idx := int(skillLevel) - 1
				if idx >= 0 && idx < len(data) {
					summ.HP = int(data[idx].X)
				}
			}
		}

		plr.addSummon(summ)
		plr.addBuff(skillID, skillLevel, delay)
		plr.inst.send(packetPlayerSkillAnimation(plr.ID, false, skillID, skillLevel))

	default:
		// Always Send a self animation so client shows casting even for non-buffs.
		plr.addBuff(skillID, skillLevel, delay)
		plr.inst.send(packetPlayerSkillAnimation(plr.ID, false, skillID, skillLevel))
	}

	// Apply MP cost/cooldown, if any (reuses the same flow as attack skills).
	plr.useSkill(skillID, skillLevel, 0)
	plr.Send(packetPlayerNoChange()) // catch all for things like GM teleport
}

func (server *Server) playerStopSkill(conn mnet.Client, reader mpacket.Reader) {
	skillID := reader.ReadInt32()

	plr, err := server.players.GetFromConn(conn)
	if err != nil || plr.buffs == nil {
		return
	}

	cb := plr.buffs

	cb.expireBuffNow(skillID)
	cb.plr.inst = plr.inst
	cb.AuditAndExpireStaleBuffs()
}

func (server *Server) playerCancelBuff(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil || plr.buffs == nil {
		return
	}

	cb := plr.buffs

	const grace = 1500 * time.Millisecond
	now := time.Now().Add(grace).UnixMilli()

	// Collect all timed sources that should be expired by server clock
	toExpire := make([]int32, 0, len(cb.expireAt))
	for src, ts := range cb.expireAt {
		if ts > 0 && ts <= now {
			toExpire = append(toExpire, src)
		}
	}

	for _, src := range toExpire {
		cb.expireBuffNow(src)
	}

	// Final sweep for any edge cases
	cb.plr.inst = plr.inst
	cb.AuditAndExpireStaleBuffs()
}

func (server Server) playerSummonMove(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	summonID := reader.ReadInt32()
	summ := plr.getSummon(summonID)
	if summ == nil || summ.IsPuppet {
		return
	}

	moveData, finalData := parseMovement(reader)
	moveBytes := generateMovementBytes(moveData)

	summ.Pos = pos{x: finalData.x, y: finalData.y}
	summ.Stance = finalData.stance
	summ.Foothold = finalData.foothold

	field, ok := server.fields[plr.mapID]
	if !ok {
		return
	}
	inst, err := field.getInstance(plr.inst.id)
	if err != nil {
		return
	}

	inst.sendExcept(packetSummonMove(plr.ID, summonID, moveBytes), conn)
}

func (server *Server) playerSummonDamage(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	summonID := reader.ReadInt32()
	summ := plr.getSummon(summonID)
	if summ == nil || !summ.IsPuppet {
		return
	}

	_ = int8(reader.ReadByte())
	damage := reader.ReadInt32()
	mobID := reader.ReadInt32()
	_ = reader.ReadByte()

	field, ok := server.fields[plr.mapID]
	if ok {
		if inst, e := field.getInstance(plr.inst.id); e == nil {
			inst.send(packetSummonDamage(plr.ID, summonID, damage, mobID))
		}
	}

	if summ.HP-int(damage) < 0 {
		plr.buffs.expireBuffNow(summ.SkillID)
	} else {
		summ.HP -= int(damage)
	}
}

func (server *Server) playerSummonAttack(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	data, valid := getAttackInfo(reader, *plr, attackSummon)
	if !valid || len(data.attackInfo) == 0 {
		return
	}

	field, ok := server.fields[plr.mapID]
	if !ok {
		return
	}
	inst, err := field.getInstance(plr.inst.id)
	if err != nil {
		return
	}

	mobDamages := make(map[int32][]int32, len(data.attackInfo))
	for _, at := range data.attackInfo {
		if at.spawnID <= 0 || len(at.damages) == 0 {
			continue
		}
		mobDamages[at.spawnID] = append(mobDamages[at.spawnID], at.damages...)
	}
	if len(mobDamages) == 0 {
		return
	}

	anim := data.action
	if anim == 0 && len(data.attackInfo) > 0 {
		anim = data.attackInfo[0].frameIndex
	}

	inst.sendExcept(packetSummonAttack(plr.ID, data.summonType, anim, byte(len(mobDamages)), mobDamages), conn)
	for spawnID, damages := range mobDamages {
		for _, d := range damages {
			if d > 0 {
				inst.lifePool.mobDamaged(spawnID, plr, d)
			}
		}
	}
}

func (server *Server) playerQuestOperation(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	act := reader.ReadByte()
	questID := reader.ReadInt16()

	switch act {
	case constant.QuestStarted:
		if !plr.tryStartQuest(questID) {
			plr.Send(packetPlayerNoChange())
		}
	case constant.QuestCompleted:
		if !plr.tryCompleteQuest(questID) {
			plr.Send(packetPlayerNoChange())
		}
	case constant.QuestForfeit:
		plr.quests.remove(questID)
		deleteQuest(plr.ID, questID)
		clearQuestMobKills(plr.ID, questID)
		plr.Send(packetQuestRemove(questID))
	case constant.QuestLostItem:
		count := reader.ReadInt16()
		questItem := reader.ReadInt16()
		if count > 0 {
			if it, err := CreateItemFromID(int32(questItem), count); err == nil {
				_, _ = plr.GiveItem(it)
			} else {
				log.Printf("[QuestPkt] lostItem give failed: err=%v", err)
			}
		} else if count < 0 {
			if !plr.removeItemsByID(int32(questItem), int32(-count)) {
				log.Printf("[QuestPkt] lostItem remove failed: Item=%d need=%d", questItem, -count)
			}
		}

	default:
		log.Println("Unknown quest operation type:", act)
	}
}

func (server *Server) playerFame(conn mnet.Client, reader mpacket.Reader) {
	source, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	targetID := reader.ReadInt32()
	up := reader.ReadBool()

	if targetID == source.ID {
		return
	}

	target, err := server.players.GetFromID(targetID)
	if err != nil || target == nil || target.mapID != source.mapID {
		source.Send(packetFameError(constant.FameIncorrectUser))
		return
	}

	if source.level < 15 {
		source.Send(packetFameError(constant.FameUnderLevel))
		return
	}

	if fameHasRecentActivity(source.ID, 24*time.Hour) {
		source.Send(packetFameError(constant.FameThisDay))
		return
	}

	if fameHasRecentActivity(source.ID, 30*24*time.Hour) {
		source.Send(packetFameError(constant.FameThisMonth))
		return
	}

	delta := int16(1)
	if !up {
		delta = -1
	}
	target.setFame(target.fame + delta)

	if err := fameInsertLog(source.ID, target.ID); err != nil {
		log.Println("fameInsertLog:", err)
	}

	target.Send(packetFameNotifyVictim(source.Name, up))
	source.Send(packetFameNotifySource(target.Name, up, target.fame))
}

func (server *Server) playerHitReactor(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	spawnID := reader.ReadInt32()
	_ = reader.ReadInt32() // stance
	_ = reader.ReadInt16() // delay

	plr.inst.reactorPool.triggerHit(spawnID, 0, server, plr)

}

func (server *Server) playerUseStorage(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	const (
		actionWithdraw        byte = 4
		actionDeposit              = 5
		actionStoreMesos           = 6
		actionExit                 = 7
		storageInvFullOrNot        = 9
		encWithdraw                = 8
		encDeposit                 = 10
		storageNotEnoughMesos      = 12
		storageIsFull              = 13
		storageDueToAnError        = 14
		storageSuccess             = 15
	)

	accountID := conn.GetAccountID()
	if accountID == 0 {
		return
	}

	isRechargeable := func(itemID int32) bool {
		base := itemID / 10000
		return base == 207
	}

	switch reader.ReadByte() {
	case actionWithdraw:
		tab := reader.ReadByte()
		slot := reader.ReadByte()

		stIdx, it := plr.storageInventory.getBySectionSlot(tab, slot)
		if stIdx < 0 || it == nil || it.ID == 0 {
			plr.Send(packetNpcStorageResult(storageDueToAnError))
			return
		}

		out := *it
		if !out.isStackable() || out.amount <= 0 {
			out.amount = 1
		}
		out.dbID = 0
		out.slotID = 0

		if err, _ := plr.GiveItem(out); err != nil {
			plr.Send(packetNpcStorageResult(storageInvFullOrNot))
			return
		}

		plr.storageInventory.removeAt(byte(stIdx))
		if err := plr.storageInventory.save(plr.accountID); err != nil {
			plr.Send(packetNpcStorageResult(storageDueToAnError))
			return
		}

		sectionItems := plr.storageInventory.getItemsInSection(tab)
		plr.Send(packetNpcStorageItemsChanged(encWithdraw, plr.storageInventory.maxSlots, tab, 0, sectionItems))

	case actionDeposit:
		srcSlot := reader.ReadInt16()
		itemID := reader.ReadInt32()
		amt := reader.ReadInt16()
		if amt <= 0 {
			plr.Send(packetNpcStorageResult(storageDueToAnError))
			return
		}

		tab := byte(itemID / 1000000)
		itemOnChar, getErr := plr.getItem(tab, srcSlot)
		if getErr != nil || itemOnChar.ID != itemID {
			plr.Send(packetNpcStorageResult(storageDueToAnError))
			return
		}

		if !plr.storageInventory.slotsAvailable() {
			plr.Send(packetNpcStorageResult(storageIsFull))
			return
		}

		if isRechargeable(itemID) {
			amt = itemOnChar.amount
		} else if !itemOnChar.isStackable() {
			amt = 1
		} else if amt > itemOnChar.amount {
			amt = itemOnChar.amount
		}

		storeCopy := itemOnChar
		storeCopy.amount = amt
		storeCopy.dbID = 0
		storeCopy.slotID = 0

		if _, remErr := plr.takeItem(itemID, srcSlot, amt, tab); remErr != nil {
			plr.Send(packetNpcStorageResult(storageDueToAnError))
			return
		}

		if !plr.storageInventory.addItem(storeCopy) {
			_, _ = plr.GiveItem(storeCopy)
			plr.Send(packetNpcStorageResult(storageIsFull))
			return
		}

		if err := plr.storageInventory.save(plr.accountID); err != nil {
			_, _ = plr.GiveItem(storeCopy)
			plr.Send(packetNpcStorageResult(storageDueToAnError))
			return
		}

		sectionItems := plr.storageInventory.getItemsInSection(tab)
		plr.Send(packetNpcStorageItemsChanged(encDeposit, plr.storageInventory.maxSlots, tab, 0, sectionItems))

	case actionStoreMesos:
		val := reader.ReadInt32()
		if val < 0 {
			store := -val
			if store <= 0 || plr.mesos < store {
				plr.Send(packetNpcStorageResult(storageNotEnoughMesos))
				return
			}
			plr.giveMesos(-store)
			if err := plr.storageInventory.changeMesos(store); err != nil {
				plr.giveMesos(store)
				plr.Send(packetNpcStorageResult(storageDueToAnError))
				return
			}
			if err := plr.storageInventory.save(plr.accountID); err != nil {
				_ = plr.storageInventory.changeMesos(-store)
				plr.giveMesos(store)
				plr.Send(packetNpcStorageResult(storageDueToAnError))
				return
			}
			plr.Send(packetNpcStorageMesosChanged(storageSuccess, plr.storageInventory.mesos, plr.storageInventory.maxSlots))
		} else if val > 0 {
			withdraw := val
			if err := plr.storageInventory.changeMesos(-withdraw); err != nil {
				plr.Send(packetNpcStorageResult(storageDueToAnError))
				return
			}
			plr.giveMesos(withdraw)
			if err := plr.storageInventory.save(plr.accountID); err != nil {
				_ = plr.storageInventory.changeMesos(withdraw)
				plr.giveMesos(-withdraw)
				plr.Send(packetNpcStorageResult(storageDueToAnError))
				return
			}
			plr.Send(packetNpcStorageMesosChanged(storageSuccess, plr.storageInventory.mesos, plr.storageInventory.maxSlots))
		} else {
			plr.Send(packetNpcStorageResult(storageIsFull))
		}

	case actionExit:
		return
	}
}

func (server *Server) playerHandleMessenger(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	mode := reader.ReadByte()

	switch mode {
	case constant.MessengerEnter:
		messengerID := reader.ReadInt32()
		var petAcc int32
		var vis []internal.KV
		var hid []internal.KV

		base := make(map[byte]int32)
		cash := make(map[byte]int32)
		cashW := int32(0)

		for _, it := range plr.equip {
			if it.slotID >= 0 {
				continue
			}

			slot := byte(-it.slotID)
			if it.slotID < -100 {
				slot = byte(-(it.slotID + 100))
			}

			if slot == 11 {
				if it.slotID < -100 {
					cashW = it.ID
				}
				continue
			}

			if it.slotID < -100 {
				cash[slot] = it.ID
			} else {
				if _, ok := base[slot]; !ok {
					base[slot] = it.ID
				}
			}
		}

		order := func(m map[byte]int32) []byte {
			ks := make([]byte, 0, len(m))
			for k := range m {
				ks = append(ks, k)
			}
			sort.Slice(ks, func(i, j int) bool { return ks[i] < ks[j] })
			return ks
		}

		for _, k := range order(base) {
			if v, ok := cash[k]; ok {
				vis = append(vis, internal.KV{K: k, V: v})
				hid = append(hid, internal.KV{K: k, V: base[k]})
			} else {
				vis = append(vis, internal.KV{K: k, V: base[k]})
			}
		}

		for _, k := range order(cash) {
			if _, used := base[k]; !used {
				vis = append(vis, internal.KV{K: k, V: cash[k]})
			}
		}

		p := internal.PacketMessengerEnter(plr.ID, messengerID, plr.face, plr.hair, cashW, petAcc, server.id, plr.gender, plr.skin, plr.Name, vis, hid)
		server.world.Send(p)
	case constant.MessengerLeave:
		p := internal.PacketMessengerLeave(plr.ID)
		server.world.Send(p)
	case constant.MessengerInvite:
		invitee := reader.ReadString(reader.ReadInt16())
		p := internal.PacketMessengerInvite(plr.ID, server.id, plr.Name, invitee)
		server.world.Send(p)
	case constant.MessengerBlocked:
		invitee := reader.ReadString(reader.ReadInt16())
		inviter := reader.ReadString(reader.ReadInt16())
		blockMode := reader.ReadByte()
		p := internal.PacketMessengerBlocked(plr.ID, server.id, blockMode, plr.Name, invitee, inviter)
		server.world.Send(p)
	case constant.MessengerChat:
		message := reader.ReadString(reader.ReadInt16())
		p := internal.PacketMessengerChat(plr.ID, server.id, plr.Name, message)
		server.world.Send(p)
	case constant.MessengerAvatar:
		var petAcc int32
		var vis []internal.KV
		var hid []internal.KV

		base := make(map[byte]int32)
		cash := make(map[byte]int32)
		cashW := int32(0)

		for _, it := range plr.equip {
			if it.slotID >= 0 {
				continue
			}

			slot := byte(-it.slotID)
			if it.slotID < -100 {
				slot = byte(-(it.slotID + 100))
			}

			if slot == 11 {
				if it.slotID < -100 {
					cashW = it.ID
				}
				continue
			}

			if it.slotID < -100 {
				cash[slot] = it.ID
			} else {
				if _, ok := base[slot]; !ok {
					base[slot] = it.ID
				}
			}
		}

		order := func(m map[byte]int32) []byte {
			ks := make([]byte, 0, len(m))
			for k := range m {
				ks = append(ks, k)
			}
			sort.Slice(ks, func(i, j int) bool { return ks[i] < ks[j] })
			return ks
		}

		for _, k := range order(base) {
			if v, ok := cash[k]; ok {
				vis = append(vis, internal.KV{K: k, V: v})
				hid = append(hid, internal.KV{K: k, V: base[k]})
			} else {
				vis = append(vis, internal.KV{K: k, V: base[k]})
			}
		}

		for _, k := range order(cash) {
			if _, used := base[k]; !used {
				vis = append(vis, internal.KV{K: k, V: cash[k]})
			}
		}

		p := internal.PacketMessengerAvatar(plr.gender, plr.skin, server.id, plr.ID, plr.face, plr.hair, cashW, petAcc, plr.Name, vis, hid)
		server.world.Send(p)
	default:
		return
	}

}

func (server *Server) playerPetSpawn(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	slot := reader.ReadInt16()

	petItem, err := plr.getItem(5, slot)
	if !petItem.pet || err != nil {
		plr.Send(packetPlayerNoChange())
		return
	}

	if petItem.petData == nil {
		sn, _ := nx.GetCommoditySNByItemID(petItem.ID)
		petItem.petData = newPet(petItem.ID, sn, petItem.dbID)
		savePet(&petItem)
	}

	petEquipped := plr.petCashID != 0
	changePet := petEquipped && plr.petCashID == int64(petItem.petData.sn)

	if petEquipped {
		plr.inst.send(packetPetRemove(plr.ID, constant.PetRemoveNone))
		plr.petCashID = 0
	}

	if !changePet {
		plr.petCashID = int64(petItem.petData.sn)

		if plr.pet == nil || plr.pet.sn != petItem.petData.sn {
			plr.pet = petItem.petData
		}

		plr.pet.pos = plr.pos
		plr.inst.send(packetPetSpawn(plr.ID, plr.pet))

		if plr.pet.spawnDate == 0 {
			plr.Send(packetPlayerPetUpdate(plr.pet.sn))
		}
		plr.pet.spawnDate = time.Now().Unix()
		plr.pet.spawned = true
	}

	plr.MarkDirty(DirtyPet, time.Millisecond*300)
	plr.Send(packetPlayerNoChange())
}

func (server *Server) playerPetMove(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	moveData, finalData := parseMovement(reader)
	moveBytes := generateMovementBytes(moveData)

	plr.pet.updateMovement(finalData)

	field, ok := server.fields[plr.mapID]

	if !ok {
		return
	}

	inst, err := field.getInstance(plr.inst.id)

	if err != nil {
		return
	}

	inst.movePlayerPet(plr.ID, moveBytes, plr)
}

func (server *Server) playerPetAction(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	actType := reader.ReadByte()
	act := reader.ReadByte()
	msg := reader.ReadRestAsString()

	plr.inst.send(packetPetAction(plr.ID, actType, act, msg))
}

func (server *Server) playerPetInteraction(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	if plr.pet == nil || !plr.pet.spawned {
		return
	}

	doMultiplier := reader.ReadByte()
	interactionID := reader.ReadByte()

	petItem := plr.pet
	success := handlePetInteraction(plr, petItem, interactionID, doMultiplier == 1)
	plr.Send(packetPetInteraction(plr.ID, interactionID, success, false))
}

func (server *Server) playerPetLoot(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil || plr.pet == nil || !plr.pet.spawned {
		return
	}

	reader.Skip(4) // unused pos
	dropID := reader.ReadInt32()

	err, drop := plr.inst.dropPool.findDropFromID(dropID)
	if err != nil {
		// Pet spams pickup so just silently return :)
		return
	}

	if plr.pet.pos.x-drop.finalPos.x > 800 || plr.pet.pos.y-drop.finalPos.y > 600 {
		// Hax
		log.Printf("Player: %s pet tried to pickup an item from far away", plr.Name)
		plr.Send(packetDropNotAvailable())
		plr.Send(packetInventoryDontTake())
		return
	}

	if !plr.petCanTakeDrop(drop) {
		return
	}

	if drop.mesos > 0 {
		amount := int32(plr.inst.dropPool.rates.mesos * float32(drop.mesos))
		plr.giveMesos(amount)
	} else {
		err, _ = plr.GiveItem(drop.item)
		if err != nil {
			plr.Send(packetInventoryFull())
			plr.Send(packetInventoryDontTake())
			return
		}

	}

	plr.inst.dropPool.playerAttemptPickup(drop, plr, 5)
}
