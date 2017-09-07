package login

import (
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

func requestID() gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.WORLD_REQUEST_ID)

	return p
}

func assignedWorldID(worldID byte) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.WORLD_REQUEST_ID)
	p.WriteByte(worldID)

	return p
}
