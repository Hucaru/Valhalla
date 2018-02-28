package handlers

import (
	"log"

	"github.com/Hucaru/gopacket"
)

func HandleChannelPacket(conn *clientChanConn, reader gopacket.Reader) {
	log.Println(reader)
}
