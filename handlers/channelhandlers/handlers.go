package channelhandlers

import (
	"log"

	"github.com/Hucaru/Valhalla/consts/opcodes"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/mnet"
)

func HandlePacket(conn mnet.MConnChannel, reader maplepacket.Reader) {
	switch reader.ReadByte() {
	case opcodes.Recv.Ping:

	case opcodes.Recv.ChannelPlayerLoad:
		playerConnect(conn, reader)
	case opcodes.Recv.ChannelUserPortal:
		playerUsePortal(conn, reader)
	case opcodes.Recv.ChannelEnterCashShop:
	case opcodes.Recv.ChannelPlayerMovement:
		playerMovement(conn, reader)
	case opcodes.Recv.ChannelMeleeSkill:
		playerMeleeSkill(conn, reader)
	case opcodes.Recv.ChannelRangedSkill:
		playerRangedSkill(conn, reader)
	case opcodes.Recv.ChannelMagicSkill:
		playerMagicSkill(conn, reader)
	case opcodes.Recv.ChannelDmgRecv:
		playerTakeDamage(conn, reader)
	case opcodes.Recv.ChannelPlayerSendAllChat:
		chatSendAll(conn, reader)
	case opcodes.Recv.ChannelSlashCommands:
	case opcodes.Recv.ChannelCharacterUIWindow:
	case opcodes.Recv.ChannelEmoticon:
	case opcodes.Recv.ChannelNpcDialogue:
	case opcodes.Recv.ChannelNpcDialogueContinue:
	case opcodes.Recv.ChannelNpcShop:
	case opcodes.Recv.ChannelInvMoveItem:
	case opcodes.Recv.ChannelChangeStat:
	case opcodes.Recv.ChannelPassiveRegen:
	case opcodes.Recv.ChannelSkillUpdate:
	case opcodes.Recv.ChannelSpecialSkill:
		playerSpecialSkill(conn, reader)
	case opcodes.Recv.ChannelCharacterInfo:
	case opcodes.Recv.ChannelLieDetectorResult:
	case opcodes.Recv.ChannelPartyInfo:
	case opcodes.Recv.ChannelGuildManagement:
	case opcodes.Recv.ChannelGuildReject:
	case opcodes.Recv.ChannelAddBuddy:
	case opcodes.Recv.ChannelMobControl:
		mobControl(conn, reader)
	case opcodes.Recv.ChannelNpcMovement:
		npcMovement(conn, reader)
	default:
		log.Println("Unkown packet:", reader)
	}
}
