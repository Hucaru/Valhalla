package server

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/player"
)

func packetMessageScrollingHeader(msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBroadcastMessage)
	p.WriteByte(4)
	p.WriteBool(bool(len(msg) > 0))
	p.WriteString(msg)

	return p
}

func packetPlayerAvatarSummaryWindow(charID int32, plr player.Data) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelAvatarInfoWindow)
	p.WriteInt32(plr.ID())
	p.WriteByte(plr.Level())
	p.WriteInt16(plr.Job())
	p.WriteInt16(plr.Fame())

	p.WriteString(plr.Guild())

	p.WriteBool(false) // if has pet
	p.WriteByte(0)     // wishlist count

	return p
}
