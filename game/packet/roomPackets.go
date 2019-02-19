package packet

import (
	opcodes "github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/mpacket"
)

func RoomShowWindow(roomType, boardType, maxPlayers, roomSlot byte, roomTitle string, chars []def.Character) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x05)
	p.WriteByte(roomType)
	p.WriteByte(maxPlayers)
	p.WriteByte(roomSlot)

	for i, c := range chars {
		p.WriteByte(byte(i))
		p.Append(writeDisplayCharacter(c))
		p.WriteInt32(0) // not sure what this is - memory card game seed? board settings?
		p.WriteString(c.Name)
	}

	p.WriteByte(0xFF)

	if roomType == 0x03 {
		return p
	}

	for i, c := range chars {
		p.WriteByte(byte(i))

		p.WriteInt32(0) // not sure what this is!?
		p.WriteInt32(c.MiniGameWins)
		p.WriteInt32(c.MiniGameDraw)
		p.WriteInt32(c.MiniGameLoss)
		p.WriteInt32(2000) // Points in the ui. What does it represent?
	}

	p.WriteByte(0xFF)
	p.WriteString(roomTitle)
	p.WriteByte(boardType)
	p.WriteByte(0)

	return p
}

func RoomJoin(roomType, roomSlot byte, char def.Character) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x04)
	p.WriteByte(roomSlot)
	p.Append(writeDisplayCharacter(char))
	p.WriteInt32(0) //?
	p.WriteString(char.Name)

	if roomType == 0x03 {
		return p
	}

	p.WriteInt32(1) // not sure what this is!?
	p.WriteInt32(char.MiniGameWins)
	p.WriteInt32(char.MiniGameDraw)
	p.WriteInt32(char.MiniGameLoss)
	p.WriteInt32(2000) // Points in the ui. What does it represent?

	return p
}

func RoomLeave(roomSlot byte, leaveCode byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x0A)
	p.WriteByte(roomSlot)
	p.WriteByte(leaveCode)

	return p
}

func RoomChat(sender, message string, roomSlot byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x06)
	p.WriteByte(8)        // msg type
	p.WriteByte(roomSlot) //
	p.WriteString(sender + " : " + message)

	return p
}

func RoomYellowChat(msgType byte, name string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x06)
	p.WriteByte(7)
	// expelled: 0, x's turn: 1, forfeit: 2, handicap request: 3, left: 4,
	// called to leave: 5, cancelled leave: 6, entered: 7, can't start lack of mesos:8
	// has matched cards: 9
	p.WriteByte(msgType)
	p.WriteString(name)

	return p
}

func RoomInvite(roomType byte, name string, roomID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x02)
	p.WriteByte(roomType)
	p.WriteString(name)
	p.WriteInt32(roomID)

	return p
}

func RoomInviteResult(resultCode byte, name string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x03)
	p.WriteByte(resultCode)
	p.WriteString(name)

	return p
}

func RoomShowAccept() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x0F)

	return p
}

func RoomRequestTie() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x2a)

	return p
}

func RoomRejectTie() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x2b)

	return p
}

func RoomRequestUndo() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x2e)

	return p
}

func RoomRejectUndo() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x2f)
	p.WriteByte(0x00)

	return p
}

func RoomUndo(x, y int32, p1 bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x2f)
	p.WriteByte(0x01)
	// the following bugs out of p1, p2, p1 and p2 requests undo. This will undo p1 move and set the board into a buged out state
	p.WriteByte(0x01)
	p.WriteBool(p1)

	return p
}

func RoomReady() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x32)

	return p
}

func RoomUnready() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x33)

	return p
}

func RoomOmokStart(ownerStart bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x35)
	p.WriteBool(ownerStart)

	return p
}

func RoomMemoryStart(ownerStart bool, boardType int32, cards []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x35)
	p.WriteBool(ownerStart)
	p.WriteByte(0x0C)

	for i := 0; i < len(cards); i++ {
		p.WriteInt32(int32(cards[i]))
	}

	return p
}

func RoomGameResult(draw bool, winningSlot byte, forfeit bool, chars []def.Character) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
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

	p.WriteInt32(chars[0].MiniGameWins)
	p.WriteInt32(chars[0].MiniGameDraw)
	p.WriteInt32(chars[0].MiniGameLoss)
	p.WriteInt32(2000)
	p.WriteInt32(1)
	p.WriteInt32(chars[1].MiniGameWins)
	p.WriteInt32(chars[1].MiniGameDraw)
	p.WriteInt32(chars[1].MiniGameLoss)
	p.WriteInt32(2000)

	return p
}

func RoomGameSkip(isOwner bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x37)
	p.WriteBool(isOwner)

	return p
}

func RoomPlaceOmokPiece(x, y int32, piece byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x38)
	p.WriteInt32(x)
	p.WriteInt32(y)
	p.WriteByte(piece)

	return p
}

func RoomOmokInvalidPlaceMsg() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x39)
	p.WriteByte(0x0)

	return p
}

func RoomSelectCard(turn, cardID, firstCardPick byte, result byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
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

func roomEnterErrorMsg(errorCode byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcodes.SendChannelRoom)
	p.WriteByte(0x05)
	p.WriteByte(0x00)
	p.WriteByte(errorCode)

	return p
}

func RoomClosed() mpacket.Packet {
	return roomEnterErrorMsg(0x01)
}

func RoomFull() mpacket.Packet {
	return roomEnterErrorMsg(0x02)
}

func RoomBusy() mpacket.Packet {
	return roomEnterErrorMsg(0x03)
}

func RoomNotAllowedWhenDead() mpacket.Packet {
	return roomEnterErrorMsg(0x04)
}

func RoomNotAllowedDuringEvent() mpacket.Packet {
	return roomEnterErrorMsg(0x05)
}

func RoomThisCharacterNotAllowed() mpacket.Packet {
	return roomEnterErrorMsg(0x06)
}

func RoomNoTradeAtm() mpacket.Packet {
	return roomEnterErrorMsg(0x07)
}

func RoomMiniRoomNotHere() mpacket.Packet {
	return roomEnterErrorMsg(0x08)
}

func RoomTradeRequireSameMap() mpacket.Packet {
	return roomEnterErrorMsg(0x09)
}

func RoomcannotCreateMiniroomHere() mpacket.Packet {
	return roomEnterErrorMsg(0x0a)
}

func RoomCannotStartGameHere() mpacket.Packet {
	return roomEnterErrorMsg(0x0b)
}

func RoomPersonalStoreFMOnly() mpacket.Packet {
	return roomEnterErrorMsg(0x0c)
}
func RoomGarbageMsgAboutFloorInFm() mpacket.Packet {
	return roomEnterErrorMsg(0x0d)
}

func RoomMayNotEnterStore() mpacket.Packet {
	return roomEnterErrorMsg(0x0e)
}

func RoomStoreMaintenance() mpacket.Packet {
	return roomEnterErrorMsg(0x0F)
}

func RoomCannotEnterTournament() mpacket.Packet {
	return roomEnterErrorMsg(0x10)
}

func RoomGarbageTradeMsg() mpacket.Packet {
	return roomEnterErrorMsg(0x11)
}

func RoomNotEnoughMesos() mpacket.Packet {
	return roomEnterErrorMsg(0x12)
}

func RoomIncorrectPassword() mpacket.Packet {
	return roomEnterErrorMsg(0x13)
}
