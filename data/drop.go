package data

type mapleDrop struct {
	// character.Item
	dropID uint32
}

func (d *mapleDrop) SetDropID(dropID uint32) {
	d.dropID = dropID
}

func (d *mapleDrop) GetDropID() uint32 {
	return d.dropID
}
