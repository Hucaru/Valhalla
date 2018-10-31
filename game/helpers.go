package game

import (
	"math/rand"

	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/packets"
	"github.com/Hucaru/Valhalla/types"

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

func GetPlayerFromConn(conn mnet.MConnChannel) Player {
	return players[conn]
}

func GetPlayersFromMapID(id int32) []Player {
	players := []Player{}

	for _, v := range players {
		if v.char.CurrentMap == id {
			players = append(players, v)
		}
	}

	return players
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

func GetMobFromMapAndSpawnID(mapID, spawnID int32) *types.Mob {
	for i, m := range maps[mapID].mobs {
		if m.SpawnID == spawnID {
			return &maps[mapID].mobs[i]
		}
	}

	return nil
}

func findNewControllerExcept(mapID int32, conn mnet.MConnChannel) mnet.MConnChannel {
	for c, v := range players {
		if v.char.CurrentMap == mapID {
			if c == conn {
				continue
			}

			return c
		}
	}

	return nil
}

func MobAssignNewController(conn mnet.MConnChannel, mob *types.Mob) {
	newController := findNewControllerExcept(players[conn].char.CurrentMap, conn)

	if newController == nil {
		return
	}

	conn.Send(packets.MobEndControl(*mob))
	mob.Controller = newController
	mob.Controller.Send(packets.MobControl(*mob))
}
