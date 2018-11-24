package world

import (
	"log"
	"net"

	"github.com/Hucaru/Valhalla/mpacket"
)

func HandlePacket(loginConn net.Conn, reader mpacket.Reader) {
	switch reader.ReadByte() {
	default:
		log.Println("Unkown packet:", reader)
	}
}

func worldnfo(reader mpacket.Reader) {

}

func channelID(reader mpacket.Reader) {

}

func newPlayer(conn net.Conn, reader mpacket.Reader) {
	//conn.Send(packet.ServerWorldInformation())
}
