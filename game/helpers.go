package game

import (
	"math/rand"

	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/packets"

	"github.com/Hucaru/Valhalla/mnet"

	"github.com/Hucaru/Valhalla/nx"
)

func AddPlayer(player Player) {
	players[player.MConnChannel] = player
	player.sendMapItems()
}

func RemovePlayer(conn mnet.MConnChannel) {
	p := players[conn]
	for _, player := range players {
		if player.Char().CurrentMap == p.Char().CurrentMap {
			player.Send(packets.MapPlayerLeft(p.Char().ID))
		}
	}

	delete(players, conn)
}

func GetPlayerFromConn(conn mnet.MConnChannel) Player {
	return players[conn]
}

func SendToMap(mapID int32, p maplepacket.Packet) {

}

func SendToMapExcept(mapID int32, p maplepacket.Packet, conn mnet.MConnChannel) {

}

func GetRandomSpawnPortal(mapID int32) (nx.Portal, byte) {
	portals := []nx.Portal{}
	inds := []int{}

	for i, p := range nx.Maps[mapID].Portals {
		if p.IsSpawn {
			portals = append(portals, p)
			inds = append(inds, i)
		}
	}

	ind := rand.Intn(len(portals))
	return portals[ind], byte(inds[ind])
}
