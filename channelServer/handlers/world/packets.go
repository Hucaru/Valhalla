package world

import (
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

func sendRequestID() gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.CHANNEL_REQUEST_ID)

	return p
}

func sendSavedRegistration(channel byte) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.CHANNEL_USE_SAVED_IDs)
	p.WriteByte(channel)

	return p
}
