package mobs

import (
	"github.com/Hucaru/Valhalla/channelServer/playerConn"
	"github.com/Hucaru/gopacket"
)

func HandleMovement(reader gopacket.Reader, conn *playerConn.Conn) {
	// mobID := reader.ReadUint32()

	// mapID := conn.GetCharacter().GetCurrentMap()
	// mobData := server.MobsGetFromMapAndID(mapID, mobID)

	// moveID := reader.ReadUint16()
	// useSkill := reader.ReadByte()
	// skill := reader.ReadByte()

	// reader.ReadInt16()
	// reader.ReadInt16()

	// mobData.LifeData.X = reader.ReadInt16()
	// mobData.LifeData.Y = reader.ReadInt16()

	// conn.Write(ControlMoveMob(mobID, moveID, useSkill, mobData.MobData.MaxMp))
	// server.SendPacketToMap(mapID, MoveMob(mobID, useSkill, skill, reader.GetBuffer()[13:]))
}
