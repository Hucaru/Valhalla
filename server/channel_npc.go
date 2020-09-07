package server

import (
	"fmt"
	"log"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/script"
	"github.com/dop251/goja"
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

	// test script
	const tmp = `
	var state = 0
	var styles = [31050, 31040, 31000, 31060, 31090, 31020, 31130, 31120, 31140, 31330, 31010]
	var goods = [ [1332020],[1332020, 1],[1332009, 0] ]

	function run(npc, player) {
		if (npc.next()) {
			state++
		} else if (npc.back()) {
			state--
		}

		if (state == 2) {
			if (npc.yes()) {
				state = 3
			} else if (npc.no()) {
				state = 4
			}
		} else if (state == 4) {
			if (npc.selection() == 1) {
				state = 0
			} else if (npc.selection() == 2) {
				state = 5
			} else if (npc.selection() == 3) {
				state = 6
			} else if (npc.selection() == 4) {
				state = 7
			} else if (npc.selection() == 5) {
				state = 8
			}
		}

		switch(state) {
		case 0:
			npc.sendBackNext("first", false, true)
			break
		case 1:
			npc.sendBackNext("second", true, false)
			break
		case 2:
			npc.sendYesNo("finished")
			break
		case 3:
			npc.sendOK("selection:" + npc.selection() + ", input number:" + npc.inputNumber() + ", input text: " + npc.inputString())
			npc.terminate()
			break
		case 4:
			npc.sendSelection("Select from one of the following:\r\n#L1#Back to start #l\r\n#L2#Styles#l\r\n#L3#Input number#l#L4#Input text#l\r\n#L5#Shop#l")
			break
		case 5:
			npc.sendStyles("Select from the following", styles)
			state = 3
			break
		case 6:
			npc.sendInputNumber("Input a number:", 100, 0, 100)
			state = 3
			break
		case 7:
			npc.sendInputText("Input text:", "default", 0, 100)
			state = 3
			break
		case 8:
			npc.sendShop(goods)
			break
		default:
			npc.sendOK("state " + state)
			npc.terminate()
		}
	}
	`
	npcProgram, err := goja.Compile("npc", tmp, false)

	if err != nil {
		fmt.Println("script compile error:", err)
		return
	}

	// Start npc session
	controller, err := script.CreateNewNpcController(npcData.ID(), conn, npcProgram)

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
	controller.ClearFlags()

	terminate := false

	msgType := reader.ReadByte()

	switch msgType {
	case 0: // next/back
		value := reader.ReadByte()

		switch value {
		case 0: // back
			controller.SetNextBack(false, true)
		case 1: // next
			controller.SetNextBack(true, false)
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
			controller.SetYesNo(false, true)
		case 1: // yes, ok
			controller.SetYesNo(true, false)
		default:
			log.Println("unknown yes/no:", value)
		}
	case 2: // string input
		if reader.ReadBool() {
			controller.SetTextInput(reader.ReadString(reader.ReadInt16()))
		} else {
			terminate = true
		}
	case 3: // number input
		if reader.ReadBool() {
			controller.SetNumberInput(reader.ReadInt32())
		} else {
			terminate = true
		}
	case 4: // select option
		if reader.ReadBool() {
			controller.SetOptionSelect(reader.ReadInt32())
		} else {
			terminate = true
		}
	case 5: // style window (no way to discern between cancel button and end chat selection)
		if reader.ReadBool() {
			controller.SetOptionSelect(int32(reader.ReadByte()))
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
	operation := reader.ReadByte()
	switch operation {
	case 0: // buy
		index := reader.ReadInt16()
		itemID := reader.ReadInt32()
		amount := reader.ReadInt16()
		fmt.Println("Buying:", itemID, "[", index, "], amount:", amount)
	case 1: // sell
		slotPos := reader.ReadInt16()
		itemID := reader.ReadInt32()
		amount := reader.ReadInt16()
		fmt.Println("Selling:", itemID, "[", slotPos, "], amount:", amount)
	default:
		log.Println("Unkown shop operation packet:", reader)
	}
}
