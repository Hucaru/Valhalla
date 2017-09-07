package world

import (
	"log"
	"net"
	"os"

	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

var WorldServer chan connection.Message

func StartListening() {
	WorldServer = make(chan connection.Message)

	listener, err := net.Listen("tcp", "0.0.0.0"+":"+"8485")

	log.Println("Listening for world connections")

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error in accepting server", err)
		}

		defer conn.Close()

		worldConnection := newConnection(conn)

		log.Println("New world connection from", worldConnection)
		go connection.HandleNewConnection(worldConnection, func(p gopacket.Reader) {
			handleWorldPacket(worldConnection, p)
		}, constants.INTERSERVER_HEADER_SIZE, false)
	}
}

func handleWorldPacket(conn *Connection, reader gopacket.Reader) {
	switch reader.ReadByte() {
	case constants.WORLD_REQUEST_ID:
		worldRequestsID(conn, reader)
	case constants.WORLD_UPDATE_WORLD:
		WorldServer <- connection.NewMessage(reader.GetBuffer(), nil)
	case constants.WORLD_UPDATE_CHANNEL:
		WorldServer <- connection.NewMessage(reader.GetBuffer(), nil)
	case constants.WORLD_NEW_CHANNEL:
		WorldServer <- connection.NewMessage(reader.GetBuffer(), nil)
	case constants.WORLD_DELETE_CHANNEL:
		WorldServer <- connection.NewMessage(reader.GetBuffer(), nil)
	default:
		log.Println("UNKNOWN PACKET FROM WORLD SERVER", reader)
	}
}

func worldRequestsID(conn *Connection, reader gopacket.Reader) {
	p := gopacket.NewPacket()
	p.WriteByte(constants.WORLD_REQUEST_ID)

	resultChan := make(chan gopacket.Packet)

	WorldServer <- connection.NewMessage(p, resultChan)

	id := <-resultChan
	r := gopacket.NewReader(&id)
	conn.setWorldID(r.ReadByte())

	conn.Write([]byte{constants.WORLD_REQUEST_ID, conn.getWorldID()})
}
