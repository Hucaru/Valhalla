package game

import (
	"strings"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (server *ChannelServer) chatSendAll(conn mnet.Client, reader mpacket.Reader) {
	msg := reader.ReadString(reader.ReadInt16())

	if strings.Index(msg, "/") == 0 && conn.GetAdminLevel() > 0 {
		server.gmCommand(conn, msg)
	} else {
		player := server.players[conn]
		char := player.char

		server.fields[char.mapID].send(packetMessageAllChat(char.id, conn.GetAdminLevel() > 0, msg), player.instanceID)
	}
}

func (server *ChannelServer) gmCommand(conn mnet.Client, msg string) {

}
