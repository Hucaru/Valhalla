package channel

import (
	"fmt"
	"log"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/mpacket"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
)

type guildContract struct {
	leader    *player
	guildName string
	signers   map[int32]bool
	accepted  int
}

func createGuildContract(plr *player, name string) *guildContract {
	return &guildContract{leader: plr, guildName: name, signers: make(map[int32]bool)}
}

func (c *guildContract) send() {
	for _, v := range c.leader.party.players {
		if v != nil && v.mapID == c.leader.mapID {
			if v.id != c.leader.id {
				c.signers[v.id] = false
			}
			v.send(packetGuildContract(c.leader.party.ID, c.leader.name, c.guildName))
		}
	}
}

func (c *guildContract) sign(playerID int32, accept bool) bool {
	if _, ok := c.signers[playerID]; ok && accept {
		c.signers[playerID] = true
		c.accepted++
	}

	return c.accepted == 1
}

func (c guildContract) error() {
	c.leader.send(packetGuildAgreementProblem())
}

// Note: since players all have to be on same channel and same map when contract signing we don't need to send an interserver message to add players
func (c guildContract) addPlayers(guild *guild, players *players) {
	guild.addPlayer(c.leader, c.leader.id, c.leader.name, int32(c.leader.job), int32(c.leader.level), 1)

	for id := range c.signers {
		plr, err := players.getFromID(id)

		if err != nil {
			log.Println(err)
			continue
		}

		err = guild.addPlayer(plr, plr.id, plr.name, int32(plr.job), int32(plr.level), 5)

		if err != nil {
			log.Println(err)
		}
	}
}

// TODO: login server needs to send a deleted character event so that world server can update currently playing players

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

func createGuild(name string, worldID int32) (*guild, error) {
	master := "Master"
	jsMaster := "Jr. Master"
	member := "Member"

	query := "INSERT INTO guilds (name, worldID, notice, master, jrMaster, member1, member2, member3) VALUES (?, ?, '', ?, ?, ?, ?, ?)"
	res, err := common.DB.Exec(query, name, worldID, master, jsMaster, member, member, member)

	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()

	if err != nil {
		return nil, err
	}

	newGuild := &guild{
		id:       int32(id),
		name:     name,
		master:   master,
		jrMaster: jsMaster,
		member1:  member,
		member2:  member,
		member3:  member,
	}

	return newGuild, nil
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

func (g guild) updateAvatar(plr *player) {
	plr.inst.sendExcept(packetMapPlayerLeft(plr.id), plr.conn)
	plr.inst.sendExcept(packetMapPlayerEnter(plr), plr.conn)
}

func (g *guild) addPlayer(plr *player, playerID int32, name string, jobID, level int32, rank int32) error {
	index := -1
	for i, v := range g.levels {
		if v == 0 {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("Guild at capacity")
	}

	_, err := common.DB.Exec("UPDATE characters SET guildID=? WHERE id=?", g.id, playerID)

	if err != nil {
		return err
	}

	g.players[index] = plr
	g.playerID[index] = playerID
	g.names[index] = name
	g.jobs[index] = jobID
	g.levels[index] = level
	g.online[index] = true
	g.ranks[index] = rank

	if plr != nil {
		plr.guild = g
		g.updateAvatar(plr)
		plr.send(packetGuildInfo(g))
		g.broadcast(packetGuildPlayerJoined(plr))
	}

	return nil
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

func (g guild) disband() error {
	for _, plr := range g.players {
		if plr != nil {
			plr.guild = nil
			g.updateAvatar(plr)
		}
	}

	g.broadcast(packetGuildInfo(nil))
	return nil
}

func packetGuildEnterName() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x01)

	return p
}

func packetGuildInviteCard(guildID int32, name string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x05)
	p.WriteInt32(guildID)
	p.WriteString(name)

	return p
}

func packetGuildContract(partyID int32, masterName, guildName string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x03)
	p.WriteInt32(partyID)
	p.WriteString(masterName)
	p.WriteString(guildName)

	return p
}

func packetGuildCreateEmblem() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x11)

	return p
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

	for _, v := range guild.levels {
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

func packetGuildAgreementProblem() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x1f)

	return p
}

func packetGuildPlayerJoined(plr *player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x27)
	p.WriteInt32(plr.guild.id)
	p.WriteInt32(plr.id)
	p.WritePaddedString(plr.name, 13)
	p.WriteInt32(int32(plr.job))
	p.WriteInt32(int32(plr.level))
	p.WriteInt32(5)
	p.WriteInt32(1) // online
	p.WriteInt32(0) // ?

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
0x01 - type name of guild dialogue box

0x02 - not accepted due to unkown reason

0x03 - guild contract accept decline ui to other people
/packet 30030000000000000000

0x04 - not accepted due to unkown reason

0x05 - guild invite card
i32 - guild id
string - inviter name

0x06 - 0x10 - not accepted due to unkown reason

0x11 - create emblem ui

0x12 - 0x19 - not accepted due to unkown reason

0x1a - guild info

0x1b - not accepted due to unkown reason

0x1c - name already in use npc message ui

0x1f - problem in gathering agreements npc message ui

0x20 - ?, dc without packet buffer length error

0x21 - already joined guild

0x22 - not accepted due to unknown reason

0x23 - cannot make guild due to level requirement message

0x24 - someone has diagreed to form guild, come back when you meet the right people npc ui message

0x25 - not accepted due to unkown reason

0x26 - problem has happened during process of forming guild npc ui

0x27 - player joined guild

0x28 - already joined the guild message

0x29 - the guild you are trying to join has reached maximum capacity

0x2a - character cannot be found in current channel

0x2b - not accepted due to unkown reason

0x2c - deleted characters removed from guild

0x2d - you are not in the guild

0x2e - not accepted due to unkown reason

0x2f - deleted character removed from guild

0x30 - you are not in the guild

0x31 - not accepted due to unkown reason

0x32 - disband guild npc ui, removes player from guild as well

0x33 - not accepted due to unkown reason

0x34 - npc dialogue box saying problem has occured during disbandon

0x35 - name is currently not accepted guild request invites

0x36 - name is taking care of another invitation

0x37 - name has denied your invitation

0x38 - admin cannot make guild message

0x39 - not accepted due to unkown reason

0x3a - guild capacity npc dialogue box (ui not updated)
i32 - guildID
i8 - capacity

0x3b - guild capacity problem dialogue box

0x3c -
i32 - guildID
i32
i32
i32

0x3e - update rank titles (dialogue box comes up saying it has been saved) ui is updated
i32 - guildID
name  - master
name
name
name
name - member

0x3d - ?

0x3e - it is saved dialogue message

0x3f - the guild request has not been accepted for unkown reason

0x40 - ?

0x41 - the guild request has not been accepted for unkown reason

0x42 - it is compelete dialogue message, removes the emblem

0x43 - the guild request has not been accepted for unkown reason

0x44 - update notice

0x45 - 0x47 - the guild request has not been accepted for unkown reason

0x48 -
i32 - guildID
i32 - ?

0x49 - some ui thing?
i32 - guildID?
i32 - amount
for amount:
	name
	i32 - points?

0x4a - less than 5 members remaning, guild quest will end in 5 seconds

0x4b - user that registered has disconnected, quest will end in 5 seconds

0x4c - guild quest status and position in queue
i8 - channelID
i32 - position in queue, 1 is enter now, 2 is head to quest map to wait, 3 and up is currently one guild participating and you are n - 1 on waiting list

0x4d - 0x4f - the guild request has not been accepted for unkown reason

*/
