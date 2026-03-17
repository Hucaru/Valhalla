package channel

import (
	"log"
	"math"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mpacket"
)

type tradeRoom struct {
	room
	mesos     map[int32]int32
	items     map[int32]map[byte]Item
	confirmed map[int32]bool

	finalized bool
}

func newTradeRoom(id int32) *tradeRoom {
	r := room{roomID: id, roomType: constant.MiniRoomTypeTrade}
	return &tradeRoom{room: r, mesos: make(map[int32]int32), items: make(map[int32]map[byte]Item), confirmed: make(map[int32]bool), finalized: false}
}

func (r *tradeRoom) addPlayer(plr *Player) bool {
	if !r.room.addPlayer(plr) {
		return false
	}

	plr.Send(packetRoomShowWindow(r.roomType, constant.MiniRoomTypeTrade, byte(constant.RoomMaxPlayers), byte(len(r.players)-1), "", r.players))

	if len(r.players) > 1 {
		r.sendExcept(packetRoomJoin(r.roomType, byte(len(r.players)-1), r.players[len(r.players)-1]), plr)
	}

	r.items[plr.ID] = make(map[byte]Item)
	r.confirmed[plr.ID] = false

	return true
}

func (r *tradeRoom) removePlayer(plr *Player) {
	r.closeWithReason(constant.RoomLeaveTradeCancelled, true)
}

func (r tradeRoom) sendInvite(plr *Player) {
	plr.Send(packetRoomInvite(constant.MiniRoomTypeTrade, r.players[0].Name, r.roomID))
}

func (r tradeRoom) reject(code byte, name string) {
	r.send(packetRoomInviteResult(code, name))
}

func (r *tradeRoom) canInsertItem(plrID int32, tradeSlot byte) bool {
	if tradeSlot < 1 || tradeSlot > 9 {
		return false
	}
	if r.items == nil {
		return false
	}
	if _, ok := r.items[plrID]; !ok || r.items[plrID] == nil {
		return false
	}
	if _, exists := r.items[plrID][tradeSlot]; exists {
		return false
	}
	return true
}

func (r *tradeRoom) insertItem(tradeSlot byte, plrID int32, item Item) bool {
	if tradeSlot < 1 || tradeSlot > 9 {
		log.Printf("trade: invalid slot %d from player %d\n", tradeSlot, plrID)
		return false
	}

	if r.items == nil || r.items[plrID] == nil {
		log.Printf("trade: missing items map for player %d\n", plrID)
		return false
	}

	if _, exists := r.items[plrID][tradeSlot]; exists {
		log.Printf("trade: slot %d already occupied for player %d\n", tradeSlot, plrID)
		return false
	}

	r.items[plrID][tradeSlot] = item
	isUser0 := r.players[0].ID == plrID
	r.players[0].Send(packetRoomTradePutItem(tradeSlot, !isUser0, item))
	r.players[1].Send(packetRoomTradePutItem(tradeSlot, isUser0, item))
	return true
}

func (r *tradeRoom) updateMesos(amount, plrID int32) {
	r.mesos[plrID] += amount
	isUser0 := r.players[0].ID == plrID
	r.players[0].Send(packetRoomTradePutMesos(r.mesos[plrID], !isUser0))
	r.players[1].Send(packetRoomTradePutMesos(r.mesos[plrID], isUser0))
}

func (r *tradeRoom) acceptTrade(plr *Player) bool {
	r.confirmed[plr.ID] = true

	for _, user := range r.players {
		if user.ID != plr.ID {
			user.Send(packetRoomTradeAccept())
		}
	}

	if r.confirmed[r.players[0].ID] && r.confirmed[r.players[1].ID] {
		r.completeTrade()
	}

	return r.finalized
}

func (r *tradeRoom) completeTrade() {
	if len(r.players) < 2 || r.players[0] == nil || r.players[1] == nil {
		r.closeWithReason(constant.MiniRoomTradeFail, true)
		return
	}

	p1 := r.players[0]
	p2 := r.players[1]

	type tradeChange struct {
		plr         *Player
		mesosChange int32
		itemsToGive []Item
	}

	changes := []tradeChange{
		{plr: p1, mesosChange: r.mesos[p2.ID]},
		{plr: p2, mesosChange: r.mesos[p1.ID]},
	}

	for _, item := range r.items[p1.ID] {
		changes[1].itemsToGive = append(changes[1].itemsToGive, item)
	}

	for _, item := range r.items[p2.ID] {
		changes[0].itemsToGive = append(changes[0].itemsToGive, item)
	}

	if int64(p1.mesos)+int64(changes[0].mesosChange) > int64(math.MaxInt32) ||
		int64(p2.mesos)+int64(changes[1].mesosChange) > int64(math.MaxInt32) ||
		changes[0].mesosChange < 0 || changes[1].mesosChange < 0 {
		r.closeWithReason(constant.MiniRoomTradeFail, true)
		return
	}

	var undo []func()

	for _, change := range changes {
		for _, item := range change.itemsToGive {
			insertedItem, err := change.plr.GiveItem(item)
			if err != nil {
				for _, fn := range undo {
					fn()
				}

				log.Printf("Trade error: failed to give item %v to %s: %v", item.ID, change.plr.Name, err)
				r.closeWithReason(constant.MiniRoomTradeInventoryFull, true)
				return
			}

			undo = append(undo, func() {
				if _, err := change.plr.takeItem(insertedItem.ID, insertedItem.slotID, insertedItem.amount, insertedItem.invID); err != nil {
					log.Printf("Trade rollback warning: failed to remove item %v from %s: %v", insertedItem.ID, change.plr.Name, err)
				}
			})
		}
	}

	applyMesosTax := func(mesos int32) int32 {
		remainingMesos := 1.0

		if mesos < 50000 {
			remainingMesos = 1
		} else if mesos < 100000 {
			remainingMesos = 0.995
		} else if mesos < 1000000 {
			remainingMesos = 0.99
		} else if mesos < 5000000 {
			remainingMesos = 0.98
		} else if mesos < 10000000 {
			remainingMesos = 0.97
		} else {
			remainingMesos = 0.96
		}

		return int32(float64(mesos) * remainingMesos)
	}

	for _, change := range changes {
		if change.mesosChange > 0 {
			mesosChange := applyMesosTax(change.mesosChange)
			change.plr.giveMesos(mesosChange)
		}
	}

	r.finalized = true
	r.closeWithReason(constant.MiniRoomTradeSuccess, false)
}

func (r *tradeRoom) closeWithReason(reason byte, rollback bool) {
	if rollback {
		r.rollback()
	}
	for i, plr := range r.players {
		if plr != nil {
			plr.Send(packetRoomLeave(byte(i), reason))
		}
	}
}

func (r *tradeRoom) rollback() {
	if r.finalized {
		return
	}

	for _, player := range r.players {
		if player == nil {
			continue
		}
		if m := r.mesos[player.ID]; m != 0 {
			player.giveMesos(m)
			r.mesos[player.ID] = 0
		}
		if bag, ok := r.items[player.ID]; ok {
			for slot, item := range bag {
				if _, err := player.GiveItem(item); err != nil {
					log.Println("tradeRoom rollback failed:", err)
				}
				delete(bag, slot)
			}
		}
	}

	r.finalized = true
}

func packetRoomTradePutItem(tradeSlot byte, user bool, item Item) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.MiniRoomTradePutItem)
	p.WriteBool(user)
	p.WriteByte(tradeSlot)
	p.WriteBytes(item.StorageBytes())
	return p
}

func packetRoomTradePutMesos(amount int32, user bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.MiniRoomTradePutMesos)
	p.WriteBool(user)
	p.WriteInt32(amount)
	return p
}

func packetRoomTradeAccept() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.MiniRoomTradeAccept)

	return p
}

func packetRoomTradeRequireSameMap() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterTradeSameMap)
}
