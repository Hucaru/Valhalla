package chat

import (
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/gopacket"
)

func noticePacket(msg string) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_BROADCAST_MESSAGE)
	p.WriteByte(0)
	p.WriteString(msg)

	return p
}

func sendDialogueBoxPacket(msg string) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_BROADCAST_MESSAGE)
	p.WriteByte(1)
	p.WriteString(msg)

	return p
}

// Need to figure out how to display the username and message atm it bastardises it.
func broadvastChannelMessagePacket(senderName string, msg string, channel byte, isSameChannel bool) gopacket.Packet {
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

func bubblessChatPacket(msgType byte, sender string, msg string) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_BUBBLESS_CHAT)
	p.WriteByte(msgType) // 0x00 buddy chat, 0x01 - party, 0x02 - guild
	p.WriteString(sender)
	p.WriteString(msg)

	return p
}

func whisperPacket(sender string, message string) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_WHISPER)
	p.WriteString(sender)
	p.WriteByte(0x00) // Some kind of channel flag, zero is same channel, not sure what non zero means for packet
	p.WriteString(message)

	return p
}

func allChatPacket(senderID uint32, isAdmin bool, msg string) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ALL_CHAT_MSG)
	p.WriteUint32(senderID)
	p.WriteBool(isAdmin)
	p.WriteString(msg)

	return p
}
