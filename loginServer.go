package main

import (
	"log"
	"net"
	"os"

	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/handlers"
	"github.com/Hucaru/gopacket"
)

const (
	protocol = "tcp"
	address  = "0.0.0.0"
	port     = "8484"
)

func main() {
	log.Println("LoginServer")

	listener, err := net.Listen("tcp", "0.0.0.0:8484")

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
		clientConnection := handlers.NewLoginConnection(connection.NewClientConnection(conn))

		log.Println("New client connection from", clientConnection)

		go connection.HandleNewConnection(clientConnection, func(p gopacket.Reader) {
			handlers.HandleLoginPacket(clientConnection, p)
		}, constants.CLIENT_HEADER_SIZE, true)
	}
}
