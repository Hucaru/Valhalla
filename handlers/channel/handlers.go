package channel

import (
	"log"

	opcodes "github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func HandlePacket(conn mnet.MConnChannel, reader mpacket.Reader) {
	switch mpacket.Opcode(reader.ReadByte()) {
	case opcodes.RecvPing:
	case opcodes.RecvChannelPlayerLoad:
		playerConnect(conn, reader)
	case opcodes.RecvChannelUserPortal:
		playerUsePortal(conn, reader)
	case opcodes.RecvChannelEnterCashShop:
	case opcodes.RecvChannelPlayerMovement:
		playerMovement(conn, reader)
	case opcodes.RecvChannelPlayerStand:
		playerStand(conn, reader)
	case opcodes.RecvChannelPlayerUserChair:
		playerUseChair(conn, reader)
	case opcodes.RecvChannelMeleeSkill:
		playerMeleeSkill(conn, reader)
	case opcodes.RecvChannelRangedSkill:
		playerRangedSkill(conn, reader)
	case opcodes.RecvChannelMagicSkill:
		playerMagicSkill(conn, reader)
	case opcodes.RecvChannelDmgRecv:
		playerTakeDamage(conn, reader)
	case opcodes.RecvChannelPlayerSendAllChat:
		chatSendAll(conn, reader)
	case opcodes.RecvChannelSlashCommands:
		chatSlashCommand(conn, reader)
	case opcodes.RecvChannelCharacterUIWindow:
		handleUIWindow(conn, reader)
	case opcodes.RecvChannelEmote:
		playerEmote(conn, reader)
	case opcodes.RecvChannelNpcDialogue:
		npcChatStart(conn, reader)
	case opcodes.RecvChannelNpcDialogueContinue:
		npcChatContinue(conn, reader)
	case opcodes.RecvChannelNpcShop:
	case opcodes.RecvChannelInvMoveItem:
		playerMoveInventoryItem(conn, reader)
	case opcodes.RecvChannelAddStatPoint:
		playerAddStatPoint(conn, reader)
	case opcodes.RecvChannelPassiveRegen:
		playerPassiveRegen(conn, reader)
	case opcodes.RecvChannelAddSkillPoint:
		playerAddSkillPoint(conn, reader)
	case opcodes.RecvChannelSpecialSkill:
		playerSpecialSkill(conn, reader)
	case opcodes.RecvChannelCharacterInfo:
		playerRequestAvatarInfoWindow(conn, reader)
	case opcodes.RecvChannelLieDetectorResult:
	case opcodes.RecvChannelPartyInfo:
	case opcodes.RecvChannelGuildManagement:
	case opcodes.RecvChannelGuildReject:
	case opcodes.RecvChannelAddBuddy:
	case opcodes.RecvChannelMobControl:
		mobControl(conn, reader)
	case opcodes.RecvChannelNpcMovement:
		npcMovement(conn, reader)
	default:
		log.Println("Unkown packet:", reader)
	}
}
