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
	defer conn.Close()

	clientConn := common.NewClientConnection(conn)

	sendHandshake(clientConn)

	fmt.Println("Handshake sent")

	sizeToRead := 2

	for {
		buffer := common.NewPacket(sizeToRead)

		err := clientConn.Read(buffer)

		if err != nil {
			fmt.Println("Error in reading from connection", err)
			return
		}
	}
}

func sendHandshake(client common.Connection) error {
	packet := common.NewPacket(0)

	packet.WriteShort(28)
	packet.WriteString("")
	packet.WriteInt(1)
	packet.WriteInt(2)
	packet.WriteByte(8)

	err := client.Write(packet)

	return err
}
