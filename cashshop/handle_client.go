package cashshop

import (
	"log"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/common"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/constant"
	"github.com/Hucaru/Valhalla/internal"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
)

func (server *Server) HandleClientPacket(conn mnet.Client, reader mpacket.Reader) {
	op := reader.ReadByte()

	switch op {
	case opcode.RecvPing:
	case opcode.RecvClientMigrate:
		server.handlePlayerConnect(conn, reader)
	case opcode.RecvCashShopOperation:
		server.handleCashShopOperation(conn, reader)
	case opcode.RecvChannelUserPortal:
		server.leaveCashShopToChannel(conn, reader)

	default:
		log.Println("UNKNOWN CASHSHOP PACKET (", op, "):", reader)
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

	server.players.Add(&plr)

	// Load cash shop storage
	storage, err := server.GetOrLoadStorage(accountID)
	if err != nil {
		log.Println("Failed to load cash shop storage for account", accountID, ":", err)
	}

	server.world.Send(internal.PacketChannelPlayerConnected(plr.ID, plr.Name, server.id, false, 0, 0))

	plr.Send(packetCashShopSet(&plr))

	// Send cash shop storage items to player (before wishlist and amounts, matching OpenMG order)
	if storage != nil {
		plr.Send(packetCashShopLoadLocker(storage, accountID, plr.ID))
	}

	//plr.Send(packetCashShopWishList(nil, false))
	plr.Send(packetCashShopUpdateAmounts(plr.GetNX(), plr.GetMaplePoints()))
}

func (server *Server) leaveCashShopToChannel(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.GetFromConn(conn)
	if err != nil || plr == nil {
		return
	}

	var prevChanID int8
	if err := common.DB.QueryRow("SELECT previousChannelID FROM characters WHERE ID=?", plr.ID).Scan(&prevChanID); err != nil {
		log.Println("Failed to fetch previousChannelID:", err)
	}

	targetChan := plr.ChannelID
	if prevChanID >= 0 && int(prevChanID) < len(server.channels) && server.channels[byte(prevChanID)].Port != 0 {
		targetChan = byte(prevChanID)
	}

	if _, err := common.DB.Exec("UPDATE characters SET migrationID=?, inCashShop=0 WHERE ID=?", targetChan, plr.ID); err != nil {
		log.Println("Failed to set migrationID:", err)
		return
	}

	var ip []byte
	var port int16

	if int(targetChan) < len(server.channels) {
		ip = server.channels[targetChan].IP
		port = server.channels[targetChan].Port
	}

	if len(ip) != 4 || port == 0 {
		log.Printf("Target channel %d missing IP/port, searching for fallback...", targetChan)

		log.Println("Sent request to world for updated channel information")
		server.world.Send(internal.PacketCashShopRequestChannelInfo())

		found := false
		for i, ch := range server.channels {
			if len(ch.IP) == 4 && ch.Port != 0 {
				targetChan = byte(i)
				ip = ch.IP
				port = ch.Port
				found = true
				log.Printf("Using fallback channel %d", targetChan)
				break
			}
		}

		if !found {
			log.Println("No valid fallback channels available")
			return
		}
	}

	p := mpacket.CreateWithOpcode(opcode.SendChannelChange)
	p.WriteBool(true)
	p.WriteBytes(ip)
	p.WriteInt16(port)
	conn.Send(p)
}

func (server *Server) handleCashShopOperation(conn mnet.Client, reader mpacket.Reader) {
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

		// Get cash shop storage
		storage, storageErr := server.GetOrLoadStorage(conn.GetAccountID())
		if storageErr != nil {
			log.Println("Failed to get cash shop storage:", storageErr)
			plr.Send(packetCashShopError(opcode.SendCashShopBuyFailed, constant.CashShopErrorUnknown))
			return
		}

		// Add item to storage instead of inventory
		slotIdx, added := storage.addItem(newItem, sn)
		if !added {
			log.Println("Failed to add item to cash shop storage")
			plr.Send(packetCashShopError(opcode.SendCashShopBuyFailed, constant.CashShopErrorExceededNumberOfCashItems))
			return
		}

		// Save storage
		if saveErr := storage.save(); saveErr != nil {
			log.Println("Failed to save cash shop storage:", saveErr)
			plr.Send(packetCashShopError(opcode.SendCashShopBuyFailed, constant.CashShopErrorUnknown))
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

		// Send buy success packet with the specific item that was just added
		addedItem, ok := storage.getItemBySlot(int16(slotIdx + 1))
		if ok {
			plr.Send(packetCashShopBuyDone(*addedItem, conn.GetAccountID(), plr.ID))
		}

	case opcode.RecvCashShopBuyPackage, opcode.RecvCashShopGiftPackage:
		currencySel := reader.ReadByte()
		pkgSN := reader.ReadInt32()

		commodity, ok := nx.GetCommodity(pkgSN)
		if !ok || commodity.Price <= 0 {
			plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))
			return
		}

		pkgMap := nx.GetPackages()
		pkgItems, ok := pkgMap[pkgSN]
		if !ok || len(pkgItems) == 0 {
			// Fallbacks: some data sets key packages by ItemID or Commodity index instead of SN
			if commodity.ItemID != 0 {
				pkgItems, ok = pkgMap[commodity.ItemID]
			}
			if (!ok || len(pkgItems) == 0) && commodity.Index != 0 {
				pkgItems, ok = pkgMap[commodity.Index]
			}
		}
		if !ok || len(pkgItems) == 0 {
			plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))
			return
		}

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

		itemsToGive := make([]channel.Item, 0, len(pkgItems))
		for _, entry := range pkgItems {
			var itemID int32
			count := int16(1)

			if itCommodity, ok := nx.GetCommodity(entry); ok && itCommodity.ItemID != 0 {
				itemID = itCommodity.ItemID
				if itCommodity.Count > 0 {
					count = int16(itCommodity.Count)
				}
			} else {
				// CashPackage.img can list raw item IDs instead of SNs
				itemID = entry
				if snByItem, ok := nx.GetCommoditySNByItemID(itemID); ok {
					if itCommodity, ok := nx.GetCommodity(snByItem); ok && itCommodity.Count > 0 {
						count = int16(itCommodity.Count)
					}
				}
			}

			if itemID == 0 {
				plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))
				return
			}

			newItem, e := channel.CreateItemFromID(itemID, count)
			if e != nil {
				plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))
				return
			}
			itemsToGive = append(itemsToGive, newItem)
		}

		if !plr.CanReceiveItems(itemsToGive) {
			plr.Send(packetCashShopError(opcode.SendCashShopBuyFailed, constant.CashShopErrorCheckFullInventory))
			return
		}

		for _, it := range itemsToGive {
			if err, _ := plr.GiveItem(it); err != nil {
				plr.Send(packetCashShopUpdateAmounts(plrNX, plrMaplePoints))
				return
			}
		}

		switch currencySel {
		case constant.CashShopNX:
			plrNX -= price
			plr.SetNX(plrNX)
		case constant.CashShopMaplePoints:
			plrMaplePoints -= price
			plr.SetMaplePoints(plrMaplePoints)
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

	case opcode.RecvCashShopMoveLtoS:
		cashItemID := reader.ReadInt64()
		_ = reader.ReadByte()
		targetSlot := reader.ReadInt16()

		storage, storageErr := server.GetOrLoadStorage(conn.GetAccountID())
		if storageErr != nil {
			plr.Send(packetCashShopError(opcode.SendCashShopMoveLtoSFailed, constant.CashShopErrorUnknown))
			return
		}

		var foundIdx = -1
		var foundItem *CashShopItem
		for i := range storage.items {
			if storage.items[i].item.ID == 0 {
				continue
			}
			if storage.items[i].cashID == cashItemID {
				foundIdx = i
				foundItem = &storage.items[i]
				break
			}
		}

		if foundIdx == -1 || foundItem == nil {
			plr.Send(packetCashShopError(opcode.SendCashShopMoveLtoSFailed, constant.CashShopErrorUnknown))
			return
		}

		removedItem, ok := storage.removeAt(foundIdx)
		if !ok {
			plr.Send(packetCashShopError(opcode.SendCashShopMoveLtoSFailed, constant.CashShopErrorUnknown))
			return
		}

		item := removedItem.item
		item.SetCashID(cashItemID)
		item.SetCashSN(removedItem.sn)

		err, givenItem := plr.GiveItem(item)
		if err != nil {
			if _, restored := storage.addItem(item, removedItem.sn); !restored {
				log.Println("CRITICAL: Restore to storage failed. Item may be lost. player:", plr.ID, "accountID:", conn.GetAccountID(), "itemID:", item.ID)
			} else {
				if saveErr := storage.save(); saveErr != nil {
					log.Println("Failed to save restored storage:", saveErr)
				}
			}
			plr.Send(packetCashShopError(opcode.SendCashShopMoveLtoSFailed, constant.CashShopErrorCheckFullInventory))
			return
		}

		if saveErr := storage.save(); saveErr != nil {
			plr.Send(packetCashShopError(opcode.SendCashShopMoveLtoSFailed, constant.CashShopErrorUnknown))
			return
		}

		plr.Send(packetCashShopMoveLtoSDone(givenItem, targetSlot))

	case opcode.RecvCashShopMoveStoL:
		// Move from slot (inventory) to locker (storage)
		cashItemID := reader.ReadInt64()
		invType := reader.ReadByte()

		storage, storageErr := server.GetOrLoadStorage(conn.GetAccountID())
		if storageErr != nil {
			log.Println("Failed to get storage:", storageErr)
			plr.Send(packetCashShopError(opcode.SendCashShopMoveStoLFailed, constant.CashShopErrorUnknown))
			return
		}

		item, itemSlot, findErr := plr.GetItemByCashID(invType, cashItemID)
		if findErr != nil {
			plr.Send(packetCashShopError(opcode.SendCashShopMoveStoLFailed, constant.CashShopErrorUnknown))
			return
		}

		expectedInvType := byte(item.ID / 1000000)
		if expectedInvType != invType {
			plr.Send(packetCashShopError(opcode.SendCashShopMoveStoLFailed, constant.CashShopErrorUnknown))
			return
		}

		sn := item.GetCashSN()
		if sn == 0 {
			plr.Send(packetCashShopError(opcode.SendCashShopMoveStoLFailed, constant.CashShopErrorUnknown))
			return
		}

		takenItem, takeErr := plr.TakeItemFromStorage(item.ID, itemSlot, 1, invType)
		if takeErr != nil {
			plr.Send(packetCashShopError(opcode.SendCashShopMoveStoLFailed, constant.CashShopErrorUnknown))
			return
		}

		slotIdx, added := storage.addItemWithCashID(takenItem, sn, cashItemID)
		if !added {
			if err, _ := plr.GiveItem(takenItem); err != nil {
				log.Println("CRITICAL: Failed to return item to player after add failure:", err)
			}
			plr.Send(packetCashShopError(opcode.SendCashShopMoveStoLFailed, constant.CashShopErrorExceededNumberOfCashItems))
			return
		}

		if saveErr := storage.save(); saveErr != nil {
			log.Println("Failed to save storage:", saveErr)
			plr.Send(packetCashShopError(opcode.SendCashShopMoveStoLFailed, constant.CashShopErrorUnknown))
			return
		}

		addedItem, ok := storage.getItemBySlot(int16(slotIdx + 1))
		if ok {
			plr.Send(packetCashShopMoveStoLDone(*addedItem, conn.GetAccountID()))
		} else {
			plr.Send(packetCashShopError(opcode.SendCashShopMoveStoLFailed, constant.CashShopErrorUnknown))
		}

	default:
		log.Println("Unknown Cash Shop Packet(", sub, "): ", reader)
	}

}
