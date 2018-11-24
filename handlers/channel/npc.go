package channel

import (
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/game/packet"
)

func npcMovement(conn mnet.MConnChannel, reader maplepacket.Reader) {
	data := reader.GetRestAsBytes()
	id := reader.ReadInt32()

	conn.Send(packet.NPCMovement(data))
	conn.Send(packet.NPCSetController(id, true))
}
