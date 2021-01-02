package channel

import (
	"github.com/Hucaru/Valhalla/common/mpacket"
	"github.com/Hucaru/Valhalla/common/opcode"
)

type guild struct {
	id       int32
	capacity int32
	notice   string
	name     string

	rank1 string
	rank2 string
	rank3 string
	rank4 string
	rank5 string

	names  []string
	jobs   []int32
	levels []int32
	online []bool
	ranks  []int32

	logoBg, logoBgColour, logo int16
	logoColour                 byte
}

func packetGuildInfo(id int32, name string, memberCount byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x1a)

	if len(name) == 0 {
		p.WriteByte(0x00) // removes player from guild
		return p
	}

	p.WriteBool(true) // In guild
	p.WriteInt32(1)   // guild id (value cannot be zero)
	p.WriteString(name)

	// 5 ranks each have a title
	p.WriteString("rank1")
	p.WriteString("rank2")
	p.WriteString("rank3")
	p.WriteString("rank4")
	p.WriteString("rank5")

	capacity := 250             // maximum
	p.WriteByte(byte(capacity)) // member count

	// iterate over all members and output ids
	for i := 0; i < capacity; i++ {
		p.WriteInt32(int32(i + 1))
	}

	// iterate over all members and input their info
	for i := 0; i < capacity; i++ {
		p.WritePaddedString("[GM]Hucaru", 13) // name
		p.WriteInt32(510)                     // job
		p.WriteInt32(255)                     // level

		if i > 4 {
			p.WriteInt32(5) // rank starts at 1
		} else {
			p.WriteInt32(int32(i + 1)) // rank starts at 1
		}

		if i%2 == 0 {
			p.WriteInt32(1) // online or not
		} else {
			p.WriteInt32(0)
		}

		p.WriteInt32(int32(i)) // ?
	}

	p.WriteInt32(int32(capacity)) // capacity
	p.WriteInt16(1030)            // logo background
	p.WriteByte(3)                // logo bg colour
	p.WriteInt16(4017)            // logo
	p.WriteByte(2)                // logo colour
	p.WriteString("notice")       // notice
	p.WriteInt32(9999)            // ?

	return p
}
