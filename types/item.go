package types

import "github.com/google/uuid"

type Item struct {
	UUID         uuid.UUID
	InvID        byte
	SlotID       int16
	ItemID       int32
	ExpireTime   uint64
	Amount       int16
	CreatorName  string
	Flag         int16
	UpgradeSlots byte
	ReqLevel     byte
	ScrollLevel  byte
	Str          int16
	Dex          int16
	Int          int16
	Luk          int16
	ReqStr       int16
	ReqDex       int16
	ReqInt       int16
	ReqLuk       int16
	HP           int16
	MP           int16
	Watk         int16
	Matk         int16
	Wdef         int16
	Mdef         int16
	Accuracy     int16
	Avoid        int16
	Hands        int16
	Speed        int16
	Jump         int16
}
