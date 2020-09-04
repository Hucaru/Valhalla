package server

import (
	"log"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/field/roompool/room"
)

const (
	roomCreate                = 0
	roomSendInvite            = 2
	roomReject                = 3
	roomAccept                = 4
	roomChat                  = 6
	roomCloseWindow           = 10
	roomUnkownOp              = 11
	roomInsertItem            = 13
	roomMesos                 = 14
	roomAcceptTrade           = 16
	roomRequestTie            = 42
	roomRequestTieResult      = 43
	roomForfeit               = 44
	roomRequestUndo           = 46
	roomRequestUndoResult     = 47
	roomRequestExitDuringGame = 48
	roomUndoRequestExit       = 49
	roomReadyButtonPressed    = 50
	roomUnready               = 51
	roomOwnerExpells          = 52
	roomGameStart             = 53
	roomChangeTurn            = 55
	roomPlacePiece            = 56
	roomSelectCard            = 60
)

const (
	roomTypeOmok         = 0x01
	roomTypeMemory       = 0x02
	roomTypeTrade        = 0x03
	roomTypePersonalShop = 0x04
)

func (server ChannelServer) roomWindow(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[plr.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(plr.InstanceID())
	pool := inst.RoomPool()

	if err != nil {
		return
	}

	operation := reader.ReadByte()

	switch operation {
	case roomCreate:
		switch roomType := reader.ReadByte(); roomType {
		case roomTypeOmok:
			name := reader.ReadString(reader.ReadInt16())

			var password string
			if reader.ReadBool() {
				password = reader.ReadString(reader.ReadInt16())
			}

			boardType := reader.ReadByte()

			r, valid := room.NewOmok(inst.NextID(), name, password, boardType).(room.Room)

			if !valid {
				return
			}

			if r.AddPlayer(plr) {
				pool.AddRoom(r)
			}
		case roomTypeMemory:
			name := reader.ReadString(reader.ReadInt16())

			var password string
			if reader.ReadBool() {
				password = reader.ReadString(reader.ReadInt16())
			}

			boardType := reader.ReadByte()

			r, valid := room.NewMemory(inst.NextID(), name, password, boardType).(room.Room)

			if !valid {
				return
			}

			if r.AddPlayer(plr) {
				pool.AddRoom(r)
			}
		case roomTypeTrade:
			r, valid := room.NewTrade(inst.NextID()).(room.Room)

			if !valid {
				return
			}

			if r.AddPlayer(plr) {
				pool.AddRoom(r)
			}
		case roomTypePersonalShop:
			log.Println("Personal shop not implemented")
		default:
			log.Println("Unknown room type", roomType)
		}
	case roomSendInvite:
		id := reader.ReadInt32()

		plr2, err := inst.GetPlayerFromID(id)

		if err != nil {
			plr.Send(room.PacketRoomTradeRequireSameMap())
			return
		}

		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if trade, valid := r.(room.Trade); valid {
			trade.SendInvite(plr2)
		}
	case roomReject:
		id := reader.ReadInt32()
		code := reader.ReadByte()

		r, err := pool.GetRoom(id)

		if err != nil {
			return
		}

		if trade, valid := r.(room.Trade); valid {
			trade.Reject(code, plr.Name())
		}
	case roomAccept:
		id := reader.ReadInt32()

		r, err := pool.GetRoom(id)

		if err != nil {
			plr.Send(room.PacketRoomTradeRequireSameMap())
			return
		}

		if reader.ReadBool() {
			password := reader.ReadString(reader.ReadInt16())

			if game, valid := r.(room.Game); valid {
				if !game.CheckPassword(password, plr) {
					return
				}
			}
		}

		r.AddPlayer(plr)

		if _, valid := r.(room.Game); valid {
			pool.UpdateGameBox(r)
		}
	case roomChat:
		msg := reader.ReadString(reader.ReadInt16())

		if len(msg) > 0 {
			r, err := pool.GetPlayerRoom(plr.ID())

			if err != nil {
				return
			}

			r.ChatMsg(plr, msg)
		}
	case roomCloseWindow:
		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if game, valid := r.(room.Game); valid {
			game.KickPlayer(plr, 0x0)

			if r.Closed() {
				pool.RemoveRoom(r.ID())
			} else {
				pool.UpdateGameBox(r)
			}
		} else if trade, valid := r.(room.Trade); valid {
			trade.RemovePlayer(plr)
			pool.RemoveRoom(r.ID())
		}
	case roomInsertItem:
		// invTab := reader.ReadByte()
		// itemSlot := reader.ReadInt16()
		// quantity := reader.ReadInt16()
		// tradeWindowSlot := reader.ReadByte()
	case roomMesos:
		// amount := reader.ReadInt32()
	case roomAcceptTrade:
	case roomRequestTie:
		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if game, valid := r.(room.Game); valid {
			game.RequestTie(plr)
		}
	case roomRequestTieResult:
		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if game, valid := r.(room.Game); valid {
			tie := reader.ReadBool()
			game.RequestTieResult(tie, plr)

			if r.Closed() {
				pool.RemoveRoom(r.ID())
			} else {
				pool.UpdateGameBox(r)
			}
		}
	case roomForfeit:
		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if game, valid := r.(room.Game); valid {
			game.Forfeit(plr)

			if r.Closed() {
				pool.RemoveRoom(r.ID())
			} else {
				pool.UpdateGameBox(r)
			}
		}
	case roomRequestUndo:
		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if game, valid := r.(room.Omok); valid {
			game.RequestUndo(plr)
		}
	case roomRequestUndoResult:
		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if game, valid := r.(room.Omok); valid {
			undo := reader.ReadBool()
			game.RequestUndoResult(undo, plr)
		}
	case roomRequestExitDuringGame:
		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if game, valid := r.(room.Game); valid {
			game.RequestExit(true, plr)
		}
	case roomUndoRequestExit:
		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if game, valid := r.(room.Game); valid {
			game.RequestExit(false, plr)
		}
	case roomReadyButtonPressed:
		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if game, valid := r.(room.Game); valid {
			game.Ready(plr)
		}
	case roomUnready:
		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if game, valid := r.(room.Game); valid {
			game.Unready(plr)
		}
	case roomOwnerExpells:
		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if game, valid := r.(room.Game); valid {
			game.Expel()
			pool.UpdateGameBox(r)
		}
	case roomGameStart:
		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if game, valid := r.(room.Game); valid {
			game.Start()
			pool.UpdateGameBox(r)
		}
	case roomChangeTurn:
		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if game, valid := r.(room.Game); valid {
			game.ChangeTurn()
		}
	case roomPlacePiece:
		x := reader.ReadInt32()
		y := reader.ReadInt32()
		piece := reader.ReadByte()

		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if game, valid := r.(room.Omok); valid {
			if game.PlacePiece(x, y, piece, plr) {
				pool.UpdateGameBox(r)
			}

			if r.Closed() {
				pool.RemoveRoom(r.ID())
			}
		}
	case roomSelectCard:
		turn := reader.ReadByte()
		cardID := reader.ReadByte()

		r, err := pool.GetPlayerRoom(plr.ID())

		if err != nil {
			return
		}

		if game, valid := r.(room.Memory); valid {
			if game.SelectCard(turn, cardID, plr) {
				pool.UpdateGameBox(r)
			}

			if r.Closed() {
				pool.RemoveRoom(r.ID())
			}
		}
	default:
		log.Println("Unknown room operation", operation)
	}
}
