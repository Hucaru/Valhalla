package droppool

import (
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/pos"
)

type field interface {
	Send(mpacket.Packet) error
}

// Data structure for the pool
type Data struct {
	instance field
	drops    []drop
}

// CreateNewPool for drops
func CreateNewPool(inst field) Data {
	return Data{instance: inst}
}

// CreateMobDrop from a mobID from a player at a given location
func (data *Data) CreateMobDrop(mesos bool, itemID int32, dropFrom pos.Data, finalPos pos.Data) {

}
