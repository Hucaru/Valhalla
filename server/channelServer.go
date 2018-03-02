package server

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/data"
	"github.com/Hucaru/Valhalla/handlers"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/player"
	"github.com/Hucaru/gopacket"
)

func Channel(configFile string) {
	log.Println("ChannelServer")

	start := time.Now()
	nx.Parse("Data.nx")
	elapsed := time.Since(start)

	log.Println("Loaded and parsed nx in", elapsed)

	player.RegisterCharactersObj(data.GetCharsPtr())

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
		clientConnection := handlers.NewChanConnection(connection.NewClientConnection(conn))

		log.Println("New client connection from", clientConnection)

		go connection.HandleNewConnection(clientConnection, func(p gopacket.Reader) {
			handlers.HandleChannelPacket(clientConnection, p)
		}, constants.CLIENT_HEADER_SIZE, true)
	}
}
