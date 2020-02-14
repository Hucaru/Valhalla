package server

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/movement"
)

func (server ChannelServer) mobControl(conn mnet.Client, reader mpacket.Reader) {
	mobSpawnID := reader.ReadInt32()
	moveID := reader.ReadInt16()
	bits := reader.ReadByte()
	action := reader.ReadInt8()
	skillData := reader.ReadUint32()

	actualAction := int(byte(action >> 1))

	if action < 0 {
		actualAction = -1
	}

	skillPossible := (bits & 0x0F) != 0

	player, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[player.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(player.InstanceID())

	if err != nil {
		return
	}

	mob := inst.GetMob(mobSpawnID)

	if mob == nil {
		return
	}

	if mob.Controller().Conn() != conn {
		return
	}

	skillDelay := int16(skillData >> 16)
	skillID := byte(skillData)
	skillLevel := byte(skillData >> 8)

	if actualAction >= 21 && actualAction <= 25 {
		mob.PerformSkill(skillDelay, skillLevel, skillID)
	} else if actualAction > 12 && actualAction < 20 {
		mob.PerformAttack(byte(actualAction - 12))
	}

	moveData, finalData := movement.ParseMovement(reader)

	if !moveData.ValidateMob(*mob) {
		return
	}

	mob.AcknowledgeController(moveID, finalData, skillPossible, skillID, skillID)
	moveBytes := movement.GenerateMovementBytes(moveData)
	inst.UpdateMob(mobSpawnID, skillPossible, byte(action), skillData, moveBytes)
}
