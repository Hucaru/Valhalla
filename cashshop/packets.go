package cashshop

import (
	"log"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

func (server *Server) playerCashShopPurchase(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)
	if err != nil {
		return
	}

	plrNX := plr.GetNX()
	plrMaplePoints := plr.GetMaplePoints()

	sub := reader.ReadByte()
	switch sub {
	case 0x02: // buy item
		currencySel := reader.ReadByte() // 0x00 = NX, 0x01 = Maple Points
		sn := reader.ReadInt32()

		commodity, ok := nx.GetCommodity(sn)
		if !ok || commodity.ItemID == 0 {
			// Unknown SN
			plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))
			return
		}

		if commodity.OnSale == 0 || commodity.Price <= 0 {
			plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))
			return
		}

		// block super megaphones
		if commodity.ItemID >= 5390000 && commodity.ItemID <= 5390002 {
			plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))
			return
		}

		// Determine quantity
		count := int16(1)
		if commodity.Count > 0 {
			count = int16(commodity.Count)
		}

		// Check funds
		price := commodity.Price
		switch currencySel {
		case 0x00:
			if plrNX < price {
				plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))
				return
			}
		case 0x01:
			if plrMaplePoints < price {
				plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))
				return
			}
		default:
			plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))
			return
		}

		newItem, e := channel.CreateItemFromID(commodity.ItemID, count)
		if e != nil {
			plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))
			return
		}

		if err := plr.GiveItem(newItem); err != nil {
			plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))
			return
		}

		switch currencySel {
		case 0x00:
			plrNX -= price
			plr.SetNX(plrNX)
		case 0x01:
			plrMaplePoints -= price
			plr.SetMaplePoints(plrMaplePoints)
		}

		plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))

	default:
		log.Println("Unknown Cash Shop Packet: ", reader)
	}

}
