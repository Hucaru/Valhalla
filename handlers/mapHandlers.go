package handlers

import (
	"log"

	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/connection"
	"github.com/Hucaru/Valhalla/maplepacket"
)

func handleUsePortal(conn *connection.Channel, reader maplepacket.Reader) {
	reader.ReadByte()
	entryType := reader.ReadInt32()

	switch entryType {
	case 0:
		channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
			channel.Maps.GetMap(char.GetCurrentMap()).RemovePlayer(conn)
			if char.GetHP() == 0 {
				portal, pID := channel.Maps.GetMap(char.GetCurrentMap()).GetRandomSpawnPortal()
				char.Character.SetHP(50)
				char.ChangeMap(channel.Maps.GetMap(char.GetCurrentMap()).GetReturnMap(), portal, pID)
			} else {
				// hacker?
			}
		})
	case -1:
		portalName := reader.ReadString(int(reader.ReadInt16()))

		channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
			channel.Maps.GetMap(char.GetCurrentMap()).RemovePlayer(conn)

			for _, v := range channel.Maps.GetMap(char.GetCurrentMap()).GetPortals() {
				if v.GetName() == portalName {
					for i, portal := range channel.Maps.GetMap(v.GetToMap()).GetPortals() {
						if portal.GetName() == v.GetToPortal() {
							mapID := v.GetToMap()
							char.ChangeMap(mapID, portal, byte(i))
							break
						}
					}
					break
				}
			}
		})
	default:
		log.Println("Unknown map entry type, packet is:", reader)
	}
}
