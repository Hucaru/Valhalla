package packets

import (
	"github.com/Hucaru/Valhalla/maplepacket"
)

func ClientHandshake(mapleVersion int16, recv, send []byte) maplepacket.Packet {
	p := maplepacket.NewPacket()

	p.WriteInt16(13)
	p.WriteInt16(mapleVersion)
	p.WriteString("")
	p.Append(recv)
	p.Append(send)
	p.WriteByte(8)

	return p

}
