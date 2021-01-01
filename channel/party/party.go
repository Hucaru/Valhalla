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
	GiveEXP(int32, bool, bool)
}

// Data containing the party information
type Data struct {
	id              int32
	serverChannelID int32

	players   [constant.MaxPartySize]player
	channelID [constant.MaxPartySize]int32
	playerID  [constant.MaxPartySize]int32
	name      [constant.MaxPartySize]string
	mapID     [constant.MaxPartySize]int32
	job       [constant.MaxPartySize]int32
	level     [constant.MaxPartySize]int32
}

// NewParty with a leader
func NewParty(id int32, plr player, channelID byte, playerID, mapID, job, level int32, playerName string, serverChannelID int32) Data {
	result := Data{id: id, serverChannelID: serverChannelID}
	result.players[0] = plr

	result.channelID[0] = int32(channelID)
	result.playerID[0] = playerID
	result.name[0] = playerName
	result.mapID[0] = mapID
	result.job[0] = job
	result.level[0] = level

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
func (d *Data) AddPlayer(plr player, channelID int32, id int32, name string, mapID, job, level int32) {
	for i, v := range d.playerID {
		if v == 0 {
			d.players[i] = plr

			if plr != nil {
				plr.SetParty(d)
			}

			d.channelID[i] = channelID
			d.playerID[i] = id
			d.name[i] = name
			d.mapID[i] = mapID
			d.job[i] = job
			d.level[i] = level

			d.Broadcast(packetPlayerJoin(d.id, name, d))

			return
		}
	}
}

// RemovePlayer from party
func (d *Data) RemovePlayer(plr player, playerID int32, kick bool) {
	for i, v := range d.playerID {
		if v == playerID {
			if i == 0 {
				d.Broadcast(packetLeaveParty(d.id, playerID, false, kick, "", d))

				for _, p := range d.players {
					if p != nil {
						p.SetParty(nil)
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
				plr.SetParty(nil)
			}

			d.Broadcast(packetLeaveParty(d.id, playerID, true, kick, name, d))

			d.players[i] = nil

			return
		}
	}
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

// UpdatePlayerMap for party window
func (d *Data) UpdatePlayerMap(playerID, mapID int32) {
	for i, v := range d.playerID {
		if v == playerID {
			d.mapID[i] = mapID

			d.Broadcast(packetUpdateParty(d.id, d))

			return
		}
	}
}

// SetPlayerChannel updates the party on what server the player is on or if onffline
func (d *Data) SetPlayerChannel(plr player, playerID int32, cashshop bool, offline bool, channelID int32) {
	for i, v := range d.playerID {
		if v == playerID {
			if cashshop {
				d.channelID[i] = -1
			} else if offline {
				d.channelID[i] = -2
			} else {
				d.channelID[i] = channelID
			}

			d.players[i] = plr

			d.Broadcast(packetUpdateParty(d.id, d))
			return
		}
	}
}

// UpdateJobLevel for the party window
func (d *Data) UpdateJobLevel(playerID, job, level int32) {
	for i, v := range d.playerID {
		if v == playerID {
			d.job[i] = job
			d.level[i] = level

			d.Broadcast(packetUpdateJobLevel(playerID, job, level))
			return
		}
	}
}

// Leader checks if the provided id is the party leader
func (d Data) Leader(id int32) bool {
	return d.playerID[0] == id
}

// Member checks if id is a party member
func (d Data) Member(id int32) bool {
	for _, v := range d.playerID {
		if v == id {
			return true
		}
	}

	return false
}

// GiveExp to party members
func (d Data) GiveExp(playerID, amount int32, sameMap bool) {
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
			if id != playerID && d.players[i] != nil && d.mapID[i] == mapID {
				nPlayers++
			}

			if nPlayers == 0 {
				return
			}
		}
	}

	for _, plr := range d.players {
		if plr != nil && sameMap && plr.MapID() == mapID {
			plr.GiveEXP(amount, false, true)
		}
	}
}
