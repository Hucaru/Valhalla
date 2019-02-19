package game

import (
	"fmt"
	"math/rand"

	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/game/packet"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type GameRoomAsserter interface {
	Expel()
	Start()
	GetPassword() string
	SendOpponent(conn mnet.MConnChannel, p mpacket.Packet)
	Tie() bool
	GiveUp(conn mnet.MConnChannel) bool
	LeaveAfterGame(conn mnet.MConnChannel)
	UndoLeaveAfterGame(conn mnet.MConnChannel)
	ChangeTurn()
}

type GameRoom struct {
	*baseRoom
	Name, Password string
	BoardType      byte
	leaveAfterGame [2]bool
	p1Turn         bool
	InProgress     bool
}

type OmokRoom struct {
	*GameRoom
	board        [15][15]byte
	previousTurn [2][2]int32
}

type MemoryRoom struct {
	*GameRoom
	cards         []byte
	firstCardPick byte
	matches       [2]int
}

func (rc *roomContainer) CreateMemoryRoom(name, password string, boardType byte) int32 {
	id := rc.getNewRoomID()

	r := &MemoryRoom{}
	r.GameRoom = &GameRoom{Name: name, Password: password, BoardType: boardType, p1Turn: true}
	r.GameRoom.baseRoom = &baseRoom{ID: id, RoomType: RoomTypeMemory, maxPlayers: memoryMaxPlayers}

	Rooms[id] = r

	return id
}

func (rc *roomContainer) CreateOmokRoom(name, password string, boardType byte) int32 {
	id := rc.getNewRoomID()

	r := &OmokRoom{}
	r.GameRoom = &GameRoom{Name: name, Password: password, BoardType: boardType, p1Turn: true}
	r.GameRoom.baseRoom = &baseRoom{ID: id, RoomType: RoomTypeOmok, maxPlayers: omokMaxPlayers}

	Rooms[id] = r
	return id
}

func (r *GameRoom) GetPassword() string {
	return r.Password
}

func (r *GameRoom) IsOwner(conn mnet.MConnChannel) bool {
	if len(r.players) > 0 && r.players[0] != nil && r.players[0] == conn {
		return true
	}

	return false
}

func (r *GameRoom) BroadcastExcept(p mpacket.Packet, conn mnet.MConnChannel) {
	for _, v := range r.players {
		if v == conn {
			continue
		}

		v.Send(p)
	}
}

func (r *GameRoom) AddPlayer(conn mnet.MConnChannel) {
	if roomPos, ok := r.baseRoom.AddPlayer(conn); ok {
		player := Players[conn]
		if roomPos == 0 {
			Maps[player.Char().MapID].Send(packet.MapShowGameBox(player.Char().ID, r.ID, byte(r.RoomType), r.BoardType, r.Name, bool(len(r.Password) > 0), r.InProgress, 0x01), player.InstanceID)
		}

		displayInfo := []def.Character{}

		for _, v := range r.players {
			displayInfo = append(displayInfo, Players[v].Char())
		}

		if len(displayInfo) > 0 {
			conn.Send(packet.RoomShowWindow(byte(r.RoomType), r.BoardType, byte(r.maxPlayers), roomPos, r.Name, displayInfo))
		}
	}
}

func (r *GameRoom) RemovePlayer(conn mnet.MConnChannel, msgCode byte) bool {
	var closeRoom bool

	if roomSlot := r.baseRoom.RemovePlayer(conn); roomSlot > -1 {
		player := Players[conn]

		if roomSlot == 0 {
			Maps[player.Char().MapID].Send(packet.MapRemoveGameBox(player.Char().ID), player.InstanceID)

			closeRoom = true

			for i, v := range r.players {
				v.Send(packet.RoomLeave(byte(i), 0))
				Players[v].RoomID = 0
			}
		} else {
			conn.Send(packet.RoomLeave(byte(roomSlot), msgCode))
			r.Broadcast(packet.RoomLeave(byte(roomSlot), msgCode))

			if msgCode == 5 {

				r.Broadcast(packet.RoomYellowChat(0, player.Char().Name))
			}

			r.players = append(r.players[:roomSlot], r.players[roomSlot+1:]...)

			for i := roomSlot; i < len(r.players)-1; i++ {
				// Update player positions from index roomSlot onwards (not + 1 as we have removed the gone player)
			}
		}

		return closeRoom
	}

	return false
}

func (r *GameRoom) Start() {
	if len(r.players) == 0 {
		return
	}

	r.InProgress = true

	player := Players[r.players[0]]
	Maps[player.Char().MapID].Send(packet.MapShowGameBox(player.Char().ID, r.ID, byte(r.RoomType), r.BoardType, r.Name, bool(len(r.Password) > 0), r.InProgress, 0x01), player.InstanceID)
}

func (r *GameRoom) Expel() {
	if len(r.players) > 1 {
		r.RemovePlayer(r.players[1], 5)
	}
}

func (r *GameRoom) gameEnd(draw, forfeit bool, conn mnet.MConnChannel) bool {
	// Update box on map
	r.InProgress = false
	player := Players[r.players[0]]
	Maps[player.Char().MapID].Send(packet.MapShowGameBox(player.Char().ID, r.ID, byte(r.RoomType), r.BoardType, r.Name, bool(len(r.Password) > 0), r.InProgress, 0x01), player.InstanceID)

	var slotID byte = 0x00
	if !r.p1Turn {
		slotID = 1
	}

	if forfeit {
		// this button can be pressed at anytime, therefore cannot rely on player turn
		for i, v := range r.players[0:2] {
			if v == conn {
				slotID = byte(i)
			}
		}

		if slotID == 0 { // for forfeits slot id is inversed
			Players[r.players[0]].SetMinigameLoss(Players[r.players[0]].Char().MiniGameLoss + 1)
			Players[r.players[1]].SetMinigameWins(Players[r.players[1]].Char().MiniGameWins + 1)
			slotID = 0x1
		} else {
			Players[r.players[1]].SetMinigameLoss(Players[r.players[1]].Char().MiniGameLoss + 1)
			Players[r.players[0]].SetMinigameWins(Players[r.players[0]].Char().MiniGameWins + 1)
			slotID = 0x0
		}

	} else if draw {
		Players[r.players[0]].SetMinigameDraw(Players[r.players[0]].Char().MiniGameDraw + 1)
		Players[r.players[1]].SetMinigameDraw(Players[r.players[1]].Char().MiniGameDraw + 1)
	} else {
		Players[r.players[slotID]].SetMinigameWins(Players[r.players[slotID]].Char().MiniGameWins + 1)

		if slotID == 1 {
			Players[r.players[0]].SetMinigameLoss(Players[r.players[0]].Char().MiniGameLoss + 1)
		} else {
			Players[r.players[1]].SetMinigameLoss(Players[r.players[1]].Char().MiniGameLoss + 1)
		}

	}

	displayInfo := []def.Character{}

	for _, v := range r.players {
		displayInfo = append(displayInfo, Players[v].Char())
	}

	r.Broadcast(packet.RoomGameResult(draw, slotID, forfeit, displayInfo))

	for i, v := range r.leaveAfterGame {
		if v {
			return r.RemovePlayer(r.players[i], 0x00)
		}
	}

	return false
}

func (r *GameRoom) SendOpponent(conn mnet.MConnChannel, p mpacket.Packet) {
	for i, v := range r.players[0:2] {
		if v == conn {
			if i == 0 && r.players[1] != nil {
				r.players[1].Send(p)
			} else if i == 1 && r.players[0] != nil {
				r.players[0].Send(p)
			} else {
				return
			}
		}
	}
}

func (r *GameRoom) Tie() bool {
	return r.gameEnd(true, false, nil)
}

func (r *GameRoom) GiveUp(conn mnet.MConnChannel) bool {
	return r.gameEnd(false, true, conn)
}

func (r *GameRoom) LeaveAfterGame(conn mnet.MConnChannel) {
	for i, v := range r.players[0:2] {
		if v == conn {
			r.leaveAfterGame[i] = true

			player, err := Players.GetFromConn(conn)

			if err != nil {
				return
			}

			r.Broadcast(packet.RoomYellowChat(0x05, player.Char().Name))

			return
		}
	}
}

func (r *GameRoom) UndoLeaveAfterGame(conn mnet.MConnChannel) {
	for i, v := range r.players[0:2] {
		if v == conn {
			r.leaveAfterGame[i] = false

			player, err := Players.GetFromConn(conn)

			if err != nil {
				return
			}

			r.Broadcast(packet.RoomYellowChat(0x06, player.Char().Name))

			return
		}
	}
}

func (r *GameRoom) ChangeTurn() {
	r.Broadcast(packet.RoomGameSkip(r.p1Turn))
	r.p1Turn = !r.p1Turn
}

func checkOmokDraw(board [15][15]byte) bool {
	for i := 0; i < 15; i++ {
		for j := 0; j < 15; j++ {
			if board[i][j] > 0 {
				return false
			}
		}
	}

	return true
}

func checkOmokWin(board [15][15]byte, piece byte) bool {
	// Check horizontal
	for i := 0; i < 15; i++ {
		for j := 0; j < 11; j++ {
			if board[j][i] == piece &&
				board[j+1][i] == piece &&
				board[j+2][i] == piece &&
				board[j+3][i] == piece &&
				board[j+4][i] == piece {
				return true
			}
		}
	}

	// Check vertical
	for i := 0; i < 11; i++ {
		for j := 0; j < 15; j++ {
			if board[j][i] == piece &&
				board[j][i+1] == piece &&
				board[j][i+2] == piece &&
				board[j][i+3] == piece &&
				board[j][i+4] == piece {
				return true
			}
		}
	}

	// Check diagonal 1
	for i := 4; i < 15; i++ {
		for j := 0; j < 11; j++ {
			if board[j][i] == piece &&
				board[j+1][i-1] == piece &&
				board[j+2][i-2] == piece &&
				board[j+3][i-3] == piece &&
				board[j+4][i-4] == piece {
				return true
			}
		}
	}

	// Check diagonal 2
	for i := 0; i < 11; i++ {
		for j := 0; j < 11; j++ {
			if board[j][i] == piece &&
				board[j+1][i+1] == piece &&
				board[j+2][i+2] == piece &&
				board[j+3][i+3] == piece &&
				board[j+4][i+4] == piece {
				return true
			}
		}
	}

	return false
}

func (r *OmokRoom) Start() {
	r.GameRoom.Start()
	r.board = [15][15]byte{}
	r.Broadcast(packet.RoomOmokStart(r.p1Turn))
}

func (r *OmokRoom) PlacePiece(x, y int32, piece byte) bool {
	if x > 14 || y > 14 || x < 0 || y < 0 {
		return false
	}

	if r.board[x][y] != 0 {
		if r.p1Turn {
			r.players[0].Send(packet.RoomOmokInvalidPlaceMsg())
		} else {
			r.players[1].Send(packet.RoomOmokInvalidPlaceMsg())
		}

		return false
	}

	r.board[x][y] = piece

	if r.p1Turn {
		r.previousTurn[0][0] = x
		r.previousTurn[0][1] = y
	} else {
		r.previousTurn[1][0] = x
		r.previousTurn[1][1] = y
	}

	r.Broadcast(packet.RoomPlaceOmokPiece(x, y, piece))

	win := checkOmokWin(r.board, piece)
	draw := checkOmokDraw(r.board)

	if win || draw {
		r.board = [15][15]byte{}
		if r.gameEnd(draw, false, nil) {
			return true
		}
	}

	r.ChangeTurn()

	return false
}

func (r *OmokRoom) UndoTurn(conn mnet.MConnChannel) {
	for i, v := range r.players[0:2] {
		if v != conn {
			r.board[r.previousTurn[i][0]][r.previousTurn[i][1]] = 0
			r.Broadcast(packet.RoomUndo(r.previousTurn[i][0], r.previousTurn[i][1], r.p1Turn))
			r.p1Turn = !r.p1Turn
			return
		}
	}
}

func (r *MemoryRoom) Start() {
	r.GameRoom.Start()
	r.shuffleCards()
	r.Broadcast(packet.RoomMemoryStart(r.p1Turn, int32(r.BoardType), r.cards))
}

func (r *MemoryRoom) shuffleCards() {
	loopCounter := byte(0)
	switch r.BoardType {
	case 0:
		loopCounter = 6
	case 1:
		loopCounter = 10
	case 2:
		loopCounter = 15
	default:
		fmt.Println("Cannot shuffle unkown card type")
	}

	r.cards = make([]byte, 0)

	for i := byte(0); i < loopCounter; i++ {
		r.cards = append(r.cards, i, i)
	}

	rand.Shuffle(len(r.cards), func(i, j int) {
		r.cards[i], r.cards[j] = r.cards[j], r.cards[i]
	})
}

func (r *MemoryRoom) SelectCard(turn, cardID byte, conn mnet.MConnChannel) bool {
	if int(cardID) >= len(r.cards) {
		return false
	}

	if turn == 1 {
		r.firstCardPick = cardID
		r.BroadcastExcept(packet.RoomSelectCard(turn, cardID, r.firstCardPick, turn), conn)
	} else if r.cards[r.firstCardPick] == r.cards[cardID] {
		player, err := Players.GetFromConn(conn)

		if err != nil {
			return false
		}

		if r.p1Turn {
			r.matches[0]++
			r.Broadcast(packet.RoomSelectCard(turn, cardID, r.firstCardPick, 0xFF))
			// increment player matched card number
		} else {
			r.matches[1]++
			r.Broadcast(packet.RoomSelectCard(turn, cardID, r.firstCardPick, 0xFF))
			// increment player matched card number
		}

		r.Broadcast(packet.RoomYellowChat(0x09, player.Char().Name))

		return r.checkCardWin()
	} else if r.p1Turn {
		r.Broadcast(packet.RoomSelectCard(turn, cardID, r.firstCardPick, 0))
		r.p1Turn = !r.p1Turn
	} else {
		r.Broadcast(packet.RoomSelectCard(turn, cardID, r.firstCardPick, 1))
		r.p1Turn = !r.p1Turn
	}

	return false
}

func (r *MemoryRoom) checkCardWin() bool {
	totalMatches := r.matches[0] + r.matches[1]

	win, draw := false, false

	switch r.BoardType {
	case 0:
		if totalMatches == 6 {
			if r.matches[0] == r.matches[1] {
				draw = true
			} else { // current player must have won
				win = true
			}
		}
	case 1:
		if totalMatches == 10 {
			if r.matches[0] == r.matches[1] {
				draw = true
			} else {
				win = true
			}
		}
	case 2:
		if totalMatches == 15 {
			if r.matches[0] == r.matches[1] {
				draw = true
			} else {
				win = true
			}
		}
	}

	if win || draw {
		if r.gameEnd(draw, false, nil) {
			return true
		}
		r.matches[0], r.matches[1] = 0, 0
	}

	return false
}
