package login

import (
	"log"
	"net"
	"time"

	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/gopacket"
)

var connected chan bool

var LoginServer chan connection.Message
var LoginServerMsg chan gopacket.Packet

func Handle(port uint16, validWorld chan bool) {
	LoginServer = make(chan connection.Message)
	LoginServerMsg = make(chan gopacket.Packet)
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

		<-connected

		savedWorldID = loginConnection.GetWorldID()
		savedChannelID = loginConnection.GetchannelID()
		useSavedIDs = true
	}
}

func manager(conn *Connection, port uint16, worldID byte, channelID byte, useSaved bool) {
	ip := getInterfaceIP()

	if useSaved {
		conn.Write(sendID(worldID, channelID, 1, ip, port))
		conn.SetWorldID(worldID)
		conn.SetChannelID(channelID)
		log.Println("Re-registered with login server using old IDs:", worldID, "-", channelID)
	} else {
		m := <-LoginServer
		reader := m.Reader
		conn.SetWorldID(reader.ReadByte())
		conn.SetChannelID(reader.ReadByte())
		conn.Write(sendID(conn.GetWorldID(), conn.GetchannelID(), 1, ip, port))
	}
}

func getInterfaceIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
