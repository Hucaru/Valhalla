package maps

import (
	"github.com/Hucaru/Valhalla/channelServer/mobs"
	"github.com/Hucaru/Valhalla/channelServer/npc"
	"github.com/Hucaru/Valhalla/channelServer/playerConn"
	"github.com/Hucaru/Valhalla/channelServer/server"
	"github.com/Hucaru/Valhalla/common/nx"
	"github.com/Hucaru/gopacket"
)

func RegisterNewPlayer(conn *playerConn.Conn, mapID uint32) {
	server.AddPlayerToMap(conn, mapID)
	displayMapObjects(conn, mapID)
}

func PlayerLeftGame(conn *playerConn.Conn) {
	mapID := conn.GetCharacter().GetCurrentMap()
	server.RemovePlayerFromMap(conn, mapID)
	alertMapPlayerLeft(conn, mapID)
}

func PlayerChangeMap(conn *playerConn.Conn, newMapID uint32) {
	previousMapID := conn.GetCharacter().GetCurrentMap()

	char := conn.GetCharacter()
	char.SetCurrentMap(newMapID)
	char.SetPreviousMap(previousMapID)

	alertMapPlayerLeft(conn, previousMapID)

	server.RemovePlayerFromMap(conn, previousMapID)
	server.AddPlayerToMap(conn, newMapID)

	displayMapObjects(conn, newMapID)
}

func alertMapPlayerLeft(conn *playerConn.Conn, mapID uint32) {
	server.PerformMapReadWork(func(maps map[uint32][]*playerConn.Conn) {
		for _, v := range maps[mapID] {
			if v != conn {
				v.Write(playerLeftField(conn.GetCharacter().GetCharID()))
			}
		}
	})
}

func PlayerMove(conn *playerConn.Conn, p gopacket.Packet) {
	server.SendPacketToMap(conn.GetCharacter().GetCurrentMap(), playerMove(conn.GetCharacter().GetCharID(), p))
}

func displayMapObjects(conn *playerConn.Conn, mapID uint32) {
	// Spawn pet

	server.PerformMapReadWork(func(maps map[uint32][]*playerConn.Conn) {
		for _, v := range maps[mapID] {
			if v != conn {
				v.Write(playerEnterField(conn.GetCharacter())) // send new player enter map to existing player
				conn.Write(playerEnterField(v.GetCharacter())) // send existing player enter map to new player
				// show existing player pet?
			}
		}
	})

	// show npcs
	life := nx.Maps[mapID].Life
	for i, v := range life {
		if v.Npc {
			conn.Write(npc.SpawnNPC(uint32(i), v))
		}
	}

	// send mob data
	mobs.PlayerEnterMap(conn, mapID)

	// show droped items

	// show player shops

	// show omok games

	// show kites

	// if map undergoing weather effect send it
}
