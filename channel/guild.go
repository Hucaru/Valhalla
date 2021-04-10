package channel

import (
	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/mpacket"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
)

// TODO: login server needs to send a deleted character event so that they can leave the guild for playing players

type guild struct {
	id       int32
	capacity byte
	name     string
	notice   string

	master   string
	jrMaster string
	member1  string
	member2  string
	member3  string

	logoBg, logo             int16
	logoBgColour, logoColour byte

	points int32

	players  [constant.MaxGuildSize]*player
	playerID [constant.MaxGuildSize]int32
	names    [constant.MaxGuildSize]string
	jobs     [constant.MaxGuildSize]int32
	levels   [constant.MaxGuildSize]int32
	online   [constant.MaxGuildSize]bool
	ranks    [constant.MaxGuildSize]int32
}

func loadGuildFromDb(guildID int32) (*guild, error) {
	loadedGuild := &guild{}

	row, err := common.DB.Query("SELECT id, guildRankID, name, job, level, channelID FROM characters WHERE guildID=?", guildID)

	if err != nil {
		return nil, err
	}

	defer row.Close()

	var i int32
	var channelID int32
	for row.Next() {
		err = row.Scan(&loadedGuild.playerID[i], &loadedGuild.ranks[i], &loadedGuild.names[i], &loadedGuild.jobs[i], &loadedGuild.levels[i], &channelID)

		if channelID > -1 {
			loadedGuild.online[i] = true
		}

		i++
	}

	query := "id,capacity,name,notice,master,jrMaster,member1,member2,member3,logoBg,logoBgColour,logo,logoColour,points"
	common.DB.QueryRow("SELECT "+query+" FROM guilds WHERE id=?", guildID).Scan(&loadedGuild.id, &loadedGuild.capacity,
		&loadedGuild.name, &loadedGuild.notice, &loadedGuild.master, &loadedGuild.jrMaster, &loadedGuild.member1,
		&loadedGuild.member2, &loadedGuild.member3, &loadedGuild.logoBg, &loadedGuild.logoBgColour, &loadedGuild.logo,
		&loadedGuild.logoColour, &loadedGuild.points)

	return loadedGuild, nil
}

func (g *guild) broadcast(p mpacket.Packet) {
	for _, v := range g.players {
		if v == nil {
			continue
		}
		v.send(p)
	}
}

func (g *guild) broadcastExcept(p mpacket.Packet, plr *player) {
	for _, v := range g.players {
		if v == nil || v == plr {
			continue
		}
		v.send(p)
	}
}

func (g *guild) playerOnline(playerID int32, plr *player, online, changeChannel bool) {
	for i, id := range g.playerID {
		if id == playerID {
			g.online[i] = online
			g.players[i] = plr

			if plr != nil {
				plr.send(packetGuildInfo(g))
			}

			if !changeChannel {
				g.broadcastExcept(packetGuildPlayerOnlineNotice(g.id, id, online), plr)
			}

			return
		}
	}
}

func (g guild) canUnload() bool {
	for _, v := range g.online {
		if v {
			return false
		}
	}

	return true
}

func (g guild) disband() {
	// for _, plr := range g.players {
	// 	if plr != nil {
	// 		plr.inst.sendExcept(, plr.conn) // remove guild from under player avatar
	// 	}
	// }

	g.broadcast(packetGuildInfo(nil))
}

func packetGuildInfo(guild *guild) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x1a)

	if guild == nil {
		p.WriteByte(0x00) // removes player from guild
		return p
	}

	p.WriteBool(true) // In guild
	p.WriteInt32(guild.id)
	p.WriteString(guild.name)

	// 5 ranks each have a title
	p.WriteString(guild.master)
	p.WriteString(guild.jrMaster)
	p.WriteString(guild.member1)
	p.WriteString(guild.member2)
	p.WriteString(guild.member3)

	var memberCount byte

	for _, v := range guild.playerID {
		if v > 0 {
			memberCount++
		}
	}

	p.WriteByte(memberCount)

	// iterate over all members and output ids
	for i := byte(0); i < memberCount; i++ {
		p.WriteInt32(guild.playerID[i])
	}

	// iterate over all members and input their info
	for i := byte(0); i < memberCount; i++ {
		p.WritePaddedString(guild.names[i], 13)
		p.WriteInt32(guild.jobs[i])
		p.WriteInt32(guild.levels[i])
		p.WriteInt32(guild.ranks[i])

		if guild.online[i] {
			p.WriteInt32(1)
		} else {
			p.WriteInt32(0)
		}
		p.WriteInt32(0) // ?
	}

	p.WriteInt32(int32(guild.capacity))
	p.WriteInt16(guild.logoBg)
	p.WriteByte(guild.logoBgColour)
	p.WriteInt16(guild.logo)
	p.WriteByte(guild.logoColour)
	p.WriteString(guild.notice)
	p.WriteInt32(guild.points)

	return p
}

func packetGuildPlayerOnlineNotice(guildID, playerID int32, online bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x3d)
	p.WriteInt32(guildID)
	p.WriteInt32(playerID)
	p.WriteBool(online)

	return p
}

func packetGuildDisbandNpc(guildID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x032)
	p.WriteInt32(guildID)

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
0x3a - guild capacity npc dialogue box (ui not updated)
i32 - guildID
i8 - capacity

0x3b - guild capacity problem dialogue box

0x3c -
i32 - guildID
i32
i32
i32

0x34 - npc dialogue box saying problem has occured during disbandon

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
