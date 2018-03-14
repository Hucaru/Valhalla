package maps

import (
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/data"
	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/gopacket"
)

var mapsPtr interfaces.Maps

var charsPtr interfaces.Characters

// RegisterCharactersObj -
func RegisterCharactersObj(chars interfaces.Characters) {
	charsPtr = chars
}

// RegisterMapsObj -
func RegisterMapsObj(mapList interfaces.Maps) {
	mapsPtr = mapList

	startRespawnMonitors()
}

// RegisterNewPlayerCallback -
func RegisterNewPlayerCallback(conn interfaces.ClientConn) {
	conn.AddCloseCallback(func() {
		PlayerLeaveMap(conn, charsPtr.GetOnlineCharacterHandle(conn).GetCurrentMap())
		charsPtr.RemoveOnlineCharacter(conn)
	})
}

// SendPacketToMap -
func SendPacketToMap(mapID uint32, p gopacket.Packet) {
	if len(p) > 0 {

		players := mapsPtr.GetMap(mapID).GetPlayers()

		for _, v := range players {
			if v != nil { // check this is still an open socket
				v.Write(p)
			}
		}
	}
}

// PlayerEnterMap -
func PlayerEnterMap(conn interfaces.ClientConn, mapID uint32) {
	m := mapsPtr.GetMap(mapID)

	for _, v := range m.GetPlayers() {
		v.Write(playerEnterMapPacket(charsPtr.GetOnlineCharacterHandle(conn)))
		conn.Write(playerEnterMapPacket(charsPtr.GetOnlineCharacterHandle(v)))
	}

	m.AddPlayer(conn)

	// Send npcs
	for i, v := range m.GetNpcs() {
		if !v.GetIsAlive() {
			continue
		}
		conn.Write(showNpcPacket(uint32(i), v))
	}

	// Send mobs
	for _, v := range m.GetMobs() {
		if !v.GetIsAlive() {
			continue
		}

		if v.GetController() == nil {
			v.SetController(conn)
			conn.Write(controlMobPacket(v.GetSpawnID(), v, false, false))
		}

		conn.Write(showMobPacket(v.GetSpawnID(), v, false))
	}
}

// PlayerLeaveMap -
func PlayerLeaveMap(conn interfaces.ClientConn, mapID uint32) {
	m := mapsPtr.GetMap(mapID)

	m.RemovePlayer(conn)

	// Remove player as mob controller
	for _, v := range m.GetMobs() {
		if v.GetController() == conn {
			v.SetController(nil)
		}

		conn.Write(endMobControlPacket(v.GetSpawnID()))
	}

	if len(m.GetPlayers()) > 0 {
		newController := m.GetPlayers()[0]
		for _, v := range m.GetMobs() {
			if v.GetIsAlive() {
				newController.Write(controlMobPacket(v.GetSpawnID(), v, false, false))
			}

		}
	}

	SendPacketToMap(mapID, playerLeftMapPacket(charsPtr.GetOnlineCharacterHandle(conn).GetCharID()))
}

func GetRandomSpawnPortal(mapID uint32) (interfaces.Portal, byte) {
	var portals []interfaces.Portal
	for _, portal := range mapsPtr.GetMap(mapID).GetPortals() {
		if portal.GetIsSpawn() {
			portals = append(portals, portal)
		}
	}
	rand.Seed(time.Now().UnixNano())
	pos := rand.Intn(len(portals))
	return portals[pos], byte(pos)
}

func DamageMobs(mapID uint32, conn interfaces.ClientConn, damages map[uint32][]uint32) []uint32 {
	m := mapsPtr.GetMap(mapID)

	var exp []uint32

	validDamages := make(map[uint32][]uint32)

	// check spawn id to make sure all are valid
	for k, dmgs := range damages {
		for _, v := range m.GetMobs() {
			if v.GetSpawnID() == k {
				validDamages[k] = dmgs
			}
		}
	}

	for k, dmgs := range validDamages {
		mob := m.GetMobFromID(k)

		if mob.GetController() != conn {
			if mob.GetController() != nil {
				mob.GetController().Write(endMobControlPacket(mob.GetSpawnID()))
			}
			mob.SetController(conn)
			conn.Write(controlMobPacket(mob.GetSpawnID(), mob, false, true)) // does mob need to be agroed?
		}

		for _, dmg := range dmgs {
			newHP := int32(int32(mob.GetHp()) - int32(dmg))

			if newHP < 1 {

				conn.Write(endMobControlPacket(mob.GetSpawnID()))
				SendPacketToMap(mapID, removeMobPacket(mob.GetSpawnID(), 1))
				exp = append(exp, mob.GetEXP())
				mob.SetIsAlive(false)

				if mob.GetMobTime() > 0 {
					mob.SetDeathTime(time.Now().Unix())
				}

				break // mob is dead, no need to process further dmg packets

			} else {
				mob.SetHp(uint16(newHP))
				// show hp bar
			}
		}
	}

	return exp
}

func startRespawnMonitors() {
	for mapID := range nx.Maps {

		go func(mapID uint32) {
			ticker := time.NewTicker(10 * time.Second)
			m := mapsPtr.GetMap(mapID)

			for {
				<-ticker.C
				for _, mob := range m.GetMobs() {
					if mob == nil {
						// GC has not collected nil refs from slice
						continue
					}

					if !mob.GetRespawns() && !mob.GetIsAlive() {
						m.RemoveMob(mob)
						continue
					}
					respawn := false

					// normal mobs
					if !mob.GetIsAlive() && !mob.GetBoss() && mob.GetMobTime() == 0 {
						respawn = true
					}

					// bosses and long respawn mbos e.g. iron hog at pig beach
					if !mob.GetIsAlive() && (mob.GetBoss() || mob.GetMobTime() > 0) {
						if (time.Now().Unix() - mob.GetDeathTime()) > mob.GetMobTime() {
							respawn = true
						}
					}

					if respawn {
						// not sure on how official maple did respawns, here is my attempt
						mob.SetX(mob.GetX())
						mob.SetY(mob.GetY())
						mob.SetFoothold(mob.GetSFoothold()) // I suspect this is the only setting that matters
						mob.SetHp(mob.GetMaxHp())
						mob.SetMp(mob.GetMaxMp())

						if len(m.GetPlayers()) > 0 {
							newController := m.GetPlayers()[0]
							newController.Write(controlMobPacket(mob.GetSpawnID(), mob, true, false))
						}
						SendPacketToMap(uint32(mapID), showMobPacket(mob.GetSpawnID(), mob, true))
						mob.SetIsAlive(true)
					}
				}
			}
		}(mapID)
	}
}

func SpawnMob(mapID, mobID uint32, x, y, foothold int16, respawns bool, controller interfaces.ClientConn) {
	m := mapsPtr.GetMap(mapID)

	if _, exists := nx.Mob[mobID]; exists {
		newMob := data.CreateMobFromID(mobID)
		newMob.SetX(x)
		newMob.SetY(y)
		newMob.SetSX(x)
		newMob.SetSY(y)
		newMob.SetFoothold(foothold)
		newMob.SetSFoothold(foothold)
		newMob.SetSpawnID(m.GetNextMobSpawnID())
		newMob.SetRespawns(respawns)

		m.AddMob(newMob)

		if len(m.GetPlayers()) > 0 {
			newController := m.GetPlayers()[0]
			newController.Write(controlMobPacket(newMob.GetSpawnID(), newMob, true, true))
		}
		SendPacketToMap(mapID, showMobPacket(newMob.GetSpawnID(), newMob, true))
		newMob.SetIsAlive(true)
	}
}
