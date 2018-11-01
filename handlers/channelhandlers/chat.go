package channelhandlers

import (
	"encoding/hex"
	"log"
	"strconv"
	"strings"

	"github.com/Hucaru/Valhalla/nx"

	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/packets"

	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/mnet"
)

func chatSendAll(conn mnet.MConnChannel, reader maplepacket.Reader) {
	msg := reader.ReadString(int(reader.ReadInt16()))

	if strings.Index(msg, "/") == 0 && conn.GetAdminLevel() > 0 {
		gmCommand(conn, msg)
	} else {
		char := game.GetPlayerFromConn(conn).Char()
		game.SendToMap(char.CurrentMap, packets.MessageAllChat(char.ID, conn.GetAdminLevel() > 0, msg))
	}
}

func chatSlashCommand(conn mnet.MConnChannel, reader maplepacket.Reader) {

}

func gmCommand(conn mnet.MConnChannel, msg string) {
	ind := strings.Index(msg, "/")
	command := strings.SplitN(msg[ind+1:], " ", -1)

	switch command[0] {
	case "packet":
		if len(command) < 2 {
			return
		}
		packet := string(command[1])
		data, err := hex.DecodeString(packet)

		if err != nil {
			log.Println("Eror in decoding string for gm command packet:", packet)
			break
		}
		log.Println("Sent packet:", hex.EncodeToString(data))
		conn.Send(data)
	case "map":
		var val int
		var err error
		var mapName string

		if len(command) == 2 {
			val, err = strconv.Atoi(command[1])
			mapName = command[1]
		} else if len(command) == 3 {
			val, err = strconv.Atoi(command[2])
			mapName = command[2]
		}

		if err != nil {
			// Check to see if name matches pre-recorded
			switch mapName {
			// Maple island
			case "amherst":
				val = 1010000
			case "southperry":
				val = 60000
			// Victoria island
			case "lith":
				val = 104000000
			case "henesys":
				val = 100000000
			case "kerning":
				val = 103000000
			case "perion":
				val = 102000000
			case "ellinia":
				val = 101000000
			case "sleepy":
				val = 105040300
			case "gm":
				val = 180000000
			// Ossyria
			case "orbis":
				val = 200000000
			case "elnath":
				val = 211000000
			case "ludi":
				val = 220000000
			case "omega":
				val = 221000000
			case "aqua":
				val = 230000000
			// Misc
			case "balrog":
				val = 105090900
			default:
				return
			}
		}

		mapID := int32(val)

		if _, ok := nx.Maps[mapID]; !ok {
			return
		}

		player := game.GetPlayerFromConn(conn)
		p, id := game.GetRandomSpawnPortal(mapID)
		player.ChangeMap(mapID, p, id)

	case "notice":
		if len(command) < 2 {
			return
		}
		char := game.GetPlayerFromConn(conn).Char()
		game.SendToMap(char.CurrentMap, packets.MessageNotice(strings.Join(command[1:], " ")))
	default:
		log.Println("Unkown GM command:", msg)
	}
}
