package login

import (
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

func sendID(worldId byte, channelID byte, population int32, ip []byte, port uint16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.CHANNEL_REGISTER)
	p.WriteByte(worldId)
	p.WriteByte(channelID)
	p.WriteInt32(population)
	p.WriteBytes(ip)
	p.WriteUint16(port)

	return p
}
