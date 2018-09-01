package server

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/handlers"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/nx"
)

func Channel() {
	log.Println("ChannelServer")

	start := time.Now()
	nx.Parse("wizetData")
	elapsed := time.Since(start)

	log.Println("Loaded and parsed Wizet data in", elapsed)

	channel.GenerateMaps()
	channel.GenerateNPCs()
	channel.GenerateMobs()

	listener, err := net.Listen("tcp", "0.0.0.0:8686")

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer connection.Db.Close()
	connection.ConnectToDb()

	log.Println("Client listener ready")

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error in accepting client", err)
		}

		defer conn.Close()
		clientConnection := connection.NewChannel(connection.NewClient(conn))

		log.Println("New client connection from", clientConnection)

		go connection.HandleNewConnection(clientConnection, func(p maplepacket.Reader) {
			handlers.HandleChannelPacket(clientConnection, p)
		}, constants.ClientHeaderSize, true)
	}
}
