package channel

import (
	"log"
	"net"
	"os"

	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

var ChannelServer chan connection.Message
var ChannelSender chan connection.Message

func StartListening() {
	ChannelServer = make(chan connection.Message)
	ChannelSender = make(chan connection.Message)

	listener, err := net.Listen("tcp", "0.0.0.0"+":"+"8486")

	log.Println("Listening for channel connections")

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

		channelConnection := newConnection(conn)
		go sendToChannel(channelConnection)

		log.Println("New channel connection from", channelConnection)
		go connection.HandleNewConnection(channelConnection, func(p gopacket.Reader) {
			handlechannelPacket(channelConnection, p)
		}, constants.INTERSERVER_HEADER_SIZE, false)
	}
}

func sendToChannel(conn *Connection) {
	for {
		m := <-ChannelSender
		conn.Write(m.Reader.GetBuffer())
	}
}

func handlechannelPacket(conn *Connection, reader gopacket.Reader) {
	switch reader.ReadByte() {
	case constants.CHANNEL_REGISTER:
		reader.ReadByte()
		conn.setChannelID(reader.ReadByte())
		ChannelServer <- connection.NewMessage(reader.GetBuffer(), nil)
	case constants.CHANNEL_UPDATE:
		ChannelServer <- connection.NewMessage(reader.GetBuffer(), nil)
	case constants.CHANNEL_DROPPED:
		ChannelServer <- connection.NewMessage(reader.GetBuffer(), nil)
	default:
		log.Println("UNKNOWN PACKET FROM CHANNEL SERVER", reader)
	}
}
