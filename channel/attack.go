package channel

import (
	"github.com/Hucaru/Valhalla/channel/message"
	"github.com/Hucaru/Valhalla/channel/player"
	"github.com/Hucaru/Valhalla/channel/pos"
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mnet"
	"github.com/Hucaru/Valhalla/mpacket"
)

func (server Server) playerMeleeSkill(conn mnet.Client, reader mpacket.Reader) {
	plr, err := server.players.getFromConn(conn)

	if err != nil {
		conn.Send(message.PacketMessageRedText(err.Error()))
		return
	}

	data, valid := getAttackInfo(reader, *plr, attackMelee)

	if !valid {
		return
	}

	field, ok := server.fields[plr.MapID()]

	if !ok {
		return
	}

	inst, err := field.GetInstance(plr.InstanceID())

	if err != nil {
		conn.Send(message.PacketMessageRedText(err.Error()))
		return
	}

	// if player in party extract

	packetSkillMelee := func(char player.Data, ad attackData) mpacket.Packet {
		p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerUseMeleeSkill)
		p.WriteInt32(char.ID())
		p.WriteByte(ad.targets*0x10 + ad.hits)
		p.WriteByte(ad.skillLevel)

		if ad.skillLevel != 0 {
			p.WriteInt32(ad.skillID)
		}

		if ad.facesLeft {
			p.WriteByte(ad.action | (1 << 7))
		} else {
			p.WriteByte(ad.action | 0)
		}

		p.WriteByte(ad.attackType)

		p.WriteByte(char.Skills()[ad.skillID].Mastery)
		p.WriteInt32(ad.projectileID)

		for _, info := range ad.attackInfo {
			p.WriteInt32(info.spawnID)
			p.WriteByte(info.hitAction)

			if ad.isMesoExplosion {
				p.WriteByte(byte(len(info.damages)))
			}

			for _, dmg := range info.damages {
				p.WriteInt32(dmg)
			}
		}

		return p
	}

	inst.SendExcept(packetSkillMelee(*plr, data), conn)

	for _, attack := range data.attackInfo {
		inst.LifePool().MobDamaged(attack.spawnID, plr, attack.damages...)
	}
}

// Following logic lifted from WvsGlobal
const (
	attackMelee = iota
	attackRanged
	attackMagic
	attackSummon
)

type attackInfo struct {
	spawnID                                                int32
	hitAction, foreAction, frameIndex, calcDamageStatIndex byte
	facesLeft                                              bool
	hitPosition, previousMobPosition                       pos.Data
	hitDelay                                               int16
	damages                                                []int32
}

type attackData struct {
	skillID, summonType, totalDamage, projectileID int32
	isMesoExplosion, facesLeft                     bool
	option, action, attackType                     byte
	targets, hits, skillLevel                      byte

	attackInfo []attackInfo
	playerPos  pos.Data
}

func getAttackInfo(reader mpacket.Reader, player player.Data, attackType int) (attackData, bool) {
	data := attackData{}

	if player.HP() == 0 {
		return data, false
	}

	// speed hack check
	if false && (reader.Time-player.LastAttackPacketTime() < 350) {
		return data, false
	}

	player.SetLastAttackPacketTime(reader.Time)

	if attackType != attackSummon {
		tByte := reader.ReadByte()
		skillID := reader.ReadInt32()

		if _, ok := player.Skills()[skillID]; !ok && skillID != 0 {
			return data, false
		}

		data.skillID = skillID

		if data.skillID != 0 {
			data.skillLevel = player.Skills()[skillID].Level
		}

		// if meso explosion data.IsMesoExplosion = true

		data.targets = tByte / 0x10
		data.hits = tByte % 0x10
		data.option = reader.ReadByte()

		tmp := reader.ReadByte()

		data.action = tmp & 0x7F
		data.facesLeft = (tmp >> 7) == 1
		data.attackType = reader.ReadByte()
	} else {
		data.summonType = reader.ReadInt32()
		data.attackType = reader.ReadByte()
		data.targets = 1
		data.hits = 1
	}

	reader.Skip(4) //checksum info?

	if attackType == attackRanged {
		projectileSlot := reader.ReadInt16() // star/arrow slot
		if projectileSlot == 0 {
			// if soul arrow is not set check for hacks
		} else {
			data.projectileID = -1

			for _, item := range player.Use() {
				if item.SlotID() == projectileSlot {
					data.projectileID = item.ID()
				}
			}
		}
		reader.ReadByte() // ?
		reader.ReadByte() // ?
		reader.ReadByte() // ?
	}

	data.attackInfo = make([]attackInfo, data.targets)

	for i := byte(0); i < data.targets; i++ {
		attack := attackInfo{}
		attack.spawnID = reader.ReadInt32()
		attack.hitAction = reader.ReadByte()

		tmp := reader.ReadByte()
		attack.foreAction = tmp & 0x7F
		attack.facesLeft = (tmp >> 7) == 1
		attack.frameIndex = reader.ReadByte()

		if !data.isMesoExplosion {
			attack.calcDamageStatIndex = reader.ReadByte()
		}

		attack.hitPosition.SetX(reader.ReadInt16())
		attack.hitPosition.SetY(reader.ReadInt16())

		attack.previousMobPosition.SetX(reader.ReadInt16())
		attack.previousMobPosition.SetY(reader.ReadInt16())

		if attackType == attackSummon {
			reader.Skip(1)
		}

		if data.isMesoExplosion {
			data.hits = reader.ReadByte()
		} else if attackType != attackSummon {
			attack.hitDelay = reader.ReadInt16()
		}

		attack.damages = make([]int32, data.hits)

		for j := byte(0); j < data.hits; j++ {
			dmg := reader.ReadInt32()
			data.totalDamage += dmg
			attack.damages[j] = dmg
		}
		data.attackInfo[i] = attack
	}

	data.playerPos.SetX(reader.ReadInt16())
	data.playerPos.SetY(reader.ReadInt16())

	return data, true
}
