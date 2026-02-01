package channel

import (
	"math"
	"slices"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mpacket"
)

type shopItem struct {
	item         Item
	price        int32
	bundles      int16
	bundleAmount int16
	reserved     int16
	hidden       bool
}

type shopRoom struct {
	room
	name    string
	private bool
	open    bool
	items   []*shopItem
	mesos   int32
}

func newShopRoom(id int32, name string, private bool) *shopRoom {
	r := room{roomID: id, roomType: constant.MiniRoomTypePlayerShop}
	return &shopRoom{
		room:    r,
		name:    name,
		private: private,
		items:   make([]*shopItem, 0),
		mesos:   0,
	}
}

func (r *shopRoom) addPlayer(plr *Player) bool {
	if !r.open && len(r.players) >= 1 {
		plr.Send(packetRoomStoreMaintenance())
		return false
	}

	if !r.room.addPlayer(plr) {
		return false
	}

	if len(r.players) >= constant.ShopMaxPlayers {
		return false
	}

	plr.Send(packetRoomShowWindow(r.roomType, constant.MiniRoomTypePlayerShop, byte(constant.ShopMaxPlayers), byte(len(r.players)-1), r.name, r.players))

	if len(r.players) > 1 {
		r.sendExcept(packetRoomJoin(r.roomType, byte(len(r.players)-1), r.players[len(r.players)-1]), plr)
		r.send(packetRoomShopRefresh(r))
	}

	return true
}

func (r *shopRoom) removePlayer(plr *Player) {
	for i, v := range r.players {
		sameConn := (v.Conn != nil && plr.Conn != nil && v.Conn == plr.Conn)
		if v.ID == plr.ID || sameConn {
			r.players = append(r.players[:i], r.players[i+1:]...)
			plr.Send(packetRoomLeave(byte(i), constant.MiniRoomLeaveReason))

			if i == constant.RoomOwnerSlot {
				for j := range r.players {
					r.players[j].Send(packetRoomLeave(byte(j+1), constant.MiniRoomClosed))
				}
				r.players = []*Player{}
			} else {
				r.send(packetRoomLeave(byte(i), constant.MiniRoomLeaveReason))
			}
			return
		}
	}
}

func (r *shopRoom) ownerPlayer() *Player {
	for _, plr := range r.players {
		if plr != nil && plr.ID == r.ownerPlayerID {
			return plr
		}
	}
	return nil
}

func (r *shopRoom) closeShop(reason byte) {
	for i, plr := range r.players {
		if plr == nil {
			continue
		}
		plr.Send(packetRoomLeave(byte(i), reason))
	}

	owner := r.ownerPlayer()

	if owner != nil && len(r.items) > 0 {
		for _, si := range r.items {
			if si == nil || si.item.dbID == 0 {
				continue
			}

			if cur, err := owner.getItem(si.item.invID, si.item.slotID); err == nil && cur.dbID == si.item.dbID {
				owner.Send(packetInventoryAddItem(cur, true))
			}
		}
	}

	r.items = make([]*shopItem, 0)
	r.players = []*Player{}
}

func (r *shopRoom) checkOpen() bool {
	for _, plr := range r.players {
		if plr.ID == r.ownerPlayerID {
			if plr.Conn == nil {
				r.closeShop(constant.MiniRoomClosed)
				return false
			}
		}
	}
	return true
}

func (r *shopRoom) reservedForDBID(dbID int64) int16 {
	var total int32
	for _, si := range r.items {
		if si == nil {
			continue
		}
		if si.item.dbID == dbID {
			total += int32(si.reserved)
		}
	}
	if total <= 0 {
		return 0
	}
	if total > int32(math.MaxInt16) {
		return math.MaxInt16
	}
	return int16(total)
}

func (r *shopRoom) refreshOwnerStackVisual(owner *Player, invID byte, slotID int16, dbID int64, forceAdd bool) {
	if owner == nil || dbID == 0 {
		return
	}

	cur, err := owner.getItem(invID, slotID)
	if err != nil || cur.dbID != dbID {
		return
	}

	res := r.reservedForDBID(dbID)
	left := int32(cur.amount) - int32(res)

	if left <= 0 {
		owner.Send(packetInventoryRemoveItem(cur))
		for _, si := range r.items {
			if si != nil && si.item.dbID == dbID {
				si.hidden = true
			}
		}
		return
	}

	vis := cur
	vis.amount = int16(left)

	wasHidden := forceAdd
	if !wasHidden {
		for _, si := range r.items {
			if si != nil && si.item.dbID == dbID && si.hidden {
				wasHidden = true
				break
			}
		}
	}

	if wasHidden {
		owner.Send(packetInventoryAddItem(vis, true))
	} else {
		owner.Send(packetInventoryModifyItemAmount(vis))
	}

	for _, si := range r.items {
		if si != nil && si.item.dbID == dbID {
			si.hidden = false
		}
	}
}

func (r *shopRoom) addItem(item Item, bundles, bundleAmount int16, price int32) bool {
	owner := r.ownerPlayer()
	if owner == nil {
		return false
	}

	if bundles <= 0 || bundleAmount <= 0 {
		return false
	}

	cur, err := owner.getItem(item.invID, item.slotID)
	if err != nil || cur.dbID != item.dbID || cur.ID != item.ID {
		return false
	}

	if cur.isRechargeable() {
		bundles = 1
		bundleAmount = cur.amount
	}

	need := bundles * bundleAmount
	if need <= 0 || need > cur.amount {
		return false
	}

	if !cur.isRechargeable() && cur.isStackable() {
		for _, si := range r.items {
			if si == nil {
				continue
			}
			if si.item.dbID != cur.dbID || si.item.ID != cur.ID {
				continue
			}
			if si.price != price || si.bundleAmount != bundleAmount {
				continue
			}

			alreadyReserved := r.reservedForDBID(cur.dbID)
			available := int32(cur.amount) - int32(alreadyReserved)
			if int32(need) > available {
				return false
			}

			si.reserved += need
			si.bundles += bundles

			r.refreshOwnerStackVisual(owner, cur.invID, cur.slotID, cur.dbID, false)
			return true
		}
	}

	si := &shopItem{
		item:         cur,
		price:        price,
		bundles:      bundles,
		bundleAmount: bundleAmount,
		reserved:     need,
		hidden:       false,
	}

	totalReserved := r.reservedForDBID(cur.dbID) + need
	if totalReserved > cur.amount {
		return false
	}

	r.items = append(r.items, si)

	r.refreshOwnerStackVisual(owner, cur.invID, cur.slotID, cur.dbID, false)
	return true
}

func (r *shopRoom) buyItem(slot byte, quantity int16, buyerID int32) (byte, bool) {
	if int(slot) < 0 || int(slot) >= len(r.items) {
		return constant.PlayerShopNotEnoughInStock, false
	}

	si := r.items[slot]
	if si == nil {
		return constant.PlayerShopNotEnoughInStock, false
	}

	if quantity <= 0 || si.bundles < quantity {
		return constant.PlayerShopNotEnoughInStock, false
	}

	totalCost := int64(si.price) * int64(quantity)
	if totalCost > int64(math.MaxInt32) {
		return constant.PlayerShopPriceTooHighForTrade, false
	}

	realAmount := quantity * si.bundleAmount
	if realAmount <= 0 || realAmount > si.reserved {
		return constant.PlayerShopNotEnoughInStock, false
	}

	var buyer *Player
	var owner *Player
	for _, plr := range r.players {
		if plr.ID == buyerID {
			buyer = plr
		}
		if plr.ID == r.ownerID() {
			owner = plr
		}
	}

	if buyer == nil {
		return constant.PlayerShopInventoryFull, false
	}
	if owner == nil {
		return constant.PlayerShopNotEnoughInStock, false
	}
	if int64(buyer.mesos) < totalCost {
		return constant.PlayerShopBuyerNotEnoughMoney, false
	}

	cur, err := owner.getItem(si.item.invID, si.item.slotID)
	if err != nil || cur.dbID != si.item.dbID || cur.ID != si.item.ID {
		return constant.PlayerShopNotEnoughInStock, false
	}

	totalReserved := r.reservedForDBID(cur.dbID)
	if totalReserved > cur.amount {
		return constant.PlayerShopNotEnoughInStock, false
	}

	if cur.isRechargeable() {
		if quantity != 1 || si.bundles != 1 {
			return constant.PlayerShopNotEnoughInStock, false
		}
		realAmount = cur.amount
		if realAmount <= 0 || realAmount != si.reserved {
			return constant.PlayerShopNotEnoughInStock, false
		}
	}

	purchased := cur
	purchased.amount = realAmount
	purchased.dbID = 0
	purchased.slotID = 0

	if _, err := buyer.GiveItem(purchased); err != nil {
		return constant.PlayerShopInventoryFull, false
	}

	if _, err := owner.TakeItemSilent(cur.ID, cur.slotID, realAmount, cur.invID); err != nil {
		return constant.PlayerShopNotEnoughInStock, false
	}

	buyer.takeMesos(int32(totalCost))
	owner.giveMesos(int32(totalCost))

	si.reserved -= realAmount
	si.bundles -= quantity

	if si.bundles <= 0 || si.reserved <= 0 {
		i := int(slot)
		r.items = slices.Delete(r.items, i, i+1)
	}

	r.refreshOwnerStackVisual(owner, cur.invID, cur.slotID, cur.dbID, false)
	return 0, true
}

func (r *shopRoom) removeItem(slot byte) bool {
	owner := r.ownerPlayer()
	if owner == nil {
		return false
	}

	if int(slot) < 0 || int(slot) >= len(r.items) {
		return false
	}

	si := r.items[slot]
	if si == nil {
		return false
	}

	dbID := si.item.dbID
	invID := si.item.invID
	invSlot := si.item.slotID
	forceAdd := si.hidden

	i := int(slot)
	r.items = slices.Delete(r.items, i, i+1)

	r.refreshOwnerStackVisual(owner, invID, invSlot, dbID, forceAdd)
	return true
}

func (r *shopRoom) displayBytes() []byte {
	p := mpacket.NewPacket()

	if len(r.players) == 0 {
		return p
	}

	p.WriteInt32(r.players[0].ID)
	p.WriteByte(r.roomType)
	p.WriteInt32(r.roomID)
	p.WriteString(r.name)
	p.WriteBool(r.private)
	p.WriteByte(0)
	p.WriteByte(byte(len(r.players)))
	p.WriteByte(constant.ShopMaxPlayers)
	p.WriteBool(r.open)

	return p
}

func packetRoomShopRefresh(shop *shopRoom) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomShopRefresh)
	p.WriteByte(byte(len(shop.items)))

	for _, shopItem := range shop.items {
		p.WriteInt16(shopItem.bundles)
		p.WriteInt16(shopItem.bundleAmount)
		p.WriteInt32(shopItem.price)
		p.Append(shopItem.item.StorageBytes())
	}

	return p
}

func packetShopItemResult(msg byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.MiniRoomPlayerShopItemResult)
	p.WriteByte(msg)

	return p
}

func packetRoomShopSoldItem(slot byte, quantity int16, buyerName string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.MiniRoomPlayerShopSoldItem)
	p.WriteByte(slot)
	p.WriteInt16(quantity)
	p.WriteString(buyerName)
	return p
}

func packetRoomShopRemoveItem(remaining byte, slot int16) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomShopMoveItemToInv)
	p.WriteByte(remaining)
	p.WriteInt16(slot)
	return p
}

func packetRoomStoreMaintenance() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterStoreMaint)
}
