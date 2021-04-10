package channel

import (
	"github.com/Hucaru/Valhalla/common/mpacket"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/internal"
)

// TODO: login server needs to send a deleted character event so that they can leave the party for playing players

type party struct {
	serverChannelID int32
	players         [constant.MaxPartySize]*player
	internal.Party
}

func (d party) broadcast(p mpacket.Packet) {
	for _, v := range d.players {
		if v == nil {
			continue
		}

		v.send(p)
	}
}

func (d *party) addExistingPlayer(plr *player) bool {
	for i, id := range d.PlayerID {
		if id == plr.id {
			d.players[i] = plr
			plr.party = d
			return true
		}
	}

	return false
}

func (d *party) addPlayer(plr *player, index int32, reader *mpacket.Reader) {
	if plr != nil {
		d.players[index] = plr
		plr.party = d
	}

	d.SerialisePacket(reader)
	d.broadcast(packetPartyPlayerJoin(d.ID, d.Name[index], d))
}

func (d *party) removePlayer(index int32, kick bool, reader *mpacket.Reader) {
	playerID := d.PlayerID[index]
	name := d.Name[index]

	d.SerialisePacket(reader)

	if index == 0 {
		d.broadcast(packetPartyLeave(d.ID, playerID, false, kick, "", d))

		for _, p := range d.players {
			if p != nil {
				p.party = nil
			}
		}
	} else {
		if d.players[index] != nil {
			d.players[index].party = nil
		}

		d.broadcast(packetPartyLeave(d.ID, playerID, true, kick, name, d))

		d.players[index] = nil
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

func (d *party) updateOnlineStatus(index int32, plr *player, reader *mpacket.Reader) {
	d.SerialisePacket(reader)
	d.players[index] = plr

	if plr != nil {
		plr.party = d
	}

	d.broadcast(packetPartyUpdate(d.ID, d))
}

func (d *party) updateInfo(index int32, reader *mpacket.Reader) {
	mapID := d.MapID[index]
	d.SerialisePacket(reader)

	if mapID != d.MapID[index] {
		d.broadcast(packetPartyUpdate(d.ID, d))
	} else {
		playerID := d.PlayerID[index]
		job := d.Job[index]
		level := d.Level[index]

		d.broadcast(packetPartyUpdateJobLevel(playerID, job, level))
	}
}

func (d party) giveExp(playerID, amount int32, sameMap bool) {
	var mapID int32 = 0

	for i, id := range d.PlayerID {
		if id == playerID {
			mapID = d.MapID[i]
			break
		}
	}

	if sameMap {
		nPlayers := 0

		for i, id := range d.PlayerID {
			if id != playerID && d.players[i] != nil && d.MapID[i] == mapID && d.players[i].hp > 0 {
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

	for i, v := range party.Level {
		if v != 0 {
			validOffsets = append(validOffsets, i)
		}
	}

	paddAmount := constant.MaxPartySize - len(validOffsets)

	for _, v := range validOffsets {
		p.WriteInt32(party.PlayerID[v])
	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(0)
	}

	for _, v := range validOffsets {
		p.WritePaddedString(party.Name[v], 13)
	}

	for i := 0; i < paddAmount; i++ {
		p.WritePaddedString("", 13)
	}

	for _, v := range validOffsets {
		p.WriteInt32(party.Job[v])
	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(0)
	}

	for _, v := range validOffsets {
		p.WriteInt32(party.Level[v])
	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(0)
	}

	for _, v := range validOffsets {
		p.WriteInt32(party.ChannelID[v]) // -1 - cashshop, -2 - offline
	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(0)
	}

	for _, v := range validOffsets {
		if party.ChannelID[v] != party.serverChannelID {
			p.WriteInt32(-1)
		} else {
			p.WriteInt32(party.players[v].mapID)
		}

	}

	for i := 0; i < paddAmount; i++ {
		p.WriteInt32(0)
	}

	p.WriteInt32(party.PlayerID[0])

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

func packetPartyCreateUnkownError() mpacket.Packet {
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
