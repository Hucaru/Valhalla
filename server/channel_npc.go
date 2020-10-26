package server

import (
	"fmt"
	"log"
	"strconv"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/server/item"
	"github.com/Hucaru/Valhalla/server/script"
)

func (server *ChannelServer) npcMovement(conn mnet.Client, reader mpacket.Reader) {
	data := reader.GetRestAsBytes()
	id := reader.ReadInt32()

	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[plr.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(plr.InstanceID())

	if err != nil {
		return
	}

	inst.LifePool().NpcAcknowledge(id, plr, data)
}

func (server *ChannelServer) npcChatStart(conn mnet.Client, reader mpacket.Reader) {
	npcSpawnID := reader.ReadInt32()

	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[plr.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(plr.InstanceID())

	if err != nil {
		return
	}

	npcData, err := inst.LifePool().GetNPCFromSpawnID(npcSpawnID)

	if err != nil {
		return
	}

	// Start npc session
	var controller *script.NpcChatController

	if program, ok := server.npcScriptStore.Get(strconv.Itoa(int(npcData.ID()))); ok {
		controller, err = script.CreateNewNpcController(npcData.ID(), conn, program, server.warpPlayer, server.fields)
	} else {
		if program, ok := server.npcScriptStore.Get("default"); ok {
			controller, err = script.CreateNewNpcController(npcData.ID(), conn, program, server.warpPlayer, server.fields)
		}
	}

	if controller == nil {
		log.Println("Unable to find npc script for:", npcData.ID(), ".... default.js not found")
		return
	}

	if err != nil {
		log.Println("script init:", err)
	}

	server.npcChat[conn] = controller
	if controller.Run(plr) {
		delete(server.npcChat, conn)
	}
}

func (server *ChannelServer) npcChatContinue(conn mnet.Client, reader mpacket.Reader) {
	if _, ok := server.npcChat[conn]; !ok {
		return
	}

	controller := server.npcChat[conn]
	controller.State().ClearFlags()

	terminate := false

	msgType := reader.ReadByte()

	switch msgType {
	case 0: // next/back
		value := reader.ReadByte()

		switch value {
		case 0: // back
			controller.State().SetNextBack(false, true)
		case 1: // next
			controller.State().SetNextBack(true, false)
		case 255: // 255/0xff end chat
			terminate = true
		default:
			terminate = true
			log.Println("unknown next/back:", value)
		}
	case 1: // yes/no, ok
		value := reader.ReadByte()

		switch value {
		case 0: // no
			controller.State().SetYesNo(false, true)
		case 1: // yes, ok
			controller.State().SetYesNo(true, false)
		case 255: // 255/0xff end chat
			terminate = true
		default:
			log.Println("unknown yes/no:", value)
		}
	case 2: // string input
		if reader.ReadBool() {
			controller.State().SetTextInput(reader.ReadString(reader.ReadInt16()))
		} else {
			terminate = true
		}
	case 3: // number input
		if reader.ReadBool() {
			controller.State().SetNumberInput(reader.ReadInt32())
		} else {
			terminate = true
		}
	case 4: // select option
		if reader.ReadBool() {
			controller.State().SetOptionSelect(reader.ReadInt32())
		} else {
			terminate = true
		}
	case 5: // style window (no way to discern between cancel button and end chat selection)
		if reader.ReadBool() {
			controller.State().SetOptionSelect(int32(reader.ReadByte()))
		} else {
			terminate = true
		}
	case 6:
		fmt.Println("pet window:", reader)
	default:
		log.Println("Unkown npc chat continue packet:", reader)
	}

	plr, err := server.players.getFromConn(conn)

	if err != nil {
		delete(server.npcChat, conn)
		return
	}

	if terminate || controller.Run(plr) {
		delete(server.npcChat, conn)
	}
}

func (server *ChannelServer) npcShop(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	operation := reader.ReadByte()
	switch operation {
	case 0: // buy
		index := reader.ReadInt16()
		itemID := reader.ReadInt32()
		amount := reader.ReadInt16()

		newItem, err := item.CreateAverageFromID(itemID, amount)

		if err != nil {
			return
		}

		if controller, ok := server.npcChat[conn]; ok {
			goods := controller.State().Goods()

			if int(index) < len(goods) && index > -1 {
				if len(goods[index]) == 1 { // Default price
					item, err := nx.GetItem(itemID)

					if err != nil {
						return
					}

					plr.GiveMesos(-1 * item.Price)
				} else if len(goods[index]) == 2 { // Custom price
					plr.GiveMesos(-1 * goods[index][1])
				} else {
					return // bad shop slice
				}

				plr.GiveItem(newItem, server.db)
				plr.Send(script.PacketShopContinue()) //check if needed
			}

		}
	case 1: // sell
		slotPos := reader.ReadInt16()
		itemID := reader.ReadInt32()
		amount := reader.ReadInt16()

		fmt.Println("Selling:", itemID, "[", slotPos, "], amount:", amount)

		item, err := nx.GetItem(itemID)

		if err != nil {
			return
		}

		invID := getInventoryID(itemID)

		plr.TakeItem(itemID, slotPos, amount, invID, server.db)

		plr.GiveMesos(item.Price)
		plr.Send(script.PacketShopContinue()) // check if needed
	case 3: // exit
		if _, ok := server.npcChat[conn]; ok {
			delete(server.npcChat, conn) // delete here as we need access to shop goods
		}
	default:
		log.Println("Unkown shop operation packet:", reader)
	}
}

func getInventoryID(id int32) byte {
	return byte(id / 1000000)
}
