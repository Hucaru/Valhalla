package channel

import (
	"github.com/Hucaru/Valhalla/common/mpacket"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/internal"
)

// TODO: login server needs to send a deleted character event so that they can leave the guild for playing players

type guild struct {
	players []*player
	internal.Guild
}

func (g *guild) broadcast(p mpacket.Packet) {
	for _, v := range g.players {
		if v == nil {
			continue
		}
		v.send(p)
	}
}

func (g *guild) addPlayer() {

}

func (g *guild) removePlayer() {

}

func (g *guild) updateInfo(plr *player, index int32, reader *mpacket.Reader) {
	// pull out pre update information needed to deduce what type of update this is
	g.SerialisePacket(reader)

	if plr != nil {
		plr.guild = g
		plr.send(packetGuildInfo(g))
		// plr.inst.send() // Show guild to other players
	}
}

func packetGuildInfo(guild *guild) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x1a)

	if guild == nil {
		p.WriteByte(0x00) // removes player from guild
		return p
	}

	p.WriteBool(true) // In guild
	p.WriteInt32(1)   // guild id (value cannot be zero)
	p.WriteString(guild.Name)

	// 5 ranks each have a title
	p.WriteString(guild.Master)
	p.WriteString(guild.JrMaster)
	p.WriteString(guild.Member1)
	p.WriteString(guild.Member2)
	p.WriteString(guild.Member3)

	var memberCount byte

	for _, v := range guild.PlayerID {
		if v > 0 {
			memberCount++
		}
	}

	p.WriteByte(memberCount)

	// iterate over all members and output ids
	for i := byte(0); i < memberCount; i++ {
		p.WriteInt32(guild.PlayerID[i])
	}

	// iterate over all members and input their info
	for i := byte(0); i < memberCount; i++ {
		p.WritePaddedString(guild.Names[i], 13)
		p.WriteInt32(guild.Jobs[i])
		p.WriteInt32(guild.Levels[i])
		p.WriteInt32(guild.Ranks[i])

		if guild.Online[i] {
			p.WriteInt32(1)
		} else {
			p.WriteInt32(0)
		}
		p.WriteInt32(0) // ?
	}

	p.WriteInt32(int32(guild.Capacity))
	p.WriteInt16(guild.LogoBg)
	p.WriteByte(guild.LogoBgColour)
	p.WriteInt16(guild.Logo)
	p.WriteByte(guild.LogoColour)
	p.WriteString(guild.Notice)
	p.WriteInt32(guild.Points)

	return p
}

func packetGuildPlayerOnlineNotice(guildID, playerIndex int32, online bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x3d)
	p.WriteInt32(guildID)
	p.WriteInt32(playerIndex)
	p.WriteBool(online)

	return p
}

func packetGuildInviteNotAccepting(name string) mpacket.Packet {
	return packetGuildInviteResult(name, 0x35)
}

func packetGuildInviteHasAnother(name string) mpacket.Packet {
	return packetGuildInviteResult(name, 0x36)
}

func packetGuildInviteRejected(name string) mpacket.Packet {
	return packetGuildInviteResult(name, 0x37)
}

func packetGuildInviteResult(name string, code byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(code)
	p.WriteString(name)

	return p
}

/*
0x32 - guildDisbanded npc dialogue box (updates ui)
i32 - guildID

0x3a - guild capacity npc dialogue box (ui not updated)
i32 - guildID
i8 - capacity

0x3c -
i32 - guildID
i32
i32
i32

0x34 - npc dialogue box saying problem has occured during disbandon

0x3b - npc dialogue box saying problem has occured during capacity increase

0x38 - admin cannot make guild message

0x49 -
i32
i32 - amount
for amount:
	name
	i32

0x4a - less than 5 members remaning, guild quest will end in 5 seconds

0x4b - user that registered has disconnected, quest will end in 5 seconds

0x4c - guild quest status and position in queue
i8 - channelID
i32 - position in queue

0x48 -
i32 - guildID
i32 - ?

0x3c -
i32 - guildID
i32
i8

0x3b -
i32 - guildID
name

between 0x3f - 0x47 -
i32 - guildID
i16
i8
i16
i8

0x3e - update rank titles (dialogue box comes up saying it has been saved) ui is updated
i32 - guildID
name  - master
name
name
name
name - member

0x30 - you are not in the guild

0x29 - the guild you are trying to join has reached maximum capacity

*/
