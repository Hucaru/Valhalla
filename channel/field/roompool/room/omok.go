package room

const roomTypeOmok = 0x01

// Omok behaviours
type Omok interface {
	PlacePiece(int32, int32, byte, player) bool
	RequestUndo(player)
	RequestUndoResult(bool, player)
}

type omok struct {
	game

	board [15][15]byte

	p1History [2][2]int32
	p2History [2][2]int32

	p1Plays int
	p2Plays int
}

// NewOmok returns a new Omok
func NewOmok(id int32, name, password string, boardType byte) Room {
	r := room{id: id, roomType: roomTypeOmok}
	g := game{room: r, name: name, password: password, boardType: boardType, ownerStart: false}
	return &omok{game: g}
}

// PlacePiece on game board
func (r *omok) PlacePiece(x, y int32, piece byte, plr player) bool {
	if x > 14 || y > 14 || x < 0 || y < 0 {
		return false
	}

	// Turns are out of sync with client probably due to hacking
	if r.p1Turn && plr.Conn() != r.players[0].Conn() {
		r.players[1].Send(packetRoomOmokInvalidPlaceMsg())
	} else if !r.p1Turn && plr.Conn() != r.players[1].Conn() {
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

		if r.Closed() { // If owner exit as part of game leave
			return false
		}

		return true
	}

	r.p1Turn = !r.p1Turn

	return false
}

// RequestUndo to the last move the player made
func (r *omok) RequestUndo(plr player) {
	for i, v := range r.players {
		if v.Conn() != plr.Conn() {
			if (i == 0 && r.p1Plays == 0) || (i == 1 && r.p2Plays == 0) {
				return
			}

			v.Send(packetRoomRequestUndo())
			return
		}
	}
}

// RequestUndoResult is the choice the other player made to the request
func (r *omok) RequestUndoResult(undo bool, plr player) {
	if undo {
		for i, v := range r.players {
			if v.Conn() != plr.Conn() {
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
			if v.Conn() != plr.Conn() {
				v.Send(packetRoomRejectUndo())
				return
			}
		}
	}
}

// Start button pressed
func (r *omok) Start() {
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
