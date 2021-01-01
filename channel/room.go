package channel

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
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

const roomMaxPlayers = 2

type roomer interface {
	id() int32
	setID(int32)
	addPlayer(*player) bool
	closed() bool
	present(int32) bool
	chatMsg(*player, string)
	ownerID() int32
}

type room struct {
	roomID        int32
	ownerPlayerID int32
	roomType      byte
	players       []*player
}

func (r room) id() int32 {
	return r.roomID
}

func (r *room) setID(id int32) {
	r.roomID = id
}

func (r room) ownerID() int32 {
	return r.ownerPlayerID
}

func (r *room) addPlayer(plr *player) bool {
	if len(r.players) == 0 {
		r.ownerPlayerID = plr.id
	} else if len(r.players) == 2 {
		plr.send(packetRoomFull())
		return false
	}

	for _, v := range r.players {
		if v == plr {
			return false
		}
	}

	r.players = append(r.players, plr)

	return true
}

func (r *room) removePlayer(plr *player) bool {
	for i, v := range r.players {
		if v.conn == plr.conn {
			r.players = append(r.players[:i], r.players[i+1:]...) // preserve order for slot numbers
			return true
		}
	}

	return false
}

func (r room) send(p mpacket.Packet) {
	for _, v := range r.players {
		v.send(p)
	}
}

func (r room) sendExcept(p mpacket.Packet, plr *player) {
	for _, v := range r.players {
		if v.conn == plr.conn {
			continue
		}
		v.send(p)
	}
}

func (r room) closed() bool {
	if len(r.players) == 0 {
		return true
	}

	return false
}

func (r room) chatMsg(plr *player, msg string) {
	for i, v := range r.players {
		if v.conn == plr.conn {
			r.send(packetRoomChat(plr.name, msg, byte(i)))
		}
	}
}

func (r room) present(id int32) bool {
	for _, v := range r.players {
		if v.id == id {
			return true
		}
	}

	return false
}

type gameRoomer interface {
	checkPassword(string, *player) bool
	ready(*player)
	unready(*player)
	start()
	displayBytes() []byte
	kickPlayer(*player, byte) bool
	expel() bool
	changeTurn()
	requestTie(*player)
	requestTieResult(bool, *player)
	forfeit(*player)
	requestExit(bool, *player)
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

func (r *gameRoom) addPlayer(plr *player) bool {
	if !r.room.addPlayer(plr) {
		return false
	}

	plr.send(packetRoomShowWindow(r.roomType, r.boardType, byte(roomMaxPlayers), byte(len(r.players)-1), r.name, r.players))

	if len(r.players) > 1 {
		r.sendExcept(packetRoomJoin(r.roomType, byte(len(r.players)-1), r.players[len(r.players)-1]), plr)
	}

	return true
}

func (r gameRoom) checkPassword(password string, plr *player) bool {
	if password != r.password {
		plr.send(packetRoomIncorrectPassword())
		return false
	}
	return true
}

func (r *gameRoom) kickPlayer(plr *player, reason byte) bool {
	for i, v := range r.players {
		if v.conn == plr.conn {
			if r.inProgress {
				r.gameEnd(false, true, plr, 0)
			}

			if !r.room.removePlayer(plr) {
				return false
			}

			plr.send(packetRoomLeave(byte(i), reason))

			if i == 0 { // owner is always at index 0
				for j := range r.players {
					fmt.Println(packetRoomLeave(byte(j+1), 0x0))
					r.send(packetRoomLeave(byte(j+1), 0x0))
				}
				r.players = []*player{} // sets the room into a closed state
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
		r.send(packetRoomYellowChat(0, r.players[1].name))
		r.kickPlayer(r.players[1], 0x5)

		return true
	}

	return false
}

func (r *gameRoom) ready(plr *player) {
	for i, v := range r.players {
		if v.conn == plr.conn && i == 1 {
			r.send(packetRoomReady())
		}
	}
}

func (r *gameRoom) unready(plr *player) {
	for i, v := range r.players {
		if v.conn == plr.conn && i == 1 {
			r.send(packetRoomUnready())
		}
	}
}

func (r *gameRoom) changeTurn() {
	r.p1Turn = !r.p1Turn
	r.send(packetRoomGameSkip(r.p1Turn))
}

func (r *gameRoom) gameEnd(draw, forfeit bool, plr *player, winningSlot byte) {
	r.inProgress = false

	if forfeit {
		if plr.conn == r.players[0].conn {
			winningSlot = 0x01
		} else {
			winningSlot = 0x00
		}
	}

	r.assignPoints(draw, winningSlot)
	r.assignWinLossDraw(draw, winningSlot)
	r.send(packetRoomGameResult(draw, winningSlot, forfeit, r.players))

	if r.exit[0] == true {
		for _, v := range r.players {
			r.kickPlayer(v, 0)
		}
	} else if r.exit[1] == true {
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

func (r *gameRoom) requestTie(plr *player) {
	for _, v := range r.players {
		if v.conn != plr.conn {
			v.send(packetRoomRequestTie())
			return
		}
	}
}

func (r *gameRoom) requestTieResult(tie bool, plr *player) {
	if tie == true {
		r.gameEnd(true, false, nil, 0)
	} else {
		for _, v := range r.players {
			if v.conn != plr.conn {
				v.send(packetRoomRejectTie())
				return
			}
		}
	}
}

func (r *gameRoom) forfeit(plr *player) {
	for _, v := range r.players {
		if v.conn == plr.conn {
			r.gameEnd(false, true, plr, 0)
			return
		}
	}
}

func (r *gameRoom) requestExit(exit bool, plr *player) {
	for i, v := range r.players {
		if v.conn == plr.conn {
			r.exit[i] = exit
			return
		}
	}
}

func (r gameRoom) displayBytes() []byte {
	p := mpacket.NewPacket()

	p.WriteInt32(r.players[0].id)
	p.WriteByte(r.roomType)
	p.WriteInt32(r.roomID)
	p.WriteString(r.name)
	p.WriteBool(len(r.password) > 0)
	p.WriteByte(r.boardType)
	p.WriteByte(byte(len(r.players))) // number that is seen in the box? Player count?
	p.WriteByte(2)                    // ?
	p.WriteBool(r.inProgress)         //Sets some korean text, does it mean game is ongoing?

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
	r := room{roomID: id, roomType: roomTypeOmok}
	g := gameRoom{room: r, name: name, password: password, boardType: boardType, ownerStart: false}
	return &omokRoom{gameRoom: g}
}

func (r *omokRoom) placePiece(x, y int32, piece byte, plr *player) bool {
	if x > 14 || y > 14 || x < 0 || y < 0 {
		return false
	}

	// Turns are out of sync with client probably due to hacking
	if r.p1Turn && plr.conn != r.players[0].conn {
		r.players[1].send(packetRoomOmokInvalidPlaceMsg())
	} else if !r.p1Turn && plr.conn != r.players[1].conn {
		r.players[0].send(packetRoomOmokInvalidPlaceMsg())
	}

	if r.board[x][y] != 0 {
		if r.p1Turn {
			r.players[0].send(packetRoomOmokInvalidPlaceMsg())
		} else {
			r.players[1].send(packetRoomOmokInvalidPlaceMsg())
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

		if r.closed() { // If owner exit as part of game leave
			return false
		}

		return true
	}

	r.p1Turn = !r.p1Turn

	return false
}

func (r *omokRoom) requestUndo(plr *player) {
	for i, v := range r.players {
		if v.conn != plr.conn {
			if (i == 0 && r.p1Plays == 0) || (i == 1 && r.p2Plays == 0) {
				return
			}

			v.send(packetRoomRequestUndo())
			return
		}
	}
}

// RequestUndoResult is the choice the other player made to the request
func (r *omokRoom) requestUndoResult(undo bool, plr *player) {
	if undo {
		for i, v := range r.players {
			if v.conn != plr.conn {
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
			if v.conn != plr.conn {
				v.send(packetRoomRejectUndo())
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

type memoryRoom struct {
	gameRoom

	cards         []byte
	firstCardPick byte
	matches       [2]int
}

func newMemoryRoom(id int32, name, password string, boardType byte) roomer {
	r := room{roomID: id, roomType: roomTypeMemory}
	g := gameRoom{room: r, name: name, password: password, boardType: boardType, ownerStart: false}
	return &memoryRoom{gameRoom: g}
}

func (r *memoryRoom) selectCard(turn, cardID byte, plr *player) bool {
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
		r.send(packetRoomYellowChat(0x09, plr.name))

		win, draw := r.checkCardWin()

		if win || draw {
			var winningSlot byte = 0x00

			if r.matches[1] > r.matches[0] {
				winningSlot = 0x01
			}

			r.gameEnd(draw, false, nil, winningSlot)

			if r.closed() { // If owner exit as part of game leave
				return false
			}

			return true
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
	case 0:
		if totalMatches == 6 {
			if r.matches[0] == r.matches[1] {
				draw = true
			} else {
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

	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	random.Shuffle(len(r.cards), func(i, j int) {
		r.cards[i], r.cards[j] = r.cards[j], r.cards[i]
	})
}

type tradeRoom struct {
	room
}

func newTradeRoom(id int32) roomer {
	r := room{roomID: id, roomType: roomTypeTrade}
	return &tradeRoom{room: r}
}

func (r *tradeRoom) addPlayer(plr *player) bool {
	if !r.room.addPlayer(plr) {
		return false
	}

	plr.send(packetRoomShowWindow(r.roomType, 0x00, byte(roomMaxPlayers), byte(len(r.players)-1), "", r.players))

	if len(r.players) > 1 {
		r.sendExcept(packetRoomJoin(r.roomType, byte(len(r.players)-1), r.players[len(r.players)-1]), plr)
	}

	return true
}

func (r *tradeRoom) removePlayer(plr *player) {
	// Note: since anyone leaving the room causes it to close we don't need to remove players
	for i, v := range r.players {
		if v.conn != plr.conn {
			v.send(packetRoomLeave(byte(i), 0x02))
		}
	}
}

func (r tradeRoom) sendInvite(plr *player) {
	plr.send(packetRoomInvite(roomTypeTrade, r.players[0].name, r.roomID))
}

func (r tradeRoom) reject(code byte, name string) {
	r.send(packetRoomInviteResult(code, name))
}

func (r *tradeRoom) insertItem() {

}

func (r *tradeRoom) addMesos(amount int32, plr *player) {

}

func (r *tradeRoom) swapItems() bool {
	return true
}

func packetRoomShowWindow(roomType, boardType, maxPlayers, roomSlot byte, roomTitle string, players []*player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x05)
	p.WriteByte(roomType)
	p.WriteByte(maxPlayers)
	p.WriteByte(roomSlot)

	for i, v := range players {
		p.WriteByte(byte(i))
		p.Append(v.displayBytes())
		p.WriteInt32(0) // not sure what this is - memory card game seed? board settings?
		p.WriteString(v.name)
	}

	p.WriteByte(0xFF)

	if roomType == 0x03 {
		return p
	}

	for i, v := range players {
		p.WriteByte(byte(i))
		p.WriteInt32(0) // not sure what this is!?
		p.WriteInt32(v.miniGameWins)
		p.WriteInt32(v.miniGameDraw)
		p.WriteInt32(v.miniGameLoss)
		p.WriteInt32(v.miniGamePoints)
	}

	p.WriteByte(0xFF)
	p.WriteString(roomTitle)
	p.WriteByte(boardType)
	p.WriteByte(0)

	return p
}

func packetRoomJoin(roomType, roomSlot byte, plr *player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x04)
	p.WriteByte(roomSlot)
	p.Append(plr.displayBytes())
	p.WriteInt32(0) //?
	p.WriteString(plr.name)

	if roomType == 0x03 || roomType == 0x04 {
		return p
	}

	p.WriteInt32(1) // not sure what this is!?
	p.WriteInt32(plr.miniGameWins)
	p.WriteInt32(plr.miniGameDraw)
	p.WriteInt32(plr.miniGameLoss)
	p.WriteInt32(plr.miniGamePoints)

	return p
}

func packetRoomLeave(roomSlot byte, leaveCode byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x0A)
	p.WriteByte(roomSlot)
	p.WriteByte(leaveCode) // 2 - trade cancelled, 6 - trade success

	return p
}

func packetRoomYellowChat(msgType byte, name string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x06)
	p.WriteByte(7)
	// expelled: 0, x's turn: 1, forfeit: 2, handicap request: 3, left: 4,
	// called to leave/expeled: 5, cancelled leave: 6, entered: 7, can't start lack of mesos:8
	// has matched cards: 9
	p.WriteByte(msgType)
	p.WriteString(name)

	return p
}

func packetRoomReady() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x32)

	return p
}

func packetRoomChat(sender, message string, roomSlot byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x06)
	p.WriteByte(8)        // msg type
	p.WriteByte(roomSlot) //
	p.WriteString(sender + " : " + message)

	return p
}

func packetRoomUnready() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x33)

	return p
}

func packetRoomOmokStart(ownerStart bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x35)
	p.WriteBool(ownerStart)

	return p
}

func packetRoomMemoryStart(ownerStart bool, boardType int32, cards []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x35)
	p.WriteBool(ownerStart)
	p.WriteByte(0x0C)

	for i := 0; i < len(cards); i++ {
		p.WriteInt32(int32(cards[i]))
	}

	return p
}

func packetRoomGameSkip(isOwner bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x37)

	if isOwner {
		p.WriteByte(0)
	} else {
		p.WriteByte(1)
	}

	return p
}

func packetRoomPlaceOmokPiece(x, y int32, piece byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x38)
	p.WriteInt32(x)
	p.WriteInt32(y)
	p.WriteByte(piece)

	return p
}

func packetRoomOmokInvalidPlaceMsg() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x39)
	p.WriteByte(0x0)

	return p
}

func packetRoomSelectCard(turn, cardID, firstCardPick byte, result byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x3c)
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
	p.WriteByte(0x2a)

	return p
}

func packetRoomRejectTie() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x2b)

	return p
}

func packetRoomRequestUndo() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x2e)

	return p
}

func packetRoomRejectUndo() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x2f)
	p.WriteByte(0x00)

	return p
}

func packetRoomUndo(piece, slot byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x2f)
	p.WriteByte(0x01)
	p.WriteByte(piece) // 0x00 seems to do nothing?
	p.WriteByte(slot)

	return p
}

func packetRoomGameResult(draw bool, winningSlot byte, forfeit bool, plr []*player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x36)

	if !draw && !forfeit {
		p.WriteBool(draw)
	} else if draw {
		p.WriteBool(draw)
	} else if forfeit {
		p.WriteByte(2)
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

func packetRoomInvite(roomType byte, name string, roomID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x02)
	p.WriteByte(roomType)
	p.WriteString(name)
	p.WriteInt32(roomID)

	return p
}

func packetRoomInviteResult(resultCode byte, name string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x03)
	p.WriteByte(resultCode)
	p.WriteString(name)

	return p
}

func packetRoomShowAccept() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x0F)

	return p
}

func packetRoomEnterErrorMsg(errorCode byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(0x05)
	p.WriteByte(0x00)
	p.WriteByte(errorCode)

	return p
}

func packetRoomClosed() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x01)
}

func packetRoomFull() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x02)
}

func packetRoomBusy() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x03)
}

func packetRoomNotAllowedWhenDead() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x04)
}

func packetRoomNotAllowedDuringEvent() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x05)
}

func packetRoomThisCharacterNotAllowed() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x06)
}

func packetRoomNoTradeAtm() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x07)
}

func packetRoomMiniRoomNotHere() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x08)
}

func packetRoomTradeRequireSameMap() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x09)
}

func packetRoomcannotCreateMiniroomHere() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x0a)
}

func packetRoomCannotStartGameHere() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x0b)
}

func packetRoomPersonalStoreFMOnly() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x0c)
}
func packetRoomGarbageMsgAboutFloorInFm() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x0d)
}

func packetRoomMayNotEnterStore() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x0e)
}

func packetRoomStoreMaintenance() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x0F)
}

func packetRoomCannotEnterTournament() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x10)
}

func packetRoomGarbageTradeMsg() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x11)
}

func packetRoomNotEnoughMesos() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x12)
}

func packetRoomIncorrectPassword() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x13)
}
