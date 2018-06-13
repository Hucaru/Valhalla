package handlers

import (
	"log"

	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/maplepacket"
)

// HandleLoginPacket -
func HandleLoginPacket(conn *connection.ClientLoginConn, reader maplepacket.Reader) {
	switch reader.ReadByte() {
	case constants.RECV_RETURN_TO_LOGIN_SCREEN:
		handleReturnToLoginScreen(conn, reader)

	case constants.RECV_LOGIN_REQUEST:
		handleLoginRequest(conn, reader)

	case constants.RECV_LOGIN_CHECK_LOGIN:
		handleGoodLogin(conn, reader)

	case constants.RECV_LOGIN_WORLD_SELECT:
		handleWorldSelect(conn, reader)

	case constants.RECV_LOGIN_CHANNEL_SELECT:
		handleChannelSelect(conn, reader)

	case constants.RECV_LOGIN_NAME_CHECK:
		handleNameCheck(conn, reader)

	case constants.RECV_LOGIN_NEW_CHARACTER:
		handleNewCharacter(conn, reader)

	case constants.RECV_LOGIN_DELETE_CHAR:
		handleDeleteCharacter(conn, reader)

	case constants.RECV_LOGIN_SELECT_CHARACTER:
		handleSelectCharacter(conn, reader)

	default:
		log.Println("UNKNOWN LOGIN PACKET:", reader)
	}

}
