package main

import (
	"fmt"
	"net"
	"os"

	"github.com/Hucaru/Valhalla/common/connection"
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/packet"
	"github.com/Hucaru/Valhalla/loginServer/handlers"
	"github.com/Hucaru/Valhalla/loginServer/loginConn"
)

const (
	protocol = "tcp"
	address  = "0.0.0.0"
	port     = "8484"
)

func main() {
	fmt.Println("LoginServer")

	// TODO: Write config reader

	listener, err := net.Listen(protocol, address+":"+port)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer connection.Db.Close()
	connection.ConnectToDb()
	fmt.Println("Listener ready")

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error in accepting client", err)
		}

		defer conn.Close()
		loginConnection := loginConn.NewConnection(conn)

		go connection.HandleNewConnection(loginConnection, func(p packet.Packet) {
			handlers.HandlePacket(loginConnection, p)
		}, constants.CLIENT_HEADER_SIZE)
	}
}
