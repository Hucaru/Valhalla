package handlers

import (
	"log"

	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/login"
	"github.com/Hucaru/gopacket"
)

// HandleLoginPacket -
func HandleLoginPacket(conn *connection.ClientLoginConn, reader gopacket.Reader) {
	switch reader.ReadByte() {
	case constants.RECV_RETURN_TO_LOGIN_SCREEN:
		login.HandleReturnToLoginScreen(conn, reader)

	case constants.RECV_LOGIN_REQUEST:
		login.HandleLoginRequest(conn, reader)

	case constants.RECV_LOGIN_CHECK_LOGIN:
		login.HandleGoodLogin(conn, reader)

	case constants.RECV_LOGIN_WORLD_SELECT:
		login.HandleWorldSelect(conn, reader)

	case constants.RECV_LOGIN_CHANNEL_SELECT:
		login.HandleChannelSelect(conn, reader)

	case constants.RECV_LOGIN_NAME_CHECK:
		login.HandleNameCheck(conn, reader)

	case constants.RECV_LOGIN_NEW_CHARACTER:
		login.HandleNewCharacter(conn, reader)

	case constants.RECV_LOGIN_DELETE_CHAR:
		login.HandleDeleteCharacter(conn, reader)

	case constants.RECV_LOGIN_SELECT_CHARACTER:
		login.HandleSelectCharacter(conn, reader)

	default:
		log.Println("UNKNOWN LOGIN PACKET:", reader)
	}

}
