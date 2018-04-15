package server

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/Hucaru/Valhalla/inventory"
	"github.com/Hucaru/Valhalla/skills"

	"github.com/Hucaru/Valhalla/command"
	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/data"
	"github.com/Hucaru/Valhalla/handlers"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/maps"
	"github.com/Hucaru/Valhalla/message"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/player"
)

func Channel() {
	log.Println("ChannelServer")

	start := time.Now()
	nx.Parse("Data.nx")
	elapsed := time.Since(start)

	log.Println("Loaded and parsed nx in", elapsed)

	data.GenerateMapsObject()

	player.RegisterCharactersObj(data.GetCharsPtr())
	message.RegisterCharactersObj(data.GetCharsPtr())
	maps.RegisterCharactersObj(data.GetCharsPtr())
	maps.RegisterMapsObj(data.GetMapsPtr())
	skills.RegisterCharactersObj(data.GetCharsPtr())
	inventory.RegisterCharactersObj(data.GetCharsPtr())
	command.RegisterCharactersObj(data.GetCharsPtr())
	command.RegisterMapsObj(data.GetMapsPtr())

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
		clientConnection := connection.NewChanConnection(connection.NewClientConnection(conn))

		log.Println("New client connection from", clientConnection)

		go connection.HandleNewConnection(clientConnection, func(p maplepacket.Reader) {
			handlers.HandleChannelPacket(clientConnection, p)
		}, constants.CLIENT_HEADER_SIZE, true)
	}
}
