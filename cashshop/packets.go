package cashshop

import (
	"log"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

func (server *Server) playerCashShopPurchase(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil {
		return
	}

	plrNX := plr.GetNX()
	plrMaplePoints := plr.GetMaplePoints()

	sub := reader.ReadByte()
	switch sub {
	case opcode.RecvCashShopBuyItem:
		currencySel := reader.ReadByte()
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

		// Determine quantity
		count := int16(1)
		if commodity.Count > 0 {
			count = int16(commodity.Count)
		}

		// Check funds
		price := commodity.Price
		switch currencySel {
		case constant.CashShopNX:
			if plrNX < price {
				plr.Send(packetCashShopError(opcode.SendCashShopBuyFailed, constant.CashShopErrorNotEnoughCash))
				return
			}
		case constant.CashShopMaplePoints:
			if plrMaplePoints < price {
				plr.Send(packetCashShopError(opcode.SendCashShopBuyFailed, constant.CashShopErrorNotEnoughCash))
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

		if err, _ := plr.GiveItem(newItem); err != nil {
			plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))
			return
		}

		switch currencySel {
		case constant.CashShopNX:
			plrNX -= price
			plr.SetNX(plrNX)
		case constant.CashShopMaplePoints:
			plrMaplePoints -= price
			plr.SetMaplePoints(plrMaplePoints)
		default:
			log.Println("Unknown currency type: ", currencySel)
			return
		}

		plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))

	case opcode.RecvCashShopGiftItem:
	case opcode.RecvCashShopUpdateWishlist:
	case opcode.RecvCashShopIncreaseSlots:
		currencySel := reader.ReadByte()
		invType := reader.ReadByte()

		price := int32(4000)

		switch currencySel {
		case constant.CashShopNX:
			if plrNX < price {
				plr.Send(packetCashShopError(opcode.SendCashShopIncSlotCountFailed, constant.CashShopErrorNotEnoughCash))
				return
			}
			if err := plr.IncreaseSlotSize(invType, 4); err != nil {
				plr.Send(packetCashShopError(opcode.SendCashShopIncSlotCountFailed, constant.CashShopErrorUnknown))
				return
			}
			plrNX -= price
			plr.SetNX(plrNX)
		case constant.CashShopMaplePoints:
			if plrMaplePoints < price {
				plr.Send(packetCashShopError(opcode.SendCashShopIncSlotCountFailed, constant.CashShopErrorNotEnoughCash))
				return
			}
			if err := plr.IncreaseSlotSize(invType, 4); err != nil {
				plr.Send(packetCashShopError(opcode.SendCashShopIncSlotCountFailed, constant.CashShopErrorUnknown))
				return
			}
			plrMaplePoints -= price
			plr.SetMaplePoints(plrMaplePoints)
		default:
			log.Println("Unknown currency type: ", currencySel)
			return
		}

		plr.Send(packetCashShopIncreaseInv(invType, plr.GetSlotSize(invType)))
		plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))
	default:
		log.Println("Unknown Cash Shop Packet(", sub, "): ", reader)
	}

}
