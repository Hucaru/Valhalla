package main

import (
	"fmt"
	"net"
	"os"

	"github.com/Hucaru/Valhalla/common"
)

const (
	protocol = "tcp"
	address  = "0.0.0.0"
	port     = "8484"
)

func main() {
	fmt.Println("LoginServer")

	listener, err := net.Listen(protocol, address+":"+port)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Listener ready")

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error in accepting client", err)
		}

		fmt.Println("New Connection")

		go handleClientConnection(conn)
	}
}

// Move this out into a seperate file and generalise to accept a master handler (contains a switch, can then take and place in common)
func handleClientConnection(conn net.Conn) {
	defer conn.Close()

	clientConn := common.NewClientConnection(conn)

	SendHandshake(clientConn)

	fmt.Println("Handshake sent")

	sizeToRead := 2

	for {
		buffer := common.NewPacket(sizeToRead)

		err := clientConn.Read(buffer)

		if err != nil {
			fmt.Println("Error in reading from connection", err)
		}
	}
}
