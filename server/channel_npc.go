package server

import (
	"fmt"

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

	// conn.Send(script.PacketChatSelection(npcData.ID(), "#e#h ##n the NPC #bchat #dsystem #gis #r#enot #k#nimplemented\r\n#L1#Option 1 #l#L2#Option 2 #l#L3#Option 3 #l"))
	// conn.Send(script.PacketChatStyleWindow(npcData.ID(), "styles", []int32{31050, 31040, 31000, 31060, 31090, 31020, 31130, 31120, 31140, 31330, 31010}))

	// test script
	const tmp = `
	var state = 0

	function run(npc, player) {
		if (npc.next()) {
			state++
		} else if (npc.back()) {
			state--
		}

		if (state == 0) {
			npc.sendBackNext("first", false, true)
		} else if (state == 1) {
			npc.sendBackNext("second", true, false)
		} else if (state == 2 ) {
			npc.sendOK("finished")
			npc.terminate()
		} else {
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
		fmt.Println("script init:", err)
	}

	server.npcChat[conn] = controller
	if controller.Run(plr) {
		delete(server.npcChat, conn)
		fmt.Println("deleted on first run")
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
		opcode := reader.ReadByte()

		if opcode == 0 { //back
			controller.SetNextBack(false, true)
		} else if opcode == 1 { // next
			controller.SetNextBack(true, false)
		} else if opcode == 0xff { // 255/0xff end chat
			terminate = true
		} else {
			fmt.Println("unknown next/back:", opcode)
		}
	case 1: // yes/no, ok
		fmt.Println("yes/no:", reader.ReadByte())
	case 2: // string input
		fmt.Println("text input - input made:", reader.ReadBool(), "text:", reader.ReadString(reader.ReadInt16()))
		// no input is end chat button
	case 3: // number input
		fmt.Println("number input - input made:", reader.ReadBool(), "number:", reader.ReadInt32())
		// no input is end chat button
	case 4: // select option
		fmt.Println("select option - option selected:", reader.ReadBool(), "option index:", reader.ReadInt32())
		// no selection is end chat button
	case 5:
		fmt.Println("style selection - option selected:", reader.ReadBool(), "option index:", reader.ReadByte())
		// no selection is end chat button
	case 6:
		fmt.Println("pet window:", reader)
	default:
		fmt.Println("Unkown npc chat continue packet:", reader)
	}

	plr, err := server.players.getFromConn(conn)

	if err != nil {
		delete(server.npcChat, conn)
		fmt.Println("deleted in player get error")
		return
	}

	if terminate || controller.Run(plr) {
		delete(server.npcChat, conn)
		fmt.Println("deleted in continue")
	}
}
