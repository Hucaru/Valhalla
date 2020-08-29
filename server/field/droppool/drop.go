package droppool

import (
	"github.com/Hucaru/Valhalla/server/item"
	"github.com/Hucaru/Valhalla/server/pos"
)

const (
	DropTimeoutNonOwner      = 0
	DropTimeoutNonOwnerParty = 1
	DropFreeForAll           = 2
	DropExplosiveFreeForAll  = 3
)

type drop struct {
	ID      int32
	ownerID int32
	partyID int32

	mesos int32
	item  item.Data

	expireTime  int64
	timeoutTime int64
	neverExpire bool

	originPos pos.Data
	finalPos  pos.Data

	dropType byte
}
