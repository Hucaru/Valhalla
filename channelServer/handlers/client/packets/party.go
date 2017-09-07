package packet

import (
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

func sendPartyJoinResutlMsg(msg byte) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PARTY_INFO)
	p.WriteByte(msg) // 8 you have already joined a party, 10 party you're trying to join in full

	return p
}

func sendPartyInviteNotification(charName string, id int32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PARTY_INFO)
	p.WriteByte(0x04)
	p.WriteInt32(id) // ? is it the party id?
	p.WriteString(charName)

	return p
}

func sendPartyInviteResult(result byte, invited string) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PARTY_INFO)
	p.WriteByte(result) // 0x15 has denied invite, 0x14 - taking care of another invitation, 0x13 currently blocking party invites
	p.WriteString(invited)

	return p
}
