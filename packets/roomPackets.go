package packets

import (
	"fmt"

	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func RoomShowWindow(roomType, boardType, maxPlayers, roomSlot byte, roomTitle string, chars []character.Character) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x05)
	p.WriteByte(roomType)
	p.WriteByte(maxPlayers)
	p.WriteByte(roomSlot)

	for i, c := range chars {
		p.WriteByte(byte(i))
		p.Append(writeDisplayCharacter(c))
		p.WriteInt32(0) // not sure what this is - memory card game seed? board settings?
		p.WriteString(c.GetName())
	}

	p.WriteByte(0xFF)

	if roomType == 0x03 {
		return p
	}

	for i, c := range chars {
		p.WriteByte(byte(i))

		p.WriteInt32(0) // not sure what this is!?
		switch roomType {
		case 0x01:
			p.WriteInt32(c.GetOmokWins())
			p.WriteInt32(c.GetOmokTies())
			p.WriteInt32(c.GetOmokLosses())
		case 0x02:
			p.WriteInt32(c.GetMemoryWins())
			p.WriteInt32(c.GetMemoryTies())
			p.WriteInt32(c.GetMemoryLosses())
		default:
			fmt.Println("Unknown game type", roomType)
		}

		p.WriteInt32(2000) // Points in the ui. What does it represent?
	}

	p.WriteByte(0xFF)
	p.WriteString(roomTitle)
	// extends packet
	p.WriteByte(0) // p1 omok piece?
	p.WriteByte(0) // p2 omok piece?

	return p
}

func RoomJoin(roomType, roomSlot byte, char character.Character) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x04)
	p.WriteByte(roomSlot)
	p.Append(writeDisplayCharacter(char))
	p.WriteInt32(0) //?
	p.WriteString(char.GetName())

	if roomType == 0x03 {
		return p
	}

	p.WriteInt32(0) // not sure what this is!?
	switch roomType {
	case 0x01:
		p.WriteInt32(char.GetOmokWins())
		p.WriteInt32(char.GetOmokTies())
		p.WriteInt32(char.GetOmokLosses())
	case 0x02:
		p.WriteInt32(char.GetMemoryWins())
		p.WriteInt32(char.GetMemoryTies())
		p.WriteInt32(char.GetMemoryLosses())
	default:
		fmt.Println("Unknown game type", roomType)
	}

	p.WriteInt32(2000) // Points in the ui. What does it represent?

	return p
}

func RoomLeave(roomSlot byte, leaveCode byte) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x0A)
	p.WriteByte(roomSlot)
	p.WriteByte(leaveCode)

	return p
}

func RoomChat(sender, message string, roomSlot byte) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x06)
	p.WriteByte(8)        // msg type
	p.WriteByte(roomSlot) //
	p.WriteString(sender + " : " + message)

	return p
}

func RoomYellowChat(msgType byte, name string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x06)
	p.WriteByte(7)
	p.WriteByte(msgType) // expelled: 0, x's turn: 1, forfeit: 2, handicap request: 3, left: 4
	p.WriteString(name)

	return p
}

func RoomShowAccept() maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x0F)

	return p
}

func RoomInvite(roomType byte, name string, roomID int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x02)
	p.WriteByte(roomType)
	p.WriteString(name)
	p.WriteInt32(roomID)

	return p
}

func RoomInviteResult(resultCode byte, name string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x03)
	p.WriteByte(resultCode)
	p.WriteString(name)

	return p
}

func RoomReady() maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x32)

	return p
}

func RoomUnReady() maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x33)

	return p
}

func RoomRequestTie() maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x2a)

	return p
}

func RoomRejectTie() maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x2b)

	return p
}

func RoomRequestUndo() maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x2e)

	return p
}

func RoomRejectUndo() maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x2e)

	return p
}

func RoomOmokStart(p1Turn bool) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x35)
	p.WriteBool(p1Turn)

	return p
}

func RoomPlaceOmokPiece(x, y int32, piece byte) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x38)
	p.WriteInt32(x)
	p.WriteInt32(y)
	p.WriteByte(piece)

	return p
}

func RoomInvalidPlaceMsg() maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x39)

	return p
}

func RoomShowMapBox(charID, roomID int32, roomType, boardType byte, name string, hasPassword, koreanText bool) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM_BOX)
	p.WriteInt32(charID)
	p.WriteByte(roomType)
	p.WriteInt32(roomID)
	p.WriteString(name)
	p.WriteBool(hasPassword)
	p.WriteByte(boardType)
	// win loss record since room opened?
	p.WriteByte(1)
	p.WriteByte(2)
	p.WriteBool(koreanText)

	return p
}

func RoomRemoveBox(charID int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM_BOX)
	p.WriteInt32(charID)
	p.WriteInt32(0)

	return p
}

func roomEnterErrorMsg(errorCode byte) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x05)
	p.WriteByte(0x00)
	p.WriteByte(errorCode)

	return p
}

func RoomClosed() maplepacket.Packet {
	return roomEnterErrorMsg(0x01)
}

func RoomFull() maplepacket.Packet {
	return roomEnterErrorMsg(0x02)
}

func RoomBusy() maplepacket.Packet {
	return roomEnterErrorMsg(0x03)
}

func RoomNotAllowedWhenDead() maplepacket.Packet {
	return roomEnterErrorMsg(0x04)
}

func RoomNotAllowedDuringEvent() maplepacket.Packet {
	return roomEnterErrorMsg(0x05)
}

func RoomThisCharacterNotAllowed() maplepacket.Packet {
	return roomEnterErrorMsg(0x06)
}

func RoomNoTradeAtm() maplepacket.Packet {
	return roomEnterErrorMsg(0x07)
}

func RoomMiniRoomNotHere() maplepacket.Packet {
	return roomEnterErrorMsg(0x08)
}

func RoomTradeRequireSameMap() maplepacket.Packet {
	return roomEnterErrorMsg(0x09)
}

func RoomcannotCreateMiniroomHere() maplepacket.Packet {
	return roomEnterErrorMsg(0x0a)
}

func RoomCannotStartGameHere() maplepacket.Packet {
	return roomEnterErrorMsg(0x0b)
}

func RoomPersonalStoreFMOnly() maplepacket.Packet {
	return roomEnterErrorMsg(0x0c)
}
func RoomGarbageMsgAboutFloorInFm() maplepacket.Packet {
	return roomEnterErrorMsg(0x0d)
}

func RoomMayNotEnterStore() maplepacket.Packet {
	return roomEnterErrorMsg(0x0e)
}

func RoomStoreMaintenance() maplepacket.Packet {
	return roomEnterErrorMsg(0x0F)
}

func RoomCannotEnterTournament() maplepacket.Packet {
	return roomEnterErrorMsg(0x10)
}

func RoomGarbageTradeMsg() maplepacket.Packet {
	return roomEnterErrorMsg(0x11)
}

func RoomNotEnoughMesos() maplepacket.Packet {
	return roomEnterErrorMsg(0x12)
}

func RoomIncorrectPassword() maplepacket.Packet {
	return roomEnterErrorMsg(0x13)
}
