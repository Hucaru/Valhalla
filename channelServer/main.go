package main

import (
	"log"
	"os"
	"time"

	"github.com/Hucaru/Valhalla/channelServer/handlers/client"
	"github.com/Hucaru/Valhalla/channelServer/handlers/login"
	"github.com/Hucaru/Valhalla/channelServer/handlers/world"
	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/nx"
	"github.com/Hucaru/gopacket"
)

const (
	protocol = "tcp"
	address  = "0.0.0.0"
	port     = 8684
)

func main() {
	log.Println("Channel Server")

	start := time.Now()
	nx.Parse("Data.nx")
	elapsed := time.Since(start)

	log.Println("Loaded and parsed nx in", elapsed)

	listener, err, port := connection.CreateServerListener(protocol, address, port)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	validWorld := make(chan bool)

	go world.Handle(validWorld)
	go login.Handle(port, validWorld)

	defer connection.Db.Close()
	connection.ConnectToDb()

	log.Println("Client listener ready")

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error in accepting client", err)
		}

		defer conn.Close()
		channelConnection := client.NewConnection(conn)

		log.Println("New Client connection from", channelConnection)

		go connection.HandleNewConnection(channelConnection, func(p gopacket.Reader) {
			client.HandlePacket(channelConnection, p)
		}, constants.CLIENT_HEADER_SIZE, true)
	}
}
