package client

import (
	"encoding/hex"
	"log"
	"strconv"
	"strings"

	"github.com/Hucaru/Valhalla/channelServer/handlers/player"
	"github.com/Hucaru/Valhalla/channelServer/handlers/playerConn"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/nx"
	"github.com/Hucaru/gopacket"
)

func HandlePacket(conn *playerConn.Conn, reader gopacket.Reader) {
	opcode := reader.ReadByte()

	switch opcode {
	case constants.RECV_CHANNEL_PLAYER_LOAD:
		player.HandlePlayerEnterGame(reader, conn)
	case constants.RECV_CHANNEL_MOVEMENT:
	case constants.RECV_CHANNEL_PLAYER_SEND_ALL_CHAT:
		handlePlayerSendAllChat(reader, conn)
	case constants.RECV_CHANNEL_ADD_BUDDY:
	default:
		log.Println("UNKNOWN CHANNEL PACKET:", reader)
	}
}

func handlePlayerSendAllChat(reader gopacket.Reader, conn *playerConn.Conn) {
	msg := reader.ReadString(int(reader.ReadInt16()))
	ind := strings.Index(msg, "!")

	if ind == 0 && conn.IsAdmin() {
		command := strings.SplitN(msg[ind+1:], " ", -1)
		switch command[0] {
		case "packet":
			packet := string(command[1])
			data, err := hex.DecodeString(packet)

			if err != nil {
				log.Println("Eror in decoding string for gm command packet:", packet)
				break
			}
			log.Println("Sent packet:", hex.EncodeToString(data))
			conn.Write(data)
		case "warp":
			val, err := strconv.Atoi(command[1])

			if err != nil {
				panic(err)
			}

			id := uint32(val)

			if _, ok := nx.Maps[id]; ok {
				player.ChangeMap(conn, uint32(id), 0, nx.GetRandomSpawnPortal(id), conn.GetCharacter().HP)
			} else {
				// check if player id in else if
			}
		default:
			log.Println("Unkown GM command", command)
		}

	}
}
