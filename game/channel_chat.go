package game

import (
	"encoding/hex"
	"log"
	"strconv"
	"strings"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (server *ChannelServer) chatSendAll(conn mnet.Client, reader mpacket.Reader) {
	msg := reader.ReadString(reader.ReadInt16())

	if strings.Index(msg, "/") == 0 && conn.GetAdminLevel() > 0 {
		server.gmCommand(conn, msg)
	} else {
		player, _ := server.players.getFromConn(conn)
		char := player.char

		server.fields[char.mapID].send(packetMessageAllChat(char.id, conn.GetAdminLevel() > 0, msg), player.instanceID)
	}
}

func (server *ChannelServer) gmCommand(conn mnet.Client, msg string) {
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
		data = append(make([]byte, 4), data...)
		conn.Send(data)
	case "map":
	case "notice":
		if len(command) < 2 {
			return
		}

		for _, v := range server.players {
			v.conn.Send(packetMessageNotice(strings.Join(command[1:], " ")))
		}
	case "msgBox":
		if len(command) < 2 {
			return
		}

		for _, v := range server.players {
			v.conn.Send(packetMessageDialogueBox(strings.Join(command[1:], " ")))
		}
	case "scrollHeader":
	case "kill":
	case "revive":
	case "cody":
	case "admin":
	case "shop":
	case "createInstance":
	case "changeInstance":
	case "deleteInstance":
	case "hp":
	case "mp":
	case "exp":
	case "level":
	case "job":
		var val int
		var err error
		var jobName string

		if len(command) == 2 {
			val, err = strconv.Atoi(command[1])
			jobName = command[1]
		} else if len(command) == 3 {
			val, err = strconv.Atoi(command[2])
			jobName = command[2]
		}

		if err != nil {
			// Check to see if name matches pre-recorded
			switch jobName {
			case "Beginner":
				val = 0
			case "Warrior":
				val = 100
			case "Fighter":
				val = 110
			case "Crusader":
				val = 111
			case "Page":
				val = 120
			case "WhiteKnight":
				val = 121
			case "Spearman":
				val = 130
			case "DragonKnight":
				val = 131
			case "Magician":
				val = 200
			case "FirePoisonWizard":
				val = 210
			case "FirePoisonMage":
				val = 211
			case "IceLightWizard":
				val = 220
			case "IceLightMage":
				val = 221
			case "Cleric":
				val = 230
			case "Priest":
				val = 231
			case "Bowman":
				val = 300
			case "Hunter":
				val = 310
			case "Ranger":
				val = 311
			case "Crossbowman":
				val = 320
			case "Sniper":
				val = 321
			case "Thief":
				val = 400
			case "Assassin":
				val = 410
			case "Hermit":
				val = 411
			case "Bandit":
				val = 420
			case "ChiefBandit":
				val = 421
			case "Gm":
				val = 500
			case "SuperGm":
				val = 510
			default:
				return
			}
		}

		jobID := int16(val)

		player, _ := server.players.getFromConn(conn)
		player.setJob(jobID)
	case "item":
	case "spawn":
	default:
		conn.Send(packetMessageNotice("Unkown gm command " + command[0]))
	}
}
