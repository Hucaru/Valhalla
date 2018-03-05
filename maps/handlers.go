package maps

import (
	"log"

	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/gopacket"
)

func PlayerEnterMap(conn interfaces.ClientConn, mapID uint32) {
	m := mapsPtr.GetMap(mapID)

	for _, v := range m.GetPlayers() {
		v.Write(playerEnterMapPacket(charsPtr.GetOnlineCharacterHandle(conn)))
		conn.Write(playerEnterMapPacket(charsPtr.GetOnlineCharacterHandle(v)))
	}

	m.AddPlayer(conn)

	// Send npcs
	for i, v := range m.GetNpcs() {
		conn.Write(showNpcPacket(uint32(i), v))
	}

	// Send mobs
	for i, v := range m.GetMobs() {
		if v.GetController() == nil {
			v.SetController(conn)
			// Send control packet
		}
		conn.Write(showMobPacket(uint32(i), v, false))
	}
}

func PlayerLeaveMap(conn interfaces.ClientConn, mapID uint32) {
	mapsPtr.GetMap(mapID).RemovePlayer(conn)
	SendPacketToMap(mapID, playerLeftMapPacket(charsPtr.GetOnlineCharacterHandle(conn).GetCharID()))
	// Remove player as controller
	// find new controller for mobs if players left in map
}

func HandlePlayerChangeMap(conn interfaces.ClientConn, reader gopacket.Reader) {
	char := charsPtr.GetOnlineCharacterHandle(conn)

	PlayerLeaveMap(conn, char.GetCurrentMap())

	var mapID uint32

	reader.ReadByte()
	entryType := reader.ReadInt32()

	switch entryType {
	case 0:
		if char.GetHP() == 0 {
			mapID = mapsPtr.GetMap(char.GetCurrentMap()).GetReturnMap()
			portal, pID := getRandomSpawnPortal(mapID)

			char.SetX(portal.GetX())
			char.SetY(portal.GetY())

			char.SetHP(50)

			conn.Write(changeMapPacket(mapID, 1, pID, char.GetHP())) // replace 1 with channel id

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

						conn.Write(changeMapPacket(mapID, 1, byte(i), char.GetHP())) // replace 1 with channel id

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
