package message

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

// PacketMessageRedText - sends red error message to client chat window
func PacketMessageRedText(msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(9)
	p.WriteString(msg)

	return p
}

// PacketMessageGuildPointsChange - sends guild point change message to client chat window
func PacketMessageGuildPointsChange(ammount int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(6)
	p.WriteInt32(ammount)

	return p
}

// PacketMessageFameChange - sends fame amount to client chat window
func PacketMessageFameChange(ammount int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(4)
	p.WriteInt32(ammount)

	return p
}

// PacketMessageItemExpired - sends the [item name] has passed its expeiration date and will be removed from your inventory
func PacketMessageItemExpired(itemID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(2)
	p.WriteInt32(itemID)
	return p
}

// PacketMessageItemExpired2 - alternate msg to above
func PacketMessageItemExpired2(itemID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(8)
	p.WriteByte(1)
	p.WriteInt32(itemID)
	return p
}

// PacketMessageMesosChangeChat - mesos amount change client chat window message
func PacketMessageMesosChangeChat(ammount int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(5)
	p.WriteInt32(ammount)

	return p
}

// PacketMessageUnableToPickUp - unable to pick up message to client chat window
func PacketMessageUnableToPickUp(itemNotAvailable bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(0)
	if itemNotAvailable {
		p.WriteByte(0xFE)
	} else {
		p.WriteByte(0xFF)
	}

	return p
}

// PacketMessageDropPickUp - pick up drop client chat window message
func PacketMessageDropPickUp(isMesos bool, itemID, ammount int32) mpacket.Packet {
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

// PacketMessageExpGained - exp gained client chat or side message
func PacketMessageExpGained(whiteText, appearInChat bool, ammount int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelInfoMessage)
	p.WriteByte(3)
	p.WriteBool(whiteText)
	p.WriteInt32(ammount)
	p.WriteBool(appearInChat)

	return p
}

// PacketMessageNotice - blue notice client chat message
func PacketMessageNotice(msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBroadcastMessage)
	p.WriteByte(0)
	p.WriteString(msg)

	return p
}

// PacketMessageDialogueBox - pop up dialogue box
func PacketMessageDialogueBox(msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBroadcastMessage)
	p.WriteByte(1)
	p.WriteString(msg)

	return p
}

// PacketCannotChangeChannel - red text
func PacketCannotChangeChannel() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelChangeServer)
	p.WriteByte(1)

	return p
}

// PacketMessageWhiteBar - white bar message, is this gm chat messages?
func PacketMessageWhiteBar(msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBroadcastMessage)
	p.WriteByte(2)
	p.WriteString(msg) // not sure how string is formated

	return p
}

//PacketMessageBroadcastChannel - Need to figure out how to display the username and  atm it bastardises it.
func PacketMessageBroadcastChannel(senderName string, msg string, channel byte, ear bool) mpacket.Packet {
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

// PacketMessageScrollingHeader - scroll message a the top
func PacketMessageScrollingHeader(msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBroadcastMessage)
	p.WriteByte(4)
	p.WriteBool(bool(len(msg) > 0))
	p.WriteString(msg)

	return p
}

// PacketMessageBubblessChat - user chat that has no bubble
func PacketMessageBubblessChat(msgType byte, sender string, msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBubblessChat)
	p.WriteByte(msgType) // 0x00 buddy chat, 0x01 - party, 0x02 - guild
	p.WriteString(sender)
	p.WriteString(msg)

	return p
}

// PacketMessageWhisper - whispher msg in client chat window
func PacketMessageWhisper(sender string, message string, channel byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelWhisper)
	p.WriteByte(0x12)
	p.WriteString(sender)
	p.WriteByte(channel)
	p.WriteString(message)

	return p
}

// PacketMessageFindResult - send the result of using the /find comand
func PacketMessageFindResult(character string, isAdmin, inCashShop, sameChannel bool, mapID int32) mpacket.Packet {
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

// PacketMessageAllChat - sends general chat message to client
func PacketMessageAllChat(senderID int32, isAdmin bool, msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelAllChatMsg)
	p.WriteInt32(senderID)
	p.WriteBool(isAdmin)
	p.WriteString(msg)

	return p
}

// PacketMessageGmBan - "You have entered an invalid character name"
func PacketMessageGmBan(good bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelEmployee)
	p.WriteByte(6)
	p.WriteByte(1)

	return p
}

// PacketMessageGmRemoveFromRanks -"You have successfully removed the name from the ranks"
func PacketMessageGmRemoveFromRanks() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelEmployee)
	p.WriteByte(6)
	p.WriteByte(0)

	return p
}

// PacketMessageGmWarning -
func PacketMessageGmWarning(good bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelEmployee)
	p.WriteByte(14)
	if good {
		p.WriteByte(1)
	} else {
		p.WriteByte(0)
	}

	return p
}

// PacketMessageGmBlockedAccess - "You have successfully blocked access"
func PacketMessageGmBlockedAccess() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelEmployee)
	p.WriteByte(4)
	p.WriteByte(0)

	return p
}

// PacketMessageGmUnblock - "The unblocking has been successful"
func PacketMessageGmUnblock() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelEmployee)
	p.WriteByte(5)
	p.WriteByte(0)

	return p
}

// PacketMessageGmWrongNpc -
func PacketMessageGmWrongNpc() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelEmployee)
	p.WriteByte(8)
	p.WriteInt16(0)

	return p
}

// PacketShowCountdown - Displays a countdown on top of screen with given time in seconds
func PacketShowCountdown(time int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCountdown)
	p.WriteByte(2)
	p.WriteInt32(time)

	return p
}

// PacketHideCountdown - hides the countdown from the player
func PacketHideCountdown() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCountdown)
	p.WriteByte(0)
	p.WriteInt32(0)

	return p
}
