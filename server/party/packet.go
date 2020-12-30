package party

import (
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

func packetPlayerJoin(partyID int32, name string, party *Data) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPartyInfo)
	p.WriteByte(0x0e)
	p.WriteInt32(partyID)
	p.WriteString(name)

	validOffsets := make([]int, 0, constant.MaxPartySize)

	for i, v := range party.level {
		if v != 0 {
			validOffsets = append(validOffsets, i)
		}
	}

	paddAmount := constant.MaxPartySize - len(validOffsets)

	for _, v := range validOffsets {
		p.WriteInt32(party.playerID[v])
	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(0)
	}

	for _, v := range validOffsets {
		p.WritePaddedString(party.name[v], 13)
	}

	for i := 0; i < paddAmount; i++ {
		p.WritePaddedString("", 13)
	}

	for _, v := range validOffsets {
		p.WriteInt32(party.job[v])
	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(0)
	}

	for _, v := range validOffsets {
		p.WriteInt32(party.level[v])
	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(0)
	}

	for _, v := range validOffsets {
		p.WriteInt32(party.channelID[v]) // -1 - cashshop, -2 - offline
	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(0)
	}

	for _, v := range validOffsets {
		p.WriteInt32(party.mapID[v])
	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(0)
	}

	p.WriteInt32(party.playerID[0])

	// Mystic door
	for range validOffsets {
		p.WriteInt32(-1)
		p.WriteInt32(-1)
		p.WriteInt32(0) // x
		p.WriteInt32(0) // y
	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(-1)
		p.WriteInt32(-1)
		p.WriteInt64(0) // int64?
	}

	// for _, v := range validOffsets {
	// 	p.WriteInt16(party.hp[v])
	// 	p.WriteInt16(party.maxHP[v])
	// }

	// for i := 0; i < paddAmount; i++ {
	// 	p.WriteInt32(0)
	// }

	return p
}
