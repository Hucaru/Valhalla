package cashshop

import (
	"log"

	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (server *Server) HandleClientPacket(conn mnet.Client, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case opcode.RecvPing:
	case opcode.RecvCashShopPurchase:
		server.playerCashShopPurchase(conn, reader)
	case opcode.RecvChannelUserPortal:
		server.leaveCashShopToChannel(conn, reader)

	default:
		log.Println("UNKNOWN CASHSHOP PACKET:", reader)
	}
}

func (server *Server) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.ChannelPlayerConnect:
		server.handlePlayerConnectedNotifications(conn, reader)
	case opcode.ChannePlayerDisconnect:
		server.handlePlayerDisconnectNotifications(conn, reader)
	case opcode.ChannelConnectionInfo:
	case opcode.CashShopOk:
		server.handleWorldConnection(conn, reader)
	case opcode.CashShopBad:
		log.Panicln("CashShop Unabled to connect to world")
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

	for i := byte(0); i < total; i++ {
		server.channels[i].IP = reader.ReadBytes(4)
		server.channels[i].Port = reader.ReadInt16()
	}
}

func (server *Server) handlePlayerConnectedNotifications(conn mnet.Server, reader mpacket.Reader) {
	playerID := reader.ReadInt32()
	name := reader.ReadString(reader.ReadInt16())
	channelID := reader.ReadByte()
	_ = reader.ReadBool()
	_ = reader.ReadInt32() // mapID
	_ = reader.ReadInt32()

	log.Printf("Player %s (ID: %d) connected to CashShop from Channel %d\n", name, playerID, channelID)
	plr, err := server.players.getFromID(playerID)
	if err != nil {
		log.Println(err)
		return
	}

	plr.ChannelID = channelID

	plr.Send(packetCashShopSet(plr))
	plr.Send(packetCashShopUpdateAmounts(plr.GetNX(), plr.GetMaplePoints()))
	plr.Send(packetCashShopWishList(nil, true))
}

func (server *Server) handlePlayerDisconnectNotifications(conn mnet.Server, reader mpacket.Reader) {
	playerID := reader.ReadInt32()
	name := reader.ReadString(reader.ReadInt16())
	_ = reader.ReadInt32()

	log.Printf("Player %s (ID: %d) disconnected from CashShop\n", name, playerID)
}

func (server *Server) leaveCashShopToChannel(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)
	if err != nil || plr == nil {
		return
	}

	// Persist migration target for the channel server handoff
	if _, err := common.DB.Exec("UPDATE characters SET migrationID=? WHERE ID=?", plr.ChannelID, plr.ID); err != nil {
		log.Println("cashshop: failed to set migrationID:", err)
		return
	}

	ip := server.channels[plr.ChannelID].IP
	port := server.channels[plr.ChannelID].Port

	p := mpacket.CreateWithOpcode(opcode.SendChannelChange)
	p.WriteBool(true)
	p.WriteBytes(ip)
	p.WriteInt16(port)
	conn.Send(p)
}
