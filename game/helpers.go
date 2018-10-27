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

	delete(players, conn)

	for _, player := range players {
		if player.Char().CurrentMap == p.Char().CurrentMap {
			player.Send(packets.MapPlayerLeft(p.Char().ID))
		}
	}

}

func GetPlayerFromConn(conn mnet.MConnChannel) Player {
	return players[conn]
}

func SendToMap(mapID int32, p maplepacket.Packet) {
	for _, player := range players {
		if player.Char().CurrentMap == mapID {
			tmp := make(maplepacket.Packet, len(p))
			copy(tmp, p)
			player.Send(tmp)
		}
	}
}

func SendToMapExcept(mapID int32, p maplepacket.Packet, exception mnet.MConnChannel) {
	for conn, player := range players {
		if conn == exception {
			continue
		} else if player.Char().CurrentMap == mapID {
			tmp := make(maplepacket.Packet, len(p))
			copy(tmp, p)
			player.Send(tmp)
		}
	}
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
