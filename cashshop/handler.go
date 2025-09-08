package cashshop

import (
	"log"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (server *Server) HandleClientPacket(conn mnet.Client, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case opcode.RecvPing:
	case opcode.RecvClientMigrate:
		server.handlePlayerConnect(conn, reader)
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

	for i := byte(0); i < total; i++ {
		server.channels[i].IP = reader.ReadBytes(4)
		server.channels[i].Port = reader.ReadInt16()
	}
}
func (server *Server) handlePlayerConnect(conn mnet.Client, reader mpacket.Reader) {
	charID := reader.ReadInt32()

	// Fetch channelID, migrationID and accountID in a single query
	var (
		migrationID byte
		channelID   int8
		accountID   int32
	)
	err := common.DB.QueryRow(
		"SELECT channelID, migrationID, accountID FROM characters WHERE ID=?",
		charID,
	).Scan(&channelID, &migrationID, &accountID)
	if err != nil {
		log.Println("playerConnect query error:", err)
		return
	}

	if migrationID != 50 {
		log.Println("cashshop:playerConnect: invalid migrationID:", migrationID)
		return
	}

	conn.SetAccountID(accountID)

	var adminLevel int
	err = common.DB.QueryRow("SELECT adminLevel FROM accounts WHERE accountID=?", conn.GetAccountID()).Scan(&adminLevel)

	if err != nil {
		log.Println(err)
		return
	}

	conn.SetAdminLevel(adminLevel)

	_, err = common.DB.Exec("UPDATE characters SET migrationID=? WHERE ID=?", -1, charID)

	if err != nil {
		log.Println(err)
		return
	}

	plr := channel.LoadPlayerFromID(charID, conn)

	server.players = append(server.players, &plr)

	server.world.Send(internal.PacketChannelPlayerConnected(plr.ID, plr.Name, server.id, false, 0, 0))

	plr.Send(packetCashShopSet(&plr))
	plr.Send(packetCashShopUpdateAmounts(plr.GetNX(), plr.GetMaplePoints()))
	plr.Send(packetCashShopWishList(nil, true))
}

func (server *Server) handlePlayerDisconnect(conn mnet.Server, reader mpacket.Reader) {
	playerID := reader.ReadInt32()
	name := reader.ReadString(reader.ReadInt16())
	_ = reader.ReadInt32()

	plr, err := server.players.getFromID(playerID)
	if err != nil {
		log.Println(err)
		return
	}
	err = server.players.removeFromConn(plr.Conn)
	if err != nil {
		return
	}

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
