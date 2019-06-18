package game

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

type instance struct {
	id      int
	fieldID int32
	npcs    []npc
	players []mnet.Client
	server  *ChannelServer
}

func (i *instance) addPlayer(conn mnet.Client) error {
	for _, npc := range i.npcs {
		conn.Send(packetNpcShow(npc))
		conn.Send(packetNpcSetController(npc.spawnID, true))
	}

	for _, other := range i.players {
		other.Send(packetMapPlayerEnter(*i.server.sessions[conn]))
		conn.Send(packetMapPlayerEnter(*i.server.sessions[other]))
	}

	i.players = append(i.players, conn)
	return nil
}

func (i *instance) removePlayer(conn mnet.Client) error {
	return nil
}

func (i *instance) delete() error {
	return nil
}

func (i instance) send(p mpacket.Packet) error {
	return nil
}

func (i instance) sendExcept(p mpacket.Packet, exception mnet.Client) error {
	return nil
}
