package cashshop

import (
	"log"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (server *Server) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.ChannelPlayerConnect:

	case opcode.ChannePlayerDisconnect:
		server.handlePlayerDisconnect(conn, reader)
	case opcode.ChannelConnectionInfo:
		server.handleChannelConnectionInfo(conn, reader)
	case opcode.CashShopOk:
		server.handleWorldConnection(conn, reader)
	case opcode.CashShopBad:
		log.Panicln("CashShop unable to connect to world")
	default:
		log.Println("UNKNOWN SERVER PACKET:", reader)
	}
}

func (server Server) handleWorldConnection(conn mnet.Server, reader mpacket.Reader) {
	reader.ReadString(reader.ReadInt16())
	log.Printf("Connected to world %s\n", conn)
}

func (server *Server) handleChannelConnectionInfo(conn mnet.Server, reader mpacket.Reader) {
	total := reader.ReadByte()

	for i := range total {
		server.channels[i].IP = reader.ReadBytes(4)
		server.channels[i].Port = reader.ReadInt16()
	}
}

func (server *Server) handlePlayerDisconnect(conn mnet.Server, reader mpacket.Reader) {
	playerID := reader.ReadInt32()
	name := reader.ReadString(reader.ReadInt16())
	_ = reader.ReadInt32()

	plr, err := server.players.GetFromID(playerID)
	if err != nil {
		log.Println(err)
		return
	}
	err = server.players.RemoveFromConn(plr.Conn)
	if err != nil {
		return
	}

	if _, err := common.DB.Exec("UPDATE characters SET inCashShop=0 WHERE ID=?", playerID); err != nil {
		return
	}

	log.Printf("Player %s (ID: %d) disconnected from CashShop\n", name, playerID)
}
