package handlers

import (
	"log"
	"strconv"

	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/loginServer/handlers/channel"
	"github.com/Hucaru/Valhalla/loginServer/handlers/world"
	"github.com/Hucaru/gopacket"
)

type channelData struct {
	ID         byte
	Name       string
	Population int32
	WorldID    byte
	IP         []byte
	Port       uint16
}

type worldData struct {
	ID         byte
	Name       string
	Ribbon     byte
	Event      string
	Channels   [20]channelData
	Warning    byte
	Population byte
	AdminOnly  bool
}

var LoginServer chan connection.Message

func Manager() {
	LoginServer = make(chan connection.Message)

	worlds := make([]worldData, 14)

	for i := range worlds {
		worlds[i] = worldData{ID: 0xFF, AdminOnly: false, Event: "", Name: world.WORLD_NAMES[i], Warning: 0x0, Population: 0x0, Ribbon: 0x0}

		for j := range worlds[i].Channels {
			worlds[i].Channels[j].ID = 0xFF
		}
	}

	for {
		select {
		case m := <-world.WorldServer:
			reader := m.Reader

			switch reader.ReadByte() {
			case constants.WORLD_DROPPED:
				handleDroppedWorld(reader, &worlds)
			case constants.WORLD_REQUEST_ID:
				handelWorldRequestID(reader, &worlds, m.ReturnChan)
			case constants.WORLD_UPDATE_WORLD:
				handleUpdateWorldInfo(reader, worlds)
			default:
			}

		case m := <-channel.ChannelServer:
			reader := m.Reader
			switch reader.ReadByte() {
			case constants.CHANNEL_REGISTER:
				handleChannelRegistration(reader, &worlds)
			case constants.CHANNEL_UPDATE:
				handleChannelRegistration(reader, &worlds)
			case constants.CHANNEL_DROPPED:
				handleChannelDropped(reader, &worlds)
			default:

			}

		case m := <-LoginServer:
			reader := m.Reader

			switch reader.ReadByte() {
			case constants.LOGIN_REQUEST_WORLD_SUMMARY:
				m.ReturnChan <- SendWorlds(worlds)
			case constants.LOGIN_REQUEST_WORLD_STATUS:
				worldID := reader.ReadInt16()
				m.ReturnChan <- SendWorldInfo(worlds[worldID].Warning, worlds[worldID].Population)
			case constants.LOGIN_REQUEST_MIGRATION_INFO:
				handleLoginRequestMigration(reader, worlds, m.ReturnChan)
			default:
				log.Println("UNRECOGNISED INTERNAL REQUEST PACKET", reader)
			}
		default:
		}
	}
}

func handleChannelRegistration(reader gopacket.Reader, worlds *[]worldData) {
	worldID := reader.ReadByte()
	channelID := reader.ReadByte()
	population := reader.ReadInt32()
	ip := reader.ReadBytes(4)
	port := reader.ReadUint16()

	newChannel := channelData{ID: channelID,
		Name:       world.WORLD_NAMES[worldID] + "-" + strconv.Itoa(int(channelID)),
		Population: population,
		WorldID:    worldID,
		IP:         ip,
		Port:       port}

	if worldID < 15 && channelID < 21 {
		(*worlds)[worldID].Channels[channelID-1] = newChannel
		log.Println("New channel registered:", worldID, "-", channelID)
	} else {
		log.Println("Channel sent invalid registration IDs:", worldID, "-", channelID)
	}

}

func handleChannelUpdate(reader gopacket.Reader, worlds *[]worldData) {
	worldID := reader.ReadByte()
	channelID := reader.ReadByte()
	population := reader.ReadInt32()
	maxPopulation := reader.ReadInt32()

	percentage := int32(1200.0 * (float64(population) / float64(maxPopulation)))
	(*worlds)[worldID].Channels[channelID].Population = percentage
}

func handleChannelDropped(reader gopacket.Reader, worlds *[]worldData) {
	worldID := reader.ReadByte()
	channelID := reader.ReadByte()

	if worldID < 15 && channelID < 21 && channelID > 0 {
		(*worlds)[worldID].Channels[channelID-1].ID = 0xFF
		log.Println("Dropped channel:", worldID, "-", channelID)
	} else {
		log.Println("Invalid dropped IDs:", worldID, "-", channelID)
	}
}

func handleDroppedWorld(reader gopacket.Reader, worlds *[]worldData) {
	worldID := reader.ReadByte()

	for i := range *worlds {
		if (*worlds)[i].ID == worldID {
			(*worlds)[i].ID = 0xFF
			break
		}
	}

	log.Println("Dropped world:", worldID)
}

func handelWorldRequestID(reader gopacket.Reader, worlds *[]worldData, returnChan chan gopacket.Packet) {
	result := gopacket.NewPacket()

	id := byte(0xFF)

	for i := range *worlds {
		if (*worlds)[i].ID == byte(0xFF) {
			id = byte(i)
			break
		}
	}

	if id != 0xFF { // Valid world request
		(*worlds)[id].ID = id

		log.Println("New world with id:", id)
	} else {
		log.Println("Rejected world server: max number of worlds reached")
	}

	result.WriteByte(id)
	returnChan <- result
}

func handleUpdateWorldInfo(reader gopacket.Reader, worlds []worldData) {
	worldID := reader.ReadByte()
	switch reader.ReadByte() {
	case constants.WORLD_POPULATION:
		worlds[worldID].Population = reader.ReadByte()

	case constants.WORLD_EVENT:
		worlds[worldID].Event = reader.ReadString(int(reader.ReadInt16()))

	case constants.WORLD_ADMIN_ONLY:
		if reader.ReadByte() == 1 {
			worlds[worldID].AdminOnly = true
		} else {
			worlds[worldID].AdminOnly = false
		}

	case constants.WORLD_WARNING:
		worlds[worldID].Warning = reader.ReadByte()

	case constants.WORLD_RIBBON:
		worlds[worldID].Ribbon = reader.ReadByte()
	default:
		log.Println("Unkown update world packet from world server", reader)
	}
}

func handleLoginRequestMigration(reader gopacket.Reader, worlds []worldData, returnChan chan gopacket.Packet) {
	worldID := reader.ReadByte()
	channelID := reader.ReadByte()
	charID := reader.ReadInt32()
	ch := worlds[int(worldID)].Channels[int(channelID)]

	var migratingWorldID, migratingChannelID int8
	err := connection.Db.QueryRow("SELECT isMigratingWorld,isMigratingChannel FROM characters where id=?", charID).Scan(&migratingWorldID, &migratingChannelID)

	if err != nil {
		panic(err.Error())
	}

	if migratingWorldID > -1 && migratingChannelID > -1 {
		log.Println(charID, "Is trying to enter the game when already logged in", worldID, ":", channelID)
	}

	log.Println(migratingWorldID, migratingChannelID)

	records, err := connection.Db.Query("UPDATE characters set isMigratingWorld=?, isMigratingChannel=? WHERE id=?", worldID, channelID, charID)

	defer records.Close()

	if err != nil {
		panic(err.Error())
	}

	returnChan <- SendMigrationInfo(ch.IP, ch.Port)
}
