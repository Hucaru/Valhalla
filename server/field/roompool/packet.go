package roompool

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

func packetMapShowGameBox(displayBytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoomBox)
	p.WriteBytes(displayBytes)

	return p
}

func packetMapRemoveGameBox(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoomBox)
	p.WriteInt32(charID)
	p.WriteInt32(0)

	return p
}
