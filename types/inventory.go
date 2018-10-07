package types

type Inventory struct {
	EquipSlotSize byte
	UseSlotSize   byte
	SetupSlotSize byte
	EtcSlotSize   byte
	CashSlotSize  byte

	Equip []Item
	Use   []Item
	SetUp []Item
	Etc   []Item
	Cash  []Item
}
