package player

import (
	"encoding/hex"
	"log"
	"strconv"

	"github.com/Hucaru/Valhalla/channelServer/maps"
	"github.com/Hucaru/Valhalla/channelServer/playerConn"
	"github.com/Hucaru/Valhalla/common/nx"
)

func dealWithCommand(conn *playerConn.Conn, command []string) {
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
			return
		}

		id := uint32(val)

		if _, exist := nx.Maps[id]; exist {
			portal := maps.GetRandomSpawnPortal(id)

			if len(command) > 2 {
				pos, err := strconv.Atoi(command[2])

				if err == nil {
					portal = maps.GetPortalByID(uint32(id), byte(pos))
				}
			}

			maps.PlayerChangeMap(conn, uint32(id), portal.ID, conn.GetCharacter().GetHP())
		} else {
			// check if player id in else if
		}
	case "job":
		val, err := strconv.Atoi(command[1])

		if err != nil {
			return
		}

		PlayerChangeJob(conn, uint16(val))
	case "level":
		val, err := strconv.Atoi(command[1])

		if err != nil {
			return
		}

		PlayerSetLevel(conn, byte(val))
	case "exp":
		val, err := strconv.Atoi(command[1])

		if err != nil {
			return
		}

		PlayerAddExp(conn, uint32(val))
	case "hp":
		val, err := strconv.Atoi(command[1])

		if err != nil {
			return
		}

		PlayerSetHP(conn, uint16(val))
	case "mp":
		val, err := strconv.Atoi(command[1])

		if err != nil {
			return
		}

		PlayerSetMP(conn, uint16(val))
	default:
		log.Println("Unkown GM command", command)
	}
}
