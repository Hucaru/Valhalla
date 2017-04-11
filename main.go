package main

import (
	"fmt"
	"net"
	"os"
)

const (
	protocol = "tcp"
	address  = "0.0.0.0"
	port     = "8484"
)

func main() {
	fmt.Println("Test")

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

func handleClientConnection(conn net.Conn) {
	// Send init packet
	conn.Write([]byte{13, 0, 28, 0, 0, 0, 28, 62, 13, 176, 236, 76, 141, 116, 8})
	fmt.Println("Handshake sent")
	// Send recv bullshit
	for {
		buffer := make([]byte, 2)

		_, err := conn.Read(buffer)

		if err != nil {
			fmt.Println("Error in reading from connection")
			return
		}

		fmt.Println("Following bytes received", buffer)
	}
}
