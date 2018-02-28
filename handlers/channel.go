package handlers

import (
	"log"

	"github.com/Hucaru/Valhalla/player"

	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/gopacket"
)

func HandleChannelPacket(conn *clientChanConn, reader gopacket.Reader) {
	switch reader.ReadByte() {
	case constants.RECV_CHANNEL_PLAYER_LOAD:
		player.HandleConnect(conn, reader)
		// maps.HandleNewPlayer(conn) // use data package to get character data for avatar

	case constants.RECV_CHANNEL_MOVEMENT:
		//

	case constants.RECV_CHANNEL_MELEE_SKILL:
		//

	case constants.RECV_CHANNEL_USE_PORTAL:
		//

	case constants.RECV_CHANNEL_REQUEST_TO_ENTER_CASH_SHOP:
		//

	case constants.RECV_CHANNEL_DMG_RECV:
		//

	case constants.RECV_CHANNEL_PLAYER_SEND_ALL_CHAT:
		//

	case constants.RECV_CHANNEL_EMOTION:
		//

	case constants.RECV_CHANNEL_NPC_DIALOGUE:
		//

	case constants.RECV_CHANNEL_CHANGE_STAT:
		//

	case constants.RECV_CHANNEL_PASSIVE_REGEN:
		//

	case constants.RECV_CHANNEL_SKILL_UPDATE:
		//

	case constants.RECV_CHANNEL_SPECIAL_SKILL_USAGE:
		//

	case constants.RECV_CHANNEL_DOUBLE_CLICK_CHARACTER:
		//

	case constants.RECV_CHANNEL_LIE_DETECTOR_RESULT:
		//

	case constants.RECV_CHANNEL_PARTY_INFO:
		//

	case constants.RECV_CHANNEL_GUILD_MANAGEMENT:
		//

	case constants.RECV_CHANNEL_GUILD_REJECT:
		//

	case constants.RECV_CHANNEL_ADD_BUDDY:
		//

	case constants.RECV_CHANNEL_MOB_MOVEMENT:
		//

	default:
		log.Println("Unkown packet:", reader)
	}
}
