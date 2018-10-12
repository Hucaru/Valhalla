package handlers

import (
	"github.com/Hucaru/Valhalla/channel"
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/movement"
	"github.com/Hucaru/Valhalla/nx"
	"github.com/Hucaru/Valhalla/packets"
)

func handleMobControl(conn mnet.MConnChannel, reader maplepacket.Reader) {
	mobID := reader.ReadInt32()
	moveID := reader.ReadInt16()
	nibbles := reader.ReadByte()
	activity := reader.ReadByte()
	skillID := reader.ReadByte()
	skillLevel := reader.ReadByte()
	option := reader.ReadInt16()

	reader.ReadInt32()

	nFrags := reader.ReadByte()

	parsedActivity := int8(uint8(activity) >> 1)

	inRange := func(val, min, max int8) bool {
		if (val >= min) && (val <= max) {
			return true
		}

		return false
	}

	isAttack := inRange(parsedActivity, 12, 20)
	isSkill := inRange(parsedActivity, 21, 25)

	var attackID int8

	if isAttack {
		attackID = parsedActivity - 12
	} else {
		attackID = -1
	}

	allowedToUseSkill := false

	if (nibbles & 0x0F) != 0 {
		allowedToUseSkill = true
	}

	channel.Players.OnCharacterFromConn(conn, func(char *channel.MapleCharacter) {
		channel.Mobs.OnMob(char.GetCurrentMap(), mobID, func(mob *channel.MapleMob) {

			if isAttack || isSkill {
				if isAttack {
					attack, valid := nx.GetMobAttack(mob.GetID(), attackID)

					if !valid {
						return
					}

					if int32(mob.GetMp())-int32(attack.ConMP) > -1 {
						mob.SetMp(mob.GetMp() - attack.ConMP)
					} else {
						mob.SetMp(0)
					}

				} else {

					if skillID != mob.GetNextSkillID() || skillLevel != mob.GetNextSkillLevel() {
						mob.SetNextSkillID(0)
						mob.SetNextSkillLevel(0)
						return
					}

					mob.UseSkill()
				}
			}

			if allowedToUseSkill {
				mob.ChooseRandomSkill()
			}

			movement.ParseFragments(nFrags, mob, reader)

			conn.Write(packets.MobAck(mobID, moveID, allowedToUseSkill, int16(mob.GetMp()), mob.GetNextSkillID(), mob.GetNextSkillLevel()))

			channel.Maps.GetMap(char.GetCurrentMap()).SendPacketExcept(packets.MobMove(mobID, allowedToUseSkill, activity, skillID, skillLevel, option, reader.GetBuffer()[13:]), conn)
		})
	})

}
