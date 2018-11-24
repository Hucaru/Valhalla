package channel

import (
	"math/rand"
	"time"

	"github.com/Hucaru/Valhalla/consts/skills"
	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/game/packet"
)

func mobControl(conn mnet.MConnChannel, reader mpacket.Reader) {
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

	player, err := game.GetPlayerFromConn(conn)

	if err != nil {
		return
	}

	char := player.Char()
	mob := game.GetMapFromID(char.CurrentMap).GetMobFromID(mobSpawnID)

	if mob == nil {
		return
	}

	if mob.Controller != conn { // prevents hijack and reassigns controller to anyone except hijacker
		mob.FindNewControllerExcept(conn)
		return
	}

	// Update mob position information
	moveData, finalData := parseMovement(reader)

	if !validateMobMovement(mob.Mob, moveData) {
		return
	}

	mob.X = finalData.X
	mob.Y = finalData.Y
	mob.Foothold = finalData.Foothold
	mob.Stance = finalData.Stance

	moveBytes := generateMovementBytes(moveData)

	// Perform the action received
	if actualAction >= 21 && actualAction <= 25 {
		performSkill(&mob.Mob, int16(skillData>>16), byte(skillData>>8), byte(skillData))

	} else if actualAction > 12 && actualAction < 20 {
		attackID := byte(actualAction - 12)

		// check mob can use attack
		if attack, valid := nx.GetMobAttack(mob.ID, attackID); valid {
			mob.MP = mob.MP - attack.MPCon
			if mob.MP < 0 {
				mob.MP = 0
			}
		}

		mob.LastAttackTime = time.Now().Unix()
	}

	// Calculate the next action
	mob.CanUseSkill = skillPossible

	if !mob.CanUseSkill || (mob.StatBuff&skills.MobStat.SealSkill > 0) || (time.Now().Unix()-mob.LastSkillUseTime) < 3 {
		// there are more reasons as to why a mob cannot use a skill
		mob.SkillID = 0
	} else {
		mob.SkillID, mob.SkillLevel = chooseNextSkill(&mob.Mob)
	}

	conn.Send(packet.MobControlAcknowledge(mobSpawnID, moveID, skillPossible, int16(mob.MP), mob.SkillID, mob.SkillLevel)) // change zeros to what is calculated as next move
	game.SendToMapExcept(char.CurrentMap, packet.MobMove(mobSpawnID, skillPossible, byte(action), skillData, moveBytes), conn)
}

func chooseNextSkill(mob *def.Mob) (byte, byte) {
	var skillID, skillLevel byte

	skillsToChooseFrom := []nx.MobSkill{}

	for _, skill := range nx.GetMobSkills(mob.ID) {

		// Skill HP check
		if (mob.HP * 100 / mob.MaxHP) < skill.HP {
			continue
		}

		// Skill cooldown check
		if val, ok := mob.SkillTimes[skill.ID]; ok {
			if (val + skill.Cooldown) > time.Now().Unix() { // Is cooldown in seconds?
				continue
			}
		}

		// Check summon limit
		if skill.ID == skills.Mob.Summon {

		}

		// Determine if stats can be buffed
		if mob.StatBuff > 0 {
			alreadySet := false

			switch skill.ID {
			case skills.Mob.WeaponAttackUp:
				fallthrough
			case skills.Mob.WeaponAttackUpAoe:
				alreadySet = mob.StatBuff&skills.MobStat.PowerUp > 0

			case skills.Mob.MagicAttackUp:
				fallthrough
			case skills.Mob.MagicAttackUpAoe:
				alreadySet = mob.StatBuff&skills.MobStat.MagicUp > 0

			case skills.Mob.WeaponDefenceUp:
				fallthrough
			case skills.Mob.WeaponDefenceUpAoe:
				alreadySet = mob.StatBuff&skills.MobStat.PowerGuardUp > 0

			case skills.Mob.MagicDefenceUp:
				fallthrough
			case skills.Mob.MagicDefenceUpAoe:
				alreadySet = mob.StatBuff&skills.MobStat.MagicGuardUp > 0

			case skills.Mob.WeaponImmunity:
				alreadySet = mob.StatBuff&skills.MobStat.PhysicalImmune > 0

			case skills.Mob.MagicImmunity:
				alreadySet = mob.StatBuff&skills.MobStat.MagicImmune > 0

			// case skills.Mob.WeaponDamageReflect:

			// case skills.Mob.MagicDamageReflect:

			case skills.Mob.McSpeedUp:
				alreadySet = mob.StatBuff&skills.MobStat.Speed > 0

			default:
			}

			if alreadySet {
				continue
			}

		}

		skillsToChooseFrom = append(skillsToChooseFrom, skill)
	}

	if len(skillsToChooseFrom) > 0 {
		skill := skillsToChooseFrom[rand.Intn(len(skillsToChooseFrom))]
		skillID = skill.ID
		skillLevel = skill.Level
	}

	if skillLevel == 0 {
		skillID = 0
	}

	return skillID, skillLevel
}

func performSkill(mob *def.Mob, delay int16, skillLevel, skillID byte) {
	if skillID != mob.SkillID || (mob.StatBuff&skills.MobStat.SealSkill > 0) {
		skillID = 0
		return
	}

	var skill *nx.MobSkill

	for _, itSkill := range nx.GetMobSkills(mob.ID) {
		if itSkill.ID == skillID && itSkill.Level == skillLevel {
			skill = &itSkill
		}
	}

	if skill == nil {
		mob.SkillID = 0
		return
	}

	mob.MP = mob.MP - skill.MPCon
	if mob.MP < 0 {
		mob.MP = 0
	}

	currentTime := time.Now().Unix()

	mob.SkillTimes[skillID] = currentTime
	mob.LastSkillUseTime = currentTime

	// Handle all the different skills!
	switch skill.ID {

	}
}
