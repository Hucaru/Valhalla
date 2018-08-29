package packets

import (
	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/maplepacket"
)

var room_omok_game byte = 0x01
var room_memory_game byte = 0x02
var room_trade byte = 0x03
var room_personal_shop byte = 0x04
var room_other_shop byte = 0x05

func roomWindow(roomType, maxUsers, roomSlot byte, chars []character.Character) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x05)
	p.WriteByte(roomType)
	p.WriteByte(maxUsers)
	p.WriteByte(roomSlot)

	for i, c := range chars {
		p.WriteByte(byte(i))
		p.Append(writeDisplayCharacter(c))
		p.WriteInt32(0) // not sure what this is, room id?
		p.WriteString(c.GetName())
	}

	p.WriteByte(0xFF)

	return p
}

func RoomJoin(roomSlot byte, char character.Character) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x04)
	p.WriteByte(roomSlot)
	p.Append(writeDisplayCharacter(char))
	p.WriteString(char.GetName())

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

func RoomShowTradeWindow(roomSlot byte, chars []character.Character) maplepacket.Packet {
	return roomWindow(room_trade, 2, roomSlot, chars)
}

func RoomInviteResult(resultCode byte, name string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ROOM)
	p.WriteByte(0x03)
	p.WriteByte(resultCode)
	p.WriteString(name)

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
