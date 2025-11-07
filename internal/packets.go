package internal

import (
	"log"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mpacket"
)

func PacketChannelPopUpdate(id byte, pop int16) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelInfo)
	p.WriteByte(id)
	p.WriteByte(0) // 0 is population
	p.WriteInt16(pop)

	return p
}

func PacketChannelPlayerConnected(playerID int32, name string, channelID byte, channelChange bool, mapID, guildID int32) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerConnect)
	p.WriteInt32(playerID)
	p.WriteString(name)
	p.WriteByte(channelID)
	p.WriteBool(channelChange)
	p.WriteInt32(mapID)
	p.WriteInt32(guildID)

	return p
}

func PacketChannelPlayerDisconnect(id int32, name string, guildID int32) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannePlayerDisconnect)
	p.WriteInt32(id)
	p.WriteString(name)
	p.WriteInt32(guildID)

	return p
}

func PacketChannelBuddyEvent(op byte, recepientID, fromID int32, fromName string, channelID byte) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerBuddyEvent)
	p.WriteByte(op)

	switch op {
	case 1: // add
		fallthrough
	case 2: // accept
		p.WriteInt32(recepientID)
		p.WriteInt32(fromID)
		p.WriteString(fromName)
		p.WriteByte(channelID)
	case 3: // delete / reject
		p.WriteInt32(recepientID)
		p.WriteInt32(fromID)
		p.WriteByte(channelID)
	default:
		log.Println("unkown internal buddy event type:", op)
	}

	return p
}

func PacketChannelWhispherChat(recepientName, fromName, msg string, channelID byte) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerChatEvent)
	p.WriteByte(OpChatWhispher) // whispher
	p.WriteString(recepientName)
	p.WriteString(fromName)
	p.WriteString(msg)
	p.WriteByte(channelID)

	return p
}

func PacketChannelPlayerChat(code byte, fromName string, buffer []byte) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerChatEvent)
	p.WriteByte(code) // 1 buddy, 2 party, 3 guild
	p.WriteString(fromName)
	p.WriteBytes(buffer)

	return p
}

func PacketChannelPartyCreateRequest(playerID int32, channelID byte, mapID, job, level int32, name string) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerPartyEvent)
	p.WriteByte(OpPartyCreate)
	p.WriteInt32(playerID)
	p.WriteByte(channelID)
	p.WriteInt32(mapID)
	p.WriteInt32(job)
	p.WriteInt32(level)
	p.WriteString(name)

	return p
}

func PacketWorldPartyCreateApproved(playerID int32, success bool, party *Party) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerPartyEvent)
	p.WriteByte(OpPartyCreate)
	p.WriteInt32(playerID)
	p.WriteBool(success)

	if success {
		p.WriteBytes(party.GeneratePacket())
	}

	return p
}

func PacketChannelPartyLeave(partyID, playerID int32, kicked bool) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerPartyEvent)
	p.WriteByte(OpPartyLeaveExpel)
	p.WriteInt32(partyID)
	p.WriteInt32(playerID)
	p.WriteBool(kicked)

	return p
}

func PacketWorldPartyLeave(partyID int32, destroy, kicked bool, index int32, party *Party) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerPartyEvent)
	p.WriteByte(OpPartyLeaveExpel)
	p.WriteInt32(partyID)
	p.WriteBool(destroy)
	p.WriteBool(kicked)
	p.WriteInt32(index)
	p.WriteBytes(party.GeneratePacket())

	return p
}

func PacketChannelPartyAccept(partyID, playerID, channelID, mapID, job, level int32, name string) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerPartyEvent)
	p.WriteByte(OpPartyAccept)
	p.WriteInt32(partyID)
	p.WriteInt32(playerID)
	p.WriteInt32(channelID)
	p.WriteInt32(mapID)
	p.WriteInt32(job)
	p.WriteInt32(level)
	p.WriteString(name)

	return p
}

func PacketWorldPartyAccept(partyID, playerID, index int32, party *Party) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerPartyEvent)
	p.WriteByte(OpPartyAccept)
	p.WriteInt32(partyID)
	p.WriteInt32(playerID)
	p.WriteInt32(index)
	p.WriteBytes(party.GeneratePacket())

	return p
}

func PacketChannelPartyUpdateInfo(partyID, playerID, job, level, mapID int32, name string) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerPartyEvent)
	p.WriteByte(OpPartyInfoUpdate)
	p.WriteInt32(partyID)
	p.WriteInt32(playerID)
	p.WriteInt32(job)
	p.WriteInt32(level)
	p.WriteInt32(mapID)
	p.WriteString(name)

	return p
}

func PacketWorldPartyUpdate(partyID, playerID, index int32, onlineStatus bool, party *Party) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerPartyEvent)
	p.WriteByte(OpPartyInfoUpdate)
	p.WriteInt32(partyID)
	p.WriteInt32(playerID)
	p.WriteInt32(index)
	p.WriteBool(onlineStatus)
	p.WriteBytes(party.GeneratePacket())

	return p
}

func PacketGuildDisband(guildID int32) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerGuildEvent)
	p.WriteByte(OpGuildDisband)
	p.WriteInt32(guildID)

	return p
}

func PacketGuildRemovePlayer(guildID, playerID int32, playerName string, expelled bool) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerGuildEvent)
	p.WriteByte(OpGuildRemovePlayer)
	p.WriteInt32(guildID)
	p.WriteInt32(playerID)
	p.WriteBool(expelled) // 0 left, 1 expelled
	p.WriteString(playerName)

	return p
}

func PacketGuildUpdateEmblem(guildID int32, logoBg, logo int16, logoBgColour, logoColour byte) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerGuildEvent)
	p.WriteByte(OpGuildEmblemChange)
	p.WriteInt32(guildID)
	p.WriteInt16(logoBg)
	p.WriteInt16(logo)
	p.WriteByte(logoBgColour)
	p.WriteByte(logoColour)

	return p
}

func PacketGuildUpdateNotice(guildID int32, notice string) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerGuildEvent)
	p.WriteByte(OpGuildNoticeChange)
	p.WriteInt32(guildID)
	p.WriteString(notice)

	return p
}

func PacketGuildTitlesChange(guildID int32, master, jrMaster, member1, member2, member3 string) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerGuildEvent)
	p.WriteByte(OpGuildTitlesChange)
	p.WriteInt32(guildID)
	p.WriteString(master)
	p.WriteString(jrMaster)
	p.WriteString(member1)
	p.WriteString(member2)
	p.WriteString(member3)

	return p
}

func PacketGuildPointsUpdate(guildID, points int32) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerGuildEvent)
	p.WriteByte(OpGuildPointsUpdate)
	p.WriteInt32(guildID)
	p.WriteInt32(points)

	return p
}

func PacketGuildRankUpdate(guildID, playerID int32, rank byte) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerGuildEvent)
	p.WriteByte(OpGuildRankUpdate)
	p.WriteInt32(guildID)
	p.WriteInt32(playerID)
	p.WriteByte(rank)

	return p
}

func PacketGuildInvite(guildID int32, inviter, invitee string) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerGuildEvent)
	p.WriteByte(OpGuildInvite)
	p.WriteInt32(guildID)
	p.WriteString(inviter)
	p.WriteString(invitee)

	return p
}

func PacketGuildInviteReject(inviter, invitee string) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerGuildEvent)
	p.WriteByte(OpGuildInviteReject)
	p.WriteString(inviter)
	p.WriteString(invitee)

	return p
}

func PacketGuildInviteAccept(playerID, guildID int32, name string, jobID, level int32, online bool, rank byte) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerGuildEvent)
	p.WriteByte(OpGuildInviteAccept)
	p.WriteInt32(playerID)
	p.WriteInt32(guildID)
	p.WriteString(name)
	p.WriteInt32(jobID)
	p.WriteInt32(level)
	p.WriteBool(online)
	p.WriteByte(rank)

	return p
}

func PacketLoginDeletedCharacter(playerID int32) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.LoginDeleteCharacter)
	p.WriteInt32(playerID)

	return p
}

func PacketRateOperation(mode byte, rate float32) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChangeRate)
	p.WriteByte(mode)
	p.WriteFloat32(rate)

	return p
}

func PacketChangeExpRate(rate float32) mpacket.Packet {
	return PacketRateOperation(1, rate)
}

func PacketChangeDropRate(rate float32) mpacket.Packet {
	return PacketRateOperation(2, rate)
}

func PacketChangeMesosRate(rate float32) mpacket.Packet {
	return PacketRateOperation(3, rate)
}

func PacketUpdateLoginInfo(ribbon byte, message string) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.UpdateLoginInfo)
	p.WriteByte(ribbon)
	p.WriteString(message)
	return p
}

func PacketSyncParties(parties map[int32]*Party) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.SyncParties)
	p.WriteInt32(int32(len(parties)))
	for _, party := range parties {
		p.WriteInt32(party.ID)
		for i := 0; i < constant.MaxPartySize; i++ {
			p.WriteInt32(party.PlayerID[i])
			p.WriteInt32(party.ChannelID[i])
			p.WriteInt32(party.MapID[i])
			p.WriteInt32(party.Job[i])
			p.WriteInt32(party.Level[i])
			p.WriteString(party.Name[i])
		}
	}
	return p
}

func PacketSyncGuilds(guildIDs []int32) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.SyncGuilds)
	p.WriteInt32(int32(len(guildIDs)))
	for _, guildID := range guildIDs {
		p.WriteInt32(guildID)
	}
	return p
}

func PacketWorldMessengerSelfEnter(recipientID int32, slot byte) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(0x00)
	p.WriteInt32(recipientID)
	p.WriteByte(slot)
	return p
}
func PacketWorldMessengerEnter(recipientID, face, hair, cashW, petAcc int32, slot, channelID, gender, skin byte, name string, vis, hid []KV) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(0x01)
	p.WriteInt32(recipientID)
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
	p.WriteByte(channelID)
	p.WriteBool(true)
	return p
}
func PacketWorldMessengerLeave(recipientID int32, slot byte) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(0x02)
	p.WriteInt32(recipientID)
	p.WriteByte(slot)
	return p
}
func PacketWorldMessengerInviteResult(senderID int32, recipient string, ok bool) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(0x04)
	p.WriteInt32(senderID)
	p.WriteString(recipient)
	p.WriteBool(ok)
	return p
}
func PacketWorldMessengerBlocked(senderID int32, receiver string, mode byte) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(0x05)
	p.WriteInt32(senderID)
	p.WriteString(receiver)
	p.WriteByte(mode)
	return p
}
func PacketWorldMessengerChat(recipientID int32, msg string) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(0x06)
	p.WriteInt32(recipientID)
	p.WriteString(msg)
	return p
}
func PacketWorldMessengerAvatar(recipientID, face, hair, cashW, petAcc int32, slot, gender, skin byte, vis, hid []KV) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(0x07)
	p.WriteInt32(recipientID)
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

func PacketMessengerEnter(charID, messengerID, face, hair, cashWeapon, petAccessory int32, channelID, gender, skin byte, name string, vis, hid []KV) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(constant.MessengerEnter)
	p.WriteInt32(charID)
	p.WriteByte(channelID)
	p.WriteString(name)
	p.WriteInt32(messengerID)
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

	p.WriteInt32(cashWeapon)
	p.WriteInt32(petAccessory)

	return p
}

func PacketMessengerLeave(charID int32) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(constant.MessengerLeave)
	p.WriteInt32(charID)
	return p
}

func PacketMessengerInvite(charID int32, channelID byte, name, invitee string) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(constant.MessengerInvite)
	p.WriteInt32(charID)
	p.WriteByte(channelID)
	p.WriteString(name)
	p.WriteString(invitee)

	return p
}

func PacketMessengerBlocked(charID int32, channelID, blockMode byte, name, invitee, inviter string) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(constant.MessengerBlocked)
	p.WriteInt32(charID)
	p.WriteByte(channelID)
	p.WriteString(name)
	p.WriteString(invitee)
	p.WriteString(inviter)
	p.WriteByte(blockMode)

	return p
}

func PacketMessengerChat(charID int32, channelID byte, name, message string) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(constant.MessengerChat)
	p.WriteInt32(charID)
	p.WriteByte(channelID)
	p.WriteString(name)
	p.WriteString(message)

	return p
}

func PacketMessengerAvatar(gender, skin, channelID byte, charID, face, hair, cashWeapon, petAccessory int32, name string, vis, hid []KV) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerMessengerEvent)
	p.WriteByte(constant.MessengerAvatar)
	p.WriteInt32(charID)
	p.WriteByte(channelID)
	p.WriteString(name)
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

	p.WriteInt32(cashWeapon)
	p.WriteInt32(petAccessory)
	return p
}

func PacketCashShopRequestChannelInfo() mpacket.Packet {
	p := mpacket.CreateInternal(opcode.CashShopRequestChannelInfo)
	return p
}

func PacketChatMegaphone(chrName, msg string, whisper bool) mpacket.Packet {
	p := mpacket.CreateInternal(opcode.ChannelPlayerChatEvent)
	p.WriteByte(OpChatMegaphone)
	p.WriteString(chrName)
	p.WriteString(msg)
	p.WriteBool(whisper)

	return p
}
