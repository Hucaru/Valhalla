package maps

import (
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/interfaces"
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
			conn.Write(controlMobPacket(v.GetSpawnID(), v, false))
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
				newController.Write(controlMobPacket(v.GetSpawnID(), v, false))
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
			conn.Write(controlMobPacket(mob.GetSpawnID(), mob, false)) // does mob need to be agroed?
		}

		for _, dmg := range dmgs {
			newHP := int32(int32(mob.GetHp()) - int32(dmg))

			if newHP < 1 {

				conn.Write(endMobControlPacket(mob.GetSpawnID()))
				SendPacketToMap(mapID, removeMobPacket(mob.GetSpawnID(), 1))
				exp = append(exp, mob.GetEXP())
				mob.SetIsAlive(false)
				// add a new mob to spawn buffer

				break // mob is dead no need to process further dmg packets

			} else {
				mob.SetHp(uint16(newHP))
				// show hp bar
			}
		}
	}

	return exp
}
