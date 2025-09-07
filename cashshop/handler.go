package cashshop

import (
	"log"
	"os"

	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (server *Server) HandleClientPacket(conn mnet.Client, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case opcode.RecvPing:
	default:
		log.Println("UNKNOWN CASHSHOP PACKET:", reader)
	}
}

func (server *Server) HandleServerPacket(conn mnet.Server, reader mpacket.Reader) {
	switch reader.ReadByte() {
	case opcode.ChannelPlayerConnect:

	case opcode.ChannePlayerDisconnect:
		// Anything?
	case opcode.CashShopOk:
		// rock and roll
	case opcode.CashShopBad:
		os.Exit(1)
	default:
		log.Println("UNKNOWN SERVER PACKET:", reader)
	}
}

func (server Server) playerEnterCashShop(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	plr.Send(packetCashShopSet(plr))
	plr.Send(packetCashShopUpdateAmounts(plr.GetNX(), plr.GetMaplePoints()))
	plr.Send(packetCashShopWishList(nil, true))

}
