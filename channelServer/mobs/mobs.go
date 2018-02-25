package mobs

import (
	"sync"

	"github.com/Hucaru/Valhalla/channelServer/playerConn"
	"github.com/Hucaru/Valhalla/channelServer/server"
	"github.com/Hucaru/Valhalla/common/nx"
	"github.com/Hucaru/gopacket"
)

type monster struct {
	lifeData   nx.Life
	mobData    nx.Monster
	controller *playerConn.Conn
}

var mobsMap = make(map[uint32][]*monster)
var mobsMapMutex = &sync.RWMutex{}

func PlayerEnterMap(conn *playerConn.Conn, mapID uint32) {
	mobsMapMutex.RLock()
	_, exists := mobsMap[mapID]
	mobsMapMutex.RUnlock()

	if !exists {
		// First time someone has entered the map, load in all the monsters map should have
		newMonster := []*monster{}

		for _, v := range nx.Maps[mapID].Life {
			if v.Mob {
				monster := &monster{}
				monster.mobData = nx.Mob[v.ID]
				monster.lifeData = v
				monster.controller = nil
				newMonster = append(newMonster, monster)
			}

		}

		mobsMapMutex.Lock()
		mobsMap[mapID] = append(mobsMap[mapID], newMonster...)
		mobsMapMutex.Unlock()
	}

	// Display currently alive monsters
	mobsMapMutex.RLock()
	mobs := mobsMap[mapID]
	mobsMapMutex.RUnlock()

	for i, v := range mobs {
		if v.mobData.Hp > 0 {
			conn.Write(showMob(uint32(i), v.lifeData, false))

			// Mob has no controller
			if v.controller == nil {
				conn.Write(controlMob(uint32(i), v.lifeData, false))

				mobsMapMutex.Lock()
				mobsMap[mapID][i].controller = conn
				mobsMapMutex.Unlock()
			}
		}
	}
}

func HandleMovement(reader gopacket.Reader, conn *playerConn.Conn) {
	mapID := conn.GetCharacter().GetCurrentMap()

	mobID := reader.ReadUint32() // add validation to this
	moveID := reader.ReadUint16()
	skillUsed := bool(reader.ReadByte() != 0)
	skill := reader.ReadByte() // skill

	x := reader.ReadInt16() // ? x pos for something?
	y := reader.ReadInt16() // ? y pos for something?

	level := byte(0)
	if skillUsed {
		// Implement mob skills
	}

	var mp uint16

	mobsMapMutex.Lock()
	for i := range mobsMap[mapID] {
		if uint32(i) == mobID {
			mobsMap[mapID][i].lifeData.X = reader.ReadInt16()
			mobsMap[mapID][i].lifeData.Y = reader.ReadInt16()
			mp = mobsMap[mapID][i].mobData.Mp
		}
	}
	mobsMapMutex.Unlock()

	conn.Write(controlAck(mobID, moveID, skillUsed, skill, level, mp))
	server.SendPacketToMap(mapID, moveMob(mobID, skillUsed, skill, x, y, reader.GetBuffer()[13:]), conn)
}

func EndMobControl(conn *playerConn.Conn) {
	mapID := conn.GetCharacter().GetCurrentMap()

	server.MapGetAllPlayers(mapID)

	mobsMapMutex.Lock()
	// Remove them as controller
	for i := range mobsMap[mapID] {
		if mobsMap[mapID][i].controller == conn {
			mobsMap[mapID][i].controller = nil
			conn.Write(endControl(mobsMap[mapID][i].lifeData.ID))
			// find a new controller for mob
		}
	}

	mobsMapMutex.Unlock()
}

func respawnMonster() {
	// if monster valid for map respawn, otherwise remove from slice
}

func SpawnMonster(mapID uint32, mobID uint32) {

}
