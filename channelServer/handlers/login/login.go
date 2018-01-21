package login

import (
	"log"
	"net"
	"time"

	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

var connected chan bool

var LoginServer chan connection.Message
var LoginServerMsg chan gopacket.Packet
var InternalMsg chan connection.Message

func Handle(port uint16, validWorld chan bool) {
	LoginServer = make(chan connection.Message)
	LoginServerMsg = make(chan gopacket.Packet)
	InternalMsg = make(chan connection.Message)
	connected = make(chan bool)
	<-validWorld

	savedWorldID := byte(0xFF)
	savedChannelID := byte(0xFF)
	useSavedIDs := false

	for {

		conn, err := net.Dial("tcp", "0.0.0.0:8486")

		if err != nil {
			log.Println("Could not connect to login server attemping a retry in 3 seconds")
			duration := time.Second
			time.Sleep(duration * 3)
			continue
		}

		defer conn.Close()

		loginConnection := newConnection(conn)

		go manager(loginConnection, port, savedWorldID, savedChannelID, useSavedIDs)

		go connection.HandleNewConnection(loginConnection, func(p gopacket.Reader) {
			handleLoginPacket(loginConnection, p)
		}, constants.INTERSERVER_HEADER_SIZE, false)

		<-connected

		savedWorldID = loginConnection.GetWorldID()
		savedChannelID = loginConnection.GetchannelID()
		useSavedIDs = true
	}
}

func manager(conn *Connection, port uint16, worldID byte, channelID byte, useSaved bool) {
	// Need to have the manager be send the old connection info so that when it attempts to reconnect with login server it uses the archived info
	if useSaved {
		conn.Write(sendID(worldID, channelID, 1, []byte{192, 168, 1, 117}, port))
		conn.SetWorldID(worldID)
		conn.SetChannelID(channelID)
		log.Println("Re-registered with login server using old IDs:", worldID, "-", channelID)
	} else {
		m := <-LoginServer
		reader := m.Reader
		conn.SetWorldID(reader.ReadByte())
		conn.SetChannelID(reader.ReadByte())
		conn.Write(sendID(conn.GetWorldID(), conn.GetchannelID(), 1, []byte{192, 168, 1, 117}, port))
	}

	var pendingMigrations []uint32

	for {
		select {
		case m := <-LoginServerMsg:
			reader := gopacket.NewReader(&m)
			charID := reader.ReadUint32()
			pendingMigrations = append(pendingMigrations, charID)

			log.Println("Migration information received for character id:", charID)
		case m := <-InternalMsg:
			charID := m.Reader.ReadUint32()

			for _, v := range pendingMigrations {
				if v == charID {
					m.ReturnChan <- []byte{0x1}
					continue
				}
			}

			m.ReturnChan <- []byte{0x0}

		default:
		}
	}
}

func handleLoginPacket(conn *Connection, reader gopacket.Reader) {
	log.Println("Received message from login server")
	LoginServerMsg <- reader.GetBuffer()
}

func ValidateMigration(charID uint32) bool {
	check := gopacket.NewPacket()
	check.WriteUint32(charID)

	result := make(chan gopacket.Packet)
	InternalMsg <- connection.NewMessage(check, result)
	validMigration := <-result
	r := gopacket.NewReader(&validMigration)

	if r.ReadByte() == 0x01 {
		return true
	}

	return false
}
