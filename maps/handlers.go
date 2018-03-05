package maps

import (
	"log"

	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/Valhalla/movement"
	"github.com/Hucaru/gopacket"
)

// HandlePlayerUsePortal -
func HandlePlayerUsePortal(conn interfaces.ClientConn, reader gopacket.Reader) {
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
func HandlePlayerEmotion(conn interfaces.ClientConn, reader gopacket.Reader) {
	emotion := reader.ReadUint32()
	char := charsPtr.GetOnlineCharacterHandle(conn)
	SendPacketToMap(char.GetCurrentMap(), playerEmotionPacket(char.GetCharID(), emotion))
}

// HandleMobMovement -
func HandleMobMovement(conn interfaces.ClientConn, reader gopacket.Reader) {
	mobID := reader.ReadUint32()
	moveID := reader.ReadUint16()
	skillUsed := bool(reader.ReadByte() != 0)
	skill := reader.ReadByte()

	level := byte(0)
	if skillUsed {
		// Implement mob skills
	}

	reader.ReadInt16() // ? x pos for something?
	reader.ReadInt16() // ? y pos for something?

	reader.ReadInt32()          //
	nFrags := reader.ReadByte() // n fragments?

	var mp uint16

	mapID := charsPtr.GetOnlineCharacterHandle(conn).GetCurrentMap()
	m := mapsPtr.GetMap(mapID)

	var mob interfaces.Mob

	for i, v := range m.GetMobs() {
		if uint32(i) == mobID {
			mob = v
		}
	}

	movement.ParseFragments(nFrags, mob, reader)

	SendPacketToMap(mapID, moveMobPacket(mobID, skillUsed, skill, reader.GetBuffer()[13:]))
	conn.Write(controlAckPacket(mobID, moveID, skillUsed, skill, level, mp))
}
