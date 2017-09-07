package handlers

import (
	"log"

	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/worldServer/handlers/channel"
	"github.com/Hucaru/Valhalla/worldServer/handlers/login"
	"github.com/Hucaru/gopacket"
)

func Run() {
	channel.ReadyForChannels = make(chan bool)
	login.LoginServer = make(chan connection.Message)
	channel.ChannelServer = make(chan connection.Message)

	var worldID byte
	channels := make([]bool, 20)

	for {
		select {
		case m := <-login.LoginServer:
			reader := m.Reader

			switch reader.ReadByte() {
			case constants.WORLD_REQUEST_ID:
				worldID = reader.ReadByte()
				log.Println("Assigned world ID: ", worldID)
				channel.ReadyForChannels <- true
			default:
			}

		case m := <-channel.ChannelServer:
			reader := m.Reader

			switch reader.ReadByte() {
			case constants.CHANNEL_REQUEST_ID:
				id := byte(0xFF)

				for i, v := range channels {
					if v == false {
						id = byte(i)
						channels[id] = true
						log.Println("New channel registered:", id+1)
						break
					}
				}

				p := gopacket.NewPacket()
				p.WriteByte(constants.CHANNEL_REQUEST_ID)
				p.WriteByte(worldID)
				p.WriteByte(id + 1)
				m.ReturnChan <- p

			case constants.CHANNEL_USE_SAVED_IDs:
				id := reader.ReadByte()

				if id < 20 && id >= 0 {
					if channels[id] == true {
						log.Println("Channel failed to register using pre-registration:", id)
					} else {
						channels[id] = true
						log.Println("Picked up pre-registered channel:", id)
					}
				}

			case constants.CHANNEL_DROPPED:
				id := reader.ReadByte()
				channels[id-1] = false
				log.Println("Channel dropped:", id)

			default:
			}

		default:
		}
	}
}
