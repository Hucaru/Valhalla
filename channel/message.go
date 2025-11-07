package channel

import (
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mpacket"
)

/*
	This file contains message packets that don't rigidly belong with a data type
	e.g. party join packet is closely coupled with party struct
*/

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

func packetCannotChangeChannel() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelChangeServer)
	p.WriteByte(1)

	return p
}

// PacketMessageWhiteBar - white bar message, is this gm chat messages?
func packetMessageWhiteBar(msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBroadcastMessage)
	p.WriteByte(2)
	p.WriteString(msg) // not sure how string is formated

	return p
}

// PacketMessageBroadcastChannel - Need to figure out how to display the username and  atm it bastardises it.
func packetMessageBroadcastChannel(senderName string, msg string, channel byte, ear bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBroadcastMessage)
	p.WriteByte(constant.BroadcastMegaphone)
	p.WriteString(senderName + " : " + msg)
	p.WriteByte(channel)
	if ear {
		p.WriteByte(0x01)
	} else {
		p.WriteByte(0x00)
	}

	return p
}

func packetMessageBroadcastSuper(senderName string, msg string, channel byte, ear bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBroadcastMessage)
	p.WriteByte(constant.BroadcastSuperMegaphone)
	p.WriteString(senderName + " : " + msg)
	p.WriteByte(channel)
	if ear {
		p.WriteByte(0x01)
	} else {
		p.WriteByte(0x00)
	}

	return p
}

// PacketMessageScrollingHeader - scroll message a the top
func packetMessageScrollingHeader(msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBroadcastMessage)
	p.WriteByte(4)
	p.WriteBool(bool(len(msg) > 0))
	p.WriteString(msg)

	return p
}

// PacketMessageBubblessChat - user chat that has no bubble
func packetMessageBubblessChat(msgType byte, sender string, msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBubblessChat)
	p.WriteByte(msgType) // 0x00 buddy chat, 0x01 - party, 0x02 - guild
	p.WriteString(sender)
	p.WriteString(msg)

	return p
}

// PacketMessageWhisper - whispher msg in client chat window
func packetMessageWhisper(sender string, message string, channel byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelWhisper)
	p.WriteByte(0x12)
	p.WriteString(sender)
	p.WriteByte(channel)
	p.WriteString(message)

	return p
}

// PacketMessageFindResult - Send the result of using the /find comand
func packetMessageFindResult(character string, is, inCashShop, sameChannel bool, mapID int32) mpacket.Packet {
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

func packetMessageAllChat(senderID int32, isAdmin bool, msg string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelAllChatMsg)
	p.WriteInt32(senderID)
	p.WriteBool(isAdmin)
	p.WriteString(msg)

	return p
}

func packetMessageGmBan(good bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelEmployee)
	p.WriteByte(6)
	p.WriteByte(1)

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

func packetMessageGmWrongNpc() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelEmployee)
	p.WriteByte(8)
	p.WriteInt16(0)

	return p
}

func packetUseScroll(playerID int32, succeed bool, destroy bool, legendarySpirit bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelUseScroll)
	p.WriteInt32(playerID)
	p.WriteBool(succeed)
	p.WriteBool(destroy)

	var ls int16 = 0
	if legendarySpirit {
		ls = 1
	}
	p.WriteInt16(ls)

	return p
}

func packetMessengerSelfEnter(slot byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMessenger)
	p.WriteByte(constant.MessengerEnterResult)
	p.WriteByte(slot)
	return p
}

func packetMessengerLeave(slot byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMessenger)
	p.WriteByte(constant.MessengerLeave)
	p.WriteByte(slot)
	return p
}

func packetMessengerInvite(sender string, messengerID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMessenger)
	p.WriteByte(constant.MessengerInvite)
	p.WriteString(sender)
	p.WriteByte(0x00)
	p.WriteInt32(messengerID)
	p.WriteByte(0x00)
	return p
}

func packetMessengerInviteResult(recipient string, success bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMessenger)
	p.WriteByte(constant.MessengerInviteResult)
	p.WriteString(recipient)
	p.WriteBool(success)
	return p
}

func packetMessengerBlocked(receiver string, mode byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMessenger)
	p.WriteByte(constant.MessengerBlocked)
	p.WriteString(receiver)
	p.WriteByte(mode)
	return p
}

func packetMessengerChat(message string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMessenger)
	p.WriteByte(constant.MessengerChat)
	p.WriteString(message)
	return p
}

func packetMessengerEnter(slot, gender, skin, ch byte, face, hair, cashW, petAcc int32, name string, announce bool, vis, hid []internal.KV) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMessenger)
	p.WriteByte(constant.MessengerEnter)
	p.WriteByte(slot)
	p.WriteByte(gender)
	p.WriteByte(skin)
	p.WriteInt32(face)
	p.WriteBool(true)
	p.WriteInt32(hair)
	for _, kv := range vis {
		kv.Serialise(&p)
	}
	p.WriteInt8(-1)
	for _, kv := range hid {
		kv.Serialise(&p)
	}
	p.WriteInt8(-1)
	p.WriteInt32(cashW)
	p.WriteInt32(petAcc)
	p.WriteString(name)
	p.WriteByte(ch)
	p.WriteBool(announce)
	return p
}

func packetMessengerAvatar(slot, gender, skin byte, face, hair, cashW, petAcc int32, vis, hid []internal.KV) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMessenger)
	p.WriteByte(constant.MessengerAvatar)
	p.WriteByte(slot)
	p.WriteByte(gender)
	p.WriteByte(skin)
	p.WriteInt32(face)
	p.WriteBool(true)
	p.WriteInt32(hair)
	for _, kv := range vis {
		kv.Serialise(&p)
	}
	p.WriteInt8(-1)
	for _, kv := range hid {
		kv.Serialise(&p)
	}
	p.WriteInt8(-1)
	p.WriteInt32(cashW)
	p.WriteInt32(petAcc)
	return p
}
