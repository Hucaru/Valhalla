package party

import "github.com/Hucaru/Valhalla/mpacket"

type player interface {
	Send(mpacket.Packet)
	Name() string
}

// Data containing the party information
type Data struct {
	id        int32
	players   [4]player
	names     [4]string
	channelID [4]byte
}

// NewParty with a leader
func NewParty(id int32, plr player, channelID byte) Data {
	result := Data{id: id}
	result.players[0] = plr
	result.names[0] = plr.Name()
	result.channelID[0] = channelID
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
