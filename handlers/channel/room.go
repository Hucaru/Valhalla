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
	roomSelectCard            = 60
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

		switch game.RoomType(roomType) {
		case game.RoomTypeOmok:
			name := reader.ReadString(int(reader.ReadInt16()))

			var password string
			if reader.ReadBool() {
				password = reader.ReadString(int(reader.ReadInt16()))
			}

			boardType := reader.ReadByte()

			player.RoomID = game.Rooms.CreateOmokRoom(name, password, boardType)
			game.Rooms[player.RoomID].AddPlayer(conn)
		case game.RoomTypeMemory:
			name := reader.ReadString(int(reader.ReadInt16()))

			var password string
			if reader.ReadBool() {
				password = reader.ReadString(int(reader.ReadInt16()))
			}

			boardType := reader.ReadByte()

			roomID := game.Rooms.CreateMemoryRoom(name, password, boardType)
			game.Rooms[roomID].AddPlayer(conn)
		case game.RoomTypeTrade:
			roomID := game.Rooms.CreateTradeRoom()
			game.Rooms[roomID].AddPlayer(conn)
		case game.RoomTypePersonalShop:
		default:
			fmt.Println("Unkown room type", roomType)
		}
	case roomSendInvite:
		if _, ok := game.Rooms[player.RoomID]; !ok {
			return
		}

		charID := reader.ReadInt32()

		for _, v := range game.Players {
			if v.Char().ID == charID {
				if player.Char().MapID != v.Char().MapID {
					return
				}

				room, ok := game.Rooms[player.RoomID].(*game.TradeRoom)

				if !ok {
					return
				}

				v.Send(packet.RoomInvite(byte(room.RoomType), player.Char().Name, player.RoomID))
			}
		}
	case roomReject:
		roomID := reader.ReadInt32()
		rejectCode := reader.ReadByte()

		if _, ok := game.Rooms[roomID]; !ok {
			return
		}

		game.Rooms[roomID].Broadcast(packet.RoomInviteResult(rejectCode, player.Char().Name))

		room, ok := game.Rooms[player.RoomID].(*game.TradeRoom)

		if !ok {
			return
		}

		room.Close() // Does a reject cause the inviter to leave as well?

		delete(game.Rooms, roomID)
	case roomAccept:
		roomID := reader.ReadInt32()

		if _, ok := game.Rooms[roomID]; !ok {
			return
		}

		room, ok := game.Rooms[roomID].(*game.GameRoom)

		if !ok {
			return
		}

		if reader.ReadBool() {
			password := reader.ReadString(int(reader.ReadInt16()))
			if room.Password != password {
				conn.Send(packet.RoomIncorrectPassword())
				return
			}
		}

		room.AddPlayer(conn)
	case roomChat:
		message := reader.ReadString(int(reader.ReadInt16()))

		if _, ok := game.Rooms[player.RoomID]; !ok {
			return
		}

		game.Rooms[player.RoomID].SendMessage(player.Char().Name, message)
	case roomCloseWindow:
		roomID := player.RoomID
		if game.Rooms[player.RoomID].RemovePlayer(conn, 0) {
			delete(game.Rooms, roomID)
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
		if _, ok := game.Rooms[player.RoomID]; ok {
			game.Rooms[player.RoomID].Broadcast(packet.RoomReady())
		}
	case roomUnready:
		if _, ok := game.Rooms[player.RoomID]; ok {
			game.Rooms[player.RoomID].Broadcast(packet.RoomUnready())
		}
	case roomOwnerExpells:
		if _, ok := game.Rooms[player.RoomID]; !ok {
			return
		}

		room, ok := game.Rooms[player.RoomID].(*game.GameRoom)

		if !ok {
			return
		}

		room.Expel()
	case roomGameStart:
		if _, ok := game.Rooms[player.RoomID]; !ok {
			return
		}

		room, ok := game.Rooms[player.RoomID].(*game.GameRoom)

		if !ok {
			return
		}

		room.Start()
	case roomChangeTurn:
		if _, ok := game.Rooms[player.RoomID]; !ok {
			return
		}

		room, ok := game.Rooms[player.RoomID].(*game.OmokRoom)

		if !ok {
			return
		}

		room.ChangeTurn()
	case roomPlacePiece:
		if _, ok := game.Rooms[player.RoomID]; !ok {
			return
		}

		x := reader.ReadInt32()
		y := reader.ReadInt32()
		piece := reader.ReadByte()

		room, ok := game.Rooms[player.RoomID].(*game.OmokRoom)

		if !ok {
			return
		}

		room.PlacePiece(x, y, piece)
	case roomSelectCard:
		if _, ok := game.Rooms[player.RoomID]; !ok {
			return
		}

		turn := reader.ReadByte()
		cardID := reader.ReadByte()

		room, ok := game.Rooms[player.RoomID].(*game.MemoryRoom)

		if !ok {
			return
		}

		room.SelectCard(turn, cardID, conn)
	default:
		fmt.Println("Unknown room operation", operation)
	}
}
