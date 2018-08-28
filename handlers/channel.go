package handlers

import (
	"log"

	"github.com/Hucaru/Valhalla/connection"

	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/maplepacket"
)

// HandleChannelPacket - Purpose is to send a packet to the correct handler(s), packages should aim to use this function to communicate betweeen each other
func HandleChannelPacket(conn *connection.Channel, reader maplepacket.Reader) {
	switch reader.ReadByte() {
	case constants.RECV_PING:

	case constants.RECV_CHANNEL_PLAYER_LOAD:
		handlePlayerConnect(conn, reader)

	case constants.RECV_CHANNEL_USE_PORTAL:
		handleUsePortal(conn, reader)
	case constants.RECV_CHANNEL_REQUEST_TO_ENTER_CASH_SHOP:

	case constants.RECV_CHANNEL_PLAYER_MOVEMENT:
		handlePlayerMovement(conn, reader)

	case constants.RECV_CHANNEL_STANDARD_SKILL:
		handleStandardSkill(conn, reader)

	case constants.RECV_CHANNEL_RANGED_SKILL:
		handleRangedSkill(conn, reader)

	case constants.RECV_CHANNEL_MAGIC_SKILL:
		handleMagicSkill(conn, reader)

	case constants.RECV_CHANNEL_DMG_RECV:
		handleTakeDamage(conn, reader)

	case constants.RECV_CHANNEL_PLAYER_SEND_ALL_CHAT:
		handleAllChat(conn, reader)

	case constants.RECV_CHANNEL_SLASH_COMMANDS:
		handleSlashCommand(conn, reader)

	case constants.RECV_CHANNEL_CHARACTER_UI_WINDOW:
		handleUIWindow(conn, reader)

	case constants.RECV_CHANNEL_EMOTICON:
		handlePlayerEmoticon(conn, reader)

	case constants.RECV_CHANNEL_NPC_DIALOGUE:
		handleNPCChat(conn, reader)

	case constants.RECV_CHANNEL_NPC_DIALOGUE_CONTINUE:
		handleNPCChatContinue(conn, reader)

	case constants.RECV_CHANNEL_NPC_SHOP:
		handleNPCShop(conn, reader)

	case constants.RECV_CHANNEL_INV_MOVE_ITEM:
		handleMoveInventoryItem(conn, reader)

	case constants.RECV_CHANNEL_CHANGE_STAT:
		handleChangeStat(conn, reader)

	case constants.RECV_CHANNEL_PASSIVE_REGEN:
		handlePassiveRegen(conn, reader)

	case constants.RECV_CHANNEL_SKILL_UPDATE:
		handleUpdateSkillRecord(conn, reader)

	case constants.RECV_CHANNEL_SPECIAL_SKILL:
		handleSpecialSkill(conn, reader)

	case constants.RECV_CHANNEL_CHARACTER_INFO:
		handleRequestAvatarInfoWindow(conn, reader)

	case constants.RECV_CHANNEL_LIE_DETECTOR_RESULT:

	case constants.RECV_CHANNEL_PARTY_INFO:

	case constants.RECV_CHANNEL_GUILD_MANAGEMENT:

	case constants.RECV_CHANNEL_GUILD_REJECT:

	case constants.RECV_CHANNEL_ADD_BUDDY:

	case constants.RECV_CHANNEL_MOB_CONTROL:
		handleMobControl(conn, reader)

	case constants.RECV_CHANNEL_NPC_MOVEMENT:
		handleNPCMovement(conn, reader)

	default:
		log.Println("Unkown packet:", reader)
	}
}
