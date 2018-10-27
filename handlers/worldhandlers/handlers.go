package worldhandlers

import (
	"log"
	"net"

	"github.com/Hucaru/Valhalla/maplepacket"
)

func HandlePacket(loginConn net.Conn, reader maplepacket.Reader) {
	switch reader.ReadByte() {
	default:
		log.Println("Unkown packet:", reader)
	}
}

func worldnfo(reader maplepacket.Reader) {

}

func channelID(reader maplepacket.Reader) {

}

func newPlayer(conn net.Conn, reader maplepacket.Reader) {
	//conn.Send(packets.ServerWorldInformation())
}
