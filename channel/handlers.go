package channel

import (
	"log"
	"strconv"
	"time"

	"github.com/Hucaru/Valhalla/channel/message"
	"github.com/Hucaru/Valhalla/channel/party"
	"github.com/Hucaru/Valhalla/channel/player"
	"github.com/Hucaru/Valhalla/channel/pos"
	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// HandleClientPacket data
func (server *Server) HandleClientPacket(conn mnet.Client, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.RecvPing:
	case opcode.RecvChannelPlayerLoad:
		server.playerConnect(conn, reader)
	case opcode.RecvCHannelChangeChannel:
		server.playerChangeChannel(conn, reader)
	case opcode.RecvChannelUserPortal:
		// This opcode is used for revival UI as well.
		server.playerUsePortal(conn, reader)
	case opcode.RecvChannelEnterCashShop:
		conn.Send(message.PacketMessageDialogueBox("Shop not implemented"))
	case opcode.RecvChannelPlayerMovement:
		server.playerMovement(conn, reader)
	case opcode.RecvChannelPlayerStand:
		server.playerStand(conn, reader)
	case opcode.RecvChannelPlayerUseChair:
		server.playerUseChair(conn, reader)
	case opcode.RecvChannelMeleeSkill:
		server.playerMeleeSkill(conn, reader)
	case opcode.RecvChannelRangedSkill:
		// server.playerRangedSkill(conn, reader)
	case opcode.RecvChannelMagicSkill:
		// server.playerMagicSkill(conn, reader)
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
	case opcode.RecvChannelInvUseItem:
		server.playerUseInventoryItem(conn, reader)
	case opcode.RecvChannelAddStatPoint:
		server.playerAddStatPoint(conn, reader)
	case opcode.RecvChannelPassiveRegen:
		server.playerPassiveRegen(conn, reader)
	case opcode.RecvChannelAddSkillPoint:
		server.playerAddSkillPoint(conn, reader)
	case opcode.RecvChannelSpecialSkill:
		// server.playerSpecialSkill(conn, reader)
	case opcode.RecvChannelCharacterInfo:
		server.playerRequestAvatarInfoWindow(conn, reader)
	case opcode.RecvChannelLieDetectorResult:
	case opcode.RecvChannelPartyInfo:
		server.playerPartyInfo(conn, reader)
	case opcode.RecvChannelGuildManagement:
	case opcode.RecvChannelGuildReject:
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
	default:
		log.Println("UNKNOWN CLIENT PACKET:", reader)
	}
}

// HandleServerPacket from world
func (server *Server) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.ChannelBad:
		server.handleNewChannelBad(conn, reader)
	case opcode.ChannelOk:
		server.handleNewChannelOK(conn, reader)
	case opcode.ChannelConnectionInfo:
		server.handleChannelConnectionInfo(conn, reader)
	case opcode.ChannePlayerConnect:
		server.handlePlayerConnectedNotifications(conn, reader)
	case opcode.ChannePlayerDisconnect:
		server.handlePlayerDisconnectNotifications(conn, reader)
	case opcode.ChannelPlayerChatEvent:
		server.handleChatEvent(conn, reader)
	case opcode.ChannelPlayerBuddyEvent:
		server.handleBuddyEvent(conn, reader)
	case opcode.ChannelPlayerPartyEvent:
		server.handlePartyEvent(conn, reader)
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
	log.Println("Registered as channel", server.id, "on world", server.worldName)

	for _, p := range server.players {
		p.Send(message.PacketMessageNotice("Re-connected to world server as channel " + strconv.Itoa(int(server.id+1))))
		// TODO send largest party id for world server to compare
	}

	accountIDs, err := common.DB.Query("SELECT accountID from characters where channelID = ? and migrationID = -1", server.id)

	if err != nil {
		log.Println(err)
		return
	}

	for accountIDs.Next() {
		var accountID int
		err := accountIDs.Scan(&accountID)

		if err != nil {
			continue
		}

		_, err = common.DB.Exec("UPDATE accounts SET isLogedIn=? WHERE accountID=?", 0, accountID)

		if err != nil {
			log.Println(err)
			return
		}
	}

	accountIDs.Close()

	_, err = common.DB.Exec("UPDATE characters SET channelID=? WHERE channelID=?", -1, server.id)

	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Loged out any accounts still connected to this channel")
}

func (server *Server) handleChannelConnectionInfo(conn mnet.Server, reader mpacket.Reader) {
	total := reader.ReadByte()

	for i := byte(0); i < total; i++ {
		server.channels[i].IP = reader.ReadBytes(4)
		server.channels[i].Port = reader.ReadInt16()
	}
}

func (server *Server) handlePlayerConnectedNotifications(conn mnet.Server, reader mpacket.Reader) {
	playerID := reader.ReadInt32()
	name := reader.ReadString(reader.ReadInt16())
	channelID := reader.ReadByte()
	changeChannel := reader.ReadBool()

	plr, _ := server.players.getFromID(playerID)

	for _, party := range server.parties {
		party.SetPlayerChannel(plr, playerID, false, false, int32(channelID))
	}

	for i, v := range server.players {
		if v.ID() == playerID {
			continue
		} else if v.HasBuddy(playerID) {
			if changeChannel {
				server.players[i].Send(message.PacketBuddyChangeChannel(playerID, int32(channelID)))
				server.players[i].AddOnlineBuddy(playerID, name, int32(channelID))
			} else {
				// send online message card, then update buddy list
				server.players[i].Send(message.PacketBuddyOnlineStatus(playerID, int32(channelID)))
				server.players[i].AddOnlineBuddy(playerID, name, int32(channelID))
			}
		}
	}
}

func (server *Server) handlePlayerDisconnectNotifications(conn mnet.Server, reader mpacket.Reader) {
	playerID := reader.ReadInt32()
	name := reader.ReadString(reader.ReadInt16())

	for _, party := range server.parties {
		party.SetPlayerChannel(new(player.Data), playerID, false, true, 0)
	}

	for i, v := range server.players {
		if v.ID() == playerID {
			continue
		} else if v.HasBuddy(playerID) {
			server.players[i].AddOfflineBuddy(playerID, name)
		}
	}
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

		plr, err := server.players.getFromID(recepientID)

		if err != nil {
			return
		}

		plr.Send(message.PacketBuddyReceiveRequest(fromID, fromName, int32(channelID)))
	case 2:
		recepientID := reader.ReadInt32()
		fromID := reader.ReadInt32()
		fromName := reader.ReadString(reader.ReadInt16())
		channelID := reader.ReadByte()

		if channelID == server.id {
			return
		}

		plr, err := server.players.getFromID(recepientID)

		if err != nil {
			return
		}

		plr.AddOfflineBuddy(fromID, fromName)
		plr.Send(message.PacketBuddyOnlineStatus(fromID, int32(channelID)))
		plr.AddOnlineBuddy(fromID, fromName, int32(channelID))
	case 3:
		recepientID := reader.ReadInt32()
		fromID := reader.ReadInt32()
		channelID := reader.ReadByte()

		if channelID == server.id {
			return
		}

		plr, err := server.players.getFromID(recepientID)

		if err != nil {
			return
		}

		plr.RemoveBuddy(fromID)
	default:
		log.Println("Unknown buddy event type:", op)
	}
}

func (server Server) handleChatEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 0: // whispher
		recepientName := reader.ReadString(reader.ReadInt16())
		fromName := reader.ReadString(reader.ReadInt16())
		msg := reader.ReadString(reader.ReadInt16())
		channelID := reader.ReadByte()

		plr, err := server.players.getFromName(recepientName)

		if err != nil {
			return
		}

		plr.Send(message.PacketMessageWhisper(fromName, msg, channelID))

	case 1: // buddy
		fromName := reader.ReadString(reader.ReadInt16())
		idCount := reader.ReadByte()

		ids := make([]int32, int(idCount))

		for i := byte(0); i < idCount; i++ {
			ids[i] = reader.ReadInt32()
		}

		msg := reader.ReadString(reader.ReadInt16())

		for _, v := range ids {
			plr, err := server.players.getFromID(v)

			if err != nil {
				continue
			}

			plr.Send(message.PacketMessageBubblessChat(0, fromName, msg))
		}
	case 2: // party
		fromName := reader.ReadString(reader.ReadInt16())
		idCount := reader.ReadByte()

		ids := make([]int32, int(idCount))

		for i := byte(0); i < idCount; i++ {
			ids[i] = reader.ReadInt32()
		}

		msg := reader.ReadString(reader.ReadInt16())

		for _, v := range ids {
			plr, err := server.players.getFromID(v)

			if err != nil {
				continue
			}

			plr.Send(message.PacketMessageBubblessChat(1, fromName, msg))
		}
	case 3: // guild
		fromName := reader.ReadString(reader.ReadInt16())
		idCount := reader.ReadByte()

		ids := make([]int32, int(idCount))

		for i := byte(0); i < idCount; i++ {
			ids[i] = reader.ReadInt32()
		}

		msg := reader.ReadString(reader.ReadInt16())

		for _, v := range ids {
			plr, err := server.players.getFromID(v)

			if err != nil {
				continue
			}

			plr.Send(message.PacketMessageBubblessChat(2, fromName, msg))
		}
	default:
		log.Println("Unknown chat event type:", op)
	}
}

func (server *Server) handlePartyEvent(conn mnet.Server, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case 0:
		log.Println("Channel server should not receive party event message type: 0")
	case 1: // new party created
		channelID := reader.ReadByte()
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		mapID := reader.ReadInt32()
		job := reader.ReadInt32()
		level := reader.ReadInt32()
		name := reader.ReadString(reader.ReadInt16())

		plr, _ := server.players.getFromID(playerID)

		// TODO: Mystic door information needs to be sent here if the leader has an active door

		newParty := party.NewParty(partyID, plr, channelID, playerID, mapID, job, level, name, int32(server.id))

		server.parties[partyID] = &newParty

		if plr != nil {
			plr.SetParty(&newParty)
			plr.Send(message.PacketPartyCreate(1, -1, -1, pos.New(0, 0, 0)))
		}
	case 2: // leave party
		destroy := reader.ReadBool()
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()

		plr, _ := server.players.getFromID(playerID)

		if party, ok := server.parties[partyID]; ok {
			party.RemovePlayer(plr, playerID, false)

			if destroy {
				delete(server.parties, partyID)
			}
		}
	case 3: // accept
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		channelID := reader.ReadInt32()
		mapID := reader.ReadInt32()
		job := reader.ReadInt32()
		level := reader.ReadInt32()
		name := reader.ReadString(reader.ReadInt16())

		plr, _ := server.players.getFromID(playerID)

		if party, ok := server.parties[partyID]; ok {
			party.AddPlayer(plr, channelID, playerID, name, mapID, job, level)
		}
	case 4: // expel
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()

		plr, _ := server.players.getFromID(playerID)

		if party, ok := server.parties[partyID]; ok {
			party.RemovePlayer(plr, playerID, true)
		}
	case 5:
		partyID := reader.ReadInt32()
		playerID := reader.ReadInt32()
		job := reader.ReadInt32()
		level := reader.ReadInt32()
		reader.ReadString(reader.ReadInt16()) // name
		if party, ok := server.parties[partyID]; ok {
			party.UpdateJobLevel(playerID, job, level)
		}
	default:
		log.Println("Unkown party event type:", op)
	}
}
