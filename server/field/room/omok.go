package room

const roomTypeOmok = 0x01

// Omok behaviours
type Omok interface {
	PlacePiece(int32, int32, byte, player) bool
}

type omok struct {
	game

	board        [15][15]byte
	previousTurn [2][2]int32
}

const maxPlayers = 2

// NewOmok returns a new Omok
func NewOmok(id int32, name, password string, boardType byte) Room {
	g := game{name: name, password: password, boardType: boardType, roomType: roomTypeOmok, ownerStart: false}
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
		r.previousTurn[0][0] = x
		r.previousTurn[0][1] = y
	} else {
		r.previousTurn[1][0] = x
		r.previousTurn[1][1] = y
	}

	r.send(packetRoomPlaceOmokPiece(x, y, piece))

	win := checkOmokWin(r.board, piece)
	draw := checkOmokDraw(r.board)

	if win || draw {
		r.gameEnd(draw, false, nil)
		return true
	}

	r.p1Turn = !r.p1Turn

	return false
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
