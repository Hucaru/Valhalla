package commands

import (
	"encoding/hex"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/interop"
	"github.com/Hucaru/Valhalla/npcChat"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/packets"
)

func HandleGmCommand(conn interop.ClientConn, msg string) {
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
		conn.Write(data)
	case "warp":
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
			default:
				return
			}
		}

		mapID := uint32(val)

		if len(command) == 2 {
			channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
				if _, exist := nx.Maps[mapID]; exist {
					portal, pID := channel.Maps.GetMap(char.GetCurrentMap()).GetRandomSpawnPortal()
					char.ChangeMap(mapID, portal, pID)
				}
			})
		} else if len(command) == 3 {
			channel.Players.OnCharacterFromName(command[1], func(char *channel.MapleCharacter) {
				if _, exist := nx.Maps[mapID]; exist {
					portal, pID := channel.Maps.GetMap(char.GetCurrentMap()).GetRandomSpawnPortal()
					char.ChangeMap(mapID, portal, pID)
				}
			})
		}

	case "job":
		val, err := strconv.Atoi(command[1])

		if err != nil {
			return
		}

		channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
			char.SetJob(uint16(val))
		})

	case "level":
		if len(command) == 2 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				return
			}

			channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
				char.SetLevel(byte(val))
			})
		} else if len(command) == 3 {
			val, err := strconv.Atoi(command[2])

			if err != nil {
				return
			}

			channel.Players.OnCharacterFromName(command[1], func(char *channel.MapleCharacter) {
				char.SetLevel(byte(val))
			})
		}
	case "exp":
		if len(command) == 2 {
			val, err := strconv.Atoi(command[1])

			if err != nil {
				return
			}

			channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
				if val > 0 {
					char.GiveEXP(uint32(val))
				} else if val < 0 {
					char.TakeEXP(uint32(val))
				}
			})
		} else if len(command) == 3 {
			val, err := strconv.Atoi(command[2])

			if err != nil {
				return
			}

			channel.Players.OnCharacterFromName(command[1], func(char *channel.MapleCharacter) {
				if val > 0 {
					char.GiveEXP(uint32(val))
				} else if val < 0 {
					char.TakeEXP(uint32(val))
				}
			})

		}
	case "notice":
		if len(command) < 2 {
			return
		}

		msg := strings.Join(command[1:], " ")

		channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
			channel.Maps.GetMap(char.GetCurrentMap()).SendPacket(packets.MessageNotice(msg))
		})
	case "dialogue":
		if len(command) < 2 {
			return
		}

		msg := strings.Join(command[1:], " ")

		channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
			channel.Maps.GetMap(char.GetCurrentMap()).SendPacket(packets.MessageDialogueBox(msg))
		})
	case "mobrate":
		val, err := strconv.Atoi(command[1])

		if err != nil {
			return
		}
		if 0 < val && val < 6 {
			channel.SetRate(channel.MobRate, uint32(val))
		} else {
			conn.Write(packets.MessageDialogueBox("Enter a value between 1 and 5"))
		}
	case "exprate":
		val, err := strconv.Atoi(command[1])

		if err != nil {
			return
		}

		channel.SetRate(channel.ExpRate, uint32(val))
	case "mesorate":
		val, err := strconv.Atoi(command[1])

		if err != nil {
			return
		}

		channel.SetRate(channel.MesoRate, uint32(val))
	case "droprate":
		val, err := strconv.Atoi(command[1])

		if err != nil {
			return
		}

		channel.SetRate(channel.DropRate, uint32(val))
	case "header":
		msg := ""
		if len(command) >= 2 {
			msg = strings.Join(command[1:], " ")
		}

		channel.SetHeader(msg)

		channel.Players.OnCharacters(func(char *channel.MapleCharacter) {
			char.SendPacket(packets.MessageScrollingHeader(msg))
		})
	case "map":
		if len(command) < 2 {
			channel.Players.OnCharacters(func(char *channel.MapleCharacter) {
				char.SendPacket(packets.MessageNotice("Your current map is: " + strconv.Itoa(int(char.GetCurrentMap()))))
			})
		} else {
			switch command[1] {
			case "mobs":
				// mobs information
			case "players":
				// players information
			case "reactors":
				// reactor information
			default:
			}
		}

	case "runNPC":
		if len(command) < 2 {
			return
		}

		npcID, err := strconv.Atoi(command[1])

		if err != nil {
			return
		}

		channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
			npcChat.NewSession(conn, uint32(npcID), char)
			npcChat.GetSession(conn).Run()
		})
	case "restart":
		channel.Players.OnCharacters(func(char *channel.MapleCharacter) {
			err := character.SaveCharacter(&char.Character)

			if err != nil {
				log.Println("Unable to save character data")
			}
		})

		os.Exit(1)

	case "shop":
		items := make(map[uint32]uint32)
		items[1322013] = 1
		items[1092008] = 1
		items[1102054] = 1
		items[1082002] = 1
		items[1072004] = 1
		items[1062007] = 1
		items[1042003] = 1
		items[1032006] = 1
		items[1002140] = 1

		conn.Write(packets.NPCShop(9200000, items))

	default:
		log.Println("Unkown GM command:", msg)
	}
}
