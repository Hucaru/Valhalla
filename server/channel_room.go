package server

import (
	"log"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/field/room"
)

const (
	roomCreate                = 0
	roomSendInvite            = 2
	roomReject                = 3
	roomAccept                = 4
	roomChat                  = 6
	roomCloseWindow           = 10
	roomInsertItem            = 13
	roomMesos                 = 14
	roomAcceptTrade           = 16
	roomRequestTie            = 42
	roomRequestTieResult      = 43
	roomRequestGiveUp         = 44
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
	player, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[player.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(player.InstanceID())

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

			if r.AddPlayer(player) {
				inst.AddRoom(r)
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

			if r.AddPlayer(player) {
				inst.AddRoom(r)
			}
		case roomTypeTrade:
			r, valid := room.NewTrade(inst.NextID()).(room.Room)

			if !valid {
				return
			}

			if r.AddPlayer(player) {
				inst.AddRoom(r)
			}
		case roomTypePersonalShop:
			log.Println("Personal shop not implemented")
		default:
			log.Println("Unknown room type", roomType)
		}
	case roomSendInvite:
	case roomReject:
	case roomAccept:
	case roomChat:
	case roomCloseWindow:
	case roomInsertItem:
		// invTab := reader.ReadByte()
		// itemSlot := reader.ReadInt16()
		// quantity := reader.ReadInt16()
		// tradeWindowSlot := reader.ReadByte()
	case roomMesos:
		// amount := reader.ReadInt32()
	case roomAcceptTrade:
	case roomRequestTie:
	case roomRequestTieResult:
	case roomRequestGiveUp:
	case roomRequestUndo:
	case roomRequestUndoResult:
	case roomRequestExitDuringGame:
	case roomUndoRequestExit:
	case roomReadyButtonPressed:
	case roomUnready:
	case roomOwnerExpells:
	case roomGameStart:
	case roomChangeTurn:
	case roomPlacePiece:
	case roomSelectCard:
	default:
		log.Println("Unknown room operation", operation)
	}
}
