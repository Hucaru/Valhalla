package command

import (
	"github.com/Hucaru/gopacket"
)

type clientConn interface {
	GetUserID() uint32
	Write(gopacket.Packet) error
}

// HandleCommand -
func HandleCommand(conn clientConn, reader gopacket.Reader) {

}
