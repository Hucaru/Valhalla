package client

import (
	"encoding/hex"
	"log"
	"strconv"
	"strings"

	"github.com/Hucaru/Valhalla/channelServer/handlers/client/packets"

	"github.com/Hucaru/Valhalla/channelServer/handlers/world"
	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/nx"
	"github.com/Hucaru/gopacket"
)

func HandlePacket(conn *Connection, reader gopacket.Reader) {
	opcode := reader.ReadByte()

	switch opcode {
	case constants.RECV_CHANNEL_PLAYER_LOAD:
		handlePlayerLoad(reader, conn)
	case constants.RECV_CHANNEL_MOVEMENT:
	case constants.RECV_CHANNEL_PLAYER_SEND_ALL_CHAT:
		handlePlayerSendAllChat(reader, conn)
	case constants.RECV_CHANNEL_ADD_BUDDY:
	default:
		log.Println("UNKNOWN CHANNEL PACKET:", reader)
	}
}

func handlePlayerSendAllChat(reader gopacket.Reader, conn *Connection) {
	msg := reader.ReadString(int(reader.ReadInt16()))
	ind := strings.Index(msg, "!")

	if ind == 0 && conn.isAdmin {
		command := strings.SplitN(msg[ind+1:], " ", -1)
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
				panic(err)
			}

			id := uint32(val)

			if _, ok := nx.Maps[id]; ok {
				SendChangeMap(conn, uint32(id), 0, 0, conn.GetCharacter().HP)
			} else {
				// check if player id in else if
			}
		default:
			log.Println("Unkown GM command", command)
		}

	}
}

func handlePlayerLoad(reader gopacket.Reader, conn *Connection) {
	charID := reader.ReadUint32() // validate this and net address from the migration packet

	if !validateNewConnection(charID) {
		conn.Close()
	}

	_, channelID := world.GetAssignedIDs()

	char := character.GetCharacter(charID)
	char.Equips = character.GetCharacterItems(char.CharID)
	char.Skills = character.GetCharacterSkills(char.CharID)

	conn.SetCharacter(char)
	conn.SetChanneldID(uint32(channelID))
	conn.SetAdmin(true)

	conn.Write(packets.SpawnGame(char, uint32(channelID)))

	// npc spawn
	life := nx.Maps[char.CurrentMap].Life
	for i, v := range life {
		if v.Npc {
			conn.Write(packets.SpawnNPC(uint32(i), v))
		}
	}
}

func SendChangeMap(conn *Connection, mapID uint32, channelID uint32, mapPos byte, hp uint16) {
	conn.Write(packets.ChangeMap(mapID, channelID, mapPos, hp))

	// npc spawn
	life := nx.Maps[mapID].Life
	for i, v := range life {
		if v.Npc {
			conn.Write(packets.SpawnNPC(uint32(i), v))
		}
	}
}

func validateNewConnection(charID uint32) bool {
	var migratingWorldID, migratingChannelID int8
	err := connection.Db.QueryRow("SELECT isMigratingWorld,isMigratingChannel FROM characters where id=?", charID).Scan(&migratingWorldID, &migratingChannelID)

	if err != nil {
		panic(err.Error())
	}

	if migratingWorldID < 0 || migratingChannelID < 0 {

		return false
	}

	msg := make(chan gopacket.Packet)
	world.InterServer <- connection.NewMessage([]byte{constants.CHANNEL_GET_INTERNAL_IDS}, msg)
	result := <-msg
	r := gopacket.NewReader(&result)

	if r.ReadByte() != byte(migratingWorldID) && r.ReadByte() != byte(migratingChannelID) {
		log.Println("Received invalid migration info for character", charID, "remote hacking")
		records, err := connection.Db.Query("UPDATE characters set migratingWorldID=?, migratingChannelID=? WHERE id=?", -1, -1, charID)

		defer records.Close()

		if err != nil {
			panic(err.Error())
		}

		return false
	}

	return true
}
