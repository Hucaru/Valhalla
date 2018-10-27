package channelhandlers

import (
	"encoding/hex"
	"log"
	"strings"

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
