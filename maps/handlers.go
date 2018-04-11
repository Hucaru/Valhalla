package maps

import (
	"log"

	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/movement"
)

// HandlePlayerUsePortal -
func HandlePlayerUsePortal(conn interfaces.ClientConn, reader maplepacket.Reader) {
	char := charsPtr.GetOnlineCharacterHandle(conn)

	PlayerLeaveMap(conn, char.GetCurrentMap())

	var mapID uint32

	reader.ReadByte()
	entryType := reader.ReadInt32()

	switch entryType {
	case 0:
		if char.GetHP() == 0 {
			mapID = mapsPtr.GetMap(char.GetCurrentMap()).GetReturnMap()
			portal, pID := GetRandomSpawnPortal(mapID)

			char.SetX(portal.GetX())
			char.SetY(portal.GetY())

			char.SetHP(50)

			conn.Write(ChangeMapPacket(mapID, 1, pID, char.GetHP())) // replace 1 with channel id

			char.SetCurrentMap(mapID)
		}
	case -1:
		portalName := reader.ReadString(int(reader.ReadUint16()))

		// check portal is valid, i.e it is not closed

		for i, v := range mapsPtr.GetMap(char.GetCurrentMap()).GetPortals() {
			if v.GetName() == portalName {

				for _, portal := range mapsPtr.GetMap(v.GetToMap()).GetPortals() {
					if portal.GetName() == v.GetToPortal() {
						mapID = v.GetToMap()
						char.SetX(portal.GetX())
						char.SetY(portal.GetY())

						conn.Write(ChangeMapPacket(mapID, 1, byte(i), char.GetHP())) // replace 1 with channel id

						char.SetCurrentMap(mapID)
						break
					}
				}
				break
			}
		}
	default:
		log.Println("Unknown map entry type, packet is:", reader)
	}

	PlayerEnterMap(conn, mapID)
}

// HandlePlayerEmotion -
func HandlePlayerEmotion(conn interfaces.ClientConn, reader maplepacket.Reader) {
	emotion := reader.ReadUint32()
	char := charsPtr.GetOnlineCharacterHandle(conn)
	SendPacketToMap(char.GetCurrentMap(), playerEmotionPacket(char.GetCharID(), emotion))
}

// HandleMobMovement -
func HandleMobMovement(conn interfaces.ClientConn, reader maplepacket.Reader) {
	mobID := reader.ReadUint32()
	moveID := reader.ReadUint16()
	skillUsed := bool(reader.ReadByte() != 0)
	skill := reader.ReadByte()

	level := byte(0)
	if skillUsed {
		// Implement mob skills
	}

	projPos := reader.ReadUint32() // x & y

	reader.ReadInt32()          //
	nFrags := reader.ReadByte() // n fragments?

	var mp uint32

	mapID := charsPtr.GetOnlineCharacterHandle(conn).GetCurrentMap()
	m := mapsPtr.GetMap(mapID)

	var mob interfaces.Mob

	for _, v := range m.GetMobs() {
		if v.GetSpawnID() == mobID {
			mob = v
			mp = v.GetMp()
		}
	}

	// This should only arrise when someone 1hit kos a mob and controller movement packet reaches after attack
	if mob == nil || !mob.GetIsAlive() {
		return
	}

	movement.ParseFragments(nFrags, mob, reader)

	conn.Write(controlAckPacket(mobID, moveID, skillUsed, skill, level, uint16(mp)))
	SendPacketToMapExcept(mapID, moveMobPacket(mobID, skillUsed, skill, projPos, reader.GetBuffer()[13:]), conn)
}
