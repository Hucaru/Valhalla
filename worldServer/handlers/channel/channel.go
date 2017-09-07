package channel

import (
	"log"
	"net"
	"os"

	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

const (
	protocol = "tcp"
	address  = "0.0.0.0"
	port     = "8585"
)

var ReadyForChannels chan bool
var ChannelServer chan connection.Message

func Handle() {
	<-ReadyForChannels
	listener, err := net.Listen(protocol, address+":"+port) // Need to change this to cycle through ports

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

		go channelSender(channelConnection)

		go connection.HandleNewConnection(channelConnection, func(p gopacket.Reader) {
			handleChannelPacket(channelConnection, p)
		}, constants.INTERSERVER_HEADER_SIZE, false)
	}
}

func channelSender(channelConnection *Connection) {
	// World server induces packets to channel server
}

func handleChannelPacket(conn *Connection, reader gopacket.Reader) {
	switch reader.ReadByte() {
	case constants.CHANNEL_REQUEST_ID:
		result := make(chan gopacket.Packet)
		ChannelServer <- connection.NewMessage(reader.GetBuffer(), result)
		p := <-result
		reader := gopacket.NewReader(&p)
		reader.ReadByte()
		reader.ReadByte()
		conn.SetchannelID(reader.ReadByte())
		conn.Write(reader.GetBuffer())
	case constants.CHANNEL_USE_SAVED_IDs:
		ChannelServer <- connection.NewMessage(reader.GetBuffer(), nil)
	default:
		log.Println("UNKNOWN PACKET FROM CHANNEL SERVER:", reader)
	}
}
