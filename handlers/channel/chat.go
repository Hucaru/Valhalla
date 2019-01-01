package channel

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Hucaru/Valhalla/game/npcchat"
	"github.com/Hucaru/Valhalla/game/script"

	"github.com/Hucaru/Valhalla/nx"

	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/game/packet"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func chatSendAll(conn mnet.MConnChannel, reader mpacket.Reader) {
	msg := reader.ReadString(int(reader.ReadInt16()))

	if strings.Index(msg, "/") == 0 && conn.GetAdminLevel() > 0 {
		gmCommand(conn, msg)
	} else {
		player, ok := game.Players[conn]

		if !ok {
			return
		}

		char := player.Char()
		game.Maps[char.MapID].Send(packet.MessageAllChat(char.ID, conn.GetAdminLevel() > 0, msg), player.InstanceID)
	}
}

func chatSlashCommand(conn mnet.MConnChannel, reader mpacket.Reader) {
	cmdType := reader.ReadByte()

	switch cmdType {
	case 5: // FIND
		// length := reader.ReadInt16()
		// name := reader.ReadString(int(length))
	default:
		fmt.Println("Chat command type of", cmdType)
	}
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

		if _, err := nx.GetMap(mapID); err != nil {
			return
		}

		player, ok := game.Players[conn]

		if !ok {
			conn.Send(packet.MessageNotice(err.Error()))
			return
		}

		p, id := game.Maps[player.Char().MapID].GetRandomSpawnPortal()
		player.ChangeMap(mapID, p, id)

	case "notice":
		if len(command) < 2 {
			return
		}
		player, ok := game.Players[conn]

		if !ok {
			conn.Send(packet.MessageNotice("Error in sending notice msg"))
			return
		}

		char := player.Char()
		game.Maps[char.MapID].Send(packet.MessageNotice(strings.Join(command[1:], " ")), player.InstanceID)
	case "kill":
		if len(command) == 1 {
			player, ok := game.Players[conn]

			if !ok {
				conn.Send(packet.MessageNotice("Error in killing player"))
				return
			}

			player.Kill()
		} else {
			if command[1] == "<map>" {
				player, ok := game.Players[conn]

				if !ok {
					conn.Send(packet.MessageNotice("Error in killing players on map"))
					return
				}

				players, err := game.Maps[player.Char().MapID].GetPlayers(player.InstanceID)

				if err != nil {
					return
				}

				for _, v := range players {
					p := game.Players[v]
					p.Kill()
				}

				return
			}

			player, err := game.GetPlayerFromName(command[1])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
				return
			}

			player.Kill()
		}
	case "revive":
		if len(command) == 1 {
			player, ok := game.Players[conn]

			if !ok {
				conn.Send(packet.MessageNotice("Error in getting player"))
				return
			}

			player.Revive()
		} else {
			if command[1] == "<map>" {
				player, ok := game.Players[conn]

				if !ok {
					conn.Send(packet.MessageNotice("Error in getting player"))
					return
				}

				players, err := game.Maps[player.Char().MapID].GetPlayers(player.InstanceID)

				if err != nil {
					return
				}

				for _, v := range players {
					p := game.Players[v]
					p.Revive()
				}

				return
			}

			player, err := game.GetPlayerFromName(command[1])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
				return
			}

			player.Revive()
		}
	case "hp":
		if len(command) < 2 {
			return
		}

		player, ok := game.Players[conn]

		if !ok {
			conn.Send(packet.MessageNotice("Error in getting player"))
			return
		}

		if command[1][0] == '+' {
			ammount, err := strconv.Atoi(command[1][1:])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			player.GiveHP(int32(ammount))
		} else if command[1][0] == '-' {
			ammount, err := strconv.Atoi(command[1][1:])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			player.GiveHP(int32(-ammount))

		} else {
			ammount, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			player.SetHP(int32(ammount))
		}
	case "cody":
		if len(command) < 2 {
			return
		}

		npcchat.NewSessionWithOverride(conn, strings.Join(command[1:], " "), 9200000)
		npcchat.Run(conn)

	case "options":
		script, err := script.Get("options")

		if err != nil {
			return
		}

		npcchat.NewSessionWithOverride(conn, script, 9010000)
		npcchat.Run(conn)

	case "shop":

	case "createInstance":
		player, ok := game.Players[conn]

		if !ok {
			return
		}

		instID := game.Maps[player.Char().MapID].CreateNewInstance()
		conn.Send(packet.MessageNotice(fmt.Sprintln("New instance created with id:", instID)))
	case "changeInstance":
		if len(command) < 2 {
			return
		}

		newInstID, err := strconv.Atoi(command[1])

		if err != nil {
			conn.Send(packet.MessageNotice("Not a valid instance ID"))
		}

		player, ok := game.Players[conn]

		if !ok {
			return
		}

		player.ChangeInstance(newInstID)
	case "deleteInstance":
		if len(command) < 2 {
			return
		}

		instID, err := strconv.Atoi(command[1])

		if err != nil {
			conn.Send(packet.MessageNotice("Not a valid instance ID"))
		}

		player, ok := game.Players[conn]

		if !ok {
			return
		}

		err = game.Maps[player.Char().MapID].DeleteInstance(instID)

		if err != nil {
			conn.Send(packet.MessageNotice(err.Error()))
		}

		conn.Send(packet.MessageNotice("Instance deleted"))
	default:
		log.Println("Unkown GM command:", msg)
	}
}
