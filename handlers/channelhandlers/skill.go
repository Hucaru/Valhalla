package channelhandlers

import (
	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/packets"
	"github.com/Hucaru/Valhalla/types"

	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/mnet"
)

func playerMeleeSkill(conn mnet.MConnChannel, reader maplepacket.Reader) {
	player := game.GetPlayerFromConn(conn)
	data, valid := getAttackInfo(reader, player, attackMelee)

	if !valid {
		return
	}

	char := player.Char()

	// fix the damange values

	game.SendToMapExcept(char.CurrentMap, packets.SkillMelee(char, data), conn)
}

func playerRangedSkill(conn mnet.MConnChannel, reader maplepacket.Reader) {
	player := game.GetPlayerFromConn(conn)
	data, valid := getAttackInfo(reader, player, attackRanged)

	if !valid {
		return
	}

	char := player.Char()

	// fix the damange values

	game.SendToMapExcept(char.CurrentMap, packets.SkillRanged(char, data), conn)
}

func playerMagicSkill(conn mnet.MConnChannel, reader maplepacket.Reader) {
	player := game.GetPlayerFromConn(conn)
	data, valid := getAttackInfo(reader, player, attackMagic)

	if !valid {
		return
	}

	char := player.Char()

	// fix the damange values

	game.SendToMapExcept(char.CurrentMap, packets.SkillMagic(char, data), conn)

	for _, ai := range data.AttackInfo {
		game.DamageMob(player, player.Char().CurrentMap, ai.SpawnID, ai.Damages)
	}

	switch data.SkillID {
	default:
		conn.Send(packets.PlayerNoChange())
	}
}

func playerSpecialSkill(conn mnet.MConnChannel, reader maplepacket.Reader) {
	player := game.GetPlayerFromConn(conn)

	skillID := reader.ReadInt32()
	skillLevel := reader.ReadByte()

	char := player.Char()

	if skill, ok := char.Skills[skillID]; !ok || skill.Level != skillLevel {
		return
	}

	game.SendToMapExcept(char.CurrentMap, packets.SkillAnimation(char.ID, skillID, skillLevel), conn)

	switch skillID {
	default:
		conn.Send(packets.PlayerNoChange())
	}
}

// Following logic lifted from WvsGlobal
const (
	attackMelee = iota
	attackRanged
	attackMagic
	attackSummon
)

func getAttackInfo(reader maplepacket.Reader, player game.Player, attackType int) (types.AttackData, bool) {
	data := types.AttackData{}

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

	if attackType == attackRanged {

	}

	reader.Skip(4) // some sort of checksum?

	data.AttackInfo = make([]types.AttackInfo, data.Targets)

	for i := byte(0); i < data.Targets; i++ {
		attack := types.AttackInfo{}
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

	if data.Hits != 0 {
		// validate dmg numbers against mob info
		for _, dmg := range data.AttackInfo {
			_ = game.GetMobFromMapAndSpawnID(player.Char().CurrentMap, dmg.SpawnID)
		}

	}

	return data, true
}
