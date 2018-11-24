package channel

import (
	"log"

	"github.com/Hucaru/Valhalla/consts/opcodes"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/mnet"
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
	case opcodes.RecvChannelCharacterUIWindow:
	case opcodes.RecvChannelEmote:
		playerEmote(conn, reader)
	case opcodes.RecvChannelNpcDialogue:
	case opcodes.RecvChannelNpcDialogueContinue:
	case opcodes.RecvChannelNpcShop:
	case opcodes.RecvChannelInvMoveItem:
	case opcodes.RecvChannelChangeStat:
	case opcodes.RecvChannelPassiveRegen:
		playerPassiveRegen(conn, reader)
	case opcodes.RecvChannelSkillUpdate:
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
