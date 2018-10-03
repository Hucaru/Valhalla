package handlers

import (
	"log"

	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/consts/opcodes"
	"github.com/Hucaru/Valhalla/maplepacket"
)

// HandleChannelPacket - Purpose is to send a packet to the correct handler(s), packages should aim to use this function to communicate betweeen each other
func HandleChannelPacket(conn *connection.Channel, reader maplepacket.Reader) {
	switch reader.ReadByte() {
	case opcodes.Recv.Ping:

	case opcodes.Recv.ChannelPlayerLoad:
		handlePlayerConnect(conn, reader)

	case opcodes.Recv.ChannelUserPortal:
		handleUsePortal(conn, reader)
	case opcodes.Recv.ChannelEnterCashShop:

	case opcodes.Recv.ChannelPlayerMovement:
		handlePlayerMovement(conn, reader)

	case opcodes.Recv.ChannelStandardSkill:
		handleStandardSkill(conn, reader)

	case opcodes.Recv.ChannelRangedSkill:
		handleRangedSkill(conn, reader)

	case opcodes.Recv.ChannelMagicSkill:
		handleMagicSkill(conn, reader)

	case opcodes.Recv.ChannelDmgRecv:
		handleTakeDamage(conn, reader)

	case opcodes.Recv.ChannelPlayerSendAllChat:
		handleAllChat(conn, reader)

	case opcodes.Recv.ChannelSlashCommands:
		handleSlashCommand(conn, reader)

	case opcodes.Recv.ChannelCharacterUIWindow:
		handleUIWindow(conn, reader)

	case opcodes.Recv.ChannelEmoticon:
		handlePlayerEmoticon(conn, reader)

	case opcodes.Recv.ChannelNpcDialogue:
		handleNPCChat(conn, reader)

	case opcodes.Recv.ChannelNpcDialogueContinue:
		handleNPCChatContinue(conn, reader)

	case opcodes.Recv.ChannelNpcShop:
		handleNPCShop(conn, reader)

	case opcodes.Recv.ChannelInvMoveItem:
		handleMoveInventoryItem(conn, reader)

	case opcodes.Recv.ChannelChangeStat:
		handleChangeStat(conn, reader)

	case opcodes.Recv.ChannelPassiveRegen:
		handlePassiveRegen(conn, reader)

	case opcodes.Recv.ChannelSkillUpdate:
		handleUpdateSkillRecord(conn, reader)

	case opcodes.Recv.ChannelSpecialSkill:
		handleSpecialSkill(conn, reader)

	case opcodes.Recv.ChannelCharacterInfo:
		handleRequestAvatarInfoWindow(conn, reader)

	case opcodes.Recv.ChannelLieDetectorResult:

	case opcodes.Recv.ChannelPartyInfo:

	case opcodes.Recv.ChannelGuildManagement:

	case opcodes.Recv.ChannelGuildReject:

	case opcodes.Recv.ChannelAddBuddy:

	case opcodes.Recv.ChannelMobControl:
		handleMobControl(conn, reader)

	case opcodes.Recv.ChannelNpcMovement:
		handleNPCMovement(conn, reader)

	default:
		log.Println("Unkown packet:", reader, opcodes.Recv.ChannelPlayerLoad)
	}
}
