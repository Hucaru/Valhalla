package channel

import (
	"math/rand"
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/packets"

	"github.com/Hucaru/Valhalla/interop"
)

type mapleMap struct {
	forcedReturn int32
	returnMap    int32
	mobRate      float64
	isTown       bool
	portals      []maplePortal
	mutex        *sync.RWMutex
	players      []interop.ClientConn
}

func (m *mapleMap) GetMobRate() float64 {
	m.mutex.RLock()
	result := m.mobRate
	m.mutex.RUnlock()

	return result
}

func (m *mapleMap) GetReturnMap() int32 {
	m.mutex.RLock()
	result := m.returnMap
	m.mutex.RUnlock()

	return result
}

func (m *mapleMap) SetReturnMap(mapID int32) {
	m.mutex.Lock()
	m.returnMap = mapID
	m.mutex.Unlock()
}

func (m *mapleMap) GetPortals() []maplePortal {
	m.mutex.RLock()
	result := m.portals
	m.mutex.RUnlock()

	return result
}

func (m *mapleMap) AddPortal(portal maplePortal) {
	m.mutex.Lock()
	m.portals = append(m.portals, portal)
	m.mutex.Unlock()
}

func (m *mapleMap) GetPlayers() []interop.ClientConn {
	m.mutex.RLock()
	result := m.players
	m.mutex.RUnlock()

	return result
}

func (m *mapleMap) AddPlayer(player interop.ClientConn) {
	m.playerEnterMap(player)

	m.mutex.Lock()
	m.players = append(m.players, player)
	m.mutex.Unlock()
}

func (m *mapleMap) RemovePlayer(player interop.ClientConn) {
	index := -1

	m.mutex.RLock()
	for i, v := range m.players {
		if v == player {
			index = i
			break
		}
	}
	m.mutex.RUnlock()

	if index < 0 {
		return
	}

	m.mutex.Lock()
	m.players = append(m.players[:index], m.players[index+1:]...)
	m.mutex.Unlock()

	m.playerLeaveMap(player)
}

func (m *mapleMap) SendPacket(packet maplepacket.Packet) {
	if len(packet) > 0 {
		m.mutex.RLock()
		for _, player := range m.players {
			player.Write(packet)
		}
		m.mutex.RUnlock()
	}
}

func (m *mapleMap) SendPacketExcept(packet maplepacket.Packet, conn interop.ClientConn) {
	if len(packet) > 0 {
		m.mutex.RLock()
		for _, player := range m.players {
			if player == conn {
				continue
			}
			player.Write(packet)
		}
		m.mutex.RUnlock()
	}
}

func (m *mapleMap) GetRandomSpawnPortal() (maplePortal, byte) {
	var portals []maplePortal
	for _, portal := range m.GetPortals() {
		if portal.GetIsSpawn() {
			portals = append(portals, portal)
		}
	}
	rand.Seed(time.Now().UnixNano())
	pos := rand.Intn(len(portals))
	return portals[pos], byte(pos)
}

func (m *mapleMap) playerEnterMap(conn interop.ClientConn) {
	for _, other := range m.GetPlayers() {
		Players.OnCharacterFromConn(conn, func(char *MapleCharacter) {
			other.Write(packets.MapPlayerEnter(char.Character))
		})

		Players.OnCharacterFromConn(other, func(char *MapleCharacter) {
			conn.Write(packets.MapPlayerEnter(char.Character))
		})
	}

	Players.OnCharacterFromConn(conn, func(char *MapleCharacter) {
		for _, npc := range NPCs.GetNpcs(char.GetCurrentMap()) {
			npc.Show(conn)
		}

		Mobs.OnMobs(char.GetCurrentMap(), func(mob *MapleMob) {
			if mob.GetController() == nil {
				mob.SetController(conn, false)
			}

			mob.Show(conn)
		})
	})
}

func (m *mapleMap) playerLeaveMap(conn interop.ClientConn) {
	Players.OnCharacterFromConn(conn, func(char *MapleCharacter) {
		for _, npc := range NPCs.GetNpcs(char.GetCurrentMap()) {
			npc.Hide(conn)
		}

		Mobs.OnMobs(char.GetCurrentMap(), func(mob *MapleMob) {
			if mob.GetController() == conn {
				mob.RemoveController()

				m.mutex.RLock()
				if len(m.GetPlayers()) > 0 {
					mob.SetController(m.GetPlayers()[0], false)
				}
				m.mutex.RUnlock()
			}

			mob.Hide(conn)
		})
	})

	Players.OnCharacterFromConn(conn, func(char *MapleCharacter) {
		m.SendPacket(packets.MapPlayerLeft(char.GetCharID()))
	})

}
