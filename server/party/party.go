package party

import (
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mpacket"
)

type player interface {
	Send(mpacket.Packet)
	ID() int32
	Name() string
	MapID() int32
	Job() int16
	Level() byte
	HP() int16
	MaxHP() int16
	SetParty(*Data)
}

// Data containing the party information
type Data struct {
	id int32

	players   [constant.MaxPartySize]player
	channelID [constant.MaxPartySize]int32
	playerID  [constant.MaxPartySize]int32
	name      [constant.MaxPartySize]string
	mapID     [constant.MaxPartySize]int32
	job       [constant.MaxPartySize]int32
	level     [constant.MaxPartySize]int32
	hp        [constant.MaxPartySize]int16
	maxHP     [constant.MaxPartySize]int16
}

// NewParty with a leader
func NewParty(id int32, plr player, channelID byte) Data {
	result := Data{id: id}
	result.players[0] = plr

	result.channelID[0] = int32(channelID)
	result.playerID[0] = plr.ID()
	result.name[0] = plr.Name()
	result.mapID[0] = plr.MapID()
	result.job[0] = int32(plr.Job())
	result.level[0] = int32(plr.Level())
	result.hp[0] = plr.HP()
	result.maxHP[0] = plr.MaxHP()

	return result
}

// ID of party
func (d Data) ID() int32 {
	return d.id
}

// Broadcast to all party members
func (d Data) Broadcast(p mpacket.Packet) {
	for _, v := range d.players {
		if v == nil {
			continue
		}

		v.Send(p)
	}
}

// BroadcastExcept to all party members
func (d Data) BroadcastExcept(p mpacket.Packet, id int32) {
	for i, v := range d.players {
		if v == nil || d.playerID[i] == id {
			continue
		}

		v.Send(p)
	}
}

// AddPlayer to party
func (d *Data) AddPlayer(plr player, channelID int32, id int32, name string, mapID, job, level int32, hp, maxHP int16) {
	for i, v := range d.playerID {
		if v == 0 {
			d.channelID[i] = channelID
			d.playerID[i] = id
			d.name[i] = name
			d.mapID[i] = mapID
			d.job[i] = job
			d.level[i] = level
			d.players[i] = plr
			d.hp[i] = hp
			d.maxHP[i] = maxHP

			d.Broadcast(packetPlayerJoin(d.id, name, d))

			return
		}
	}
}

// RemovePlayer from party
func (d *Data) RemovePlayer(plr player) bool {
	for i, v := range d.players {
		if v == plr {
			d.players[i] = nil

			if i == 0 {
				return true
			}
			return false
		}
	}

	return false
}

// Full returns false if there is a space in the party
func (d Data) Full() bool {
	for _, v := range d.players {
		if v == nil {
			return false
		}
	}

	return true
}

// PlayerLoggedIn will assign the player handle to this channel's party list
func (d *Data) PlayerLoggedIn(plr player, channelID int32) {
	for i, v := range d.playerID {
		if v == plr.ID() {
			d.players[i] = plr
			d.channelID[i] = channelID
			d.mapID[i] = plr.MapID()
			d.job[i] = int32(plr.Job())
			d.level[i] = int32(plr.Level())
			d.name[i] = plr.Name()
			d.hp[i] = plr.HP()
			d.maxHP[i] = plr.MaxHP()

			plr.SetParty(d)

			d.Broadcast(packetPlayerJoin(d.id, plr.Name(), d))

			return
		}
	}
}

// PartyUpdatePlayer information
func (d *Data) PartyUpdatePlayer(channelID, playerID, mapID, job, level int32, name string, hp, maxHP int16) {
	for i, v := range d.playerID {
		if v == playerID {
			d.channelID[i] = channelID
			d.mapID[i] = mapID
			d.job[i] = job
			d.level[i] = level
			d.name[i] = name
			d.hp[i] = hp
			d.maxHP[i] = maxHP

			d.Broadcast(packetPlayerJoin(d.id, name, d))

			return
		}
	}
}
