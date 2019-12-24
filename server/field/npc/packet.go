package npc

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

func packetNpcSetController(npcID int32, isLocal bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcControl)
	p.WriteBool(isLocal)
	p.WriteInt32(npcID)

	return p
}
