package command

import (
	"encoding/hex"
	"log"
	"strconv"
	"strings"

	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/Valhalla/maps"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/gopacket"
)

type clientConn interface {
	GetUserID() uint32
	Write(gopacket.Packet) error
}

// HandleCommand -
func HandleCommand(conn interfaces.ClientConn, text string) {
	ind := strings.Index(text, "!")
	command := strings.SplitN(text[ind+1:], " ", -1)

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

		mapID := uint32(val)

		if _, exist := nx.Maps[mapID]; exist {
			portal, pID := maps.GetRandomSpawnPortal(mapID)

			char := charsPtr.GetOnlineCharacterHandle(conn)

			char.SetX(portal.GetX())
			char.SetY(portal.GetY())

			maps.PlayerLeaveMap(conn, char.GetCurrentMap())
			conn.Write(maps.ChangeMapPacket(mapID, 1, pID, char.GetHP()))
			maps.PlayerEnterMap(conn, mapID)
		} else {
			// check if player id in else if
		}
	default:
		log.Println("Unkown GM command", command)
	}
}
