package channel

import (
	"log"
	"slices"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/mpacket"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
)

type guild struct {
	id       int32
	worldID  int32
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

	players *players

	playerID []int32
	names    []string
	jobs     []int32
	levels   []int32
	online   []bool // TODO: repurpose this field for contract signing
	ranks    []byte
}

func loadGuildFromDb(guildID int32, players *players) (*guild, error) {
	loadedGuild := &guild{}

	row, err := common.DB.Query("SELECT id, guildRank, name, job, level, channelID FROM characters WHERE guildID=?", guildID)

	if err != nil {
		return nil, err
	}

	defer row.Close()

	loadedGuild.playerID = make([]int32, 0, constant.MaxGuildSize)
	loadedGuild.names = make([]string, 0, constant.MaxGuildSize)
	loadedGuild.jobs = make([]int32, 0, constant.MaxGuildSize)
	loadedGuild.levels = make([]int32, 0, constant.MaxGuildSize)
	loadedGuild.online = make([]bool, 0, constant.MaxGuildSize)
	loadedGuild.ranks = make([]byte, 0, constant.MaxGuildSize)

	var channelID int32
	var playerID int32
	var rank byte
	var name string
	var job int32
	var level int32

	for row.Next() {
		err = row.Scan(&playerID, &rank, &name, &job, &level, &channelID)

		if err != nil {
			log.Panicln(err)
		}

		loadedGuild.playerID = append(loadedGuild.playerID, playerID)
		loadedGuild.names = append(loadedGuild.names, name)
		loadedGuild.jobs = append(loadedGuild.jobs, job)
		loadedGuild.levels = append(loadedGuild.levels, level)
		loadedGuild.online = append(loadedGuild.online, channelID > -1)
		loadedGuild.ranks = append(loadedGuild.ranks, rank)
	}

	query := "SELECT id,capacity,name,notice,master,jrMaster,member1,member2,member3,logoBg,logoBgColour,logo,logoColour,points FROM guilds WHERE id=?"
	err = common.DB.QueryRow(query, guildID).Scan(&loadedGuild.id, &loadedGuild.capacity,
		&loadedGuild.name, &loadedGuild.notice, &loadedGuild.master, &loadedGuild.jrMaster, &loadedGuild.member1,
		&loadedGuild.member2, &loadedGuild.member3, &loadedGuild.logoBg, &loadedGuild.logoBgColour, &loadedGuild.logo,
		&loadedGuild.logoColour, &loadedGuild.points)

	if err != nil {
		return nil, err
	}

	loadedGuild.players = players

	return loadedGuild, nil
}

func createGuildContract(guildName string, worldID int32, players *players, master *player) *guild {
	newGuild := &guild{
		worldID: worldID,
		name:    guildName,
		players: players,
	}

	newGuild.playerID = append(newGuild.playerID, master.id)
	newGuild.names = append(newGuild.names, master.name)
	newGuild.jobs = append(newGuild.jobs, int32(master.job))
	newGuild.levels = append(newGuild.levels, int32(master.level))
	newGuild.online = append(newGuild.online, true)
	newGuild.ranks = append(newGuild.ranks, 1)

	newGuild.master = "Master"
	newGuild.jrMaster = "Jr. Master"
	newGuild.member1 = "Member"
	newGuild.member2 = "Member"
	newGuild.member3 = "Member"

	for _, plr := range master.party.players {
		if plr == master {
			continue
		}

		if plr == nil {
			continue
		}

		if plr.mapID != master.mapID {
			continue
		}

		newGuild.playerID = append(newGuild.playerID, plr.id)
		newGuild.names = append(newGuild.names, plr.name)
		newGuild.jobs = append(newGuild.jobs, int32(plr.job))
		newGuild.levels = append(newGuild.levels, int32(plr.level))
		newGuild.online = append(newGuild.online, false)
		newGuild.ranks = append(newGuild.ranks, 5)

		plr.guild = newGuild
		plr.send(packetGuildContract(master.party.ID, master.name, guildName))
	}

	master.guild = newGuild

	return newGuild
}

func (g *guild) signContract(playerID int32) error {
	for i, id := range g.playerID {
		if id == playerID {
			g.online[i] = true // switching a guild character to online during contract signing stage means they have accepted
			g.ranks[i] = 5
			break
		}
	}

	signed := 0
	for _, v := range g.online {
		if v {
			signed++
		}
	}

	if signed == len(g.online) {
		query := "INSERT INTO guilds (name, worldID, notice, master, jrMaster, member1, member2, member3) VALUES (?, ?, '', ?, ?, ?, ?, ?)"
		res, err := common.DB.Exec(query, g.name, g.worldID, g.master, g.jrMaster, g.member1, g.member2, g.member3)

		if err != nil {
			return err
		}

		guildID, err := res.LastInsertId()

		if err != nil {
			return err
		}

		g.id = int32(guildID)

		for i, id := range g.playerID {
			// add each member to guild
			query = "UPDATE characters set guildID=?, guildRank=? WHERE id=?"
			_, err := common.DB.Exec(query, g.id, g.ranks[i], id)

			if err != nil {
				return err
			}

			plr, err := g.players.getFromID(id)

			if err != nil {
				return err
			}

			g.broadcastExcept(packetGuildPlayerJoined(plr), plr)
			plr.send(packetGuildInfo(g))
			g.updateAvatar(plr)
		}

		plr, err := g.players.getFromID(g.playerID[0])

		if err != nil {
			return err
		}

		plr.giveMesos(-5e6)
	}

	return nil
}

func (g *guild) broadcast(p mpacket.Packet) {
	for _, v := range g.playerID {
		plr, err := g.players.getFromID(v)

		if err != nil {
			continue
		}

		plr.send(p)
	}
}

func (g *guild) broadcastExcept(p mpacket.Packet, pass *player) {
	for _, v := range g.playerID {
		plr, err := g.players.getFromID(v)

		if err != nil || plr == pass {
			continue
		}

		plr.send(p)
	}
}

func (g guild) updateAvatar(plr *player) {
	if plr == nil {
		return
	}

	plr.inst.sendExcept(packetMapPlayerLeft(plr.id), plr.conn)
	plr.inst.sendExcept(packetMapPlayerEnter(plr), plr.conn)
}

func (g *guild) updateEmblem(logoBg, logo int16, logoBgColour, logoColour byte) {
	g.logoBg = logoBg
	g.logo = logo
	g.logoBgColour = logoBgColour
	g.logoBgColour = logoColour

	for _, v := range g.playerID {
		plr, err := g.players.getFromID(v)

		if err != nil {
			continue
		}

		g.updateAvatar(plr)
		plr.send(packetGuildUpdateEmblem(g.id, logoBg, logo, logoBgColour, logoColour))
	}
}

func (g *guild) updateTitles(master, jrMaster, member1, member2, member3 string) {
	g.master = master
	g.jrMaster = jrMaster
	g.member1 = member1
	g.member2 = member2
	g.member3 = member3

	g.broadcast(packetGuilderTitlesUpdate(g.id, master, jrMaster, member1, member2, member3))
}

func (g *guild) setPoints(points int32) {
	g.points = points
	g.broadcast(packetGuildSetPoints(g.id, points))
}

func (g *guild) updateRank(playerID int32, rank byte) {
	for i, id := range g.playerID {
		if id == playerID {
			g.ranks[i] = rank
			g.broadcast(packetGuildRankUpdate(g.id, playerID, int32(rank)))
			break
		}
	}
}

func (g *guild) addPlayer(playerID int32, name string, jobID, level int32, online bool, rank byte) {
	plr, err := g.players.getFromID(playerID)

	if int(g.capacity) == len(g.online) {
		if err == nil {
			plr.send(packetGuildCannotJoinMaxPlayers())
		}

		return
	}

	_, err = common.DB.Exec("UPDATE characters SET guildID=?, guildRank=? WHERE id=?", g.id, 5, playerID)

	if err != nil {
		log.Fatal()
	}

	if plr != nil {
		plr.guild = g
		g.updateAvatar(plr)
		plr.send(packetGuildInfo(g))
		g.broadcast(packetGuildPlayerJoined(plr))
	}

	g.playerID = append(g.playerID, playerID)
	g.names = append(g.names, name)
	g.jobs = append(g.jobs, jobID)
	g.levels = append(g.levels, level)
	g.online = append(g.online, true)
	g.ranks = append(g.ranks, rank)
}

func (g *guild) removePlayer(playerID int32, expelled bool, name string) {
	for i, id := range g.playerID {
		if id == playerID {
			g.playerID = slices.Delete(g.playerID, i, i+1)
			g.names = slices.Delete(g.names, i, i+1)
			g.jobs = slices.Delete(g.jobs, i, i+1)
			g.levels = slices.Delete(g.levels, i, i+1)
			g.online = slices.Delete(g.online, i, i+1)
			g.ranks = slices.Delete(g.ranks, i, i+1)

			if plr, err := g.players.getFromID(id); err == nil {
				plr.guild = nil
				plr.send(packetGuildInfo(nil))
				g.updateAvatar(plr)
			}

			break
		}
	}

	g.broadcast(packetGuildRemovePlayer(g.id, playerID, name, expelled))
}

func (g *guild) playerOnline(playerID int32, plr *player, online, changeChannel bool) {
	for i, id := range g.playerID {
		if id == playerID {
			g.online[i] = online

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
	for _, v := range g.playerID {
		plr, err := g.players.getFromID(v)

		if err != nil {
			continue
		}

		plr.send(packetGuildDisbandMessage(g.id))
		plr.guild = nil
		g.updateAvatar(plr)
	}
}

func (g guild) isMaster(p *player) bool {
	for i, v := range g.playerID {
		if v == p.id && g.ranks[i] == 1 {
			return true
		}
	}

	return false
}

func packetGuildEnterName() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x01)

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

func packetGuildInviteCard(guildID int32, inviter string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x05)
	p.WriteInt32(guildID)
	p.WriteString(inviter)

	return p
}

func packetGuildEmblemEditor() mpacket.Packet {
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

	p.WriteBool(true)
	p.WriteInt32(guild.id)
	p.WriteString(guild.name)

	// 5 ranks each have a title
	p.WriteString(guild.master)
	p.WriteString(guild.jrMaster)
	p.WriteString(guild.member1)
	p.WriteString(guild.member2)
	p.WriteString(guild.member3)

	memberCount := byte(len(guild.playerID))
	p.WriteByte(memberCount)

	// The client wants the data listed in order from master to member 3
	for j := byte(1); j < 6; j++ {
		for i, rank := range guild.ranks {
			if rank != j {
				continue
			}

			p.WriteInt32(guild.playerID[i])
		}
	}

	for j := byte(1); j < 6; j++ {
		for i, rank := range guild.ranks {
			if rank != j {
				continue
			}

			p.WritePaddedString(guild.names[i], 13)
			p.WriteInt32(guild.jobs[i])
			p.WriteInt32(guild.levels[i])
			p.WriteInt32(int32(guild.ranks[i]))

			if guild.online[i] {
				p.WriteInt32(1)
			} else {
				p.WriteInt32(0)
			}

			p.WriteInt32(0) // ?
		}
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

func packetGuildNameInUse() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x1c)

	return p
}

func packetGuildAgreementProblem() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x1f)

	return p
}

func packetGuildAlreadyJoined() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x21)

	return p
}

func packetGuildCannotMakeLevel() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x23)

	return p
}

func packetGuildContractDisagree() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x24)

	return p
}

func packetGuildProblemOccurred() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x26)

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

func packetGuildCannotJoinMaxPlayers() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x29)

	return p
}

func packetGuildRemovePlayer(guildID, playerID int32, name string, expelled bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	if expelled {
		p.WriteByte(0x2f)
	} else {
		p.WriteByte(0x2c)
	}

	p.WriteInt32(guildID)
	p.WriteInt32(playerID)
	p.WriteString(name)

	return p
}

func packetGuildNotIn() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x30)

	return p
}

func packetGuildDisbandMessage(guildID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x032)
	p.WriteInt32(guildID)

	return p
}

func packetGuildInviteNotAccepting(name string) mpacket.Packet {
	return packetGuildInviteResult(name, 0x35)
}

func packetGuildInviteeHasAnother(name string) mpacket.Packet {
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

func packetGuildDisbandErrorNPC() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x34)

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

func packetGuilderTitlesUpdate(guildID int32, master, jrMaster, member1, member2, member3 string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x3e)
	p.WriteInt32(guildID)
	p.WriteString(master)
	p.WriteString(jrMaster)
	p.WriteString(member1)
	p.WriteString(member2)
	p.WriteString(member3)

	return p
}

func packetGuildRankUpdate(guildID, playerID, rank int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x40)
	p.WriteInt32(guildID)
	p.WriteInt32(playerID)
	p.WriteInt32(rank)

	return p
}

func packetGuildUpdateEmblem(guildID int32, logoBg, logo int16, logoBgColour, logoColour byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x42)
	p.WriteInt32(guildID)
	p.WriteInt16(logoBg)
	p.WriteByte(logoBgColour)
	p.WriteInt16(logo)
	p.WriteByte(logoColour)

	return p
}

func packetGuildUpdateNotice(guildID int32, notice string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x44)
	p.WriteInt32(guildID)
	p.WriteString(notice)

	return p
}

func packetGuildSetPoints(guildID, points int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelGuildInfo)
	p.WriteByte(0x48)
	p.WriteInt32(guildID)
	p.WriteInt32(points)

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

0x20 - ?, dc

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

0x2c - deleted characters removed from guild, left
i32 - guild id
i32 - player id
string - name

0x2d - you are not in the guild

0x2e - not accepted due to unkown reason

0x2f - deleted character removed from guild, expelled
i32 - guild id
i32 - player id
string - name

0x30 - you are not in the guild

0x31 - not accepted due to unkown reason

0x32 - disband guild npc ui for leader, removes player from guild as well, send to all players
i32 - guildID
i8 - ?

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

0x3c - level , job id change
i32 - guildID
i32 - playerID
i32 - level
i32 - jobID

0x3e - update rank titles (dialogue box comes up saying it has been saved) ui is updated
i32 - guildID
name  - master
name
name
name
name - member

0x3d - guild player online

0x3e - it is saved dialogue message

0x3f - the guild request has not been accepted for unkown reason

0x40 - rank change
i32 - guild id
i32 - player id
i32 - guild rank

0x41 - the guild request has not been accepted for unkown reason

0x42 - it is compelete dialogue message, removes/changes the emblem
i32 - gid
i16 - bg
i16 - bgcolour
i16 - logo
i16 - logocolour

0x43 - the guild request has not been accepted for unkown reason

0x44 - update notice

0x45 - 0x47 - the guild request has not been accepted for unkown reason

0x48 - update guild points
i32 - guildID
i32 - gp points

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

*/
