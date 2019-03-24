package channel

import (
	"github.com/Hucaru/Valhalla/game"

	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func playerMeleeSkill(conn mnet.MConnChannel, reader mpacket.Reader) {
	player, ok := game.Players[conn]

	if !ok {
		return
	}

	data, valid := getAttackInfo(reader, *player, attackMelee)

	if !valid {
		return
	}

	char := player.Char()

	// fix the damage values

	for _, attack := range data.AttackInfo {
		mob, err := game.Maps[char.MapID].GetMobFromSpawnID(attack.SpawnID, player.InstanceID)

		if err != nil || mob == nil {
			return
		}

		mob.GiveDamage(player.MConnChannel, attack.Damages)
	}

	game.Maps[char.MapID].SendExcept(game.PacketSkillMelee(char, data), conn, player.InstanceID)
	game.Maps[char.MapID].HandleDeadMobs(player.InstanceID)
}

func playerRangedSkill(conn mnet.MConnChannel, reader mpacket.Reader) {
	player, ok := game.Players[conn]

	if !ok {
		return
	}

	data, valid := getAttackInfo(reader, *player, attackRanged)

	if !valid {
		return
	}

	char := player.Char()

	// fix the damage values

	for _, attack := range data.AttackInfo {
		mob, err := game.Maps[char.MapID].GetMobFromSpawnID(attack.SpawnID, player.InstanceID)

		if err != nil || mob == nil {
			return
		}

		mob.GiveDamage(player.MConnChannel, attack.Damages)
	}

	game.Maps[char.MapID].SendExcept(game.PacketSkillRanged(char, data), conn, player.InstanceID)
	game.Maps[char.MapID].HandleDeadMobs(player.InstanceID)
}

func playerMagicSkill(conn mnet.MConnChannel, reader mpacket.Reader) {
	player, ok := game.Players[conn]

	if !ok {
		return
	}

	data, valid := getAttackInfo(reader, *player, attackMagic)

	if !valid {
		return
	}

	char := player.Char()

	// fix the damange values

	for _, attack := range data.AttackInfo {
		mob, err := game.Maps[char.MapID].GetMobFromSpawnID(attack.SpawnID, player.InstanceID)

		if err != nil || mob == nil {
			return
		}

		mob.GiveDamage(player.MConnChannel, attack.Damages)
	}

	game.Maps[char.MapID].SendExcept(game.PacketSkillMagic(char, data), conn, player.InstanceID)
	game.Maps[char.MapID].HandleDeadMobs(player.InstanceID)

	switch data.SkillID {
	default:
		conn.Send(game.PacketPlayerNoChange())
	}
}

func playerSpecialSkill(conn mnet.MConnChannel, reader mpacket.Reader) {
	player, ok := game.Players[conn]

	if !ok {
		return
	}

	skillID := reader.ReadInt32()
	skillLevel := reader.ReadByte()

	char := player.Char()

	if skill, ok := char.Skills[skillID]; !ok || skill.Level != skillLevel {
		return
	}

	game.Maps[char.MapID].SendExcept(game.PacketSkillAnimation(char.ID, skillID, skillLevel), conn, player.InstanceID)

	switch skillID {
	default:
		conn.Send(game.PacketPlayerNoChange())
	}
}

// Following logic lifted from WvsGlobal
const (
	attackMelee = iota
	attackRanged
	attackMagic
	attackSummon
)

func getAttackInfo(reader mpacket.Reader, player game.Player, attackType int) (game.AttackData, bool) {
	data := game.AttackData{}

	if player.Char().HP == 0 {
		return data, false
	}

	// speed hack check
	if false && (reader.Time-player.LastAttackPacketTime < 350) {
		return data, false
	}

	player.LastAttackPacketTime = reader.Time

	if attackType != attackSummon {
		tByte := reader.ReadByte()
		skillID := reader.ReadInt32()

		if _, ok := player.Char().Skills[skillID]; !ok && skillID != 0 {
			return data, false
		}

		data.SkillID = skillID

		if data.SkillID != 0 {
			data.SkillLevel = player.Char().Skills[skillID].Level
		}

		// if meso explosion data.IsMesoExplosion = true

		data.Targets = tByte / 0x10
		data.Hits = tByte % 0x10
		data.Option = reader.ReadByte()

		tmp := reader.ReadByte()

		data.Action = tmp & 0x7F
		data.FacesLeft = (tmp >> 7) == 1
		data.AttackType = reader.ReadByte()
	} else {
		data.SummonType = reader.ReadInt32()
		data.AttackType = reader.ReadByte()
		data.Targets = 1
		data.Hits = 1
	}

	reader.Skip(4) //checksum info?

	if attackType == attackRanged {
		projectileSlot := reader.ReadInt16() // star/arrow slot
		if projectileSlot == 0 {
			// if soul arrow is not set check for hacks
		} else {
			data.ProjectileID = -1
			for _, item := range player.Char().Inventory.Use {
				if item.SlotID == projectileSlot {
					data.ProjectileID = item.ItemID
				}
			}

		}
		reader.ReadByte() // ?
		reader.ReadByte() // ?
		reader.ReadByte() // ?
	}

	data.AttackInfo = make([]game.AttackInfo, data.Targets)

	for i := byte(0); i < data.Targets; i++ {
		attack := game.AttackInfo{}
		attack.SpawnID = reader.ReadInt32()
		attack.HitAction = reader.ReadByte()

		tmp := reader.ReadByte()
		attack.ForeAction = tmp & 0x7F
		attack.FacesLeft = (tmp >> 7) == 1
		attack.FrameIndex = reader.ReadByte()

		if !data.IsMesoExplosion {
			attack.CalcDamageStatIndex = reader.ReadByte()
		}

		attack.HitPosition.X = reader.ReadInt16()
		attack.HitPosition.Y = reader.ReadInt16()

		attack.PreviousMobPosition.X = reader.ReadInt16()
		attack.PreviousMobPosition.Y = reader.ReadInt16()

		if attackType == attackSummon {
			reader.Skip(1)
		}

		if data.IsMesoExplosion {
			data.Hits = reader.ReadByte()
		} else if attackType != attackSummon {
			attack.HitDelay = reader.ReadInt16()
		}

		attack.Damages = make([]int32, data.Hits)

		for j := byte(0); j < data.Hits; j++ {
			dmg := reader.ReadInt32()
			data.TotalDamage += dmg
			attack.Damages[j] = dmg
		}
		data.AttackInfo[i] = attack
	}

	data.PlayerPos.X = reader.ReadInt16()
	data.PlayerPos.Y = reader.ReadInt16()

	return data, true
}
