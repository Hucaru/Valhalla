package server

import (
	"fmt"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
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

	// conn.Send(script.PacketChatSelection(npcData.ID(), "#e#h ##n the NPC #bchat #dsystem #gis #r#enot #k#nimplemented\r\n#L1#Option 1 #l#L2#Option 2 #l#L3#Option 3 #l"))
	conn.Send(script.PacketChatStyleWindow(npcData.ID(), "styles", []int32{31050, 31040, 31000, 31060, 31090, 31020, 31130, 31120, 31140, 31330, 31010}))

	// Start npc session
}

func (server *ChannelServer) npcChatContinue(conn mnet.Client, reader mpacket.Reader) {
	msgType := reader.ReadByte()
	switch msgType {
	case 0: // next/back
		fmt.Println("next/back:", reader.ReadByte())
	case 1: // yes/no, ok
		fmt.Println("yes/no:", reader.ReadByte())
	case 2: // string input
		fmt.Println("text input - input made:", reader.ReadBool(), "text:", reader.ReadInt16())
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

	// Check npc session active for user, if not return
}
