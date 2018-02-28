package client

import (
	"log"

	"github.com/Hucaru/Valhalla/channelServer/player"
	"github.com/Hucaru/Valhalla/channelServer/playerConn"
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
	case constants.RECV_CHANNEL_MOVEMENT:
	case constants.RECV_CHANNEL_SKILL_USAGE:
	case constants.RECV_CHANNEL_DMG_RECV:
	case constants.RECV_CHANNEL_PLAYER_SEND_ALL_CHAT:
	case constants.RECV_CHANNEL_EMOTION:
	case constants.RECV_CHANNEL_NPC_DIALOGUE:
	case constants.RECV_CHANNEL_CHANGE_STAT:
	case constants.RECV_CHANNEL_PASSIVE_REGEN:
	case constants.RECV_CHANNEL_SKILL_UPDATE:
	case constants.RECV_CHANNEL_SPECIAL_SKILL_USAGE:
	case constants.RECV_CHANNEL_MOB_MOVEMENT:
	default:
		log.Println("UNKNOWN CHANNEL PACKET:", reader)
	}
}
