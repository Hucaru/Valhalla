package skills

import (
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/gopacket"
)

func skillAnimationPacket(charID uint32, skillID uint32, level byte) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_ANIMATION)
	p.WriteUint32(charID)
	p.WriteByte(0x01)
	p.WriteUint32(skillID)
	p.WriteByte(level)

	return p
}
