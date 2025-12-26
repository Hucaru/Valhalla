package channel

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mpacket"
)

type roomer interface {
	id() int32
	setID(int32)
	addPlayer(*Player) bool
	closed() bool
	present(int32) bool
	chatMsg(*Player, string)
	ownerID() int32
}

type room struct {
	roomID        int32
	ownerPlayerID int32
	roomType      byte
	players       []*Player
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

func (r *room) addPlayer(plr *Player) bool {
	if len(r.players) == 0 {
		r.ownerPlayerID = plr.ID
	} else if len(r.players) == constant.RoomMaxPlayers {
		plr.Send(packetRoomFull())
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

func (r *room) removePlayer(plr *Player) bool {
	for i, v := range r.players {
		if v.Conn == plr.Conn {
			r.players = append(r.players[:i], r.players[i+1:]...) // preserve order for slot numbers
			return true
		}
	}

	return false
}

func (r room) send(p mpacket.Packet) {
	for _, v := range r.players {
		v.Send(p)
	}
}

func (r room) sendExcept(p mpacket.Packet, plr *Player) {
	for _, v := range r.players {
		if v.Conn == plr.Conn {
			continue
		}
		v.Send(p)
	}
}

func (r room) closed() bool {
	return len(r.players) == 0
}

func (r room) chatMsg(plr *Player, msg string) {
	for i, v := range r.players {
		if v.Conn == plr.Conn {
			r.send(packetRoomChat(plr.Name, msg, byte(i)))
		}
	}
}

func (r room) present(id int32) bool {
	for _, v := range r.players {
		if v.ID == id {
			return true
		}
	}

	return false
}

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

type tradeRoom struct {
	room
	mesos     map[int32]int32
	items     map[int32]map[byte]Item
	confirmed map[int32]bool

	finalized bool
}

func newTradeRoom(id int32) roomer {
	r := room{roomID: id, roomType: constant.MiniRoomTypeTrade}
	return &tradeRoom{room: r, mesos: make(map[int32]int32), items: make(map[int32]map[byte]Item), confirmed: make(map[int32]bool), finalized: false}
}

func (r *tradeRoom) addPlayer(plr *Player) bool {
	if !r.room.addPlayer(plr) {
		return false
	}

	plr.Send(packetRoomShowWindow(r.roomType, 0x00, byte(constant.RoomMaxPlayers), byte(len(r.players)-1), "", r.players))

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

func (r *tradeRoom) insertItem(tradeSlot byte, plrID int32, item Item) {
	if tradeSlot < 1 || tradeSlot > 9 {
		log.Printf("trade: invalid slot %d from player %d\n", tradeSlot, plrID)
		return
	}

	if _, exists := r.items[plrID][tradeSlot]; exists {
		log.Printf("trade: slot %d already occupied for player %d\n", tradeSlot, plrID)
		return
	}

	r.items[plrID][tradeSlot] = item
	isUser0 := r.players[0].ID == plrID
	r.players[0].Send(packetRoomTradePutItem(tradeSlot, !isUser0, item))
	r.players[1].Send(packetRoomTradePutItem(tradeSlot, isUser0, item))
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

	if !p1.canReceiveItems(changes[0].itemsToGive) || !p2.canReceiveItems(changes[1].itemsToGive) {
		r.closeWithReason(constant.MiniRoomTradeInventoryFull, true)
		return
	}

	var undo []func()
	defer func() {
		if r.finalized {
			return
		}
		for i := len(undo) - 1; i >= 0; i-- {
			func(fn func()) {
				defer func() { _ = recover() }()
				fn()
			}(undo[i])
		}
	}()

	for _, it := range changes[0].itemsToGive {
		err, gi := p1.GiveItem(it)
		if err != nil {
			log.Printf("Trade error: failed to give item %v to %s: %v", it.ID, p1.Name, err)
			r.closeWithReason(constant.MiniRoomTradeInventoryFull, true)
			return
		}

		undo = append(undo, func() {
			if _, err := p1.takeItem(gi.ID, gi.slotID, gi.amount, gi.invID); err != nil {
				log.Printf("Trade rollback warning: failed to remove item %v from %s: %v", gi.ID, p1.Name, err)
			}
		})
	}

	for _, it := range changes[1].itemsToGive {
		err, gi := p2.GiveItem(it)
		if err != nil {
			log.Printf("Trade error: failed to give item %v to %s: %v", it.ID, p2.Name, err)
			r.closeWithReason(constant.MiniRoomTradeInventoryFull, true)
			return
		}

		undo = append(undo, func() {
			if _, err := p2.takeItem(gi.ID, gi.slotID, gi.amount, gi.invID); err != nil {
				log.Printf("Trade rollback warning: failed to remove item %v from %s: %v", gi.ID, p2.Name, err)
			}
		})
	}

	if changes[0].mesosChange > 0 {
		mc := changes[0].mesosChange
		p1.giveMesos(mc)
		undo = append(undo, func() { p1.giveMesos(-mc) })
	}

	if changes[1].mesosChange > 0 {
		mc := changes[1].mesosChange
		p2.giveMesos(mc)
		undo = append(undo, func() { p2.giveMesos(-mc) })
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
				if err, _ := player.GiveItem(item); err != nil {
					log.Println("tradeRoom rollback failed:", err)
				}
				delete(bag, slot)
			}
		}
	}

	r.finalized = true
}

func packetRoomShowWindow(roomType, boardType, maxPlayers, roomSlot byte, roomTitle string, players []*Player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomPacketShowWindow)
	p.WriteByte(roomType)
	p.WriteByte(maxPlayers)
	p.WriteByte(roomSlot)

	for i, v := range players {
		p.WriteByte(byte(i))
		p.Append(v.displayBytes())
		p.WriteInt32(0) // not sure what this is - memory card game seed? board settings?
		p.WriteString(v.Name)
	}

	p.WriteByte(0xFF)

	if roomType == constant.MiniRoomTypeTrade {
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

func packetRoomJoin(roomType, roomSlot byte, plr *Player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomPacketJoin)
	p.WriteByte(roomSlot)
	p.Append(plr.displayBytes())
	p.WriteInt32(0) //?
	p.WriteString(plr.Name)

	if roomType == constant.MiniRoomTypeTrade || roomType == constant.MiniRoomTypePlayerShop {
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
	p.WriteByte(constant.RoomPacketLeave)
	p.WriteByte(roomSlot)
	p.WriteByte(leaveCode) // 2 - trade cancelled, 6 - trade success

	return p
}

func packetRoomYellowChat(msgType byte, name string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.MiniRoomChat)
	p.WriteByte(constant.RoomChatTypeNotice)
	// expelled: 0, x's turn: 1, forfeit: 2, handicap request: 3, left: 4,
	// called to leave/expeled: 5, cancelled leave: 6, entered: 7, can't start lack of mesos:8
	// has matched cards: 9
	p.WriteByte(msgType)
	p.WriteString(name)

	return p
}

func packetRoomReady() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomReadyButtonPressed)

	return p
}

func packetRoomChat(sender, message string, roomSlot byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.MiniRoomChat)
	p.WriteByte(constant.RoomChatTypeChat)
	p.WriteByte(roomSlot)
	p.WriteString(sender + " : " + message)

	return p
}

func packetRoomUnready() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomUnready)

	return p
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

func packetRoomInvite(roomType byte, name string, roomID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomPacketInvite)
	p.WriteByte(roomType)
	p.WriteString(name)
	p.WriteInt32(roomID)

	return p
}

func packetRoomInviteResult(resultCode byte, name string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomPacketInviteResult)
	p.WriteByte(resultCode)
	p.WriteString(name)

	return p
}

func packetRoomShowAccept() mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomPacketShowAccept)

	return p
}

func packetRoomEnterErrorMsg(errorCode byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoom)
	p.WriteByte(constant.RoomPacketShowWindow)
	p.WriteByte(0x00)
	p.WriteByte(errorCode)

	return p
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

func packetRoomClosed() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterClosed)
}

func packetRoomFull() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterFull)
}

func packetRoomBusy() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterBusy)
}

func packetRoomNotAllowedWhenDead() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterNotAllowedDead)
}

func packetRoomNotAllowedDuringEvent() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterNotAllowedEvent)
}

func packetRoomThisCharacterNotAllowed() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterThisCharNotAllow)
}

func packetRoomNoTradeAtm() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterNoTradeATM)
}

func packetRoomMiniRoomNotHere() mpacket.Packet {
	return packetRoomEnterErrorMsg(0x08)
}

func packetRoomTradeRequireSameMap() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterTradeSameMap)
}

func packetRoomcannotCreateMiniroomHere() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterCannotCreateHere)
}

func packetRoomCannotStartGameHere() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterCannotStartHere)
}

func packetRoomPersonalStoreFMOnly() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterStoreFMOnly)
}
func packetRoomGarbageMsgAboutFloorInFm() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterGarbageFloorFM)
}

func packetRoomMayNotEnterStore() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterMayNotEnterStore)
}

func packetRoomStoreMaintenance() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterStoreMaint)
}

func packetRoomCannotEnterTournament() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.MiniRoomEnterUnableEnterTournament)
}

func packetRoomGarbageTradeMsg() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.RoomEnterGarbageTradeMsg)
}

func packetRoomNotEnoughMesos() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.MiniRoomEnterNotEnoughMesos)
}

func packetRoomIncorrectPassword() mpacket.Packet {
	return packetRoomEnterErrorMsg(constant.MiniRoomEnterIncorrectPassword)
}
