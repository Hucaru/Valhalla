package data

import "github.com/Hucaru/Valhalla/inventory"

type mapleDrop struct {
	inventory.Item
	dropID uint32
}

func (d *mapleDrop) SetDropID(dropID uint32) {
	d.dropID = dropID
}

func (d *mapleDrop) GetDropID() uint32 {
	return d.dropID
}
