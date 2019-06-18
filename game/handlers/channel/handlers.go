package channel

import (
	"log"

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func HandlePacket(conn mnet.Client, reader mpacket.Reader) {
	switch mpacket.Opcode(reader.ReadByte()) {
	case opcode.RecvPing:
	case opcode.RecvChannelPlayerLoad:
		playerConnect(conn, reader)
	case opcode.RecvChannelUserPortal:
		playerUsePortal(conn, reader)
	case opcode.RecvChannelEnterCashShop:
	case opcode.RecvChannelPlayerMovement:
		playerMovement(conn, reader)
	case opcode.RecvChannelPlayerStand:
		playerStand(conn, reader)
	case opcode.RecvChannelPlayerUserChair:
		playerUseChair(conn, reader)
	case opcode.RecvChannelMeleeSkill:
		playerMeleeSkill(conn, reader)
	case opcode.RecvChannelRangedSkill:
		playerRangedSkill(conn, reader)
	case opcode.RecvChannelMagicSkill:
		playerMagicSkill(conn, reader)
	case opcode.RecvChannelDmgRecv:
		playerTakeDamage(conn, reader)
	case opcode.RecvChannelPlayerSendAllChat:
		chatSendAll(conn, reader)
	case opcode.RecvChannelSlashCommands:
		chatSlashCommand(conn, reader)
	case opcode.RecvChannelCharacterUIWindow:
		handleUIWindow(conn, reader)
	case opcode.RecvChannelEmote:
		playerEmote(conn, reader)
	case opcode.RecvChannelNpcDialogue:
		npcChatStart(conn, reader)
	case opcode.RecvChannelNpcDialogueContinue:
		npcChatContinue(conn, reader)
	case opcode.RecvChannelNpcShop:
	case opcode.RecvChannelInvMoveItem:
		playerMoveInventoryItem(conn, reader)
	case opcode.RecvChannelAddStatPoint:
		playerAddStatPoint(conn, reader)
	case opcode.RecvChannelPassiveRegen:
		playerPassiveRegen(conn, reader)
	case opcode.RecvChannelAddSkillPoint:
		playerAddSkillPoint(conn, reader)
	case opcode.RecvChannelSpecialSkill:
		playerSpecialSkill(conn, reader)
	case opcode.RecvChannelCharacterInfo:
		playerRequestAvatarInfoWindow(conn, reader)
	case opcode.RecvChannelLieDetectorResult:
	case opcode.RecvChannelPartyInfo:
	case opcode.RecvChannelGuildManagement:
	case opcode.RecvChannelGuildReject:
	case opcode.RecvChannelAddBuddy:
	case opcode.RecvChannelMobControl:
		mobControl(conn, reader)
	case opcode.RecvChannelNpcMovement:
		npcMovement(conn, reader)
	default:
		log.Println("Unkown packet:", reader)
	}
}
