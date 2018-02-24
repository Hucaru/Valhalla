package client

import (
	"log"

	"github.com/Hucaru/Valhalla/channelServer/mobs"
	"github.com/Hucaru/Valhalla/channelServer/npc"
	"github.com/Hucaru/Valhalla/channelServer/player"
	"github.com/Hucaru/Valhalla/channelServer/playerConn"
	"github.com/Hucaru/Valhalla/channelServer/skills"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

func HandlePacket(conn *playerConn.Conn, reader gopacket.Reader) {
	opcode := reader.ReadByte()

	switch opcode {
	case constants.RECV_PING:
		// handle ping, does client expect pong?
	case constants.RECV_CHANNEL_PLAYER_LOAD:
		player.HandlePlayerEnterGame(reader, conn)
	case constants.RECV_CHANNEL_USE_PORTAL:
		player.HandlePlayerUsePortal(reader, conn)
	case constants.RECV_CHANNEL_MOVEMENT:
		player.HandlePlayerMovement(reader, conn)
	case constants.RECV_CHANNEL_SKILL_USAGE:
		skills.HandlePlayerSkillUsage(reader, conn)
	case constants.RECV_CHANNEL_PLAYER_SEND_ALL_CHAT:
		player.HandlePlayerSendAllChat(reader, conn)
	case constants.RECV_CHANNEL_EMOTION:
		player.HandlePlayerEmotion(reader, conn)
	case constants.RECV_CHANNEL_NPC_DIALOGUE:
		npc.HandleNPCDialogue(reader, conn)
	case constants.RECV_CHANNEL_CHANGE_STAT:
		player.HandlePlayerChangeStat(reader, conn)
	case constants.RECV_CHANNEL_PASSIVE_REGEN:
		player.HandlePlayerPassiveRegen(reader, conn)
	case constants.RECV_CHANNEL_SKILL_UPDATE:
		player.HandlePlayerSkillUpdate(reader, conn)
	case constants.RECV_CHANNEL_SPECIAL_SKILL_USAGE:
		skills.HandlePlayerSpecialSkillUsage(reader, conn)
	case constants.RECV_CHANNEL_MOB_MOVEMENT:
		mobs.HandleMovement(reader, conn)
	default:
		log.Println("UNKNOWN CHANNEL PACKET:", reader)
	}
}
