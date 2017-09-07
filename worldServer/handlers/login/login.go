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
var LoginSender chan gopacket.Packet

func Connect() {
	connected = make(chan bool)
	for {
		conn, err := net.Dial("tcp", "127.0.0.1:8485")

		if err != nil {
			log.Println("Could not connect to login server attemping a retry in 3 seconds")
			duration := time.Second
			time.Sleep(duration * 3)
			continue
		}

		defer conn.Close()

		loginConnection := newConnection(conn)
		loginConnection.Write(requestID())

		go connection.HandleNewConnection(loginConnection, func(p gopacket.Reader) {
			handleLoginPacket(loginConnection, p)
		}, constants.INTERSERVER_HEADER_SIZE, false)
		<-connected
	}
}

func handleLoginPacket(conn *Connection, reader gopacket.Reader) {
	switch reader.ReadByte() {
	case constants.WORLD_REQUEST_ID:
		worldID := reader.ReadByte()

		if worldID == 0xFF {
			conn.Close()
		}

		LoginServer <- connection.NewMessage(assignedWorldID(worldID), nil)

	default:
		log.Println("UNKOWN LOGIN PACKET:", reader)
	}
}
