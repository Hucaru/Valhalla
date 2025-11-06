package channel

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/constant/skill"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
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
		newParty.addPlayer(plr, 0, &reader)
		server.parties[newParty.ID] = newParty
	case internal.OpPartyLeaveExpel:
		partyID := reader.ReadInt32()
		destroy := reader.ReadBool()
		kicked := reader.ReadBool()
		index := reader.ReadInt32()

		if party, ok := server.parties[partyID]; ok {
			party.removePlayer(index, kicked, &reader)
		}

		if destroy {
			delete(server.parties, partyID)
		}

	case internal.OpPartyAccept:
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		index := reader.ReadInt32()

		if party, ok := server.parties[partyID]; ok {
			plr, _ := server.players.GetFromID(playerID)
			party.addPlayer(plr, index, &reader)
		}
	case internal.OpPartyInfoUpdate:
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		index := reader.ReadInt32()
		onlineStatus := reader.ReadBool()
		if party, ok := server.parties[partyID]; ok {
			if onlineStatus {
				plr, _ := server.players.GetFromID(playerID)
				party.updateOnlineStatus(index, plr, &reader)
			} else {
				party.updateInfo(index, &reader)
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

	readMobListAndDelay := func() {
		count := int(reader.ReadByte())
		for i := 0; i < count; i++ {
			_ = reader.ReadInt32() // mob spawn ID
		}
		_ = reader.ReadInt16() // delay
	}

	switch skill.Skill(skillID) {
	// Party buffs handled earlier remain unchanged...
	case skill.Haste, skill.BanditHaste, skill.Bless, skill.IronWill, skill.Rage, skill.GMHaste, skill.GMBless, skill.GMHolySymbol,
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

	// Self toggles and non-party buffs (boolean/ratio-type): apply to self
	case skill.DarkSight,
		skill.MagicGuard,
		skill.Invincible,
		skill.SoulArrow, skill.CBSoulArrow,
		skill.ShadowPartner, skill.GMShadowPartner,
		skill.MesoGuard,
		// Attack speed boosters (self)
		skill.SwordBooster, skill.AxeBooster, skill.PageSwordBooster, skill.BwBooster,
		skill.SpearBooster, skill.PolearmBooster,
		skill.BowBooster, skill.CrossbowBooster,
		skill.ClawBooster, skill.DaggerBooster,
		skill.SpellBooster, skill.ILSpellBooster,
		// GM Hide (mapped to invincible bit)
		skill.Hide:
		plr.addBuff(skillID, skillLevel, delay)
		plr.inst.send(packetPlayerSkillAnimation(plr.ID, true, skillID, skillLevel))

	// Debuffs on mobs: [mobCount][mobIDs...][delay]
	case skill.Threaten,
		skill.Slow, skill.ILSlow,
		skill.MagicCrash,
		skill.PowerCrash,
		skill.ArmorCrash,
		skill.ILSeal, skill.Seal,
		skill.ShadowWeb,
		skill.Doom:
		readMobListAndDelay()

	// Summons and puppet:
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

		plr.addBuff(skillID, skillLevel, delay)
		plr.addSummon(summ)
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
		plr.removeSummon(true, 0x02)
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
