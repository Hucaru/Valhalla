package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/game/packet"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type roomContainer map[int32]*Room

var Rooms = make(roomContainer)

const (
	OmokRoom         = 0x01
	MemoryRoom       = 0x02
	TradeRoom        = 0x03
	PersonalShop     = 0x04
	omokMaxPlayers   = 4
	memoryMaxPlayers = 4
	tradeMaxPlayers  = 2
)

// type Room interface {
// 	Broadcast(p mpacket.Packet)
// 	SendMessage(name, msg string)
// 	AddPlayer(conn mnet.MConnChannel)
// 	RemovePlayer(conn mnet.MConnChannel, msgCode byte)
// }

// type TradeRoom struct {
// 	ID       int32
// 	players  []mnet.MConnChannel
// 	RoomType byte

// 	accepted int
// 	items    [2][9]Item
// 	mesos    [2]int32
// }

// type GameRoom struct {
// }

type Room struct {
	ID       int32
	RoomType byte
	players  []mnet.MConnChannel

	Name, Password string
	inProgress     bool
	BoardType      byte
	leaveAfterGame [2]bool
	p1Turn         bool

	board        [15][15]byte
	previousTurn [2][2]int32

	cards         []byte
	firstCardPick byte
	matches       int

	accepted int
	items    [2][9]Item
	mesos    [2]int32

	maxPlayers int
}

var roomCounter = int32(0)

func (rc *roomContainer) getNewRoomID() int32 {
	roomCounter++

	if roomCounter == 0 {
		roomCounter = 1
	}

	return roomCounter
}

func (rc *roomContainer) CreateMemoryRoom(name, password string, boardType byte) int32 {
	id := rc.getNewRoomID()
	r := &Room{ID: id, RoomType: MemoryRoom, Name: name, Password: password, BoardType: boardType, maxPlayers: memoryMaxPlayers, p1Turn: true}
	Rooms[id] = r
	return id
}

func (rc *roomContainer) CreateOmokRoom(name, password string, boardType byte) int32 {
	id := rc.getNewRoomID()
	r := &Room{ID: id, RoomType: OmokRoom, Name: name, Password: password, BoardType: boardType, maxPlayers: omokMaxPlayers, p1Turn: true, previousTurn: [2][2]int32{}}
	Rooms[id] = r
	return id
}

func (rc *roomContainer) CreateTradeRoom() int32 {
	id := rc.getNewRoomID()
	r := &Room{ID: id, RoomType: TradeRoom, maxPlayers: tradeMaxPlayers}
	Rooms[id] = r
	return id
}

func (r *Room) IsOwner(conn mnet.MConnChannel) bool {
	if len(r.players) > 0 && r.players[0] == conn {
		return true
	}

	return false
}

func (r *Room) Broadcast(p mpacket.Packet) {
	for _, v := range r.players {
		v.Send(p)
	}
}

func (r *Room) BroadcastExcept(p mpacket.Packet, conn mnet.MConnChannel) {
	for _, v := range r.players {
		if v == conn {
			continue
		}

		v.Send(p)
	}
}

func (r *Room) SendMessage(name, msg string) {
	for roomSlot, v := range r.players {
		if Players[v].Char().Name == name {
			r.Broadcast(packet.RoomChat(name, msg, byte(roomSlot)))
			break
		}
	}
}

func (r *Room) AddPlayer(conn mnet.MConnChannel) {
	if len(r.players) == r.maxPlayers {
		conn.Send(packet.RoomFull())
	}

	r.players = append(r.players, conn)

	player := Players[conn]
	roomPos := byte(len(r.players)) - 1

	if roomPos == 0 {
		Maps[player.Char().MapID].Send(packet.MapShowGameBox(player.Char().ID, r.ID, r.RoomType, r.BoardType, r.Name, bool(len(r.Password) > 0), r.inProgress, 0x01), player.InstanceID)
	}

	r.Broadcast(packet.RoomJoin(r.RoomType, roomPos, player.Char()))

	displayInfo := []def.Character{}

	for _, v := range r.players {
		displayInfo = append(displayInfo, Players[v].Char())
	}

	if len(displayInfo) > 0 {
		conn.Send(packet.RoomShowWindow(r.RoomType, r.BoardType, byte(r.maxPlayers), roomPos, r.Name, displayInfo))
	}

	Players[conn].RoomID = r.ID
}

func (r *Room) RemovePlayer(conn mnet.MConnChannel, msgCode byte) bool {
	closeRoom := false
	roomSlot := -1

	for i, v := range r.players {
		if v == conn {
			roomSlot = i
			break
		}
	}

	if roomSlot < 0 {
		return false
	}

	player := Players[conn]
	player.RoomID = 0

	switch r.RoomType {
	case TradeRoom:
		if r.accepted > 0 {
			r.Broadcast(packet.RoomLeave(byte(roomSlot), 7))
		} else {
			r.Broadcast(packet.RoomLeave(byte(roomSlot), 2))
		}

		closeRoom = true
	case MemoryRoom:
		fallthrough
	case OmokRoom:
		if roomSlot == 0 {
			closeRoom = true

			Maps[player.Char().MapID].Send(packet.MapRemoveGameBox(player.Char().ID), player.InstanceID)

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
	default:
		fmt.Println("have not implemented remove player for room type", r.RoomType)
	}

	return closeRoom
}

func (r *Room) Expel() {
	if len(r.players) > 1 {
		r.RemovePlayer(r.players[1], 5)
	}
}

func (r *Room) shuffleCards() {
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

	for i := byte(0); i < loopCounter; i++ {
		r.cards = append(r.cards, i, i)
	}

	shuffle := func(vals []byte) {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		for len(vals) > 0 {
			n := len(vals)
			randIndex := r.Intn(n)
			vals[n-1], vals[randIndex] = vals[randIndex], vals[n-1]
			vals = vals[:n-1]
		}
	}

	shuffle(r.cards)
}

func (r *Room) Start() {
	if len(r.players) == 0 {
		return
	}

	r.inProgress = true

	player := Players[r.players[0]]
	Maps[player.Char().MapID].Send(packet.MapShowGameBox(player.Char().ID, r.ID, r.RoomType, r.BoardType, r.Name, bool(len(r.Password) > 0), r.inProgress, 0x01), player.InstanceID)

	switch r.RoomType {
	case OmokRoom:
		r.Broadcast(packet.RoomOmokStart(r.p1Turn))
	case MemoryRoom:
		r.shuffleCards()
		r.Broadcast(packet.RoomMemoryStart(r.p1Turn, int32(r.BoardType), r.cards))
	default:
		fmt.Println("Cannot start a non game room")
	}
}

func (r *Room) ChangeTurn() {
	r.Broadcast(packet.RoomGameSkip(r.p1Turn))
	r.p1Turn = !r.p1Turn
}

func (r *Room) PlacePiece(x, y int32, piece byte) {
	if x > 14 || y > 14 || x < 0 || y < 0 {
		return
	}

	if r.board[x][y] != 0 {
		if r.p1Turn {
			r.players[0].Send(packet.RoomOmokInvalidPlaceMsg())
		} else {
			r.players[1].Send(packet.RoomOmokInvalidPlaceMsg())
		}

		return
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

	forfeit := false

	if win || draw {
		r.gameEnd(draw, forfeit)
	}

	r.ChangeTurn()
}

func (r *Room) SelectCard(firstPick bool, cardID byte, conn mnet.MConnChannel) {
	if int(cardID) >= len(r.cards) {
		return
	}

	if firstPick {
		r.firstCardPick = cardID
		r.BroadcastExcept(packet.RoomSelectCard(firstPick, cardID, r.firstCardPick, 1), conn)
	} else if r.cards[r.firstCardPick] == r.cards[cardID] {
		if r.p1Turn {
			r.Broadcast(packet.RoomSelectCard(firstPick, cardID, r.firstCardPick, 2))
			// set owner points
		} else {
			r.Broadcast(packet.RoomSelectCard(firstPick, cardID, r.firstCardPick, 3))
			// set p1 points
		}
	} else if r.p1Turn {
		r.Broadcast(packet.RoomSelectCard(firstPick, cardID, r.firstCardPick, 0))
	} else {
		r.Broadcast(packet.RoomSelectCard(firstPick, cardID, r.firstCardPick, 1))
	}
}

func (r *Room) gameEnd(draw, forfeit bool) {
	// Update box on map
	r.inProgress = false
	player := Players[r.players[0]]
	Maps[player.Char().MapID].Send(packet.MapShowGameBox(player.Char().ID, r.ID, r.RoomType, r.BoardType, r.Name, bool(len(r.Password) > 0), r.inProgress, 0x01), player.InstanceID)

	// Update player records
	slotID := byte(0)
	if !r.p1Turn {
		slotID = 1
	}

	if forfeit {
		if slotID == 1 { // for forfeits slot id is inversed
			Players[r.players[0]].SetMinigameLoss(Players[r.players[0]].Char().MiniGameLoss + 1)
			Players[r.players[1]].SetMinigameWins(Players[r.players[1]].Char().MiniGameWins + 1)
		} else {
			Players[r.players[1]].SetMinigameLoss(Players[r.players[1]].Char().MiniGameLoss + 1)
			Players[r.players[0]].SetMinigameWins(Players[r.players[0]].Char().MiniGameWins + 1)
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

	r.board = [15][15]byte{}

	// kick players who have registered to leave
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
