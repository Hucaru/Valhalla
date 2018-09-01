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
	case constants.RecvPing:

	case constants.RecvChannelPlayerLoad:
		handlePlayerConnect(conn, reader)

	case constants.RecvChannelUserPortal:
		handleUsePortal(conn, reader)
	case constants.RecvChannelEnterCashShop:

	case constants.RecvChannelPlayerMovement:
		handlePlayerMovement(conn, reader)

	case constants.RecvChannelStandardSkill:
		handleStandardSkill(conn, reader)

	case constants.RecvChannelRangedSkill:
		handleRangedSkill(conn, reader)

	case constants.RecvChannelMagicSkill:
		handleMagicSkill(conn, reader)

	case constants.RecvChannelDmgRecv:
		handleTakeDamage(conn, reader)

	case constants.RecvChannelPlayerSendAllChat:
		handleAllChat(conn, reader)

	case constants.RecvChannelSlashCommands:
		handleSlashCommand(conn, reader)

	case constants.RecvChannelCharacterUIWindow:
		handleUIWindow(conn, reader)

	case constants.RecvChannelEmoticon:
		handlePlayerEmoticon(conn, reader)

	case constants.RecvChannelNpcDialogue:
		handleNPCChat(conn, reader)

	case constants.RecvChannelNpcDialogueContinue:
		handleNPCChatContinue(conn, reader)

	case constants.RecvChannelNpcShop:
		handleNPCShop(conn, reader)

	case constants.RecvChannelInvMoveItem:
		handleMoveInventoryItem(conn, reader)

	case constants.RecvChannelChangeStat:
		handleChangeStat(conn, reader)

	case constants.RecvChannelPassiveRegen:
		handlePassiveRegen(conn, reader)

	case constants.RecvChannelSkillUpdate:
		handleUpdateSkillRecord(conn, reader)

	case constants.RecvChannelSpecialSkill:
		handleSpecialSkill(conn, reader)

	case constants.RecvChannelCharacterInfo:
		handleRequestAvatarInfoWindow(conn, reader)

	case constants.RecvChannelLieDetectorResult:

	case constants.RecvChannelPartyInfo:

	case constants.RecvChannelGuildManagement:

	case constants.RecvChannelGuildReject:

	case constants.RecvChannelAddBuddy:

	case constants.RecvChannelMobControl:
		handleMobControl(conn, reader)

	case constants.RecvChannelNpcMovement:
		handleNPCMovement(conn, reader)

	default:
		log.Println("Unkown packet:", reader)
	}
}
