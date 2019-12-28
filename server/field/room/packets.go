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
		p.WriteInt32(v.MiniGamePoints())
	}

	p.WriteByte(0xFF)
	p.WriteString(roomTitle)
	p.WriteByte(boardType)
	p.WriteByte(0)

	return p
}

func packetRoomJoin(roomType, roomSlot byte, plr player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x04)
	p.WriteByte(roomSlot)
	p.Append(plr.DisplayBytes())
	p.WriteInt32(0) //?
	p.WriteString(plr.Name())

	if roomType == 0x03 {
		return p
	}

	p.WriteInt32(1) // not sure what this is!?
	p.WriteInt32(plr.MiniGameWins())
	p.WriteInt32(plr.MiniGameDraw())
	p.WriteInt32(plr.MiniGameLoss())
	p.WriteInt32(2000) // Points in the ui. What does it represent?

	return p
}

func packetRoomLeave(roomSlot byte, leaveCode byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x0A)
	p.WriteByte(roomSlot)
	p.WriteByte(leaveCode)

	return p
}

func packetRoomYellowChat(msgType byte, name string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x06)
	p.WriteByte(7)
	// expelled: 0, x's turn: 1, forfeit: 2, handicap request: 3, left: 4,
	// called to leave/expeled: 5, cancelled leave: 6, entered: 7, can't start lack of mesos:8
	// has matched cards: 9
	p.WriteByte(msgType)
	p.WriteString(name)

	return p
}

func packetRoomReady() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x32)

	return p
}

func packetRoomChat(sender, message string, roomSlot byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x06)
	p.WriteByte(8)        // msg type
	p.WriteByte(roomSlot) //
	p.WriteString(sender + " : " + message)

	return p
}

func packetRoomUnready() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x33)

	return p
}

func packetRoomOmokStart(ownerStart bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x35)
	p.WriteBool(ownerStart)

	return p
}

func packetRoomMemoryStart(ownerStart bool, boardType int32, cards []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x35)
	p.WriteBool(ownerStart)
	p.WriteByte(0x0C)

	for i := 0; i < len(cards); i++ {
		p.WriteInt32(int32(cards[i]))
	}

	return p
}

func packetRoomGameSkip(isOwner bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x37)

	if isOwner {
		p.WriteByte(0)
	} else {
		p.WriteByte(1)
	}

	return p
}

func packetRoomPlaceOmokPiece(x, y int32, piece byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x38)
	p.WriteInt32(x)
	p.WriteInt32(y)
	p.WriteByte(piece)

	return p
}

func packetRoomOmokInvalidPlaceMsg() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x39)
	p.WriteByte(0x0)

	return p
}

func packetRoomSelectCard(turn, cardID, firstCardPick byte, result byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x3c)
	p.WriteByte(turn)

	if turn == 1 {
		p.WriteByte(cardID)
	} else if turn == 0 {
		p.WriteByte(cardID)
		p.WriteByte(firstCardPick)
		p.WriteByte(result)
	}

	return p
}

func packetRoomGameResult(draw bool, winningSlot byte, forfeit bool, plr []player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x36)

	if !draw && !forfeit {
		p.WriteBool(draw)
	} else if draw {
		p.WriteBool(draw)
	} else if forfeit {
		p.WriteByte(2)
	}

	p.WriteByte(winningSlot)

	// Why is there a difference between the two?
	if draw {
		p.WriteByte(0)
		p.WriteByte(0)
		p.WriteByte(0)
	} else {
		p.WriteInt32(1) // ?
	}

	p.WriteInt32(plr[0].MiniGameWins())
	p.WriteInt32(plr[0].MiniGameDraw())
	p.WriteInt32(plr[0].MiniGameLoss())
	p.WriteInt32(plr[0].MiniGamePoints())
	p.WriteInt32(1)
	p.WriteInt32(plr[1].MiniGameWins())
	p.WriteInt32(plr[1].MiniGameDraw())
	p.WriteInt32(plr[1].MiniGameLoss())
	p.WriteInt32(plr[1].MiniGamePoints())

	return p
}

func packetRoomEnterErrorMsg(errorCode byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x05)
	p.WriteByte(0x00)
	p.WriteByte(errorCode)

	return p
}

func packetRoomClosed() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x01)
}

func packetRoomFull() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x02)
}

func packetRoomBusy() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x03)
}

func packetRoomNotAllowedWhenDead() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x04)
}

func packetRoomNotAllowedDuringEvent() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x05)
}

func packetRoomThisCharacterNotAllowed() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x06)
}

func packetRoomNoTradeAtm() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x07)
}

func packetRoomMiniRoomNotHere() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x08)
}

func packetRoomTradeRequireSameMap() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x09)
}

func packetRoomcannotCreateMiniroomHere() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x0a)
}

func packetRoomCannotStartGameHere() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x0b)
}

func packetRoomPersonalStoreFMOnly() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x0c)
}
func packetRoomGarbageMsgAboutFloorInFm() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x0d)
}

func packetRoomMayNotEnterStore() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x0e)
}

func packetRoomStoreMaintenance() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x0F)
}

func packetRoomCannotEnterTournament() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x10)
}

func packetRoomGarbageTradeMsg() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x11)
}

func packetRoomNotEnoughMesos() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x12)
}

func packetRoomIncorrectPassword() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x13)
}
