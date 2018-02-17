package player

import (
	"github.com/Hucaru/Valhalla/channelServer/handlers/npc"
	"github.com/Hucaru/Valhalla/channelServer/handlers/playerConn"
	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/Valhalla/common/nx"
)

func PlayerSpawnIn(conn *playerConn.Conn, char character.Character, channelID uint32) {
	conn.Write(spawnGame(char, channelID))

	// npc spawn
	life := nx.Maps[char.CurrentMap].Life
	for i, v := range life {
		if v.Npc {
			conn.Write(npc.SpawnNPC(uint32(i), v))
		}
	}
}

func ChangeMap(conn *playerConn.Conn, mapID uint32, channelID uint32, mapPos byte, hp uint16) {
	conn.Write(changeMap(mapID, channelID, mapPos, hp))

	// npc spawn
	life := nx.Maps[mapID].Life
	for i, v := range life {
		if v.Npc {
			conn.Write(npc.SpawnNPC(uint32(i), v))
		}
	}
}
