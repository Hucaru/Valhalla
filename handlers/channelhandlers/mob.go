package channelhandlers

import (
	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/packets"
)

func mobControl(conn mnet.MConnChannel, reader maplepacket.Reader) {
	mobSpawnID := reader.ReadInt32()

	player := game.GetPlayerFromConn(conn)
	mob := game.GetMobFromMapAndSpawnID(player.Char().CurrentMap, mobSpawnID)

	if mob == nil {
		return
	}

	moveID := reader.ReadInt16()
	bits := reader.ReadByte()

	skillPossible := (bits & 0x0F) != 0

	action := reader.ReadInt8()

	actualAction := int(action >> 1)

	if action < 0 {
		actualAction = -1
	}

	unknownData := reader.ReadInt32()

	moveData, finalData := parseMovement(reader)

	if !validateMobMovement(*mob, moveData) {
		return
	}

	mob.X = finalData.X
	mob.Y = finalData.Y
	mob.Foothold = finalData.Foothold
	mob.Stance = finalData.Stance

	moveBytes := generateMovementBytes(moveData)

	// skill level of zero must have skill id of zero
	conn.Send(packets.MobControlAcknowledge(mobSpawnID, moveID, skillPossible, int16(mob.MP), 0, 0))
	game.SendToMapExcept(player.Char().CurrentMap, packets.MobMove(mobSpawnID, skillPossible, byte(actualAction), unknownData, moveBytes), conn)
	packets.MobMove(mobSpawnID, skillPossible, byte(actualAction), unknownData, moveBytes)
}
