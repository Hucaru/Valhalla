package character

type Item struct {
	invID       byte
	slotNumber  byte
	itemID      uint32
	expiration  uint64
	amount      uint16
	creatorName string
	flag        uint16
}

func (i *Item) GetInvID() byte {
	return i.invID
}

func (i *Item) SetInvID(val byte) {
	i.invID = val
}

func (i *Item) GetSlotNumber() byte {
	return i.slotNumber
}

func (i *Item) SetSlotNumber(val byte) {
	i.slotNumber = val
}

func (i *Item) GetItemID() uint32 {
	return i.itemID
}

func (i *Item) SetItemID(val uint32) {
	i.itemID = val
}

func (i *Item) GetExpiration() uint64 {
	return i.expiration
}

func (i *Item) SetExpiration(val uint64) {
	i.expiration = val
}

func (i *Item) GetAmount() uint16 {
	return i.amount
}

func (i *Item) SetAmount(val uint16) {
	i.amount = val
}

func (i *Item) GetCreatorName() string {
	return i.creatorName
}

func (i *Item) SetCreatorName(val string) {
	i.creatorName = val
}

func (i *Item) GetFlag() uint16 {
	return i.flag
}

func (i *Item) SetFlag(val uint16) {
	i.flag = val
}
