package room

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

func packetRoomShowWindow(roomType, boardType, maxPlayers, roomSlot byte, roomTitle string, players []player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x05)
	p.WriteByte(roomType)
	p.WriteByte(maxPlayers)
	p.WriteByte(roomSlot)

	for i, v := range players {
		p.WriteByte(byte(i))
		p.Append(v.DisplayBytes())
		p.WriteInt32(0) // not sure what this is - memory card game seed? board settings?
		p.WriteString(v.Name())
	}

	p.WriteByte(0xFF)

	if roomType == 0x03 {
		return p
	}

	for i, v := range players {
		p.WriteByte(byte(i))
		p.WriteInt32(0) // not sure what this is!?
		p.WriteInt32(v.MiniGameWins())
		p.WriteInt32(v.MiniGameDraw())
		p.WriteInt32(v.MiniGameLoss())
		p.WriteInt32(2000) // Points in the ui. What does it represent?
	}

	p.WriteByte(0xFF)
	p.WriteString(roomTitle)
	p.WriteByte(boardType)
	p.WriteByte(0)

	return p
}
