package channel

import (
	"fmt"

	"github.com/Hucaru/Valhalla/game/packet"

	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
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
	roomReadyButtonPressed    = 50
	roomUnready               = 51
	roomOwnerExpells          = 52
	roomGameStart             = 53
	roomChangeTurn            = 55
	roomPlacePiece            = 56
)

func handleUIWindow(conn mnet.MConnChannel, reader mpacket.Reader) {
	operation := reader.ReadByte()

	player := game.Players[conn]

	switch operation {
	case roomCreate:
		if player.RoomID > 0 {
			return
		}

		roomType := reader.ReadByte()

		switch roomType {
		case game.OmokRoom:
			name := reader.ReadString(int(reader.ReadInt16()))

			var password string
			if reader.ReadBool() {
				password = reader.ReadString(int(reader.ReadInt16()))
			}

			boardType := reader.ReadByte()

			player.RoomID = game.Rooms.CreateOmokRoom(name, password, boardType)
			game.Rooms[player.RoomID].AddPlayer(conn)
		case game.MemoryRoom:
			name := reader.ReadString(int(reader.ReadInt16()))

			var password string
			if reader.ReadBool() {
				password = reader.ReadString(int(reader.ReadInt16()))
			}

			boardType := reader.ReadByte()

			roomID := game.Rooms.CreateMemoryRoom(name, password, boardType)
			game.Rooms[roomID].AddPlayer(conn)
		case game.TradeRoom:
		case game.PersonalShop:
		default:
			fmt.Println("Unkown room type", roomType)
		}
	case roomSendInvite:
	case roomReject:
	case roomAccept:
		roomID := reader.ReadInt32()

		if _, ok := game.Rooms[roomID]; !ok {
			return
		}

		if reader.ReadBool() {
			password := reader.ReadString(int(reader.ReadInt16()))
			if game.Rooms[roomID].Password != password {
				conn.Send(packet.RoomIncorrectPassword())
				return
			}
		}

		game.Rooms[roomID].AddPlayer(conn)
	case roomChat:
		message := reader.ReadString(int(reader.ReadInt16()))

		if _, ok := game.Rooms[player.RoomID]; !ok {
			return
		}

		game.Rooms[player.RoomID].SendMessage(player.Char().Name, message)
	case roomCloseWindow:
		if game.Rooms[player.RoomID].RemovePlayer(conn, 0) {
			delete(game.Rooms, player.RoomID)
		}
	case roomInsertItem:
	case roomMesos:
	case roomAcceptTrade:
	case roomRequestTie:
	case roomRequestTieResult:
	case roomRequestGiveUp:
	case roomRequestUndo:
	case roomRequestUndoResult:
	case roomRequestExitDuringGame:
	case roomReadyButtonPressed:
	case roomUnready:
	case roomOwnerExpells:
	case roomGameStart:
	case roomChangeTurn:
	case roomPlacePiece:
	default:
		fmt.Println("Unknown room operation", operation)
	}
}
