package server

import (
	"encoding/hex"
	"log"
	"strconv"
	"strings"

	"github.com/Hucaru/Valhalla/entity"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/server/item"
)

func (server *ChannelServer) chatSendAll(conn mnet.Client, reader mpacket.Reader) {
	msg := reader.ReadString(reader.ReadInt16())

	if strings.Index(msg, "/") == 0 && conn.GetAdminLevel() > 0 {
		server.gmCommand(conn, msg)
	} else {
		player, err := server.players.getFromConn(conn)

		if err != nil {
			return
		}

		inst, err := server.fields[player.MapID()].GetInstance(player.InstanceID())

		if err != nil {
			return
		}

		inst.Send(entity.PacketMessageAllChat(player.ID(), conn.GetAdminLevel() > 0, msg))
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
	case "mapInfo":
		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[player.MapID()]

		if !ok {
			return
		}

		for i, v := range field.Instances() {
			info := "instance " + strconv.Itoa(i) + ":"
			conn.Send(entity.PacketMessageNotice(info))
			info = v.String()
			conn.Send(entity.PacketMessageNotice(info))
		}
	case "pos":
		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		conn.Send(entity.PacketMessageNotice(player.Pos().String()))
	case "notice":
		if len(command) < 2 {
			return
		}

		for _, v := range server.players {
			v.Send(entity.PacketMessageNotice(strings.Join(command[1:], " ")))
		}
	case "msgBox":
		if len(command) < 2 {
			return
		}

		for _, v := range server.players {
			v.Send(entity.PacketMessageDialogueBox(strings.Join(command[1:], " ")))
		}
	case "header":
		if len(command) < 2 {
			server.header = ""
		} else {
			server.header = strings.Join(command[1:], " ")
		}

		for _, v := range server.players {
			v.Send(entity.PacketMessageScrollingHeader(server.header))
		}
	case "wheader": // sends to world server to propagate to all channels

	case "kill":
		player, err := server.players.getFromConn(conn)

		if len(command) == 2 {
			player, err = server.players.getFromName(command[1])
		}

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		player.SetHP(0)
	case "revive":
		player, err := server.players.getFromConn(conn)

		if len(command) == 2 {
			player, err = server.players.getFromName(command[1])
		}

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		player.SetHP(player.MaxHP())
	case "cody":
	case "admin":
	case "shop":
	case "style":
	case "createInstance":
		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[player.MapID()]

		if !ok {
			return
		}

		id := field.CreateInstance()

		conn.Send(entity.PacketMessageNotice("Created instance: " + strconv.Itoa(id)))
	case "changeInstance":
		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		var instanceID int

		if len(command) == 2 {
			instanceID, err = strconv.Atoi(command[1])
		} else if len(command) == 3 {
			player, err = server.players.getFromName(command[1])

			if err != nil {
				conn.Send(entity.PacketMessageRedText(err.Error()))
				return
			}

			instanceID, err = strconv.Atoi(command[2])
		}

		field, ok := server.fields[player.MapID()]

		if !ok {
			return
		}

		err = field.ChangePlayerInstance(player, instanceID)

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
		}

		conn.Send(entity.PacketMessageNotice("Changed instance to " + strconv.Itoa(instanceID)))
	case "deleteInstance":
		if len(command) != 2 {
			return
		}

		instanceID, err := strconv.Atoi(command[1])

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		if instanceID < 1 {
			conn.Send(entity.PacketMessageRedText("Cannot delete instance 0"))
			return
		}

		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		if player.InstanceID() == instanceID {
			conn.Send(entity.PacketMessageRedText("Cannot delete the same instance you are in"))
			return
		}

		field, ok := server.fields[player.MapID()]

		if !ok {
			return
		}

		err = field.DeleteInstance(instanceID)

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		conn.Send(entity.PacketMessageNotice("Deleted"))
	case "hp":
		player, err := server.players.getFromConn(conn)

		var amount int

		if len(command) == 3 {
			player, err = server.players.getFromName(command[1])
			amount, err = strconv.Atoi(command[2])
		} else if len(command) == 2 {
			amount, err = strconv.Atoi(command[1])
		}

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		if int16(amount) > player.MaxHP() {
			player.SetMaxHP(int16(amount))
		}

		player.SetHP(int16(amount))
	case "mp":
		player, err := server.players.getFromConn(conn)

		var amount int

		if len(command) == 3 {
			player, err = server.players.getFromName(command[1])
			amount, err = strconv.Atoi(command[2])
		} else if len(command) == 2 {
			amount, err = strconv.Atoi(command[1])
		}

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		if int16(amount) > player.MaxMP() {
			player.SetMaxMP(int16(amount))
		}

		player.SetMP(int16(amount))
	case "exp":
		player, err := server.players.getFromConn(conn)

		var amount int

		if len(command) == 3 {
			player, err = server.players.getFromName(command[1])
			amount, err = strconv.Atoi(command[2])
		} else if len(command) == 2 {
			amount, err = strconv.Atoi(command[1])
		}

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[player.MapID()]

		if !ok {
			return
		}

		inst, err := field.GetInstance(player.InstanceID())

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		player.SetEXP(int32(amount), inst)
	case "level":
		player, err := server.players.getFromConn(conn)

		var amount int

		if len(command) == 3 {
			player, err = server.players.getFromName(command[1])
			amount, err = strconv.Atoi(command[2])
		} else if len(command) == 2 {
			amount, err = strconv.Atoi(command[1])
		}

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[player.MapID()]

		if !ok {
			return
		}

		inst, err := field.GetInstance(player.InstanceID())

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		player.SetLevel(byte(amount), inst)
	case "levelup":
		player, err := server.players.getFromConn(conn)

		var amount int

		if len(command) == 3 {
			player, err = server.players.getFromName(command[1])
			amount, err = strconv.Atoi(command[2])
		} else if len(command) == 2 {
			amount, err = strconv.Atoi(command[1])
		}

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[player.MapID()]

		if !ok {
			return
		}

		inst, err := field.GetInstance(player.InstanceID())

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		player.GiveLevel(byte(amount), inst)
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

		jobID := int16(val)

		if err != nil {
			jobID = convertJobNameToID(jobName)
		}

		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		player.SetJob(jobID)
	case "item":
		var itemID int32
		var amount int16 = 1

		if len(command) > 1 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(entity.PacketMessageRedText(err.Error()))
				return
			}

			itemID = int32(val)

			if len(command) == 3 {
				val, err = strconv.Atoi(command[2])

				if err != nil {
					conn.Send(entity.PacketMessageRedText(err.Error()))
					return
				}

				amount = int16(val)
			}
		}

		item, err := item.CreateFromID(itemID, amount)

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		item.SetCreatorName(player.Name())
		err = player.GiveItem(item)

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
		}
	case "mesos":
		if len(command) == 2 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(entity.PacketMessageRedText(err.Error()))
				return
			}

			player, err := server.players.getFromConn(conn)

			if err != nil {
				conn.Send(entity.PacketMessageRedText(err.Error()))
				return
			}

			player.SetMesos(int32(val))
		}
	case "spawn":
	case "warp":
		var val int
		var err error
		var mapName string
		var id int32
		var playerName string

		if len(command) == 2 {
			val, err = strconv.Atoi(command[1])
			mapName = command[1]
		} else if len(command) == 3 {
			playerName = command[1]
			val, err = strconv.Atoi(command[2])
			mapName = command[2]
		}

		if err != nil {
			id = convertMapNameToID(mapName)
		} else {
			id = int32(val)
		}

		if _, err := nx.GetMap(id); err != nil {
			return
		}

		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[id]

		if !ok {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		inst, err := field.GetInstance(player.InstanceID())

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		portal, err := inst.GetRandomSpawnPortal()

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		if playerName != "" {
			player, err = server.players.getFromName(playerName)

			if err != nil {
				conn.Send(entity.PacketMessageRedText(err.Error()))
				return
			}
		}

		dstField, ok := server.fields[id]

		if !ok {
			conn.Send(entity.PacketMessageRedText("Invalid map id"))
			return
		}

		server.warpPlayer(player, dstField, portal)
	case "loadout":
		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(entity.PacketMessageRedText(err.Error()))
			return
		}

		items := []int32{1372010, 1402005, 1422013, 1412021, 1382016, 1432030, 1442002, 1302023, 1322045, 1312015, 1332027, 1332026, 1462017, 1472033, 1452020, 1092029, 1092025}

		for _, v := range items {
			item, err := item.CreatePerfectFromID(v, 1)

			if err != nil {
				conn.Send(entity.PacketMessageRedText(err.Error()))
				return
			}

			item.SetCreatorName(player.Name())
			err = player.GiveItem(item)

			if err != nil {
				conn.Send(entity.PacketMessageRedText(err.Error()))
			}
		}
	default:
		conn.Send(entity.PacketMessageRedText("Unkown gm command " + command[0]))
	}
}

func convertMapNameToID(name string) int32 {
	switch name {
	// Maple island
	case "amherst":
		return 1010000
	case "southperry":
		return 60000
	// Victoria island
	case "lith":
		return 104000000
	case "henesys":
		return 100000000
	case "kerning":
		return 103000000
	case "perion":
		return 102000000
	case "ellinia":
		return 101000000
	case "sleepy":
		return 105040300
	case "gm":
		return 180000000
	// Ossyria
	case "orbis":
		return 200000000
	case "elnath":
		return 211000000
	case "ludi":
		return 220000000
	case "omega":
		return 221000000
	case "aqua":
		return 230000000
	// Misc
	case "balrog":
		return 105090900
	default:
		return 180000000
	}
}

func convertJobNameToID(name string) int16 {
	switch name {
	case "Beginner":
		return 0
	case "Warrior":
		return 100
	case "Fighter":
		return 110
	case "Crusader":
		return 111
	case "Page":
		return 120
	case "WhiteKnight":
		return 121
	case "Spearman":
		return 130
	case "DragonKnight":
		return 131
	case "Magician":
		return 200
	case "FirePoisonWizard":
		return 210
	case "FirePoisonMage":
		return 211
	case "IceLightWizard":
		return 220
	case "IceLightMage":
		return 221
	case "Cleric":
		return 230
	case "Priest":
		return 231
	case "Bowman":
		return 300
	case "Hunter":
		return 310
	case "Ranger":
		return 311
	case "Crossbowman":
		return 320
	case "Sniper":
		return 321
	case "Thief":
		return 400
	case "Assassin":
		return 410
	case "Hermit":
		return 411
	case "Bandit":
		return 420
	case "ChiefBandit":
		return 421
	case "Gm":
		return 500
	case "SuperGm":
		return 510
	default:
		return 0
	}
}
