package packets

import (
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func MessageRedText(msg string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(9)
	p.WriteString(msg)

	return p
}

func MessageGuildPointsChange(ammount int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(6)
	p.WriteInt32(ammount)

	return p
}

func MessageFameChange(ammount int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(4)
	p.WriteInt32(ammount)

	return p
}

// sends the [item name] has passed its expeiration date and will be removed from your inventory
func MessageItemExpired(itemID int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(2)
	p.WriteInt32(itemID)
	return p
}

func MessageItemExpired2(itemID int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(8)
	p.WriteByte(1)
	p.WriteInt32(itemID)
	return p
}

func MessageMesosChangeChat(ammount int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(5)
	p.WriteInt32(ammount)

	return p
}

func MessageUnableToPickUp(itemNotAvailable bool) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(0)
	if itemNotAvailable {
		p.WriteByte(0xFE)
	} else {
		p.WriteByte(0xFF)
	}

	return p
}

func MessageDropPickUp(isMesos bool, itemID, ammount int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(0)

	if isMesos {
		p.WriteInt32(ammount)
		p.WriteInt32(0)
	} else {
		p.WriteInt32(itemID)
		p.WriteInt32(ammount)
	}

	return p
}

func MessageExpGained(whiteText, appearInChat bool, ammount int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_INFO_MESSAGE)
	p.WriteByte(3)
	p.WriteBool(whiteText)
	p.WriteInt32(ammount)
	p.WriteBool(appearInChat)

	return p
}

func MessageNotice(msg string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_BROADCAST_MESSAGE)
	p.WriteByte(0)
	p.WriteString(msg)

	return p
}

func MessageDialogueBox(msg string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_BROADCAST_MESSAGE)
	p.WriteByte(1)
	p.WriteString(msg)

	return p
}

func MessageWhiteBar(msg string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_BROADCAST_MESSAGE)
	p.WriteByte(2)
	p.WriteString(msg) // not sure how string is formated

	return p
}

// Need to figure out how to display the username and  atm it bastardises it.
func MessageBroadcastChannel(senderName string, msg string, channel byte, ear bool) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_BROADCAST_MESSAGE)
	p.WriteByte(3)
	p.WriteString(senderName)
	p.WriteByte(channel)
	if ear {
		p.WriteByte(0x01)
	} else {
		p.WriteByte(0x00)
	}

	return p
}

func MessageScrollingHeader(msg string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_BROADCAST_MESSAGE)
	p.WriteByte(4)
	p.WriteBool(bool(len(msg) > 0))
	p.WriteString(msg)

	return p
}

func MessageBubblessChat(msgType byte, sender string, msg string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_BUBBLESS_CHAT)
	p.WriteByte(msgType) // 0x00 buddy chat, 0x01 - party, 0x02 - guild
	p.WriteString(sender)
	p.WriteString(msg)

	return p
}

func MessageWhisper(sender string, message string, channel byte) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_WHISPER)
	p.WriteByte(0x12)
	p.WriteString(sender)
	p.WriteByte(channel)
	p.WriteString(message)

	return p
}

func MessageFindResult(character string, isAdmin, inCashShop, sameChannel bool, mapID int32) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_WHISPER)

	if isAdmin {
		p.WriteByte(0x05)
		p.WriteString("User not found")

	} else if mapID > 0 {
		p.WriteByte(0x9)
		p.WriteString(character)

		if inCashShop {
			p.WriteByte(0x02)
			p.WriteInt32(0) // ?
		} else if sameChannel {
			p.WriteByte(0x01)
			p.WriteInt32(mapID)
			p.WriteInt32(0) // ?
		} else {
			p.WriteByte(0x01)
			p.WriteInt32(mapID)
		}

		p.WriteInt32(0) // ?
	} else {
		p.WriteByte(0x0A)
		p.WriteString(character)
		p.WriteByte(0) // ?
	}

	return p
}

func MessageAllChat(senderID int32, isAdmin bool, msg string) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_ALL_CHAT_MSG)
	p.WriteInt32(senderID)
	p.WriteBool(isAdmin)
	p.WriteString(msg)

	return p
}

// Implement logic for these
func MessageGmBan(good bool) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_EMPLOYEE)
	if good {
		p.WriteByte(4)
		p.WriteByte(0)
	} else {
		p.WriteByte(6)
		p.WriteByte(1)
	}

	return p
}

func MessageGmRemoveFromRanks() maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_EMPLOYEE)
	p.WriteByte(6)
	p.WriteByte(0)

	return p
}

func MessageGmWarning(good bool) maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_EMPLOYEE)
	p.WriteByte(14)
	if good {
		p.WriteByte(1)
	} else {
		p.WriteByte(0)
	}

	return p
}

func MessageGmBlockedAccess() maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_EMPLOYEE)
	p.WriteByte(4)
	p.WriteByte(0)

	return p
}

func MessageGmUnblock() maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_EMPLOYEE)
	p.WriteByte(5)
	p.WriteByte(0)

	return p
}

// Don't know what this is used for
func MessageGmWrongNpc() maplepacket.Packet {
	p := maplepacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_EMPLOYEE)
	p.WriteByte(8)
	p.WriteInt16(0)

	return p
}
