package game

import (
	"fmt"
	"math/rand"

	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/packets"

	"github.com/Hucaru/Valhalla/mnet"

	"github.com/Hucaru/Valhalla/nx"
)

func AddPlayer(player Player) {
	players[player.MConnChannel] = player
	player.sendMapItems()
	maps[player.char.CurrentMap].addController(player.MConnChannel)
}

func RemovePlayer(conn mnet.MConnChannel) {
	p := players[conn]
	maps[p.char.CurrentMap].removeController(conn)

	delete(players, conn)

	for _, player := range players {
		if player.Char().CurrentMap == p.Char().CurrentMap {
			player.Send(packets.MapPlayerLeft(p.Char().ID))
		}
	}

}

func GetPlayerFromConn(conn mnet.MConnChannel) (Player, error) {
	if val, ok := players[conn]; ok {
		return val, nil
	}

	return Player{}, fmt.Errorf("Player from connection %s not found", conn)
}

func GetPlayerFromID(id int32) (Player, error) {
	for _, p := range players {
		if p.Char().ID == id {
			return p, nil
		}
	}

	return Player{}, fmt.Errorf("Player ID %i not found", id)
}

func GetPlayerFromName(name string) (Player, error) {
	for _, p := range players {
		if p.Char().Name == name {
			return p, nil
		}
	}

	return Player{}, fmt.Errorf("Player name %s not found", name)
}

func GetPlayersFromMapID(id int32) []Player {
	playerList := []Player{}

	for _, v := range players {
		if v.char.CurrentMap == id {
			playerList = append(playerList, v)
		}
	}

	return playerList
}

func GetMapFromID(id int32) *GameMap {
	if _, ok := maps[id]; ok {
		return maps[id]
	}

	return nil
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
