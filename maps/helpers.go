package maps

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/inventory"

	"github.com/Hucaru/Valhalla/data"
	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/gopacket"
)

var charsPtr interfaces.Characters
var mapsPtr interfaces.Maps

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

// SendPacketToMapExcept -
func SendPacketToMapExcept(mapID uint32, p gopacket.Packet, conn interfaces.ClientConn) {
	if len(p) > 0 {

		players := mapsPtr.GetMap(mapID).GetPlayers()

		for _, v := range players {
			if v != nil && v != conn { // check this is still an open socket
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

func DamageMobs(mapID uint32, conn interfaces.ClientConn, damages map[uint32][]uint32) map[interfaces.ClientConn][]uint32 {
	m := mapsPtr.GetMap(mapID)

	exp := make(map[interfaces.ClientConn][]uint32)

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
		}

		for _, dmg := range dmgs {
			newHP := int32(int32(mob.GetHp()) - int32(dmg))

			if _, exists := mob.GetDmgReceived()[conn]; !exists {
				mob.GetDmgReceived()[conn] = dmg
			} else {
				mob.GetDmgReceived()[conn] += dmg
			}

			if newHP < 1 {
				conn.Write(endMobControlPacket(mob.GetSpawnID()))
				SendPacketToMap(mapID, removeMobPacket(mob.GetSpawnID(), 1))

				mob.SetIsAlive(false)

				if mob.GetMobTime() > 0 {
					mob.SetDeathTime(time.Now().Unix())
				}

				for k := range mob.GetDmgReceived() {

					char := charsPtr.GetOnlineCharacterHandle(k)

					if char != nil && char.GetCurrentMap() == mapID {
						exp[k] = append(exp[k], mob.GetEXP()) // modify this exp based on % dmg done to mob
					}
				}

				mob.SetDmgReceived(make(map[interfaces.ClientConn]uint32))

				break // mob is dead, no need to process further dmg information for mob

			} else {
				mob.SetHp(uint32(newHP))
				conn.Write(controlMobPacket(mob.GetSpawnID(), mob, false, true))
				// if this version has general mob hp bars show it
			}
		}
	}

	return exp
}

func startRespawnMonitors() {
	for mapID := range nx.Maps {

		go func(mapID uint32) {
			m := mapsPtr.GetMap(mapID)
			mobRate := m.GetMobRate()

			if mobRate == 0 {
				mobRate = 1
			}

			ticker := time.NewTicker(time.Duration(5.0/mobRate) * time.Second)

			for {
				<-ticker.C
				for _, mob := range m.GetMobs() {
					if mob == nil {
						continue
					}

					if !mob.GetRespawns() && !mob.GetIsAlive() {
						m.RemoveMob(mob)
						continue
					}

					// normal mobs
					if !mob.GetIsAlive() && !mob.GetBoss() && mob.GetMobTime() == 0 {
						for i := uint32(0); i < constants.GetRate(constants.MobRate); i++ {
							if len(m.GetMobs()) > m.GetNumberSpawnableMobs()*int(constants.GetRate(constants.MobRate)) {
								m.RemoveMob(mob)
								break
							}

							newMob := m.GetRandomSpawnableMob(mob.GetSX(), mob.GetSY(), mob.GetSFoothold())
							newMob.SetSpawnID(m.GetNextMobSpawnID())

							m.RemoveMob(mob)
							m.AddMob(newMob)

							newMob.SetIsAlive(true)

							if len(m.GetPlayers()) > 0 {
								newController := m.GetPlayers()[0]
								newController.Write(controlMobPacket(newMob.GetSpawnID(), newMob, true, false))
							}

							SendPacketToMap(uint32(mapID), showMobPacket(newMob.GetSpawnID(), newMob, true))
						}
					}

					// bosses and long respawn mbos e.g. iron hog at pig beach or jr boogies
					if !mob.GetIsAlive() && (mob.GetBoss() || mob.GetMobTime() > 0) {
						if (time.Now().Unix() - mob.GetDeathTime()) > mob.GetMobTime() {
							// Don't keep old id as we take advantage of overflow to recycle them
							mob.SetSpawnID(m.GetNextMobSpawnID())
							mob.SetX(mob.GetSX())
							mob.SetY(mob.GetSY())
							mob.SetFoothold(mob.GetSFoothold())
							mob.SetHp(mob.GetMaxHp())
							mob.SetMp(mob.GetMaxMp())
							mob.SetIsAlive(true)

							m.RemoveMob(mob)
							m.AddMob(mob)

							if len(m.GetPlayers()) > 0 {
								newController := m.GetPlayers()[0]
								newController.Write(controlMobPacket(mob.GetSpawnID(), mob, true, false))
							}
							fmt.Println(mob)
							SendPacketToMap(uint32(mapID), showMobPacket(mob.GetSpawnID(), mob, true))
						}
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
		newMob.SetDmgReceived(make(map[interfaces.ClientConn]uint32))

		m.AddMob(newMob)

		if len(m.GetPlayers()) > 0 {
			newController := m.GetPlayers()[0]
			newController.Write(controlMobPacket(newMob.GetSpawnID(), newMob, true, true))
		}
		SendPacketToMap(mapID, showMobPacket(newMob.GetSpawnID(), newMob, true))
		newMob.SetIsAlive(true)
	}
}

func showExistingDrop() {

}

// SpawnDrop -
func SpawnDrop(item inventory.Item, dropAnimation byte, playerDrop bool, dropperID uint32, pos interfaces.Pos, expirationTime time.Duration) {
	// Add hack detection later (e.g. check person owns item & item is a dropable item)

	// spawn thread to call remove drop after expirationTime, if item in map drop table remove it
}

// RemoveDrop -
func RemoveDrop() {

}
