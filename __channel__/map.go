package channel

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/game/packet"
)

type mapleMap struct {
	forcedReturn int32
	returnMap    int32
	mobRate      float64
	isTown       bool
	portals      []maplePortal
	mutex        *sync.RWMutex
	players      []mnet.MConnChannel
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

func (m *mapleMap) GetNearestSpawnPortalID(char *MapleCharacter) int {
	var distance float64 = -1
	var ind int

	m.mutex.RLock()
	for i, v := range m.portals {
		if !v.GetIsSpawn() {
			continue
		}
		calc := math.Hypot(float64(char.GetX()-v.GetX()), float64(char.GetX()-v.GetX()))

		if distance == -1 {
			distance = calc // guaranteed to always return a portal this way
		} else if distance > calc {
			distance = calc
			ind = i
		}
	}
	m.mutex.RUnlock()

	return ind
}

func (m *mapleMap) AddPortal(portal maplePortal) {
	m.mutex.Lock()
	m.portals = append(m.portals, portal)
	m.mutex.Unlock()
}

func (m *mapleMap) OnPlayers(action func(conn mnet.MConnChannel) bool) {
	m.mutex.RLock()
	for i := range m.players {
		done := action(m.players[i])
		if done {
			break
		}
	}
	m.mutex.RUnlock()
}

func (m *mapleMap) AddPlayer(player mnet.MConnChannel) {
	m.playerEnterMap(player)

	m.mutex.Lock()
	m.players = append(m.players, player)
	m.mutex.Unlock()
}

func (m *mapleMap) RemovePlayer(player mnet.MConnChannel) {
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
			player.Send(packet)
		}
		m.mutex.RUnlock()
	}
}

func (m *mapleMap) SendPacketExcept(packet maplepacket.Packet, conn mnet.MConnChannel) {
	if len(packet) > 0 {
		m.mutex.RLock()
		for _, player := range m.players {
			if player == conn {
				continue
			}
			player.Send(packet)
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

func (m *mapleMap) playerEnterMap(conn mnet.MConnChannel) {
	m.OnPlayers(func(other mnet.MConnChannel) bool {

		Players.OnCharacterFromConn(conn, func(char *MapleCharacter) {
			other.Send(packet.MapPlayerEnter(char.Character))
		})

		Players.OnCharacterFromConn(other, func(char *MapleCharacter) {
			conn.Send(packet.MapPlayerEnter(char.Character))
		})

		return false
	})

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

		ActiveRooms.OnRoom(func(r *Room) {
			if r.MapID == char.GetCurrentMap() {
				if p, valid := r.GetBox(); valid {
					char.SendPacket(p)
				}
			}
		})
	})
}

func (m *mapleMap) playerLeaveMap(conn mnet.MConnChannel) {
	Players.OnCharacterFromConn(conn, func(char *MapleCharacter) {
		for _, npc := range NPCs.GetNpcs(char.GetCurrentMap()) {
			npc.Hide(conn)
		}

		Mobs.OnMobs(char.GetCurrentMap(), func(mob *MapleMob) {
			if mob.GetController() == conn {
				mob.RemoveController()

				m.OnPlayers(func(conn mnet.MConnChannel) bool {
					mob.SetController(conn, false)
					return true
				})
			}

			mob.Hide(conn)
		})
	})

	Players.OnCharacterFromConn(conn, func(char *MapleCharacter) {
		m.SendPacket(packet.MapPlayerLeft(char.GetCharID()))
	})

}
