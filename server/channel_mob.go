package server

import (
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/field/lifepool/mob"
	"github.com/Hucaru/Valhalla/server/movement"
	"github.com/Hucaru/Valhalla/server/player"
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

	inst, err := server.getPlayerInstance(conn, reader)
	if err != nil {
		return
	}

	moveData, finalData := movement.ParseMovement(reader)

	moveBytes := movement.GenerateMovementBytes(moveData)

	inst.LifePool().MobAcknowledge(mobSpawnID, plr, moveID, skillPossible, byte(action), skillData, moveData, finalData, moveBytes)

}

func (server ChannelServer) mobDamagePlayer(conn mnet.Client, reader mpacket.Reader, mobAttack int8) {
	damage := reader.ReadInt32() // Damage amount
	healSkillID := int32(0)

	if damage < -1 {
		return
	}

	reducedDamage := damage

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

	var mob mob.Data
	var mobSkillID, mobSkillLevel byte = 0, 0

	if mobAttack < -1 {
		mobSkillLevel = reader.ReadByte()
		mobSkillID = reader.ReadByte()
	} else {
		magicElement := int32(0)

		if reader.ReadBool() {
			magicElement = reader.ReadInt32()
			_ = magicElement
			// 0 = no element (Grendel the Really Old, 9001001)
			// 1 = Ice (Celion? blue, 5120003)
			// 2 = Lightning (Regular big Sentinel, 3000000)
			// 3 = Fire (Fire sentinel, 5200002)
		}

		spawnID := reader.ReadInt32()
		mobID := reader.ReadInt32()

		mob, err = inst.LifePool().GetMobFromID(spawnID)
		if err != nil {
			return
		}

		if mob.ID() != mobID {
			return
		}

		stance := reader.ReadByte()

		reflected := reader.ReadByte()

		reflectAction := byte(0)
		var reflectX, reflectY int16 = 0, 0

		if reflected > 0 {
			reflectAction = reader.ReadByte()
			reflectX, reflectY = reader.ReadInt16(), reader.ReadInt16()
		}

		// Magic guard dmg absorption

		// Fighter / Page power guard

		// Meso guard

		plr.DamagePlayer(int16(damage))
		inst.Send(player.PlayerReceivedDmg(plr.ID(), mobAttack, damage, reducedDamage, spawnID, mobID, healSkillID, stance, reflectAction, reflected, reflectX, reflectY))
	}
	if mobSkillID != 0 && mobSkillLevel != 0 {
		// new skill
	}

}

func (server ChannelServer) mobDistance(conn mnet.Client, reader mpacket.Reader) {
	/*
		id := reader.ReadInt32()
		distance := reader.ReadInt32()

		Unknown what this packet is for
	*/

}
