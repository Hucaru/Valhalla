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

	skillPossible := (bits & 0x0F) != 0

	plr, err := server.players.getFromConn(conn)

	if err != nil {
		return
	}

	field, ok := server.fields[plr.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(plr.InstanceID())

	if err != nil {
		return
	}

	moveData, finalData := movement.ParseMovement(reader)

	moveBytes := movement.GenerateMovementBytes(moveData)

	inst.LifePool().MobAcknowledge(mobSpawnID, plr, moveID, skillPossible, byte(action), skillData, moveData, finalData, moveBytes)

	// skillDelay := int16(skillData >> 16)
	// skillID := byte(skillData)
	// skillLevel := byte(skillData >> 8)

	// if actualAction >= 21 && actualAction <= 25 {
	// 	mob.PerformSkill(skillDelay, skillLevel, skillID)
	// } else if actualAction > 12 && actualAction < 20 {
	// 	mob.PerformAttack(byte(actualAction - 12))
	// }

	// mob.AcknowledgeController(moveID, finalData, skillPossible, skillID, skillLevel)
	// moveBytes := movement.GenerateMovementBytes(moveData)
	// inst.UpdateMob(mobSpawnID, skillPossible, byte(action), skillData, moveBytes)
}
