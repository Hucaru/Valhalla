package server

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/server/item"
	"github.com/Hucaru/Valhalla/server/message"
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

		inst.Send(message.PacketMessageAllChat(player.ID(), conn.GetAdminLevel() > 0, msg))
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
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[player.MapID()]

		if !ok {
			return
		}

		for i, v := range field.Instances() {
			info := "instance " + strconv.Itoa(i) + ":"
			conn.Send(message.PacketMessageNotice(info))
			info = v.String()
			conn.Send(message.PacketMessageNotice(info))
		}
	case "pos":
		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		conn.Send(message.PacketMessageNotice(player.Pos().String()))
	case "notice":
		if len(command) < 2 {
			return
		}

		for _, v := range server.players {
			v.Send(message.PacketMessageNotice(strings.Join(command[1:], " ")))
		}
	case "msgBox":
		if len(command) < 2 {
			return
		}

		for _, v := range server.players {
			v.Send(message.PacketMessageDialogueBox(strings.Join(command[1:], " ")))
		}
	case "header":
		if len(command) < 2 {
			server.header = ""
		} else {
			server.header = strings.Join(command[1:], " ")
		}

		for _, v := range server.players {
			v.Send(message.PacketMessageScrollingHeader(server.header))
		}
	case "wheader": // sends to world server to propagate to all channels

	case "kill":
		player, err := server.players.getFromConn(conn)

		if len(command) == 2 {
			player, err = server.players.getFromName(command[1])
		}

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		player.SetHP(0)
	case "revive":
		player, err := server.players.getFromConn(conn)

		if len(command) == 2 {
			player, err = server.players.getFromName(command[1])
		}

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
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
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[player.MapID()]

		if !ok {
			return
		}

		id := field.CreateInstance()

		conn.Send(message.PacketMessageNotice("Created instance: " + strconv.Itoa(id)))
	case "changeInstance":
		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		var instanceID int

		if len(command) == 2 {
			instanceID, err = strconv.Atoi(command[1])
		} else if len(command) == 3 {
			player, err = server.players.getFromName(command[1])

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
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
			conn.Send(message.PacketMessageRedText(err.Error()))
		}

		conn.Send(message.PacketMessageNotice("Changed instance to " + strconv.Itoa(instanceID)))
	case "deleteInstance":
		if len(command) != 2 {
			return
		}

		instanceID, err := strconv.Atoi(command[1])

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		if instanceID < 1 {
			conn.Send(message.PacketMessageRedText("Cannot delete instance 0"))
			return
		}

		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		if player.InstanceID() == instanceID {
			conn.Send(message.PacketMessageRedText("Cannot delete the same instance you are in"))
			return
		}

		field, ok := server.fields[player.MapID()]

		if !ok {
			return
		}

		err = field.DeleteInstance(instanceID)

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		conn.Send(message.PacketMessageNotice("Deleted"))
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
			conn.Send(message.PacketMessageRedText(err.Error()))
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
			conn.Send(message.PacketMessageRedText(err.Error()))
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
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		player.SetEXP(int32(amount))
	case "gexp":
		player, err := server.players.getFromConn(conn)

		var amount int

		if len(command) == 3 {
			player, err = server.players.getFromName(command[1])
			amount, err = strconv.Atoi(command[2])
		} else if len(command) == 2 {
			amount, err = strconv.Atoi(command[1])
		}

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		player.GiveEXP(int32(amount), false, false)
	case "level":
		plr, err := server.players.getFromConn(conn)

		var amount int

		if len(command) == 3 {
			plr, err = server.players.getFromName(command[1])
			amount, err = strconv.Atoi(command[2])
		} else if len(command) == 2 {
			amount, err = strconv.Atoi(command[1])
		}

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		plr.SetLevel(byte(amount))
	case "levelup":
		player, err := server.players.getFromConn(conn)

		var amount int

		if len(command) == 3 {
			player, err = server.players.getFromName(command[1])
			amount, err = strconv.Atoi(command[2])
		} else if len(command) == 2 {
			amount, err = strconv.Atoi(command[1])
		} else if len(command) == 1 {
			amount = 1
		}

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		player.GiveLevel(byte(amount))
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
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		player.SetJob(jobID)
	case "item":
		var itemID int32
		var amount int16 = 1

		if len(command) > 1 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				return
			}

			itemID = int32(val)

			if len(command) == 3 {
				val, err = strconv.Atoi(command[2])

				if err != nil {
					conn.Send(message.PacketMessageRedText(err.Error()))
					return
				}

				amount = int16(val)
			}
		}

		item, err := item.CreateFromID(itemID, amount)

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		item.SetCreatorName(player.Name())
		err = player.GiveItem(item, server.db)

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
		}
	case "mesos":
		if len(command) == 2 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				return
			}

			player, err := server.players.getFromConn(conn)

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				return
			}

			player.SetMesos(int32(val))
		}
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

		plr, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[id]

		if !ok {
			conn.Send(message.PacketMessageRedText("Could not find field ID"))
			return
		}

		inst, err := field.GetInstance(plr.InstanceID())

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		portal, err := inst.GetRandomSpawnPortal()

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		if playerName != "" {
			plr, err = server.players.getFromName(playerName)

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				return
			}
		}

		dstField, ok := server.fields[id]

		if !ok {
			conn.Send(message.PacketMessageRedText("Invalid map id"))
			return
		}

		server.warpPlayer(plr, dstField, portal)
	case "loadout":
		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		items := []int32{1372010, 1402005, 1422013, 1412021, 1382016, 1432030, 1442002, 1302023, 1322045, 1312015, 1332027, 1332026, 1462017, 1472033, 1452020, 1092029, 1092025}

		for _, v := range items {
			item, err := item.CreatePerfectFromID(v, 1)

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				return
			}

			item.SetCreatorName(player.Name())
			err = player.GiveItem(item, server.db)

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
			}
		}
	case "spawn":
		if len(command) == 2 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				return
			}

			plr, err := server.players.getFromConn(conn)

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				return
			}

			field, ok := server.fields[plr.MapID()]

			if !ok {
				conn.Send(message.PacketMessageRedText("Could not find field ID"))
				return
			}

			inst, err := field.GetInstance(plr.InstanceID())

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				return
			}

			inst.SpawnMobFromMobID(int32(val), plr.Pos(), false, true, true)
		}
	case "removeMob":
		var spawnID int32
		var deathType byte

		if len(command) > 1 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				return
			}

			spawnID = int32(val)
		}

		if len(command) == 3 {
			val, err := strconv.Atoi(command[2])

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				return
			}

			deathType = byte(val)
		}

		plr, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[plr.MapID()]

		if !ok {
			conn.Send(message.PacketMessageRedText("Could not find field ID"))
			return
		}

		inst, err := field.GetInstance(plr.InstanceID())

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		inst.RemoveMob(int32(spawnID), deathType)
	case "killMob":
		var spawnID int32

		if len(command) > 1 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				return
			}

			spawnID = int32(val)
		}

		plr, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[plr.MapID()]

		if !ok {
			conn.Send(message.PacketMessageRedText("Could not find field ID"))
			return
		}

		inst, err := field.GetInstance(plr.InstanceID())

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		m := inst.GetMob(spawnID)

		if m == nil {
			conn.Send(message.PacketMessageRedText("Invalid mob spawn id"))
			return
		}

		m.HandleDamage(plr, inst, nil, m.HP())
	case "spawnMob":
		var mobID int32
		var count int = 1

		if len(command) > 1 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				return
			}

			mobID = int32(val)
		}

		if len(command) == 3 {
			val, err := strconv.Atoi(command[2])

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				return
			}

			count = val
		}

		plr, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[plr.MapID()]

		if !ok {
			conn.Send(message.PacketMessageRedText("Could not find field ID"))
			return
		}

		inst, err := field.GetInstance(plr.InstanceID())

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		for i := 0; i < count; i++ {
			err := inst.SpawnMobFromMobID(mobID, plr.Pos(), false, true, true)

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				break
			}
		}
	case "spawnBoss":
		var mobID []int32
		var count int = 1
		var err error

		if len(command) > 1 {
			mobID, err = covnertMobNameToID(command[1])

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				return
			}
		}

		if len(command) == 3 {
			count, err = strconv.Atoi(command[2])

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				return
			}
		}

		plr, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[plr.MapID()]

		if !ok {
			conn.Send(message.PacketMessageRedText("Could not find field ID"))
			return
		}

		inst, err := field.GetInstance(plr.InstanceID())

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		for i := 0; i < count; i++ {
			for _, id := range mobID {
				err = inst.SpawnMobFromMobID(id, plr.Pos(), false, true, true)
			}

			if err != nil {
				conn.Send(message.PacketMessageRedText(err.Error()))
				break
			}
		}
	case "testMob":
		plr, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[plr.MapID()]

		if !ok {
			conn.Send(message.PacketMessageRedText("Could not find field ID"))
			return
		}

		inst, err := field.GetInstance(plr.InstanceID())

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
			return
		}
		err = inst.SpawnMobFromMobID(5100001, plr.Pos(), true, true, true)

		if err != nil {
			conn.Send(message.PacketMessageRedText(err.Error()))
		}

	default:
		conn.Send(message.PacketMessageRedText("Unkown gm command " + command[0]))
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

func covnertMobNameToID(name string) ([]int32, error) {
	switch name {
	case "balrog":
		return []int32{8130100}, nil
	case "cbalrog":
		return []int32{8150000}, nil
	case "zakum":
		return []int32{
			8800003, //Zakum's Arm 1
			8800004, //Zakum's Arm 2
			8800005, //Zakum's Arm 3
			8800006, //Zakum's Arm 4
			8800007, //Zakum's Arm 5
			8800008, //Zakum's Arm 6
			8800009, //Zakum's Arm 7
			8800010, //Zakum's Arm 8
			8800000, //Zakum1's body
		}, nil
	case "pap":
		return []int32{8500001}, nil // clock - 8500002
	case "pianus":
		return []int32{8520000}, nil // or 8510000
	case "mushmom":
		return []int32{6130101}, nil
	case "zmushmom":
		return []int32{6300005}, nil
	}

	return nil, fmt.Errorf("Unkown mob name")
}
