package message

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/pos"
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
func PacketMessageFindResult(character string, is, inCashShop, sameChannel bool, mapID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelWhisper)

	if mapID >= 0 {
		p.WriteByte(0x9)
		p.WriteString(character)

		if inCashShop {
			p.WriteByte(0x02)
			p.WriteInt32(0) // ?
			p.WriteInt32(0) // ?
		} else if sameChannel {
			p.WriteByte(0x01)
			p.WriteInt32(mapID)
			p.WriteInt32(0) // ?
			p.WriteInt32(0) // ?
		} else {
			p.WriteByte(0x03)
			p.WriteInt32(mapID)
			p.WriteInt32(0) // ?
			p.WriteInt32(0) // ?
		}
	} else {
		p.WriteByte(0x0A)
		p.WriteString(character)
		p.WriteBool(is)
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

// PacketBuddyUnkownError - error diolog message box
func PacketBuddyUnkownError() mpacket.Packet {
	return packetBuddyRequestResult(0x16)
}

// PacketBuddyPlayerFullList - shows full buddy list dialog box
func PacketBuddyPlayerFullList() mpacket.Packet {
	return packetBuddyRequestResult(0x0b)
}

// PacketBuddyOtherFullList - other player has full buddy list dialog box
func PacketBuddyOtherFullList() mpacket.Packet {
	return packetBuddyRequestResult(0x0c)
}

// PacketBuddyAlreadyAdded - already added buddy dialog box
func PacketBuddyAlreadyAdded() mpacket.Packet {
	return packetBuddyRequestResult(0x0d)
}

// PacketBuddyIsGM - cannot add gm to buddy list dialog box
func PacketBuddyIsGM() mpacket.Packet {
	return packetBuddyRequestResult(0x0e)
}

// PacketBuddyNameNotRegistered - name not regsitered dialog box
func PacketBuddyNameNotRegistered() mpacket.Packet {
	return packetBuddyRequestResult(0x0f)
}

func packetBuddyRequestResult(code byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(code)

	return p
}

// PacketBuddyReceiveRequest - buddy request notice card
func PacketBuddyReceiveRequest(fromID int32, fromName string, fromChannelID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x9)
	p.WriteInt32(fromID)
	p.WriteString(fromName)
	p.WriteInt32(fromID)
	p.WritePaddedString(fromName, 13)
	p.WriteByte(1)
	p.WriteInt32(fromChannelID)
	p.WriteBool(false) // sender in cash shop

	return p
}

// PacketBuddyOnlineStatus - buddy online status notice card
func PacketBuddyOnlineStatus(id int32, channelID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x14)
	p.WriteInt32(id)
	p.WriteInt8(0)
	p.WriteInt32(channelID)

	return p
}

// PacketBuddyChangeChannel - buddy ui change channel change
func PacketBuddyChangeChannel(id int32, channelID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBuddyInfo)
	p.WriteByte(0x14)
	p.WriteInt32(id)
	p.WriteInt8(1)
	p.WriteInt32(channelID)

	return p
}

// PacketPartyCreateUnkownError - sends the unkown error in creating party message
func PacketPartyCreateUnkownError() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPartyInfo)
	p.WriteByte(0)

	return p
}

// PacketPartyInviteNotice - shows the party invite notice
func PacketPartyInviteNotice(partyID int32, fromName string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPartyInfo)
	p.WriteByte(0x04)
	p.WriteInt32(partyID)
	p.WriteString(fromName)

	return p
}

func packetPartyMessage(op byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPartyInfo)
	p.WriteByte(op)

	return p
}

// PacketPartyAlreadyJoined - sends the player has already in a party message
func PacketPartyAlreadyJoined() mpacket.Packet {
	return packetPartyMessage(0x08)
}

// PacketPartyBeginnerCannotCreate - sends a beginner cannot create a party message
func PacketPartyBeginnerCannotCreate() mpacket.Packet {
	return packetPartyMessage(0x09)
}

// PacketPartyNotInParty - sends you have yet to join a party message
func PacketPartyNotInParty() mpacket.Packet {
	return packetPartyMessage(0x0c)
}

// PacketPartyAlreadyJoined2 - sends the player has already in a party message
func PacketPartyAlreadyJoined2() mpacket.Packet {
	return packetPartyMessage(0x0f)
}

// PacketPartyToJoinIsFull - sends the party the player is trying to join is full message
func PacketPartyToJoinIsFull() mpacket.Packet {
	return packetPartyMessage(0x10)
}

// PacketPartyUnableToFindPlayer - sends the unable to find player message
func PacketPartyUnableToFindPlayer() mpacket.Packet {
	return packetPartyMessage(0x11)
}

// PacketPartyAdminNoCreate - sends the gm cannot create party
func PacketPartyAdminNoCreate() mpacket.Packet {
	return packetPartyMessage(0x18)
}

// PacketPartyUnableToFindPlayer2 - sends the unable to find player message
func PacketPartyUnableToFindPlayer2() mpacket.Packet {
	return packetPartyMessage(0x19)
}

func packetPartyMessageName(op byte, name string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPartyInfo)
	p.WriteByte(op)
	p.WriteString(name)

	return p
}

// PacketPartyBlockingInvites - sends the player is blocking party invites message
func PacketPartyBlockingInvites(name string) mpacket.Packet {
	return packetPartyMessageName(0x13, name)
}

// PacketPartyHasOtherRequest - sends the player is taking care of another request
func PacketPartyHasOtherRequest(name string) mpacket.Packet {
	return packetPartyMessageName(0x14, name)
}

// PacketPartyRequestDenied - sends the player has denied the party request message
func PacketPartyRequestDenied(name string) mpacket.Packet {
	return packetPartyMessageName(0x15, name)
}

// PacketPartyCreate - created party message
func PacketPartyCreate(partyID int32, doorMap1, doorMap2 int32, point pos.Data) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPartyInfo)
	p.WriteByte(0x07)
	p.WriteInt32(partyID)

	if doorMap1 > -1 {
		p.WriteInt32(doorMap1)
		p.WriteInt32(doorMap2)
		p.WriteInt16(point.X())
		p.WriteInt16(point.Y())
	} else {
		p.WriteInt32(-1)
		p.WriteInt32(-1)
		p.WriteInt32(0) // empty pos
	}

	return p
}

/*
0x1b:
i32
i32
i32

0x1c:
i8
i32
i32
i16
i16

*/
