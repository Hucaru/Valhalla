package channel

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Hucaru/Valhalla/game/def"

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
		data = append(make([]byte, 4), data...)
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

		p, id, _ := game.Maps[player.Char().MapID].GetRandomSpawnPortal()
		player.ChangeMap(mapID, p, id)
	case "notice":
		if len(command) < 2 {
			return
		}

		for c := range game.Players {
			c.Send(packet.MessageNotice(strings.Join(command[1:], " ")))
		}
	case "msgBox":
		if len(command) < 2 {
			return
		}

		for c := range game.Players {
			c.Send(packet.MessageDialogueBox(strings.Join(command[1:], " ")))
		}
	case "scrollHeader":
		if len(command) < 2 {
			for c := range game.Players {
				c.Send(packet.MessageScrollingHeader(""))
			}
			return
		}

		for c := range game.Players {
			c.Send(packet.MessageScrollingHeader(strings.Join(command[1:], " ")))
		}
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

			player, err := game.Players.GetFromName(command[1])

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

			player, err := game.Players.GetFromName(command[1])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
				return
			}

			player.Revive()
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
			amount, err := strconv.Atoi(command[1][1:])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			player.GiveHP(int32(amount))
		} else if command[1][0] == '-' {
			amount, err := strconv.Atoi(command[1][1:])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			player.GiveHP(int32(-amount))

		} else {
			amount, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			player.SetMaxHP(int32(amount))
		}
	case "mp":
		if len(command) < 2 {
			return
		}

		player, ok := game.Players[conn]

		if !ok {
			conn.Send(packet.MessageNotice("Error in getting player"))
			return
		}

		if command[1][0] == '+' {
			amount, err := strconv.Atoi(command[1][1:])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			player.GiveMP(int32(amount))
		} else if command[1][0] == '-' {
			amount, err := strconv.Atoi(command[1][1:])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			player.GiveMP(int32(-amount))

		} else {
			amount, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			player.SetMaxMP(int32(amount))
		}
	case "exp":
		if len(command) < 2 {
			return
		}

		player, ok := game.Players[conn]

		if !ok {
			conn.Send(packet.MessageNotice("Error in getting player"))
			return
		}

		if command[1][0] == '+' {
			amount, err := strconv.Atoi(command[1][1:])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			player.GiveEXP(int32(amount), false, false)
		} else if command[1][0] == '-' {
			amount, err := strconv.Atoi(command[1][1:])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			player.GiveEXP(int32(-amount), false, false)

		} else {
			amount, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			player.SetEXP(int32(amount))
		}
	case "level":
		if len(command) < 2 {
			return
		}

		player, ok := game.Players[conn]

		if !ok {
			conn.Send(packet.MessageNotice("Error in getting player"))
			return
		}

		if command[1][0] == '+' {
			amount, err := strconv.Atoi(command[1][1:])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			player.GiveLevel(int8(amount))
		} else if command[1][0] == '-' {
			amount, err := strconv.Atoi(command[1][1:])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			player.GiveLevel(int8(-amount))

		} else {
			amount, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			player.SetLevel(byte(amount))
		}
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

		player, ok := game.Players[conn]

		if !ok {
			conn.Send(packet.MessageNotice(err.Error()))
			return
		}

		player.SetJob(jobID)
	case "item":
		if len(command) < 2 {
			return
		}

		itemID, err := strconv.Atoi(command[1])

		if err != nil {
			conn.Send(packet.MessageNotice(err.Error()))
		}

		player, ok := game.Players[conn]

		if !ok {
			conn.Send(packet.MessageNotice("Error in getting player"))
			return
		}

		item, err := def.CreateItemFromID(int32(itemID))

		if err != nil {
			conn.Send(packet.MessageNotice(err.Error()))
			return
		}

		if len(command) > 2 {
			amount, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packet.MessageNotice(err.Error()))
			}

			item.Amount = int16(amount)
		}

		player.GiveItem(item)
	default:
		log.Println("Unkown GM command:", msg)
	}
}
