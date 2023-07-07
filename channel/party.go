package channel

import (
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mpacket"
)

type party struct {
	id              int32
	serverChannelID int32

	players   [constant.MaxPartySize]*player
	channelID [constant.MaxPartySize]int32
	playerID  [constant.MaxPartySize]int32
	name      [constant.MaxPartySize]string
	mapID     [constant.MaxPartySize]int32
	job       [constant.MaxPartySize]int32
	level     [constant.MaxPartySize]int32
}

func newParty(id int32, plr *player, channelID byte, playerID, mapID, job, level int32, playerName string, serverChannelID int32) party {
	result := party{id: id, serverChannelID: serverChannelID}
	result.players[0] = plr

	result.channelID[0] = int32(channelID)
	result.playerID[0] = playerID
	result.name[0] = playerName
	result.mapID[0] = mapID
	result.job[0] = job
	result.level[0] = level

	return result
}

func (d party) broadcast(p mpacket.Packet) {
	for _, v := range d.players {
		if v == nil {
			continue
		}

		v.send(p)
	}
}

func (d party) broadcastExcept(p mpacket.Packet, id int32) {
	for i, v := range d.players {
		if v == nil || d.playerID[i] == id {
			continue
		}

		v.send(p)
	}
}

func (d *party) addPlayer(plr *player, channelID int32, id int32, name string, mapID, job, level int32) {
	for i, v := range d.playerID {
		if v == 0 {
			d.players[i] = plr

			if plr != nil {
				plr.party = d
			}

			d.channelID[i] = channelID
			d.playerID[i] = id
			d.name[i] = name
			d.mapID[i] = mapID
			d.job[i] = job
			d.level[i] = level

			d.broadcast(packetPartyPlayerJoin(d.id, name, d))

			return
		}
	}
}

func (d *party) removePlayer(plr *player, playerID int32, kick bool) {
	for i, v := range d.playerID {
		if v == playerID {
			if i == 0 {
				d.broadcast(packetPartyLeave(d.id, playerID, false, kick, "", d))

				for _, p := range d.players {
					if p != nil {
						p.party = nil
					}
				}

				return
			}

			name := d.name[i]

			d.channelID[i] = 0
			d.playerID[i] = 0
			d.name[i] = ""
			d.mapID[i] = 0
			d.job[i] = 0
			d.level[i] = 0

			if plr != nil {
				plr.party = nil
			}

			d.broadcast(packetPartyLeave(d.id, playerID, true, kick, name, d))

			d.players[i] = nil

			return
		}
	}
}

func (d party) full() bool {
	for _, v := range d.players {
		if v == nil {
			return false
		}
	}

	return true
}

func (d *party) updatePlayerMap(playerID, mapID int32) {
	for i, v := range d.playerID {
		if v == playerID {
			d.mapID[i] = mapID

			d.broadcast(packetPartyUpdate(d.id, d))

			return
		}
	}
}

func (d *party) setPlayerChannel(plr *player, playerID int32, cashshop bool, offline bool, channelID int32, mapID int32) {
	for i, v := range d.playerID {
		if v == playerID {
			if cashshop {
				d.channelID[i] = -1
			} else if offline {
				d.channelID[i] = -2
			} else {
				d.channelID[i] = channelID
			}

			d.mapID[i] = mapID
			d.players[i] = plr

			d.broadcast(packetPartyUpdate(d.id, d))
			return
		}
	}
}

func (d *party) updateJobLevel(playerID, job, level int32) {
	for i, v := range d.playerID {
		if v == playerID {
			d.job[i] = job
			d.level[i] = level

			d.broadcast(packetPartyUpdateJobLevel(playerID, job, level))
			return
		}
	}
}

func (d party) leader(id int32) bool {
	return d.playerID[0] == id
}

func (d party) member(id int32) bool {
	for _, v := range d.playerID {
		if v == id {
			return true
		}
	}

	return false
}

func (d party) giveExp(playerID, amount int32, sameMap bool) {
	var mapID int32 = 0

	for i, id := range d.playerID {
		if id == playerID {
			mapID = d.mapID[i]
			break
		}
	}

	if sameMap {
		nPlayers := 0

		for i, id := range d.playerID {
			if id != playerID && d.players[i] != nil && d.mapID[i] == mapID && d.players[i].hp > 0 {
				nPlayers++
			}

			if nPlayers == 0 {
				return
			}
		}
	}

	for _, plr := range d.players {
		if plr != nil && sameMap && plr.mapID == mapID {
			plr.giveEXP(amount, false, true)
		}
	}
}

func packetPartyCreate(partyID int32, doorMap1, doorMap2 int32, point pos) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPartyInfo)
	p.WriteByte(0x07)
	p.WriteInt32(partyID)

	if doorMap1 > -1 {
		p.WriteInt32(doorMap1)
		p.WriteInt32(doorMap2)
		p.WriteInt16(point.x)
		p.WriteInt16(point.y)
	} else {
		p.WriteInt32(-1)
		p.WriteInt32(-1)
		p.WriteInt32(0) // empty pos
	}

	return p
}

func packetPartyPlayerJoin(partyID int32, name string, party *party) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPartyInfo)
	p.WriteByte(0x0e)
	p.WriteInt32(partyID)
	p.WriteString(name)

	updateParty(&p, party)

	return p
}

func packetPartyUpdate(partyID int32, party *party) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPartyInfo)
	p.WriteByte(0x1a)
	p.WriteInt32(partyID)

	updateParty(&p, party)

	return p
}

func packetPartyLeave(partyID, playerID int32, keepParty, kicked bool, name string, party *party) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPartyInfo)
	p.WriteByte(0x0b)
	p.WriteInt32(partyID)
	p.WriteInt32(playerID)
	p.WriteBool(keepParty)

	if keepParty {
		p.WriteBool(kicked)
		p.WriteString(name)
		updateParty(&p, party)
	}

	return p
}

func packetPartyDoorUpdate(index byte, townID, mapID int32, point pos) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPartyInfo)
	p.WriteByte(0x1c)
	p.WriteByte(index)
	p.WriteInt32(townID)
	p.WriteInt32(mapID)
	p.WriteInt16(point.x)
	p.WriteInt16(point.y)

	return p
}

func packetPartyUpdateJobLevel(playerID, job, level int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPartyInfo)
	p.WriteByte(0x1b)
	p.WriteInt32(playerID)
	p.WriteInt32(level)
	p.WriteInt32(job)

	return p
}

func updateParty(p *mpacket.Packet, party *party) {
	validOffsets := make([]int, 0, constant.MaxPartySize)

	for i, v := range party.level {
		if v != 0 {
			validOffsets = append(validOffsets, i)
		}
	}

	paddAmount := constant.MaxPartySize - len(validOffsets)

	for _, v := range validOffsets {
		p.WriteInt32(party.playerID[v])
	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(0)
	}

	for _, v := range validOffsets {
		p.WritePaddedString(party.name[v], 13)
	}

	for i := 0; i < paddAmount; i++ {
		p.WritePaddedString("", 13)
	}

	for _, v := range validOffsets {
		p.WriteInt32(party.job[v])
	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(0)
	}

	for _, v := range validOffsets {
		p.WriteInt32(party.level[v])
	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(0)
	}

	for _, v := range validOffsets {
		p.WriteInt32(party.channelID[v]) // -1 - cashshop, -2 - offline
	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(0)
	}

	for _, v := range validOffsets {
		if party.channelID[v] != party.serverChannelID {
			p.WriteInt32(-1)
		} else {
			p.WriteInt32(party.mapID[v])
		}

	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(0)
	}

	p.WriteInt32(party.playerID[0])

	// Mystic door
	for range validOffsets {
		p.WriteInt32(-1)
		p.WriteInt32(-1)
		p.WriteInt32(0) // x
		p.WriteInt32(0) // y
	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(-1)
		p.WriteInt32(-1)
		p.WriteInt64(0) // int64?
	}
}

func packetPartyCreateUnknownError() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPartyInfo)
	p.WriteByte(0)

	return p
}

func packetPartyInviteNotice(partyID int32, fromName string) mpacket.Packet {
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

func packetPartyAlreadyJoined() mpacket.Packet {
	return packetPartyMessage(0x08)
}

func packetPartyBeginnerCannotCreate() mpacket.Packet {
	return packetPartyMessage(0x09)
}

func packetPartyNotInParty() mpacket.Packet {
	return packetPartyMessage(0x0c)
}

func packetPartyAlreadyJoined2() mpacket.Packet {
	return packetPartyMessage(0x0f)
}

func packetPartyToJoinIsFull() mpacket.Packet {
	return packetPartyMessage(0x10)
}

func packetPartyUnableToFindPlayer() mpacket.Packet {
	return packetPartyMessage(0x11)
}

func packetPartyAdminNoCreate() mpacket.Packet {
	return packetPartyMessage(0x18)
}

func packetPartyUnableToFindPlayer2() mpacket.Packet {
	return packetPartyMessage(0x19)
}

func packetPartyMessageName(op byte, name string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPartyInfo)
	p.WriteByte(op)
	p.WriteString(name)

	return p
}

func packetPartyBlockingInvites(name string) mpacket.Packet {
	return packetPartyMessageName(0x13, name)
}

func packetPartyHasOtherRequest(name string) mpacket.Packet {
	return packetPartyMessageName(0x14, name)
}

func packetPartyRequestDenied(name string) mpacket.Packet {
	return packetPartyMessageName(0x15, name)
}
