package server

import (
	"github.com/Hucaru/Valhalla/entity"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
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

	player, err := server.players.GetFromConn(conn)

	if err != nil {
		return
	}

	char := player.Char()

	field, ok := server.fields[char.MapID()]

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

	if mob.Controller() != conn {
		return
	}

	// Perform received action e.g. use skill, attack etc
	_ = actualAction

	moveData, finalData := entity.ParseMovement(reader)

	if !moveData.ValidateMob(*mob) {
		return
	}

	mob.AcknowledgeController(moveID, finalData)
	moveBytes := entity.GenerateMovementBytes(moveData)
	inst.SendExcept(entity.PacketMobMove(mobSpawnID, skillPossible, byte(action), skillData, moveBytes), conn)
}
