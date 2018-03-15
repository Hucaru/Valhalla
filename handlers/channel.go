package handlers

import (
	"log"

	"github.com/Hucaru/Valhalla/command"
	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/maps"
	"github.com/Hucaru/Valhalla/message"
	"github.com/Hucaru/Valhalla/player"
	"github.com/Hucaru/Valhalla/skills"

	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/gopacket"
)

// HandleChannelPacket - Purpose is to send a packet to the correct handler(s), packages should aim to use this function to communicate betweeen each other
func HandleChannelPacket(conn *connection.ClientChanConn, reader gopacket.Reader) {
	switch reader.ReadByte() {
	case constants.RECV_PING:
		// Is client expecting a pong?

	case constants.RECV_CHANNEL_PLAYER_LOAD:
		mapID := player.HandleConnect(conn, reader)
		maps.PlayerEnterMap(conn, mapID)
		maps.RegisterNewPlayerCallback(conn)

	case constants.RECV_CHANNEL_USE_PORTAL:
		maps.HandlePlayerUsePortal(conn, reader)

	case constants.RECV_CHANNEL_REQUEST_TO_ENTER_CASH_SHOP:
		message.HandleCashShopButton(conn, reader)

	case constants.RECV_CHANNEL_MOVEMENT:
		mapID, p := player.HandleMovement(conn, reader)
		maps.SendPacketToMap(mapID, p)

	case constants.RECV_CHANNEL_STANDARD_SKILL:
		mapID, p, damages := skills.HandleStandardSkill(conn, reader)
		maps.SendPacketToMap(mapID, p)
		exp := maps.DamageMobs(mapID, conn, damages)

		for k, v := range exp {
			player.GiveExp(k, v)
		}

	case constants.RECV_CHANNEL_RANGED_SKILL:
		mapID, p, damages := skills.HandleRangedSkill(conn, reader)
		maps.SendPacketToMap(mapID, p)
		exp := maps.DamageMobs(mapID, conn, damages)

		for k, v := range exp {
			player.GiveExp(k, v)
		}

	case constants.RECV_CHANNEL_MAGIC_SKILL:
		mapID, p, damages := skills.HandleMagicSkill(conn, reader)
		maps.SendPacketToMap(mapID, p)
		exp := maps.DamageMobs(mapID, conn, damages)

		for k, v := range exp {
			player.GiveExp(k, v)
		}

	case constants.RECV_CHANNEL_DMG_RECV:
		mapID, p := player.HandleTakeDamage(conn, reader)
		maps.SendPacketToMapExcept(mapID, p, conn)

	case constants.RECV_CHANNEL_PLAYER_SEND_ALL_CHAT:
		mapID, text, isCommand, p := message.HandleAllChat(conn, reader)
		if isCommand {
			command.HandleCommand(conn, text)
			return
		}
		maps.SendPacketToMap(mapID, p)

	case constants.RECV_CHANNEL_EMOTION:
		maps.HandlePlayerEmotion(conn, reader)

	case constants.RECV_CHANNEL_NPC_DIALOGUE:
		// npc.HandleNpcDialogue(conn, reader) // Goes off to the script engine.

	case constants.RECV_CHANNEL_CHANGE_STAT:
		player.HandleChangeStat(conn, reader)

	case constants.RECV_CHANNEL_PASSIVE_REGEN:
		player.HandlePassiveRegen(conn, reader)

	case constants.RECV_CHANNEL_SKILL_UPDATE:
		player.HandleUpdateSkillRecord(conn, reader)

	case constants.RECV_CHANNEL_SPECIAL_SKILL_USAGE:
		mapID, p := skills.HandleSpecialSkill(conn, reader)
		maps.SendPacketToMap(mapID, p)
		// Send buff to party

	case constants.RECV_CHANNEL_DOUBLE_CLICK_CHARACTER:
		// player.HandleRequestAvatarInfoWindow(conn, reader)

	case constants.RECV_CHANNEL_LIE_DETECTOR_RESULT:
		// send to the anti cheat thread

	case constants.RECV_CHANNEL_PARTY_INFO:
		//

	case constants.RECV_CHANNEL_GUILD_MANAGEMENT:
		//

	case constants.RECV_CHANNEL_GUILD_REJECT:
		//

	case constants.RECV_CHANNEL_ADD_BUDDY:
		//

	case constants.RECV_CHANNEL_MOB_MOVEMENT:
		maps.HandleMobMovement(conn, reader)

	default:
		log.Println("Unkown packet:", reader)
	}
}
