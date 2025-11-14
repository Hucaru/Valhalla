package channel

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (server *Server) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.ChannelBad:
		server.handleNewChannelBad(conn, reader)
	case opcode.ChannelOk:
		server.handleNewChannelOK(conn, reader)
	case opcode.ChannelConnectionInfo:
		server.handleChannelConnectionInfo(conn, reader)
	case opcode.ChannelPlayerConnect:
		server.handlePlayerConnectedNotifications(conn, reader)
	case opcode.ChannePlayerDisconnect:
		server.handlePlayerDisconnectNotifications(conn, reader)
	case opcode.ChannelPlayerChatEvent:
		server.handleChatEvent(conn, reader)
	case opcode.ChannelPlayerBuddyEvent:
		server.handleBuddyEvent(conn, reader)
	case opcode.ChannelPlayerPartyEvent:
		server.handlePartyEvent(conn, reader)
	case opcode.ChannelPlayerGuildEvent:
		server.handleGuildEvent(conn, reader)
	case opcode.ChangeRate:
		server.handleChangeRate(conn, reader)
	case opcode.CashShopInfo:
		server.handleCashShopInfo(conn, reader)
	case opcode.ChannelPlayerMessengerEvent:
		server.handleMessengerEvent(conn, reader)
	case opcode.LoginDeleteCharacter:
		server.handleCharacterDeleted(conn, reader)
	case opcode.SyncParties:
		server.handleSyncParties(conn, reader)

	default:
		log.Println("UNKNOWN SERVER PACKET:", reader)
	}
}

func (server *Server) handleNewChannelBad(conn mnet.Server, reader mpacket.Reader) {
	log.Println("Rejected by world server at", conn)
	timer := time.NewTimer(30 * time.Second)

	<-timer.C

	server.registerWithWorld()
}

func (server *Server) handleNewChannelOK(conn mnet.Server, reader mpacket.Reader) {
	server.worldName = reader.ReadString(reader.ReadInt16())
	server.id = reader.ReadByte()
	server.rates.exp = reader.ReadFloat32()
	server.rates.drop = reader.ReadFloat32()
	server.rates.mesos = reader.ReadFloat32()

	log.Printf("Registered as channel %d on world %s with rates: Exp - x%.2f, Drop - x%.2f, Mesos - x%.2f",
		server.id, server.worldName, server.rates.exp, server.rates.drop, server.rates.mesos)

	server.players.broadcast(packetMessageNotice("Re-connected to world server as channel " + strconv.Itoa(int(server.id+1))))

	accountIDs, err := common.DB.Query("SELECT accountID from characters where channelID = ? and migrationID = -1", server.id)

	if err != nil {
		log.Fatal(err)
	}

	defer accountIDs.Close()

	for accountIDs.Next() {
		var accountID int
		err := accountIDs.Scan(&accountID)

		if err != nil {
			continue
		}

		_, err = common.DB.Exec("UPDATE accounts SET isLogedIn=? WHERE accountID=?", 0, accountID)

		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = common.DB.Exec("UPDATE characters SET channelID=? WHERE channelID=?", -1, server.id)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Logged out any accounts still connected to this channel")
}

func (server *Server) handleChannelConnectionInfo(conn mnet.Server, reader mpacket.Reader) {
	total := reader.ReadByte()

	for i := byte(0); i < total; i++ {
		server.channels[i].IP = reader.ReadBytes(4)
		server.channels[i].Port = reader.ReadInt16()
	}
}

func (server *Server) handleCashShopInfo(conn mnet.Server, reader mpacket.Reader) {
	server.cashShop.IP = reader.ReadBytes(4)
	server.cashShop.Port = reader.ReadInt16()

	log.Println("Cash Shop Information Recieved IP:", server.cashShop.IP, "Port:", server.cashShop.Port)
}

func (server *Server) handlePlayerConnectedNotifications(conn mnet.Server, reader mpacket.Reader) {
	playerID := reader.ReadInt32()
	name := reader.ReadString(reader.ReadInt16())
	channelID := reader.ReadByte()
	changeChannel := reader.ReadBool()
	_ = reader.ReadInt32() // mapID
	guildID := reader.ReadInt32()

	plr, err := server.players.GetFromID(playerID)

	if err == nil && plr.guild != nil {
		plr.guild.playerOnline(playerID, plr, true, changeChannel)
	} else {
		if guild, ok := server.guilds[guildID]; ok {
			guild.playerOnline(playerID, nil, true, changeChannel)
		} else if guildID > -1 {

		}
	}

	server.players.observe(func(plr *Player) {
		if plr.hasBuddy(playerID) {
			if changeChannel {
				plr.Send(packetBuddyChangeChannel(playerID, int32(channelID)))
				plr.addOnlineBuddy(playerID, name, int32(channelID))
			} else {
				plr.Send(packetBuddyOnlineStatus(playerID, int32(channelID)))
				plr.addOnlineBuddy(playerID, name, int32(channelID))
			}
		}
	})
}

func (server *Server) handlePlayerDisconnectNotifications(conn mnet.Server, reader mpacket.Reader) {
	playerID := reader.ReadInt32()
	name := reader.ReadString(reader.ReadInt16())
	guildID := reader.ReadInt32()

	if guild, ok := server.guilds[guildID]; ok {
		guild.playerOnline(playerID, nil, false, false)

		if guild.canUnload() {
			delete(server.guilds, guildID)
		}
	}

	server.players.observe(func(plr *Player) {
		if plr.hasBuddy(playerID) {
			plr.addOfflineBuddy(playerID, name)
		}
	})
}

func (server *Server) handleBuddyEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 1:
		recepientID := reader.ReadInt32()
		fromID := reader.ReadInt32()
		fromName := reader.ReadString(reader.ReadInt16())
		channelID := reader.ReadByte()

		if channelID == server.id {
			return
		}

		plr, err := server.players.GetFromID(recepientID)

		if err != nil {
			return
		}

		plr.Send(packetBuddyReceiveRequest(fromID, fromName, int32(channelID)))
	case 2:
		recepientID := reader.ReadInt32()
		fromID := reader.ReadInt32()
		fromName := reader.ReadString(reader.ReadInt16())
		channelID := reader.ReadByte()

		if channelID == server.id {
			return
		}

		plr, err := server.players.GetFromID(recepientID)

		if err != nil {
			return
		}

		plr.addOfflineBuddy(fromID, fromName)
		plr.Send(packetBuddyOnlineStatus(fromID, int32(channelID)))
		plr.addOnlineBuddy(fromID, fromName, int32(channelID))
	case 3:
		recepientID := reader.ReadInt32()
		fromID := reader.ReadInt32()
		channelID := reader.ReadByte()

		if channelID == server.id {
			return
		}

		plr, err := server.players.GetFromID(recepientID)

		if err != nil {
			return
		}

		plr.removeBuddy(fromID)
	default:
		log.Println("Unknown buddy event type:", op)
	}
}

func (server *Server) handlePartyEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case internal.OpPartyCreate:
		playerID := reader.ReadInt32()

		plr, err := server.players.GetFromID(playerID)

		if !reader.ReadBool() {
			if err != nil {
				plr.Send(packetPartyCreateUnkownError())
			}

			return
		}

		newParty := &party{serverChannelID: int32(server.id)}
		newParty.SerialisePacket(&reader)
		newParty.addPlayer(plr, 0)
		server.parties[newParty.ID] = newParty
		server.updatePartyMetric()
	case internal.OpPartyLeaveExpel:
		partyID := reader.ReadInt32()
		destroy := reader.ReadBool()
		kicked := reader.ReadBool()
		index := reader.ReadInt32()

		if party, ok := server.parties[partyID]; ok {
			party.SerialisePacket(&reader)
			party.removePlayer(index, kicked)
		}

		if destroy {
			delete(server.parties, partyID)
			server.updatePartyMetric()
		}

	case internal.OpPartyAccept:
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		index := reader.ReadInt32()

		if party, ok := server.parties[partyID]; ok {
			plr, _ := server.players.GetFromID(playerID)
			party.SerialisePacket(&reader)
			party.addPlayer(plr, index)

			if plr != nil && plr.inst != nil {
				plr.inst.requestDoorPartySync(plr)
			}
		}

	case internal.OpPartyInfoUpdate:
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		index := reader.ReadInt32()
		onlineStatus := reader.ReadBool()
		if party, ok := server.parties[partyID]; ok {
			party.SerialisePacket(&reader)
			if onlineStatus {
				plr, _ := server.players.GetFromID(playerID)
				party.updateOnlineStatus(index, plr)
			} else {
				party.updateInfo(index)
			}
		}
	default:
		log.Println("Unknown party event type:", op)
	}
}

func (server *Server) handleChangeRate(conn mnet.Server, reader mpacket.Reader) {
	mode := reader.ReadByte()
	rate := reader.ReadFloat32()

	modeMap := map[byte]string{
		1: "exp",
		2: "drop",
		3: "mesos",
	}
	switch mode {
	case 1:
		server.rates.exp = rate
	case 2:
		server.rates.drop = rate
	case 3:
		server.rates.mesos = rate
	default:
		log.Println("Unknown rate mode")
		return
	}

	log.Printf("%s rate has changed to x%.2f", modeMap[mode], rate)
	server.players.broadcast(packetMessageNotice(fmt.Sprintf("%s rate has changed to x%.2f", modeMap[mode], rate)))
}

func (server Server) handleChatEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case internal.OpChatWhispher:
		recepientName := reader.ReadString(reader.ReadInt16())
		fromName := reader.ReadString(reader.ReadInt16())
		msg := reader.ReadString(reader.ReadInt16())
		channelID := reader.ReadByte()

		plr, err := server.players.GetFromName(recepientName)

		if err != nil {
			return
		}

		plr.Send(packetMessageWhisper(fromName, msg, channelID))

	case internal.OpChatBuddy:
		fromName := reader.ReadString(reader.ReadInt16())
		idCount := reader.ReadByte()

		ids := make([]int32, int(idCount))

		for i := byte(0); i < idCount; i++ {
			ids[i] = reader.ReadInt32()
		}

		msg := reader.ReadString(reader.ReadInt16())

		for _, v := range ids {
			plr, err := server.players.GetFromID(v)

			if err != nil {
				continue
			}

			plr.Send(packetMessageBubblessChat(0, fromName, msg))
		}
	case internal.OpChatParty:
		fromName := reader.ReadString(reader.ReadInt16())
		idCount := reader.ReadByte()

		ids := make([]int32, int(idCount))

		for i := byte(0); i < idCount; i++ {
			ids[i] = reader.ReadInt32()
		}

		msg := reader.ReadString(reader.ReadInt16())

		for _, v := range ids {
			plr, err := server.players.GetFromID(v)

			if err != nil {
				continue
			}

			plr.Send(packetMessageBubblessChat(1, fromName, msg))
		}
	case internal.OpChatGuild:
		fromName := reader.ReadString(reader.ReadInt16())
		idCount := reader.ReadByte()

		ids := make([]int32, int(idCount))

		for i := byte(0); i < idCount; i++ {
			ids[i] = reader.ReadInt32()
		}

		msg := reader.ReadString(reader.ReadInt16())

		for _, v := range ids {
			plr, err := server.players.GetFromID(v)

			if err != nil {
				continue
			}

			plr.Send(packetMessageBubblessChat(2, fromName, msg))
		}
	case internal.OpChatMegaphone: // Super megaphone broadcast from world server
		fromName := reader.ReadString(reader.ReadInt16())
		msg := reader.ReadString(reader.ReadInt16())
		whisper := reader.ReadBool()

		// Broadcast to all players on this channel
		server.players.broadcast(packetMessageBroadcastSuper(fromName, msg, server.id, whisper))
	default:
		log.Println("Unknown chat event type:", op)
	}
}

func (server *Server) handleMessengerEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case constant.MessengerEnter:
		recipientID := reader.ReadInt32()
		slot := reader.ReadByte()
		if plr, err := server.players.GetFromID(recipientID); err == nil {
			plr.Send(packetMessengerSelfEnter(slot))
		}
	case constant.MessengerEnterResult:
		recipientID := reader.ReadInt32()
		slot := reader.ReadByte()
		gender := reader.ReadByte()
		skin := reader.ReadByte()
		face := reader.ReadInt32()
		_ = reader.ReadBool()
		hair := reader.ReadInt32()

		var vis []internal.KV
		for {
			b := reader.ReadByte()
			if b == 0xFF {
				break
			}
			vis = append(vis, internal.KV{K: b, V: reader.ReadInt32()})
		}
		var hid []internal.KV
		for {
			b := reader.ReadByte()
			if b == 0xFF {
				break
			}
			hid = append(hid, internal.KV{K: b, V: reader.ReadInt32()})
		}

		cashW := reader.ReadInt32()
		petAcc := reader.ReadInt32()

		name := reader.ReadString(reader.ReadInt16())
		ch := reader.ReadByte()
		announce := reader.ReadBool()

		if plr, err := server.players.GetFromID(recipientID); err == nil {
			p := packetMessengerEnter(slot, gender, skin, ch, face, hair, cashW, petAcc, name, announce, vis, hid)
			plr.Send(p)
		}

	case constant.MessengerLeave:
		recipientID := reader.ReadInt32()
		slot := reader.ReadByte()
		if plr, err := server.players.GetFromID(recipientID); err == nil {
			plr.Send(packetMessengerLeave(slot))
		}
	case constant.MessengerInvite:
		inviteeID := reader.ReadInt32()
		sender := reader.ReadString(reader.ReadInt16())
		mID := reader.ReadInt32()
		nameResolution := reader.ReadBool()
		if nameResolution {
			targetName := reader.ReadString(reader.ReadInt16())
			if plr, err := server.players.GetFromName(targetName); err == nil {
				plr.Send(packetMessengerInvite(sender, mID))
			}
			return
		}
		if inviteeID != 0 {
			if plr, err := server.players.GetFromID(inviteeID); err == nil {
				plr.Send(packetMessengerInvite(sender, mID))
			}
		}
	case constant.MessengerInviteResult:
		senderID := reader.ReadInt32()
		recipient := reader.ReadString(reader.ReadInt16())
		success := reader.ReadBool()
		if plr, err := server.players.GetFromID(senderID); err == nil {
			plr.Send(packetMessengerInviteResult(recipient, success))
		}
	case constant.MessengerBlocked:
		senderID := reader.ReadInt32()
		receiver := reader.ReadString(reader.ReadInt16())
		mode := reader.ReadByte()
		if plr, err := server.players.GetFromID(senderID); err == nil {
			plr.Send(packetMessengerBlocked(receiver, mode))
		}
	case constant.MessengerChat:
		recipientID := reader.ReadInt32()
		msg := reader.ReadString(reader.ReadInt16())
		if plr, err := server.players.GetFromID(recipientID); err == nil {
			plr.Send(packetMessengerChat(msg))
		}
	case constant.MessengerAvatar:
		recipientID := reader.ReadInt32()
		slot := reader.ReadByte()
		gender := reader.ReadByte()
		skin := reader.ReadByte()
		face := reader.ReadInt32()
		_ = reader.ReadBool()
		hair := reader.ReadInt32()

		var vis []internal.KV
		for {
			b := reader.ReadByte()
			if b == 0xFF {
				break
			}
			vis = append(vis, internal.KV{K: b, V: reader.ReadInt32()})
		}
		var hid []internal.KV
		for {
			b := reader.ReadByte()
			if b == 0xFF {
				break
			}
			hid = append(hid, internal.KV{K: b, V: reader.ReadInt32()})
		}

		cashW := reader.ReadInt32()
		petAcc := reader.ReadInt32()

		if plr, err := server.players.GetFromID(recipientID); err == nil {
			p := packetMessengerAvatar(slot, gender, skin, face, hair, cashW, petAcc, vis, hid)
			plr.Send(p)
		}
	default:
	}
}

func (server Server) handleGuildEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case internal.OpGuildDisband:
		guildID := reader.ReadInt32()

		if guild, ok := server.guilds[guildID]; ok {
			guild.disband()
			delete(server.guilds, guildID)
		}
	case internal.OpGuildTitlesChange:
		guildID := reader.ReadInt32()
		master := reader.ReadString(reader.ReadInt16())
		jrMaster := reader.ReadString(reader.ReadInt16())
		member1 := reader.ReadString(reader.ReadInt16())
		member2 := reader.ReadString(reader.ReadInt16())
		member3 := reader.ReadString(reader.ReadInt16())

		if guild, ok := server.guilds[guildID]; ok {
			guild.updateTitles(master, jrMaster, member1, member2, member3)
		}
	case internal.OpGuildRankUpdate:
		guildID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		rank := reader.ReadByte()

		if guild, ok := server.guilds[guildID]; ok {
			guild.updateRank(playerID, rank)
		}
	case internal.OpGuildRemovePlayer:
		guildID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		expelled := reader.ReadBool()
		playerName := reader.ReadString(reader.ReadInt16())

		if guild, ok := server.guilds[guildID]; ok {
			guild.removePlayer(playerID, expelled, playerName)
		}
	case internal.OpGuildNoticeChange:
		guildID := reader.ReadInt32()
		notice := reader.ReadString(reader.ReadInt16())

		if guild, ok := server.guilds[guildID]; ok {
			guild.notice = notice
			guild.broadcast(packetGuildUpdateNotice(guildID, notice))
		}
	case internal.OpGuildEmblemChange:
		guildID := reader.ReadInt32()

		if guild, ok := server.guilds[guildID]; ok {
			logoBg := reader.ReadInt16()
			logo := reader.ReadInt16()
			logoBgColour := reader.ReadByte()
			logoColour := reader.ReadByte()
			guild.updateEmblem(logoBg, logo, logoBgColour, logoColour)
		}
	case internal.OpGuildPointsUpdate:
		guildID := reader.ReadInt32()
		points := reader.ReadInt32()

		if guild, ok := server.guilds[guildID]; ok {
			guild.setPoints(points)
		}
	case internal.OpGuildInvite:
		guildID := reader.ReadInt32()
		inviter := reader.ReadString(reader.ReadInt16())
		invitee := reader.ReadString(reader.ReadInt16())

		plr, err := server.players.GetFromName(invitee)

		if err != nil {
			return
		}

		plr.Send(packetGuildInviteCard(guildID, inviter))
	case internal.OpGuildInviteReject:
		inviterName := reader.ReadString(reader.ReadInt16())
		inviteeName := reader.ReadString(reader.ReadInt16())

		inviter, err := server.players.GetFromName(inviterName)

		if err != nil {
			return
		}

		inviter.Send(packetGuildInviteRejected(inviteeName))
	case internal.OpGuildInviteAccept:
		playerID := reader.ReadInt32()
		guildID := reader.ReadInt32()
		name := reader.ReadString(reader.ReadInt16())
		jobID := reader.ReadInt32()
		level := reader.ReadInt32()
		online := reader.ReadBool()
		rank := reader.ReadByte()

		var err error

		guild, ok := server.guilds[guildID]

		if !ok {
			guild, err = loadGuildFromDb(guildID, &server.players)

			if err != nil {
				return
			}
		}

		guild.addPlayer(playerID, name, jobID, level, online, rank)
	default:
		log.Println("Unkown guild event type:", op)
	}
}

func (server *Server) guildInviteResult(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 0x36: // client sends this when it receives Player is dealing with another invitation
		inviter := reader.ReadString(reader.ReadInt16())
		invitee := reader.ReadString(reader.ReadInt16())
		_, _ = inviter, invitee
	case constant.GuildRejectInvite: // reject
		inviterName := reader.ReadString(reader.ReadInt16())
		inviteeName := reader.ReadString(reader.ReadInt16())

		var guildID, playerID int32

		query := "SELECT guildID FROM characters WHERE Name=?"
		err := common.DB.QueryRow(query, inviterName).Scan(&guildID)

		if err != nil {
			log.Fatal(err)
		}

		query = "SELECT ID FROM characters WHERE Name=?"
		err = common.DB.QueryRow(query, inviteeName).Scan(&playerID)

		if err != nil {
			log.Fatal(err)
		}

		query = "DELETE FROM guild_invites WHERE playerID=? AND guildID=?"

		if _, err = common.DB.Exec(query, playerID, guildID); err != nil {
			log.Fatal(err)
		}

		server.world.Send(internal.PacketGuildInviteReject(inviterName, inviteeName))
	default:
		log.Println("Unknown guild invite operation", op, reader)
	}

}

func (server *Server) handleCharacterDeleted(conn mnet.Server, reader mpacket.Reader) {
	charID := reader.ReadInt32()

	// Remove from any party
	for _, party := range server.parties {
		for i, plrID := range party.PlayerID {
			if plrID == charID {
				party.removePlayer(int32(i), true)
				break
			}
		}
	}

	// Remove from any guild
	for _, guild := range server.guilds {
		for i, playerID := range guild.playerID {
			if playerID == charID {
				guild.removePlayer(charID, false, guild.names[i])
				break
			}
		}
	}
}

func (server *Server) handleSyncParties(conn mnet.Server, reader mpacket.Reader) {
	count := reader.ReadInt32()
	log.Printf("Syncing %d parties from world server", count)

	for i := int32(0); i < count; i++ {
		partyID := reader.ReadInt32()

		// Create or get existing party
		p, exists := server.parties[partyID]
		if !exists {
			p = &party{
				serverChannelID: int32(server.id),
			}
			server.parties[partyID] = p
		}

		p.ID = partyID

		// Read party data
		for j := 0; j < constant.MaxPartySize; j++ {
			p.PlayerID[j] = reader.ReadInt32()
			p.ChannelID[j] = reader.ReadInt32()
			p.MapID[j] = reader.ReadInt32()
			p.Job[j] = reader.ReadInt32()
			p.Level[j] = reader.ReadInt32()
			p.Name[j] = reader.ReadString(reader.ReadInt16())

			// If player is on this channel, link them
			if p.ChannelID[j] == int32(server.id) && p.PlayerID[j] > 0 {
				if plr, err := server.players.GetFromID(p.PlayerID[j]); err == nil {
					p.players[j] = plr
					plr.party = p
				}
			}
		}
	}

	log.Printf("Party sync completed: %d parties loaded", count)
}
