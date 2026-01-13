package channel

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mpacket"
)

type gameRoomer interface {
	checkPassword(string, *Player) bool
	ready(*Player)
	unready(*Player)
	start()
	displayBytes() []byte
	kickPlayer(*Player, byte) bool
	expel() bool
	changeTurn()
	requestTie(*Player)
	requestTieResult(bool, *Player)
	forfeit(*Player)
	requestExit(bool, *Player)
}

type gameRoom struct {
	room

	boardType  byte
	ownerStart bool
	p1Turn     bool
	inProgress bool
	name       string
	password   string
	exit       [2]bool
}

func (r *gameRoom) addPlayer(plr *Player) bool {
	if !r.room.addPlayer(plr) {
		return false
	}

	plr.Send(packetRoomShowWindow(r.roomType, r.boardType, byte(constant.RoomMaxPlayers), byte(len(r.players)-1), r.name, r.players))

	if len(r.players) > 1 {
		r.sendExcept(packetRoomJoin(r.roomType, byte(len(r.players)-1), r.players[len(r.players)-1]), plr)
	}

	return true
}

func (r gameRoom) checkPassword(password string, plr *Player) bool {
	if password != r.password {
		plr.Send(packetRoomIncorrectPassword())
		return false
	}
	return true
}

func (r *gameRoom) kickPlayer(plr *Player, reason byte) bool {
	for i, v := range r.players {
		if v.Conn == plr.Conn {
			if r.inProgress {
				r.gameEnd(false, true, plr, 0)
			}

			if !r.room.removePlayer(plr) {
				return false
			}

			plr.Send(packetRoomLeave(byte(i), reason))

			if i == constant.RoomOwnerSlot { // owner is always at index 0
				for j := range r.players {
					fmt.Println(packetRoomLeave(byte(j+1), 0x0))
					r.send(packetRoomLeave(byte(j+1), 0x0))
				}
				r.players = []*Player{} // sets the room into a closed state
			} else {
				fmt.Println(packetRoomLeave(byte(i), reason))
				r.send(packetRoomLeave(byte(i), reason))
			}

			return true
		}
	}

	return false
}

func (r *gameRoom) expel() bool {
	if len(r.players) > 1 {
		r.send(packetRoomYellowChat(constant.RoomYellowChatExpelled, r.players[1].Name))
		r.kickPlayer(r.players[1], constant.MiniRoomExpelled)

		return true
	}

	return false
}

func (r *gameRoom) ready(plr *Player) {
	for i, v := range r.players {
		if v.Conn == plr.Conn && i == constant.RoomGuestSlot {
			r.send(packetRoomReady())
		}
	}
}

func (r *gameRoom) unready(plr *Player) {
	for i, v := range r.players {
		if v.Conn == plr.Conn && i == constant.RoomGuestSlot {
			r.send(packetRoomUnready())
		}
	}
}

func (r *gameRoom) changeTurn() {
	r.p1Turn = !r.p1Turn
	r.send(packetRoomGameSkip(r.p1Turn))
}

func (r *gameRoom) gameEnd(draw, forfeit bool, plr *Player, winningSlot byte) {
	r.inProgress = false

	if forfeit {
		if plr.Conn == r.players[0].Conn {
			winningSlot = 0x01
		} else {
			winningSlot = 0x00
		}
	}

	r.assignPoints(draw, winningSlot)
	r.assignWinLossDraw(draw, winningSlot)
	r.send(packetRoomGameResult(draw, winningSlot, forfeit, r.players))

	if r.exit[0] {
		for _, v := range r.players {
			r.kickPlayer(v, 0)
		}
	} else if r.exit[1] {
		r.kickPlayer(r.players[1], 0)
		r.exit[1] = false // no need to clear owner entry, if they leave room closes
	}
}

func (r *gameRoom) assignWinLossDraw(draw bool, winningSlot byte) {
	if draw {
		r.players[0].miniGameDraw = r.players[0].miniGameDraw + 1
		r.players[1].miniGameDraw = r.players[1].miniGameDraw + 1
	} else {
		r.players[winningSlot].miniGameWins = r.players[winningSlot].miniGameWins + 1

		if winningSlot == 0x00 {
			r.players[1].miniGameLoss = r.players[1].miniGameLoss + 1
		} else {
			r.players[0].miniGameLoss = r.players[0].miniGameLoss + 1
		}
	}
}

// TODO: Validate points/elo calculation
func (r *gameRoom) assignPoints(draw bool, winningSlot byte) {
	p := 400 // use the same as chess
	// Rating transformation
	r0 := math.Pow10(int(r.players[0].miniGamePoints) / p)
	r1 := math.Pow10(int(r.players[1].miniGamePoints) / p)
	// Expected score
	e0 := float64(r0 / (r0 + r1))
	e1 := float64(r1 / (r0 + r1))

	var s0, s1 float64

	var k = 32.0 // ICC chess value, Change this to change how hard ratings are impacted

	if draw {
		s0, s1 = 0.5, 0.5
	} else if winningSlot == 0x00 {
		s0 = 1
		s1 = 0
	} else {
		s0 = 0
		s1 = 1
	}

	r.players[0].miniGamePoints = r.players[0].miniGamePoints + int32(k*(s0-e0))
	r.players[1].miniGamePoints = r.players[1].miniGamePoints + int32(k*(s1-e1))
}

func (r *gameRoom) requestTie(plr *Player) {
	for _, v := range r.players {
		if v.Conn != plr.Conn {
			v.Send(packetRoomRequestTie())
			return
		}
	}
}

func (r *gameRoom) requestTieResult(tie bool, plr *Player) {
	if tie {
		r.gameEnd(true, false, nil, 0)
	} else {
		for _, v := range r.players {
			if v.Conn != plr.Conn {
				v.Send(packetRoomRejectTie())
				return
			}
		}
	}
}

func (r *gameRoom) forfeit(plr *Player) {
	for _, v := range r.players {
		if v.Conn == plr.Conn {
			r.gameEnd(false, true, plr, 0)
			return
		}
	}
}

func (r *gameRoom) requestExit(exit bool, plr *Player) {
	for i, v := range r.players {
		if v.Conn == plr.Conn {
			r.exit[i] = exit
			return
		}
	}
}

func (r gameRoom) displayBytes() []byte {
	p := mpacket.NewPacket()

	p.WriteInt32(r.players[0].ID)
	p.WriteByte(r.roomType)
	p.WriteInt32(r.roomID)
	p.WriteString(r.name)
	p.WriteBool(len(r.password) > 0)
	p.WriteByte(r.boardType)
	p.WriteByte(byte(len(r.players)))    // number that is seen in the box? Player count?
	p.WriteByte(constant.RoomMaxPlayers) // ?
	p.WriteBool(r.inProgress)            //Sets some korean text, does it mean game is ongoing?

	return p
}

type omokRoom struct {
	gameRoom

	board [15][15]byte

	p1History [2][2]int32
	p2History [2][2]int32

	p1Plays int
	p2Plays int
}

func newOmokRoom(id int32, name, password string, boardType byte) roomer {
	r := room{roomID: id, roomType: constant.MiniRoomTypeOmok}
	g := gameRoom{room: r, name: name, password: password, boardType: boardType, ownerStart: false}
	return &omokRoom{gameRoom: g}
}

func (r *omokRoom) placePiece(x, y int32, piece byte, plr *Player) bool {
	if x > constant.OmokBoardSize-1 || y > constant.OmokBoardSize-1 || x < 0 || y < 0 {
		return false
	}

	// Turns are out of sync with client probably due to hacking
	if r.p1Turn && plr.Conn != r.players[0].Conn {
		r.players[1].Send(packetRoomOmokInvalidPlaceMsg())
	} else if !r.p1Turn && plr.Conn != r.players[1].Conn {
		r.players[0].Send(packetRoomOmokInvalidPlaceMsg())
	}

	if r.board[x][y] != 0 {
		if r.p1Turn {
			r.players[0].Send(packetRoomOmokInvalidPlaceMsg())
		} else {
			r.players[1].Send(packetRoomOmokInvalidPlaceMsg())
		}

		return false
	}

	r.board[x][y] = piece

	if r.p1Turn {
		i := 1 - r.p1Plays%2
		r.p1History[i][0] = x
		r.p1History[i][1] = y
		r.p1Plays++
	} else {
		i := 1 - r.p2Plays%2
		r.p2History[i][0] = x
		r.p2History[i][1] = y
		r.p2Plays++
	}

	r.send(packetRoomPlaceOmokPiece(x, y, piece))

	win := checkOmokWin(r.board, piece)
	draw := checkOmokDraw(r.board)

	if win || draw {
		var winningSlot byte = 0x00

		if !r.p1Turn {
			winningSlot = 0x01
		}

		r.gameEnd(draw, false, nil, winningSlot)

		return !r.closed()
	}

	r.p1Turn = !r.p1Turn

	return false
}

func (r *omokRoom) requestUndo(plr *Player) {
	for i, v := range r.players {
		if v.Conn != plr.Conn {
			if (i == 0 && r.p1Plays == 0) || (i == 1 && r.p2Plays == 0) {
				return
			}

			v.Send(packetRoomRequestUndo())
			return
		}
	}
}

// RequestUndoResult is the choice the other Player made to the request
func (r *omokRoom) requestUndoResult(undo bool, plr *Player) {
	if undo {
		for i, v := range r.players {
			if v.Conn != plr.Conn {
				turns := byte(1)
				slot := byte(i)

				if i == 0 {
					r.p1Plays--
					j := 1 - r.p1Plays%2
					x := r.p1History[j][0]
					y := r.p1History[j][1]
					r.board[x][y] = 0

					if r.p1Turn {
						r.p2Plays--
						k := 1 - r.p2Plays%2
						x := r.p2History[k][0]
						y := r.p2History[k][1]
						r.board[x][y] = 0
						turns = 2
					}
				} else if i == 1 {
					r.p2Plays--
					j := 1 - r.p2Plays%2
					x := r.p2History[j][0]
					y := r.p2History[j][1]
					r.board[x][y] = 0

					if !r.p1Turn {
						r.p1Plays--
						k := 1 - r.p1Plays%2
						x := r.p1History[k][0]
						y := r.p1History[k][1]
						r.board[x][y] = 0
						turns = 2
					}
				}

				if slot == 0 {
					r.p1Turn = true
				} else {
					r.p1Turn = false
				}

				r.send(packetRoomUndo(turns, slot))
				return
			}
		}
	} else {
		for _, v := range r.players {
			if v.Conn != plr.Conn {
				v.Send(packetRoomRejectUndo())
				return
			}
		}
	}
}

func (r *omokRoom) start() {
	if len(r.players) < 2 {
		return
	}

	r.board = [15][15]byte{}
	r.inProgress = true
	r.ownerStart = !r.ownerStart
	r.p1Turn = r.ownerStart
	r.send(packetRoomOmokStart(r.ownerStart))
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
	for i := 0; i < constant.OmokBoardSize; i++ {
		for j := 0; j < constant.OmokBoardSize-4; j++ {
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
	for i := 0; i < constant.OmokBoardSize-4; i++ {
		for j := 0; j < constant.OmokBoardSize; j++ {
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
	for i := 4; i < constant.OmokBoardSize; i++ {
		for j := 0; j < constant.OmokBoardSize-4; j++ {
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
	for i := 0; i < constant.OmokBoardSize-4; i++ {
		for j := 0; j < constant.OmokBoardSize-4; j++ {
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

type memoryRoom struct {
	gameRoom

	cards         []byte
	firstCardPick byte
	matches       [2]int
}

func newMemoryRoom(id int32, name, password string, boardType byte) roomer {
	r := room{roomID: id, roomType: constant.MiniRoomTypeMatchCards}
	g := gameRoom{room: r, name: name, password: password, boardType: boardType, ownerStart: false}
	return &memoryRoom{gameRoom: g}
}

func (r *memoryRoom) selectCard(turn, cardID byte, plr *Player) bool {
	if int(cardID) >= len(r.cards) {
		return false
	}
	if turn == 1 {
		r.firstCardPick = cardID
		r.sendExcept(packetRoomSelectCard(turn, cardID, cardID, turn), plr)
	} else if r.cards[r.firstCardPick] == r.cards[cardID] {

		var points byte = 2

		if r.p1Turn {
			r.matches[0]++
		} else {
			r.matches[1]++
			points = 3
		}

		r.send(packetRoomSelectCard(turn, cardID, r.firstCardPick, points))
		r.send(packetRoomYellowChat(constant.RoomYellowChatMatchedCards, plr.Name))

		win, draw := r.checkCardWin()

		if win || draw {
			var winningSlot byte = 0x00

			if r.matches[1] > r.matches[0] {
				winningSlot = 0x01
			}

			r.gameEnd(draw, false, nil, winningSlot)

			return !r.closed()
		}

	} else if r.p1Turn {
		r.send(packetRoomSelectCard(turn, cardID, r.firstCardPick, 0))
		r.p1Turn = !r.p1Turn
	} else {
		r.send(packetRoomSelectCard(turn, cardID, r.firstCardPick, 1))
		r.p1Turn = !r.p1Turn
	}

	return false
}

func (r *memoryRoom) checkCardWin() (bool, bool) {
	win, draw := false, false
	totalMatches := r.matches[0] + r.matches[1]

	switch r.boardType {
	case constant.MatchCardsSizeSmall:
		if totalMatches == constant.MatchCardsPairsSmall {
			if r.matches[0] == r.matches[1] {
				draw = true
			} else {
				win = true
			}
		}
	case constant.MatchCardsSizeMedium:
		if totalMatches == constant.MatchCardsPairsMedium {
			if r.matches[0] == r.matches[1] {
				draw = true
			} else {
				win = true
			}
		}
	case constant.MatchCardsSizeLarge:
		if totalMatches == constant.MatchCardsPairsLarge {
			if r.matches[0] == r.matches[1] {
				draw = true
			} else {
				win = true
			}
		}
	}

	return win, draw
}

func (r *memoryRoom) start() {
	if len(r.players) < 2 {
		return
	}

	r.inProgress = true
	r.ownerStart = !r.ownerStart
	r.p1Turn = r.ownerStart
	r.matches[0], r.matches[1] = 0, 0
	r.shuffleCards()
	r.send(packetRoomMemoryStart(r.ownerStart, int32(r.boardType), r.cards))
}

func (r *memoryRoom) shuffleCards() {
	loopCounter := byte(0)
	switch r.boardType {
	case constant.MatchCardsSizeSmall:
		loopCounter = constant.MatchCardsPairsSmall
	case constant.MatchCardsSizeMedium:
		loopCounter = constant.MatchCardsPairsMedium
	case constant.MatchCardsSizeLarge:
		loopCounter = constant.MatchCardsPairsLarge
	default:
		fmt.Println("Cannot shuffle unkown card type")
	}

	r.cards = make([]byte, 0)

	for i := byte(0); i < loopCounter; i++ {
		r.cards = append(r.cards, i, i)
	}

	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	random.Shuffle(len(r.cards), func(i, j int) {
		r.cards[i], r.cards[j] = r.cards[j], r.cards[i]
	})
}

func packetRoomOmokStart(ownerStart bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomGameStart)
	p.WriteBool(ownerStart)

	return p
}

func packetRoomMemoryStart(ownerStart bool, boardType int32, cards []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomGameStart)
	p.WriteBool(ownerStart)
	p.WriteByte(constant.RoomPacketMemoryStart)

	for i := 0; i < len(cards); i++ {
		p.WriteInt32(int32(cards[i]))
	}

	return p
}

func packetRoomGameSkip(isOwner bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomChangeTurn)

	if isOwner {
		p.WriteByte(0)
	} else {
		p.WriteByte(1)
	}

	return p
}

func packetRoomPlaceOmokPiece(x, y int32, piece byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomPlacePiece)
	p.WriteInt32(x)
	p.WriteInt32(y)
	p.WriteByte(piece)

	return p
}

func packetRoomOmokInvalidPlaceMsg() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomInvalidPlace)
	p.WriteByte(0x0)

	return p
}

func packetRoomSelectCard(turn, cardID, firstCardPick byte, result byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomSelectCard)
	p.WriteByte(turn)

	if turn == 1 {
		p.WriteByte(cardID)
	} else if turn == 0 {
		p.WriteByte(cardID)
		p.WriteByte(firstCardPick)
		p.WriteByte(result)
	}

	return p
}

func packetRoomRequestTie() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomRequestTie)

	return p
}

func packetRoomRejectTie() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomRequestTieResult)

	return p
}

func packetRoomRequestUndo() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomRequestUndo)

	return p
}

func packetRoomRejectUndo() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomRequestUndoResult)
	p.WriteByte(0x00)

	return p
}

func packetRoomUndo(piece, slot byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomRequestUndoResult)
	p.WriteByte(0x01)
	p.WriteByte(piece)
	p.WriteByte(slot)

	return p
}

func packetRoomGameResult(draw bool, winningSlot byte, forfeit bool, plr []*Player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomGameResult)

	if !draw && !forfeit {
		p.WriteBool(draw)
	} else if draw {
		p.WriteBool(draw)
	} else if forfeit {
		p.WriteByte(constant.GameForfeit)
	}

	p.WriteByte(winningSlot)

	// Why is there a difference between the two?
	if draw {
		p.WriteByte(0)
		p.WriteByte(0)
		p.WriteByte(0)
	} else {
		p.WriteInt32(1) // ?
	}

	p.WriteInt32(plr[0].miniGameWins)
	p.WriteInt32(plr[0].miniGameDraw)
	p.WriteInt32(plr[0].miniGameLoss)
	p.WriteInt32(plr[0].miniGamePoints)
	p.WriteInt32(1)
	p.WriteInt32(plr[1].miniGameWins)
	p.WriteInt32(plr[1].miniGameDraw)
	p.WriteInt32(plr[1].miniGameLoss)
	p.WriteInt32(plr[1].miniGamePoints)

	return p
}

func packetRoomIncorrectPassword() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.MiniRoomEnterIncorrectPassword)
}
