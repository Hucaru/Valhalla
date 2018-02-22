package world

import (
	"log"
	"net"
	"time"

	"github.com/Hucaru/Valhalla/channelServer/login"
	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

var connected chan bool
var worldServer chan connection.Message
var InterServer chan connection.Message

func Handle(validWorld chan bool) {
	connected = make(chan bool)

	savedWorldID := byte(0xFF)
	savedChannelID := byte(0xFF)
	useSavedIDs := false

	for {
		conn, err := net.Dial("tcp", "127.0.0.1:8585")

		if err != nil {
			log.Println("Could not connect to world server attemping a retry in 3 seconds")
			duration := time.Second
			time.Sleep(duration * 3)
			continue
		}

		defer conn.Close()

		worldConnection := newConnection(conn)

		go manager(worldConnection, validWorld, savedWorldID, savedChannelID, useSavedIDs)

		go connection.HandleNewConnection(worldConnection, func(p gopacket.Reader) {
			handleWorldPacket(worldConnection, p)
		}, constants.INTERSERVER_HEADER_SIZE, false)
		<-connected

		savedWorldID = worldConnection.GetWorldID()
		savedChannelID = worldConnection.GetchannelID()
		useSavedIDs = true
	}
}

func manager(conn *Connection, validWorld chan bool, worldID byte, channelID byte, useSaved bool) {
	worldServer = make(chan connection.Message)
	InterServer = make(chan connection.Message)

	if useSaved {
		conn.Write(sendSavedRegistration(channelID))
		conn.SetWorldID(worldID)
		conn.SetchannelID(channelID)
		log.Println("Re-registered with world server using old id:", channelID)
	} else {
		conn.Write(sendRequestID())
	}
	for {
		select {
		case m := <-worldServer:
			reader := m.Reader

			switch reader.ReadByte() {
			case constants.CHANNEL_REQUEST_ID:
				worldID := reader.ReadByte()
				channelID := reader.ReadByte()

				conn.SetWorldID(worldID)
				conn.SetchannelID(channelID)

				log.Println("Assigned id:", worldID, "-", channelID)

				validWorld <- true

				p := gopacket.NewPacket()
				p.WriteByte(worldID)
				p.WriteByte(channelID)

				login.LoginServer <- connection.NewMessage(p, nil)
			default:
				log.Println("UNKOWN MANAGER PACKET:", reader)
			}
		case m := <-InterServer:
			reader := m.Reader
			switch reader.ReadByte() {
			case constants.CHANNEL_GET_INTERNAL_IDS:
				p := gopacket.NewPacket()
				p.WriteByte(conn.GetWorldID())
				p.WriteByte(conn.GetchannelID())
				m.ReturnChan <- p
			}
		}
	}
}

func handleWorldPacket(conn *Connection, reader gopacket.Reader) {
	worldServer <- connection.NewMessage(reader.GetBuffer(), nil)
}

func GetAssignedIDs() (byte, byte) {
	msg := make(chan gopacket.Packet)
	InterServer <- connection.NewMessage([]byte{constants.CHANNEL_GET_INTERNAL_IDS}, msg)
	result := <-msg

	r := gopacket.NewReader(&result)
	world := r.ReadByte()
	channel := r.ReadByte() - 1

	return world, channel
}
