package game

import (
	"log"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/game/entity"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

// Channel server state
type Channel struct {
	// maps
	// players
}

// Initialise the server
func (server *Channel) Initialise(chan func()) {

}

// ClientConnected to server
func (server *Channel) ClientConnected(conn mnet.Client, keyRecv, keySend []byte) {
	conn.Send(entity.PacketClientHandshake(constant.MapleVersion, keyRecv, keySend))
}

func (server *Channel) ClientDisconnected(conn mnet.Client) {

}

// HandleClientPacket from client
func (server *Channel) HandleClientPacket(conn mnet.Client, reader mpacket.Reader) {
	switch mpacket.Opcode(reader.ReadByte()) {
	case opcode.RecvPing:
	case opcode.RecvChannelPlayerLoad:
		server.playerConnect(conn, reader)
	case opcode.RecvChannelUserPortal:
		// server.playerUsePortal(conn, reader)
	case opcode.RecvChannelEnterCashShop:
	case opcode.RecvChannelPlayerMovement:
		// server.playerMovement(conn, reader)
	case opcode.RecvChannelPlayerStand:
		// server.playerStand(conn, reader)
	case opcode.RecvChannelPlayerUserChair:
		// server.playerUseChair(conn, reader)
	case opcode.RecvChannelMeleeSkill:
		// server.playerMeleeSkill(conn, reader)
	case opcode.RecvChannelRangedSkill:
		// server.playerRangedSkill(conn, reader)
	case opcode.RecvChannelMagicSkill:
		// server.playerMagicSkill(conn, reader)
	case opcode.RecvChannelDmgRecv:
		// server.playerTakeDamage(conn, reader)
	case opcode.RecvChannelPlayerSendAllChat:
		// server.chatSendAll(conn, reader)
	case opcode.RecvChannelSlashCommands:
		// server.chatSlashCommand(conn, reader)
	case opcode.RecvChannelCharacterUIWindow:
		// server.handleUIWindow(conn, reader)
	case opcode.RecvChannelEmote:
		// server.playerEmote(conn, reader)
	case opcode.RecvChannelNpcDialogue:
		// server.npcChatStart(conn, reader)
	case opcode.RecvChannelNpcDialogueContinue:
		// server.npcChatContinue(conn, reader)
	case opcode.RecvChannelNpcShop:
	case opcode.RecvChannelInvMoveItem:
		// server.playerMoveInventoryItem(conn, reader)
	case opcode.RecvChannelAddStatPoint:
		// server.playerAddStatPoint(conn, reader)
	case opcode.RecvChannelPassiveRegen:
		// server.playerPassiveRegen(conn, reader)
	case opcode.RecvChannelAddSkillPoint:
		// server.playerAddSkillPoint(conn, reader)
	case opcode.RecvChannelSpecialSkill:
		// server.playerSpecialSkill(conn, reader)
	case opcode.RecvChannelCharacterInfo:
		// server.playerRequestAvatarInfoWindow(conn, reader)
	case opcode.RecvChannelLieDetectorResult:
	case opcode.RecvChannelPartyInfo:
	case opcode.RecvChannelGuildManagement:
	case opcode.RecvChannelGuildReject:
	case opcode.RecvChannelAddBuddy:
	case opcode.RecvChannelMobControl:
		// server.mobControl(conn, reader)
	case opcode.RecvChannelNpcMovement:
		// server.npcMovement(conn, reader)
	default:
		log.Println("Unkown packet:", reader)
	}
}

func (server *Channel) playerConnect(conn mnet.Client, reader mpacket.Reader) {

}

// HandleServerPacket from client
func (server *Channel) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {

}
