package client

import (
	"log"

	"github.com/Hucaru/Valhalla/channelServer/handlers/player"
	"github.com/Hucaru/Valhalla/channelServer/handlers/playerConn"
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
	case constants.RECV_CHANNEL_PLAYER_SEND_ALL_CHAT:
		player.HandlePlayerSendAllChat(reader, conn)
	case constants.RECV_CHANNEL_ADD_BUDDY:

	default:
		log.Println("UNKNOWN CHANNEL PACKET:", reader)
	}
}
