package server

import "github.com/Hucaru/Valhalla/mpacket"

// PacketClientHandshake sent to client on initial connection
func PacketClientHandshake(mapleVersion int16, recv, send []byte) mpacket.Packet {
	p := mpacket.NewPacket()

	p.WriteInt16(13)
	p.WriteInt16(mapleVersion)
	p.WriteString("")
	p.Append(recv)
	p.Append(send)
	p.WriteByte(8)

	return p

}
