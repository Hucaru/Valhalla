package server

import (
	"log"

	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/message"
)

// HandleClientPacket data
func (server *ChannelServer) HandleClientPacket(conn mnet.Client, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.RecvPing:
	case opcode.RecvChannelPlayerLoad:
		server.playerConnect(conn, reader)
	case opcode.RecvCHannelChangeChannel:
		server.playerChangeChannel(conn, reader)
	case opcode.RecvChannelUserPortal:
		// This opcode is used for revival UI as well.
		server.playerUsePortal(conn, reader)
	case opcode.RecvChannelEnterCashShop:
		conn.Send(message.PacketMessageDialogueBox("Shop not implemented"))
	case opcode.RecvChannelPlayerMovement:
		server.playerMovement(conn, reader)
	case opcode.RecvChannelPlayerStand:
		server.playerStand(conn, reader)
	case opcode.RecvChannelPlayerUseChair:
		server.playerUseChair(conn, reader)
	case opcode.RecvChannelMeleeSkill:
		server.playerMeleeSkill(conn, reader)
	case opcode.RecvChannelRangedSkill:
		// server.playerRangedSkill(conn, reader)
	case opcode.RecvChannelMagicSkill:
		// server.playerMagicSkill(conn, reader)
	case opcode.RecvChannelDmgRecv:
		server.playerTakeDamage(conn, reader)
	case opcode.RecvChannelPlayerSendAllChat:
		server.chatSendAll(conn, reader)
	case opcode.RecvChannelGroupChat:
		server.chatGroup(conn, reader)
	case opcode.RecvChannelSlashCommands:
		server.chatSlashCommand(conn, reader)
	case opcode.RecvChannelCharacterUIWindow:
		server.roomWindow(conn, reader)
	case opcode.RecvChannelEmote:
		server.playerEmote(conn, reader)
	case opcode.RecvChannelNpcDialogue:
		server.npcChatStart(conn, reader)
	case opcode.RecvChannelNpcDialogueContinue:
		server.npcChatContinue(conn, reader)
	case opcode.RecvChannelNpcShop:
		server.npcShop(conn, reader)
	case opcode.RecvChannelInvMoveItem:
		server.playerMoveInventoryItem(conn, reader)
	case opcode.RecvChannelInvUseItem:
		server.playerUseInventoryItem(conn, reader)
	case opcode.RecvChannelAddStatPoint:
		server.playerAddStatPoint(conn, reader)
	case opcode.RecvChannelPassiveRegen:
		server.playerPassiveRegen(conn, reader)
	case opcode.RecvChannelAddSkillPoint:
		server.playerAddSkillPoint(conn, reader)
	case opcode.RecvChannelSpecialSkill:
		// server.playerSpecialSkill(conn, reader)
	case opcode.RecvChannelCharacterInfo:
		server.playerRequestAvatarInfoWindow(conn, reader)
	case opcode.RecvChannelLieDetectorResult:
	case opcode.RecvChannelPartyInfo:
	case opcode.RecvChannelGuildManagement:
	case opcode.RecvChannelGuildReject:
	case opcode.RecvChannelBuddyOperation:
		server.playerBuddyOperation(conn, reader)
	case opcode.RecvChannelUseMysticDoor:
		server.playerUseMysticDoor(conn, reader)
	case opcode.RecvChannelMobControl:
		server.mobControl(conn, reader)
	case opcode.RecvChannelDistance:
		server.mobDistance(conn, reader)
	case opcode.RecvChannelNpcMovement:
		server.npcMovement(conn, reader)
	case opcode.RecvChannelBoatMap:
		// [mapID int32][? byte]
	default:
		log.Println("UNKNOWN CLIENT PACKET:", reader)
	}
}
