package handlers

import (
	"log"

	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/maplepacket"
)

// HandleLoginPacket -
func HandleLoginPacket(conn *connection.Login, reader maplepacket.Reader) {
	switch reader.ReadByte() {
	case constants.RecvReturnToLoginScreen:
		handleReturnToLoginScreen(conn, reader)

	case constants.RecvLoginRequest:
		handleLoginRequest(conn, reader)

	case constants.RecvLoginCheckLogin:
		handleGoodLogin(conn, reader)

	case constants.RecvLoginWorldSelect:
		handleWorldSelect(conn, reader)

	case constants.RecvLoginChannelSelect:
		handleChannelSelect(conn, reader)

	case constants.RecvLoginNameCheck:
		handleNameCheck(conn, reader)

	case constants.RecvLoginNewCharacter:
		handleNewCharacter(conn, reader)

	case constants.RecvLoginDeleteChar:
		handleDeleteCharacter(conn, reader)

	case constants.RecvLoginSelectCharacter:
		handleSelectCharacter(conn, reader)

	default:
		log.Println("UNKNOWN LOGIN PACKET:", reader)
	}

}
