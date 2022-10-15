package channel

import (
	"encoding/hex"
	"fmt"
	"github.com/Hucaru/Valhalla/internal"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

// TODO: Split these into ranks/levels (each rank can do everything the previous can):
// Admin -  Everything, can run server wide commands, can generate items, provide exp etc.
// Game Master  - can ban, can run channel wide commands, can spawn monsters
// Support  - can assist with issues such as missing passes in PQ, stuck players, misc issues
// Community - can start and moderate events
func (server *Server) gmCommand(conn mnet.Client, msg string) {
	ind := strings.Index(msg, "/")
	command := strings.SplitN(msg[ind+1:], " ", -1)

	switch command[0] {
	case "rate":
		rates := map[string]func(rate float32) mpacket.Packet{
			"exp":   internal.PacketChangeExpRate,
			"drop":  internal.PacketChangeDropRate,
			"mesos": internal.PacketChangeMesosRate,
		}

		if len(command) < 3 {
			conn.Send(packetMessageRedText("Command structure is /rate <exp | drop | mesos> <rate>"))
			return
		}

		mode := command[1]
		mFunc, ok := rates[mode]
		if !ok {
			conn.Send(packetMessageRedText("Choose between exp/drop/mesos rates"))
			return
		}

		rate := command[2]
		r, err := strconv.ParseFloat(rate, 32)
		if err != nil {
			log.Println("Failed parsing rate: ", err)
			conn.Send(packetMessageRedText("<rate> should be a number"))
			return
		}

		server.world.Send(mFunc(float32(r)))
	case "showRates":
		conn.Send(packetMessageNotice(fmt.Sprintf("Exp: x%.2f, Drop: x%.2f, Mesos: x%.2f", server.rates.exp, server.rates.drop, server.rates.mesos)))

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
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[player.mapID]

		if !ok {
			return
		}

		for i, v := range field.instances {
			info := "instance " + strconv.Itoa(i) + ":"
			conn.Send(packetMessageNotice(info))
			info = v.String()
			conn.Send(packetMessageNotice(info))
		}
	case "pos":
		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		conn.Send(packetMessageNotice(player.pos.String()))
	case "notice":
		if len(command) < 2 {
			return
		}

		for _, v := range server.players {
			v.send(packetMessageNotice(strings.Join(command[1:], " ")))
		}
	case "msgBox":
		if len(command) < 2 {
			return
		}

		for _, v := range server.players {
			v.send(packetMessageDialogueBox(strings.Join(command[1:], " ")))
		}
	case "header":
		if len(command) < 2 {
			server.header = ""
		} else {
			server.header = strings.Join(command[1:], " ")
		}

		for _, v := range server.players {
			v.send(packetMessageScrollingHeader(server.header))
		}
	case "wheader": // sends to world server to propagate to all channels

	case "kill":
		player, err := server.players.getFromConn(conn)

		if len(command) == 2 {
			player, err = server.players.getFromName(command[1])
		}

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		player.setHP(0)
	case "revive":
		player, err := server.players.getFromConn(conn)

		if len(command) == 2 {
			player, err = server.players.getFromName(command[1])
		}

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		player.setHP(player.maxHP)
	case "createInstance":
		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[player.mapID]

		if !ok {
			return
		}

		id := field.createInstance(&server.rates)

		conn.Send(packetMessageNotice("Created instance: " + strconv.Itoa(id)))
	case "changeInstance":
		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		var instanceID int

		if len(command) == 2 {
			instanceID, err = strconv.Atoi(command[1])
		} else if len(command) == 3 {
			player, err = server.players.getFromName(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			instanceID, err = strconv.Atoi(command[2])
		}

		field, ok := server.fields[player.mapID]

		if !ok {
			return
		}

		err = field.changePlayerInstance(player, instanceID)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
		}

		conn.Send(packetMessageNotice("Changed instance to " + strconv.Itoa(instanceID)))
	case "deleteInstance":
		if len(command) != 2 {
			return
		}

		instanceID, err := strconv.Atoi(command[1])

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		if instanceID < 1 {
			conn.Send(packetMessageRedText("Cannot delete instance 0"))
			return
		}

		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		if player.inst.id == instanceID {
			conn.Send(packetMessageRedText("Cannot delete the same instance you are in"))
			return
		}

		field, ok := server.fields[player.mapID]

		if !ok {
			return
		}

		err = field.deleteInstance(instanceID)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		conn.Send(packetMessageNotice("Deleted"))
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
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		if int16(amount) > player.maxHP {
			player.setMaxHP(int16(amount))
		}

		player.setHP(int16(amount))
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
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		if int16(amount) > player.maxMP {
			player.setMaxMP(int16(amount))
		}

		player.setMP(int16(amount))
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
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		player.setEXP(int32(amount))
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
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		player.giveEXP(int32(amount), false, false)
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
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		plr.setLevel(byte(amount))
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
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		player.giveLevel(byte(amount))
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
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		player.setJob(jobID)
	case "item":
		var itemID int32
		var amount int16 = 1

		if len(command) > 1 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			itemID = int32(val)

			if len(command) == 3 {
				val, err = strconv.Atoi(command[2])

				if err != nil {
					conn.Send(packetMessageRedText(err.Error()))
					return
				}

				amount = int16(val)
			}
		}

		item, err := createItemFromID(itemID, amount)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		item.creatorName = player.name
		err = player.giveItem(item)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
		}
	case "mesos":
		if len(command) == 2 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			player, err := server.players.getFromConn(conn)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			player.setMesos(int32(val))
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
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		plr, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[id]

		if !ok {
			conn.Send(packetMessageRedText("Could not find field ID"))
			return
		}

		inst, err := field.getInstance(plr.inst.id)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		portal, err := inst.getRandomSpawnPortal()

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		if playerName != "" {
			plr, err = server.players.getFromName(playerName)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}
		}

		dstField, ok := server.fields[id]

		if !ok {
			conn.Send(packetMessageRedText("Invalid map id"))
			return
		}

		server.warpPlayer(plr, dstField, portal)
	case "loadout":
		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		items := []int32{1372010, 1402005, 1422013, 1412021, 1382016, 1432030, 1442002, 1302023, 1322045, 1312015, 1332027, 1332026, 1462017, 1472033, 1452020, 1092029, 1092025}

		for _, v := range items {
			item, err := createPerfectItemFromID(v, 1)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			item.creatorName = player.name
			err = player.giveItem(item)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
			}
		}
	case "killMob":
		var spawnID int32

		if len(command) > 1 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			spawnID = int32(val)
		}

		plr, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[plr.mapID]

		if !ok {
			conn.Send(packetMessageRedText("Could not find field ID"))
			return
		}

		inst, err := field.getInstance(plr.inst.id)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		inst.lifePool.mobDamaged(spawnID, plr, math.MaxInt32)
	case "killmobs":
		var deathType byte = 1
		if len(command) > 1 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			deathType = byte(val)
		}

		plr, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[plr.mapID]

		if !ok {
			conn.Send(packetMessageRedText("Could not find field ID"))
			return
		}

		inst, err := field.getInstance(plr.inst.id)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		inst.lifePool.killMobs(deathType)
	case "spawnMob":
		var mobID int32
		var count int = 1

		if len(command) > 1 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			mobID = int32(val)
		}

		if len(command) == 3 {
			val, err := strconv.Atoi(command[2])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			count = val
		}

		plr, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[plr.mapID]

		if !ok {
			conn.Send(packetMessageRedText("Could not find field ID"))
			return
		}

		inst, err := field.getInstance(plr.inst.id)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		for i := 0; i < count; i++ {
			err := inst.lifePool.spawnMobFromID(mobID, plr.pos, false, true, true)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
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
				conn.Send(packetMessageRedText(err.Error()))
				return
			}
		}

		if len(command) == 3 {
			count, err = strconv.Atoi(command[2])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}
		}

		plr, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[plr.mapID]

		if !ok {
			conn.Send(packetMessageRedText("Could not find field ID"))
			return
		}

		inst, err := field.getInstance(plr.inst.id)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		for i := 0; i < count; i++ {
			for _, id := range mobID {
				err = inst.lifePool.spawnMobFromID(id, plr.pos, false, true, true)
			}

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				break
			}
		}
	case "testMob":
		plr, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[plr.mapID]

		if !ok {
			conn.Send(packetMessageRedText("Could not find field ID"))
			return
		}

		inst, err := field.getInstance(plr.inst.id)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		err = inst.lifePool.spawnMobFromID(5100001, plr.pos, true, true, true)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
		}
	case "portal":
		// plr, err := server.players.getFromConn(conn)

		// if err != nil {
		// 	conn.Send(packetMessageRedText(err.Error()))
		// 	return
		// }

		// field, ok := server.fields[plr.MapID()]

		// if !ok {
		// 	conn.Send(packetMessageRedText("Could not find field ID"))
		// 	return
		// }

		// inst, err := field.GetInstance(plr.InstanceID())

		// if err != nil {
		// 	conn.Send(packetMessageRedText(err.Error()))
		// 	return
		// }

		// dstField, ok := server.fields[180000000]

		// if !ok {
		// 	conn.Send(packetMessageRedText("Could not find field ID"))
		// 	return
		// }

		// inst.CreatePublicMysticDoor(dstField, plr.Pos(), time.Now().Add(time.Second*60).Unix())
	case "drop":
		plr, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[plr.mapID]

		if !ok {
			conn.Send(packetMessageRedText("Could not find field ID"))
			return
		}

		inst, err := field.getInstance(plr.inst.id)

		pool := inst.dropPool

		var mesos int32 = 1000

		items := []int32{1372010, 1402005, 1422013, 1412021, 1382016, 1432030, 1442002, 1302023, 1322045, 1312015, 1332027, 1332026, 1462017, 1472033, 1452020, 1092029, 1092025}
		drops := make([]item, len(items))

		for i, v := range items {
			item, err := createPerfectItemFromID(v, 1)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			item.creatorName = plr.name
			drops[i] = item
		}

		pool.createDrop(dropSpawnNormal, dropFreeForAll, mesos, plr.pos, true, plr.id, 0, drops...)
	case "dropr":
		var id int32 = -1
		var err error

		if len(command) > 1 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			id = int32(val)
		} else {
			conn.Send(packetMessageRedText("Supply drop id"))
			return
		}

		plr, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[plr.mapID]

		if !ok {
			conn.Send(packetMessageRedText("Could not find field ID"))
			return
		}
		inst, err := field.getInstance(plr.inst.id)
		pool := inst.dropPool
		pool.removeDrop(0, id)
	case "npco":
		// This isn't working, either incorrect opcode or script string is invalid
		p := mpacket.CreateWithOpcode(0x9F)
		p.WriteByte(2)        // amount
		p.WriteInt32(9200000) // npc id
		p.WriteString("cody") // string
		var startDate uint32 = 1 + (1 * 100) + (2001 * 10000)
		var endDate uint32 = 1 + (1 * 100) + (2099 * 10000)
		p.WriteUint32(startDate)
		p.WriteUint32(endDate)

		fmt.Println(p)
		conn.Send(p)
	case "whereami":
		player, err := server.players.getFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}
		conn.Send(packetMessageRedText(fmt.Sprintf("%d", player.mapID)))

	default:
		conn.Send(packetMessageRedText("Unkown gm command " + command[0]))
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
