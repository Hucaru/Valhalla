package channel

import (
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/internal"

	"github.com/Hucaru/Valhalla/mpacket"

	"github.com/Hucaru/Valhalla/mnet"
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
	case "setWorldMessage":
		if len(command) < 2 {
			conn.Send(packetMessageRedText("Command structure is /setWorldMessage <ribbon_number> [message]"))
			return
		}

		ribbon, err := strconv.Atoi(command[1])
		if err != nil || ribbon < 0 {
			conn.Send(packetMessageRedText("Invalid ribbon number"))
			return
		}

		message := ""
		if len(command) >= 3 {
			message = strings.Join(command[2:], " ")
		}

		server.world.Send(internal.PacketUpdateLoginInfo(byte(ribbon), message))
		conn.Send(packetMessageNotice(fmt.Sprintf("Login info updated: Ribbon=%d, Message=%s", ribbon, message)))

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
		player, err := server.players.GetFromConn(conn)

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
		player, err := server.players.GetFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		conn.Send(packetMessageNotice(player.pos.String()))
	case "notice":
		if len(command) < 2 {
			return
		}

		server.players.broadcast(packetMessageNotice(strings.Join(command[1:], " ")))
	case "msgBox":
		if len(command) < 2 {
			return
		}

		server.players.broadcast(packetMessageDialogueBox(strings.Join(command[1:], " ")))
	case "header":
		if len(command) < 2 {
			server.header = ""
		} else {
			server.header = strings.Join(command[1:], " ")
		}

		server.players.broadcast(packetMessageScrollingHeader(server.header))
	case "wheader": // sends to world server to propagate to all channels

	case "kill":
		player, err := server.players.GetFromConn(conn)

		if len(command) == 2 {
			player, err = server.players.GetFromName(command[1])
		}

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		player.setHP(0)
	case "revive":
		player, err := server.players.GetFromConn(conn)

		if len(command) == 2 {
			player, err = server.players.GetFromName(command[1])
		}

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		player.setHP(player.maxHP)
	case "createInstance":
		player, err := server.players.GetFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		field, ok := server.fields[player.mapID]

		if !ok {
			return
		}

		id := field.createInstance(&server.rates, server)

		conn.Send(packetMessageNotice("Created instance: " + strconv.Itoa(id)))
	case "changeInstance":
		player, err := server.players.GetFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		var instanceID int

		if len(command) == 2 {
			instanceID, err = strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
			}
		} else if len(command) == 3 {
			player, err = server.players.GetFromName(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			instanceID, err = strconv.Atoi(command[2])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
			}
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

		player, err := server.players.GetFromConn(conn)

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
		player, err := server.players.GetFromConn(conn)

		var amount int

		if len(command) == 3 {
			player, err = server.players.GetFromName(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

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
		player, err := server.players.GetFromConn(conn)

		var amount int

		if len(command) == 3 {
			player, err = server.players.GetFromName(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

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

	case "setMaxHP":
		player, err := server.players.GetFromConn(conn)
		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}
		if len(command) < 2 {
			conn.Send(packetMessageRedText("Usage: /setMaxHP <amount>"))
			return
		}
		val, err := strconv.Atoi(command[1])
		if err != nil {
			conn.Send(packetMessageRedText("Amount must be a number"))
			return
		}
		if val < 1 {
			conn.Send(packetMessageRedText("Max HP must be at least 1"))
			return
		}
		player.setMaxHP(int16(val))
		conn.Send(packetMessageNotice(fmt.Sprintf("Set Max HP to %d", val)))

	case "setMaxMP":
		player, err := server.players.GetFromConn(conn)
		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}
		if len(command) < 2 {
			conn.Send(packetMessageRedText("Usage: /setMaxMP <amount>"))
			return
		}
		val, err := strconv.Atoi(command[1])
		if err != nil {
			conn.Send(packetMessageRedText("Amount must be a number"))
			return
		}
		if val < 0 {
			conn.Send(packetMessageRedText("Max MP cannot be negative"))
			return
		}
		player.setMaxMP(int16(val))
		conn.Send(packetMessageNotice(fmt.Sprintf("Set Max MP to %d", val)))

	case "str", "dex", "int", "luk":
		var (
			target *Player
			val    int
			err    error
		)

		switch len(command) {
		case 2:
			target, err = server.players.GetFromConn(conn)
			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}
			val, err = strconv.Atoi(command[1])
			if err != nil {
				conn.Send(packetMessageRedText("Amount must be a number"))
				return
			}
		case 3:
			target, err = server.players.GetFromName(command[1])
			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}
			val, err = strconv.Atoi(command[2])
			if err != nil {
				conn.Send(packetMessageRedText("Amount must be a number"))
				return
			}
		default:
			conn.Send(packetMessageRedText("Usage: /str <amount> | /str <player> <amount> (same for /dex /int /luk)"))
			return
		}

		if val < 0 {
			conn.Send(packetMessageRedText("Stat cannot be negative"))
			return
		}

		switch command[0] {
		case "str":
			target.str = int16(val)
			target.MarkDirty(DirtyStr, time.Millisecond*300)
			target.Send(packetPlayerStatChange(false, constant.StrID, int32(target.str)))
		case "dex":
			target.dex = int16(val)
			target.MarkDirty(DirtyDex, time.Millisecond*300)
			target.Send(packetPlayerStatChange(false, constant.DexID, int32(target.dex)))
		case "int":
			target.intt = int16(val)
			target.MarkDirty(DirtyInt, time.Millisecond*300)
			target.Send(packetPlayerStatChange(false, constant.IntID, int32(target.intt)))
		case "luk":
			target.luk = int16(val)
			target.MarkDirty(DirtyLuk, time.Millisecond*300)
			target.Send(packetPlayerStatChange(false, constant.LukID, int32(target.luk)))
		}

	case "questFinish":
		if len(command) < 2 {
			conn.Send(packetMessageRedText("Usage: /questFinish <quest-id>"))
			return
		}
		qid64, err := strconv.ParseInt(command[1], 10, 16)
		if err != nil {
			conn.Send(packetMessageRedText("Quest ID must be a number"))
			return
		}
		questID := int16(qid64)

		plr, err := server.players.GetFromConn(conn)
		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		nowMs := time.Now().UnixMilli()
		plr.quests.complete(questID, nowMs)
		setQuestCompleted(plr.ID, questID, nowMs)

		plr.Send(packetQuestComplete(questID))
		conn.Send(packetMessageNotice(fmt.Sprintf("Quest %d completed", questID)))

	case "questUntil":
		// Example: /questUntil 1001 3
		if len(command) < 3 {
			conn.Send(packetMessageRedText("Usage: /questUntil <quest-id> <part>"))
			return
		}
		qid64, err := strconv.ParseInt(command[1], 10, 16)
		if err != nil {
			conn.Send(packetMessageRedText("Quest ID must be a number"))
			return
		}
		part, err := strconv.Atoi(command[2])
		if err != nil || part < 0 {
			conn.Send(packetMessageRedText("Part must be a non-negative number"))
			return
		}
		questID := int16(qid64)
		record := fmt.Sprintf("p%d", part)

		plr, err := server.players.GetFromConn(conn)
		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		plr.quests.add(questID, record)
		upsertQuestRecord(plr.ID, questID, record)

		plr.Send(packetQuestUpdate(questID, record))
		conn.Send(packetMessageNotice(fmt.Sprintf("Quest %d progressed to %s", questID, record)))

	case "questReset":
		if len(command) < 2 {
			conn.Send(packetMessageRedText("Usage: /questReset <quest-id>"))
			return
		}
		qid64, err := strconv.ParseInt(command[1], 10, 16)
		if err != nil {
			conn.Send(packetMessageRedText("Quest ID must be a number"))
			return
		}
		questID := int16(qid64)

		plr, err := server.players.GetFromConn(conn)
		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		delete(plr.quests.inProgress, questID)
		delete(plr.quests.completed, questID)
		delete(plr.quests.mobKills, questID)

		deleteQuest(plr.ID, questID)
		clearQuestMobKills(plr.ID, questID)

		plr.Send(packetQuestRemove(questID))
		conn.Send(packetMessageNotice(fmt.Sprintf("Quest %d has been reset", questID)))

	case "skillLv":
		var target *Player
		var err error
		args := command[1:]
		switch len(args) {
		case 2:
			target, err = server.players.GetFromConn(conn)
			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}
		case 3:
			target, err = server.players.GetFromName(args[0])
			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}
			args = args[1:]
		default:
			conn.Send(packetMessageRedText("Usage: /skillLv <skill-id> <level|max> OR /skillLv <player> <skill-id> <level|max>"))
			return
		}
		id64, err := strconv.ParseInt(args[0], 10, 32)
		if err != nil {
			conn.Send(packetMessageRedText("Skill must be a numeric ID"))
			return
		}
		skillID := int32(id64)
		levels, err := nx.GetPlayerSkill(skillID)
		if err != nil || len(levels) == 0 {
			conn.Send(packetMessageRedText(fmt.Sprintf("Unknown or invalid skill ID: %d", skillID)))
			return
		}
		var level byte
		if strings.EqualFold(args[1], "max") {
			level = byte(len(levels))
		} else {
			lv64, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				conn.Send(packetMessageRedText("Level must be a number or 'max'"))
				return
			}
			if lv64 < 0 || int(lv64) > len(levels) {
				conn.Send(packetMessageRedText(fmt.Sprintf("Invalid level, max for skill %d is %d", skillID, len(levels))))
				return
			}
			level = byte(lv64)
		}
		if level == 0 {
			delete(target.skills, skillID)
			target.MarkDirty(DirtySkills, time.Millisecond*300)
			conn.Send(packetMessageNotice(fmt.Sprintf("Removed skill %d from %s", skillID, target.Name)))
			return
		}
		ps, err := createPlayerSkillFromData(skillID, level)
		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}
		if target.skills == nil {
			target.skills = make(map[int32]playerSkill, 8)
		}
		target.skills[skillID] = ps
		target.MarkDirty(DirtySkills, time.Millisecond*300)
		conn.Send(packetMessageNotice(fmt.Sprintf("Set %s's skill %d to level %d", target.Name, skillID, ps.Level)))

	case "maxSkills":
		target, err := server.players.GetFromConn(conn)
		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}
		if target.skills == nil {
			target.skills = make(map[int32]playerSkill, 1024)
		}

		jobFamilies := []int{
			int(constant.BeginnerJobID),

			int(constant.WarriorJobID),
			int(constant.FighterJobID),
			int(constant.CrusaderJobID),
			int(constant.PageJobID),
			int(constant.WhiteKnightJobID),
			int(constant.SpearmanJobID),
			int(constant.DragonKnightJobID),

			int(constant.MagicianJobID),
			int(constant.FirePoisonWizardJobID),
			int(constant.FirePoisonMageJobID),
			int(constant.IceLightWizardJobID),
			int(constant.IceLightMageJobID),
			int(constant.ClericJobID),
			int(constant.PriestJobID),

			int(constant.BowmanJobID),
			int(constant.HunterJobID),
			int(constant.RangerJobID),
			int(constant.CrossbowmanJobID),
			int(constant.SniperJobID),

			int(constant.ThiefJobID),
			int(constant.AssassinJobID),
			int(constant.HermitJobID),
			int(constant.BanditJobID),
			int(constant.ChiefBanditJobID),

			int(constant.GmJobID),
			int(constant.SuperGmJobID),
		}

		var count int
		for _, job := range jobFamilies {
			base := job * 10000
			for idx := 0; idx <= 1999; idx++ {
				skillID := int32(base + idx)
				levels, err := nx.GetPlayerSkill(skillID)
				if err != nil || len(levels) == 0 {
					continue
				}
				maxLv := byte(len(levels))
				ps, err := createPlayerSkillFromData(skillID, maxLv)
				if err != nil {
					continue
				}
				target.skills[skillID] = ps
				count++
			}
		}

		target.MarkDirty(DirtySkills, time.Millisecond*300)
		conn.Send(packetMessageNotice(fmt.Sprintf("Maxed %d skills across all classes for %s", count, target.Name)))

	case "resetSkills":
		target, err := server.players.GetFromConn(conn)
		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}
		if target.skills != nil {
			for k := range target.skills {
				delete(target.skills, k)
			}
		}

		target.MarkDirty(DirtySkills, time.Millisecond*300)
		conn.Send(packetMessageNotice(fmt.Sprintf("Reset all skills for %s", target.Name)))

	case "exp":
		player, err := server.players.GetFromConn(conn)

		var amount int

		if len(command) == 3 {
			player, err = server.players.GetFromName(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

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
		player, err := server.players.GetFromConn(conn)

		var amount int

		if len(command) == 3 {
			player, err = server.players.GetFromName(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			amount, err = strconv.Atoi(command[2])
		} else if len(command) == 2 {
			amount, err = strconv.Atoi(command[1])
		}

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		player.giveEXP(int32(amount), false, false)
	case "ap":
		plr, err := server.players.GetFromConn(conn)

		var amount int

		if len(command) == 3 {
			plr, err = server.players.GetFromName(command[1])
			amount, err = strconv.Atoi(command[2])
		} else if len(command) == 2 {
			amount, err = strconv.Atoi(command[1])
		}

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		plr.setAP(int16(amount))
	case "sp":
		plr, err := server.players.GetFromConn(conn)

		var amount int

		if len(command) == 3 {
			plr, err = server.players.GetFromName(command[1])
			amount, err = strconv.Atoi(command[2])
		} else if len(command) == 2 {
			amount, err = strconv.Atoi(command[1])
		}

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		plr.setSP(int16(amount))
	case "level":
		plr, err := server.players.GetFromConn(conn)

		var amount int

		if len(command) == 3 {
			plr, err = server.players.GetFromName(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

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
		player, err := server.players.GetFromConn(conn)

		var amount int

		if len(command) == 3 {
			player, err = server.players.GetFromName(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

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

		player, err := server.players.GetFromConn(conn)

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

		item, err := CreateItemFromID(itemID, amount)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		player, err := server.players.GetFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		item.creatorName = player.Name
		err, _ = player.GiveItem(item)

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

			player, err := server.players.GetFromConn(conn)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			player.setMesos(int32(val))
		}
	case "nx":
		if len(command) == 2 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			player, err := server.players.GetFromConn(conn)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			player.nx += (int32(val))
		}
	case "maplepoints":
		if len(command) == 2 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			player, err := server.players.GetFromConn(conn)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			player.maplepoints += (int32(val))
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

		plr, err := server.players.GetFromConn(conn)

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
			plr, err = server.players.GetFromName(playerName)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}
		}

		dstField, ok := server.fields[id]

		if !ok {
			conn.Send(packetMessageRedText("Invalid map ID"))
			return
		}

		err = server.warpPlayer(plr, dstField, portal, true)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
		}
	case "warpTo":
		playerName := command[1]

		person, err := server.players.GetFromName(playerName)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		dstField, ok := server.fields[person.inst.fieldID]

		if !ok {
			conn.Send(packetMessageRedText("Invalid map ID"))
			return
		}

		portalID, err := person.inst.calculateNearestSpawnPortalID(person.pos)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		err = server.warpPlayer(plr, dstField, person.inst.portals[portalID], true)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
		}
	case "loadout":
		player, err := server.players.GetFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		equips := []int32{1372010, 1402005, 1422013, 1412021, 1382016, 1432030, 1442002, 1302023, 1322045, 1312015, 1332027, 1332026, 1462017, 1472033, 1452020, 1092029, 1092025}

		for _, v := range equips {
			item, err := createPerfectItemFromID(v, 1)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			item.creatorName = player.Name
			err, _ = player.GiveItem(item)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
			}
		}

		etc := []int32{4006001, 4006000, 4001017, 4031179, 4031059}

		for _, v := range etc {
			item, err := createPerfectItemFromID(v, 100)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			item.creatorName = player.Name
			err, _ = player.GiveItem(item)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
			}
		}
	case "clearDrops":
		plr, err := server.players.GetFromConn(conn)

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

		inst.dropPool.clearDrops()

	case "removeTimer":
		plr, err := server.players.GetFromConn(conn)

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

		inst.fieldTimer.Reset(0)

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

		plr, err := server.players.GetFromConn(conn)

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
	case "killAll":
		fallthrough
	case "killmobs":
		plr, err := server.players.GetFromConn(conn)

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

		for spawnID, mob := range inst.lifePool.mobs {
			inst.lifePool.mobDamaged(spawnID, plr, mob.hp)
		}

	case "spawn":
		fallthrough
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

		plr, err := server.players.GetFromConn(conn)

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
			err := inst.lifePool.spawnMobFromID(mobID, plr.pos, false, true, true, constant.MobSummonTypeInstant, plr.ID)

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

		plr, err := server.players.GetFromConn(conn)

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
				err = inst.lifePool.spawnMobFromID(id, plr.pos, false, true, true, constant.MobSummonTypeInstant, plr.ID)
			}

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				break
			}
		}
	case "partyCreate":
		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		server.world.Send(internal.PacketChannelPartyCreateRequest(plr.ID, server.id, plr.mapID, int32(plr.job), int32(plr.level), plr.Name))
	case "guildCreate":
		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		var controller *npcChatController

		if program, ok := server.npcScriptStore.scripts["2010007"]; ok {
			controller, err = createNpcChatController(2010007, conn, program, plr, server.fields, server.warpPlayer, server.world)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}
		}

		if controller == nil {
			log.Println("Unable to find guild npc script")
			return
		}

		if err != nil {
			log.Println("script init:", err)
		}

		server.npcChat[conn] = controller
		server.updateNPCInteractionMetric(1)

		if controller.run() {
			delete(server.npcChat, conn)
			server.updateNPCInteractionMetric(-1)
		}
	case "guildDisband":
		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		if plr.guild == nil {
			conn.Send(packetMessageRedText("Not in guild, cannot disband"))
		}

		server.world.Send(internal.PacketGuildDisband(plr.guild.id))
	case "guildPoints":
		plr, err := server.players.GetFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}

		if plr.guild == nil {
			conn.Send(packetMessageRedText("Not in guild, cannot disband"))
		}

		if len(command) == 2 {
			points, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
			}

			server.world.Send(internal.PacketGuildPointsUpdate(plr.guild.id, int32(points)))
		}
	case "changeBgm":
		bgm := ""

		if len(command) > 1 {
			bgm = command[1]
		}

		plr, err := server.players.GetFromConn(conn)

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

		inst.changeBgm(bgm)
	case "testMob":
		plr, err := server.players.GetFromConn(conn)

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

		err = inst.lifePool.spawnMobFromID(5100001, plr.pos, true, true, true, constant.MobSummonTypeInstant, plr.ID)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
		}
	case "portal":
		// plr, err := server.players.getFromConn(Conn)

		// if err != nil {
		// 	Conn.Send(packetMessageRedText(err.Error()))
		// 	return
		// }

		// field, ok := server.fields[plr.MapID()]

		// if !ok {
		// 	Conn.Send(packetMessageRedText("Could not find field ID"))
		// 	return
		// }

		// inst, err := field.GetInstance(plr.InstanceID())

		// if err != nil {
		// 	Conn.Send(packetMessageRedText(err.Error()))
		// 	return
		// }

		// dstField, ok := server.fields[180000000]

		// if !ok {
		// 	Conn.Send(packetMessageRedText("Could not find field ID"))
		// 	return
		// }

		// inst.CreatePublicMysticDoor(dstField, plr.Pos(), time.Now().Add(time.Second*60).Unix())
	case "drop":
		plr, err := server.players.GetFromConn(conn)

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

		pool := inst.dropPool

		var mesos int32 = 1000

		items := []int32{1372010, 1402005, 1422013, 1412021, 1382016, 1432030, 1442002, 1302023, 1322045, 1312015, 1332027, 1332026, 1462017, 1472033, 1452020, 1092029, 1092025}
		drops := make([]Item, len(items))

		for i, v := range items {
			item, err := createPerfectItemFromID(v, 1)

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			item.creatorName = plr.Name
			drops[i] = item
		}

		pool.createDrop(dropSpawnNormal, dropFreeForAll, mesos, plr.pos, true, plr.ID, 0, drops...)
	case "dropr":
		var id int32
		var err error

		if len(command) > 1 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				conn.Send(packetMessageRedText(err.Error()))
				return
			}

			id = int32(val)
		} else {
			conn.Send(packetMessageRedText("Supply drop ID"))
			return
		}

		plr, err := server.players.GetFromConn(conn)

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

		pool := inst.dropPool
		pool.removeDrop(0, id)
	case "npco":
		// This isn't working, either incorrect opcode or script string is invalid
		p := mpacket.CreateWithOpcode(0x9F)
		p.WriteByte(2)        // amount
		p.WriteInt32(9200000) // npc ID
		p.WriteString("cody") // string
		var startDate uint32 = 1 + (1 * 100) + (2001 * 10000)
		var endDate uint32 = 1 + (1 * 100) + (2099 * 10000)
		p.WriteUint32(startDate)
		p.WriteUint32(endDate)

		fmt.Println(p)
		conn.Send(p)
	case "whereami":
		player, err := server.players.GetFromConn(conn)

		if err != nil {
			conn.Send(packetMessageRedText(err.Error()))
			return
		}
		conn.Send(packetMessageRedText(fmt.Sprintf("%d", player.mapID)))

	default:
		conn.Send(packetMessageRedText("Unknown gm command " + command[0]))
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
	case "guild":
		return 200000301
	case "pap":
		return constant.MapBossPapulatus
	case "pianus":
		return constant.MapBossPianus
	case "zakum":
		return constant.MapBossZakum
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
		return []int32{constant.MobBalrog}, nil
	case "cbalrog":
		return []int32{constant.MobCrimsonBalrog}, nil
	case "zakum":
		return []int32{
			constant.MobZakumArm1,
			constant.MobZakumArm2,
			constant.MobZakumArm3,
			constant.MobZakumArm4,
			constant.MobZakumArm5,
			constant.MobZakumArm6,
			constant.MobZakumArm7,
			constant.MobZakumArm8,
			constant.MobZakum1Body,
		}, nil
	case "pap":
		return []int32{constant.MobPapalatus}, nil
	case "pianus":
		return []int32{constant.MobPianus}, nil
	case "mushmom":
		return []int32{constant.MobMushmom}, nil
	case "zmushmom":
		return []int32{constant.MobZombieMushmom}, nil
	}

	return nil, fmt.Errorf("unknown mob Name")
}
