package packets

import (
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

func sendLieDetectorInit(num1 uint16, num2 uint16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_LIE_DETECTOR_TEST)
	p.WriteByte(0x04)
	p.WriteUint16(num1)
	p.WriteUint16(num2)
	p.WriteByte(1)

	return p
}

func sendLieDetectorUpdate(num1 uint16, num2 uint16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_LIE_DETECTOR_TEST)
	p.WriteByte(0x04)
	p.WriteUint16(num1)
	p.WriteUint16(num2)
	p.WriteByte(0)

	return p
}

func sendLieDetectorUserNotFound() gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_LIE_DETECTOR_TEST)
	p.WriteByte(0x00)

	return p
}

func sendLieDetectorCantUse() gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_LIE_DETECTOR_TEST)
	p.WriteByte(0x01)

	return p
}

func sendLieDetectorUserTestedBefore() gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_LIE_DETECTOR_TEST)
	p.WriteByte(0x02)

	return p
}

func sendLieDetectorCurrentlyUndergoingTest() gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_LIE_DETECTOR_TEST)
	p.WriteByte(0x03)

	return p
}

func sendLieDetectorGuilty() gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_LIE_DETECTOR_TEST)
	p.WriteByte(0x05)

	return p
}

func sendLieDetectorNotGuilty() gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_LIE_DETECTOR_TEST)
	p.WriteByte(0x06)

	return p
}

func sendLieDetectorGoodReport() gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_LIE_DETECTOR_TEST)
	p.WriteByte(0x07)

	return p
}
