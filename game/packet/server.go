package packet

import (
	"github.com/Hucaru/Valhalla/maplepacket"
)

func ServerWorldInformation(name, msg string, ribbon, expEvent, nChannels byte, population []int32, addresses [][]byte, ports []int16) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(0x00)
	p.WriteString(name)
	p.WriteByte(ribbon)
	p.WriteString(msg)
	p.WriteByte(expEvent)
	p.WriteByte(nChannels)
	p.WriteByte(byte(len(population)))

	for i, v := range population {
		p.WriteInt32(v)
		p.WriteBytes(addresses[i])
		p.WriteInt16(ports[i])
	}

	return p
}

func ServerChannelID(id byte) maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(0x01)
	p.WriteByte(id)

	return p
}

func ServerNewPlayer() maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(0x02)
	return p
}

func ServerPlayerLeave() maplepacket.Packet {
	p := maplepacket.CreateWithOpcode(0x03)
	return p
}
