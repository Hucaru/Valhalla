package handlers

import (
	"log"

	"github.com/Hucaru/Valhalla/player"

	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/gopacket"
)

// HandleChannelPacket - Purpose is to send a packet to the correct handler(s), packages should aim to use this function to communicate betweeen each other
func HandleChannelPacket(conn *clientChanConn, reader gopacket.Reader) {
	switch reader.ReadByte() {
	case constants.RECV_PING:
		// Is client expecting a pong?

	case constants.RECV_CHANNEL_PLAYER_LOAD:
		player.HandleConnect(conn, reader)
		// maps.HandleNewPlayer(conn) // use data package to get character data for avatar

	case constants.RECV_CHANNEL_MOVEMENT:
		// p := player.HandleMovementData(conn, reader)
		// maps.SendPacketToMap(mapID, p) // if len(p) < 1 then don't bother sending as it's an empty packet

	case constants.RECV_CHANNEL_MELEE_SKILL:
		// p := skills.HandleMeleeSkill(conn, reader)
		// maps.SendPacketToMap(mapID, p)

	case constants.RECV_CHANNEL_USE_PORTAL:
		// maps.HandleUsePortal(conn, reader)

	case constants.RECV_CHANNEL_REQUEST_TO_ENTER_CASH_SHOP:
		//

	case constants.RECV_CHANNEL_DMG_RECV:
		// mapID, p := player.HandleReceivesDmg(conn, reader)
		// maps.SendPacketToMap(mapID, p)

	case constants.RECV_CHANNEL_PLAYER_SEND_ALL_CHAT:
		// mapID, text := chat.HandleAllChat(conn, reader)
		// if text[0] == "!" && conn.IsAdmin()  {
		// command.Handle(conn, text)
		// return
		// }
		// p := chat.CreateAllChatPacket(text)
		// maps.SendPacketToMap(mapID, p)

	case constants.RECV_CHANNEL_EMOTION:
		// maps.HandleCharacterEmotion(conn, reader)

	case constants.RECV_CHANNEL_NPC_DIALOGUE:
		// npc.HandleNpcDialogue(conn, reader) // Goes off to the script engine.

	case constants.RECV_CHANNEL_CHANGE_STAT:
		// player.HandleStatChange(conn, reader)

	case constants.RECV_CHANNEL_PASSIVE_REGEN:
		// player.HandlePassiveRegen(conn, reader)

	case constants.RECV_CHANNEL_SKILL_UPDATE:
		// player.HandleUpdateSkillRecord(conn, reader)

	case constants.RECV_CHANNEL_SPECIAL_SKILL_USAGE: // is this ranged or magic attack?
		//

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
		// maps.HandleMobMovement(conn, reader) // Maps owns mobs and therefore deals with them

	default:
		log.Println("Unkown packet:", reader)
	}
}
