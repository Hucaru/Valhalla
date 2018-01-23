package packets

import (
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

func sendNotice(msg string) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_BROADCAST_MESSAGE)
	p.WriteByte(0)
	p.WriteString(msg)

	return p
}

func sendDialogueBox(msg string) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_BROADCAST_MESSAGE)
	p.WriteByte(1)
	p.WriteString(msg)

	return p
}

// Need to figure out how to display the username and message atm it bastardises it.
func sendBroadvastChannelMessage(senderName string, msg string, channel byte, isSameChannel bool) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_BROADCAST_MESSAGE)
	p.WriteByte(3)
	p.WriteString(senderName)
	p.WriteByte(channel)
	if isSameChannel {
		p.WriteByte(0x01)
	} else {
		p.WriteByte(0x00)
	}

	return p
}

func sendBubblessChat(msgType byte, sender string, msg string) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_BUBBLESS_CHAT)
	p.WriteByte(msgType) // 0x00 buddy chat, 0x01 - party, 0x02 - guild
	p.WriteString(sender)
	p.WriteString(msg)

	return p
}
