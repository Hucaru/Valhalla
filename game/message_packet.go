package game

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

func packetMessageRedText(msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(9)
	p.WriteString(msg)

	return p
}

func packetMessageGuildPointsChange(ammount int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(6)
	p.WriteInt32(ammount)

	return p
}

func packetMessageFameChange(ammount int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(4)
	p.WriteInt32(ammount)

	return p
}

// sends the [item name] has passed its expeiration date and will be removed from your inventory
func packetMessageItemExpired(itemID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(2)
	p.WriteInt32(itemID)
	return p
}

func packetMessageItemExpired2(itemID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(8)
	p.WriteByte(1)
	p.WriteInt32(itemID)
	return p
}

func packetMessageMesosChangeChat(ammount int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(5)
	p.WriteInt32(ammount)

	return p
}

func packetMessageUnableToPickUp(itemNotAvailable bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(0)
	if itemNotAvailable {
		p.WriteByte(0xFE)
	} else {
		p.WriteByte(0xFF)
	}

	return p
}

func packetMessageDropPickUp(isMesos bool, itemID, ammount int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
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

func packetMessageExpGained(whiteText, appearInChat bool, ammount int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(3)
	p.WriteBool(whiteText)
	p.WriteInt32(ammount)
	p.WriteBool(appearInChat)

	return p
}

func packetMessageNotice(msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBroadcastMessage)
	p.WriteByte(0)
	p.WriteString(msg)

	return p
}

func packetMessageDialogueBox(msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBroadcastMessage)
	p.WriteByte(1)
	p.WriteString(msg)

	return p
}

func packetMessageWhiteBar(msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBroadcastMessage)
	p.WriteByte(2)
	p.WriteString(msg) // not sure how string is formated

	return p
}

// Need to figure out how to display the username and  atm it bastardises it.
func packetMessageBroadcastChannel(senderName string, msg string, channel byte, ear bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBroadcastMessage)
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

func packetMessageScrollingHeader(msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBroadcastMessage)
	p.WriteByte(4)
	p.WriteBool(bool(len(msg) > 0))
	p.WriteString(msg)

	return p
}

func packetMessageBubblessChat(msgType byte, sender string, msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBubblessChat)
	p.WriteByte(msgType) // 0x00 buddy chat, 0x01 - party, 0x02 - guild
	p.WriteString(sender)
	p.WriteString(msg)

	return p
}

func packetMessageWhisper(sender string, message string, channel byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelWhisper)
	p.WriteByte(0x12)
	p.WriteString(sender)
	p.WriteByte(channel)
	p.WriteString(message)

	return p
}

func packetMessageFindResult(character string, isAdmin, inCashShop, sameChannel bool, mapID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelWhisper)

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

func packetMessageAllChat(senderID int32, isAdmin bool, msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelAllChatMsg)
	p.WriteInt32(senderID)
	p.WriteBool(isAdmin)
	p.WriteString(msg)

	return p
}

// Implement logic for these
func packetMessageGmBan(good bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelEmployee)
	if good {
		p.WriteByte(4)
		p.WriteByte(0)
	} else {
		p.WriteByte(6)
		p.WriteByte(1)
	}

	return p
}

func packetMessageGmRemoveFromRanks() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelEmployee)
	p.WriteByte(6)
	p.WriteByte(0)

	return p
}

func packetMessageGmWarning(good bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelEmployee)
	p.WriteByte(14)
	if good {
		p.WriteByte(1)
	} else {
		p.WriteByte(0)
	}

	return p
}

func packetMessageGmBlockedAccess() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelEmployee)
	p.WriteByte(4)
	p.WriteByte(0)

	return p
}

func packetMessageGmUnblock() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelEmployee)
	p.WriteByte(5)
	p.WriteByte(0)

	return p
}

// Don't know what this is used for
func packetMessageGmWrongNpc() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelEmployee)
	p.WriteByte(8)
	p.WriteInt16(0)

	return p
}
