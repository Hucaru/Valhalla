package channelhandlers

import (
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/packets"
)

func npcMovement(conn mnet.MConnChannel, reader maplepacket.Reader) {
	data := reader.GetRestAsBytes()
	id := reader.ReadInt32()

	conn.Send(packets.NPCMovement(data))
	conn.Send(packets.NPCSetController(id, true))
}
