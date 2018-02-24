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

			// show any summoned monsters

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
	mobID := reader.ReadUint32()

	mapID := conn.GetCharacter().GetCurrentMap()

	moveID := reader.ReadUint16()
	useSkill := reader.ReadByte()
	skill := reader.ReadByte()

	reader.ReadInt16()
	reader.ReadInt16()

	var mp uint16

	mobsMapMutex.Lock()
	for i := range mobsMap[mapID] {
		if uint32(i) == mobID {
			mobsMap[mapID][i].lifeData.X = reader.ReadInt16()
			mobsMap[mapID][i].lifeData.Y = reader.ReadInt16()
			mp = mobsMap[mapID][i].mobData.Mp
			break
		}
	}
	mobsMapMutex.Unlock()

	server.SendPacketToMap(mapID, moveMob(mobID, useSkill, skill, reader.GetBuffer()[13:]))
	conn.Write(controlMoveMob(mobID, moveID, useSkill, mp))

	mobsMapMutex.RLock()
	mob := mobsMap[mapID][mobID]
	mobsMapMutex.RUnlock()

	conn.Write(controlMob(mobID, mob.lifeData, false))
}

func validDistance(charX int16, charY int16, mobX int16, mobY int16) bool {
	maxPos := 40000000000
	deltaX := mobX - charX
	deltaY := mobY - charY

	scalarDelta := deltaX*deltaX + deltaY*deltaY

	return int(scalarDelta) < maxPos
}
