package maps

import (
	"github.com/Hucaru/Valhalla/character"
	"github.com/Hucaru/Valhalla/constants"
	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/gopacket"
)

func playerEnterMapPacket(char *character.Character) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_CHARCTER_ENTER_FIELD)
	p.WriteUint32(char.GetCharID()) // player id
	p.WriteString(char.GetName())   // char name
	p.WriteUint32(0)                // map buffs?
	p.WriteUint32(0)                // map buffs?
	p.WriteUint32(0)                // map buffs?
	p.WriteUint32(0)                // map buffs?

	character.WriteDisplayCharacter(char, &p)

	p.WriteUint32(0)                // ?
	p.WriteUint32(0)                // ?
	p.WriteUint32(0)                // ?
	p.WriteUint32(char.GetCharID()) // 0 means no chair in use, stance needs to be changed to match

	p.WriteInt16(char.GetX())
	p.WriteInt16(char.GetY())

	p.WriteByte(char.GetState())
	p.WriteInt16(char.GetFoothold())
	p.WriteUint32(0) // ?

	return p
}

func ChangeMapPacket(mapID uint32, channelID uint32, mapPos byte, hp uint16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_WARP_TO_MAP)
	p.WriteUint32(channelID)
	p.WriteByte(0) // character portal counter
	p.WriteByte(0) // Is connecting
	p.WriteUint32(mapID)
	p.WriteByte(mapPos)
	p.WriteUint16(hp)
	p.WriteByte(0) // flag for more reading

	return p
}

func playerLeftMapPacket(charID uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_CHARCTER_LEAVE_FIELD)
	p.WriteUint32(charID)

	return p
}

func showMobPacket(spawnID uint32, mob interfaces.Mob, isNewSpawn bool) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_SHOW_MOB)
	p.WriteUint32(spawnID)
	p.WriteByte(0x01)
	p.WriteUint32(mob.GetID())

	p.WriteUint32(0) //?

	p.WriteInt16(mob.GetX())
	p.WriteInt16(mob.GetY())
	p.WriteByte(0x02) // direction 0x02 faces right, 0x03 faces left
	p.WriteInt16(mob.GetFoothold())
	p.WriteInt16(mob.GetFoothold())

	if isNewSpawn {
		p.WriteByte(0xFE)
	} else {
		p.WriteByte(0xFF)
	}

	p.WriteUint32(0)

	return p
}

func showNpcPacket(index uint32, npc interfaces.Npc) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_NPC_SPAWN_1)
	p.WriteUint32(index)
	p.WriteUint32(npc.GetID())
	p.WriteInt16(npc.GetX())
	p.WriteInt16(npc.GetY())

	p.WriteByte(1 - npc.GetFace())

	p.WriteInt16(npc.GetFoothold())
	p.WriteInt16(npc.GetRx0())
	p.WriteInt16(npc.GetRx1())

	p.WriteByte(constants.SEND_CHANNEL_NPC_SPAWN_2)
	p.WriteByte(0x1)
	p.WriteUint32(npc.GetID())
	p.WriteUint32(npc.GetID())
	p.WriteInt16(npc.GetX())
	p.WriteInt16(npc.GetY())

	p.WriteByte(1 - npc.GetFace())
	p.WriteByte(1 - npc.GetFace())

	p.WriteInt16(npc.GetFoothold())
	p.WriteInt16(npc.GetRx0())
	p.WriteInt16(npc.GetRx1())

	return p
}

func playerEmotionPacket(playerID uint32, emotion uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_PLAYER_EMOTION)
	p.WriteUint32(playerID)
	p.WriteUint32(emotion)

	return p
}

func controlMobPacket(spawnID uint32, mob interfaces.Mob, isNewSpawn bool) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_CONTROL_MOB)
	p.WriteByte(0x01) // if mob is agroed or not. 0x01 is not agroed, other values means agroed
	p.WriteUint32(spawnID)
	p.WriteByte(0x01)
	p.WriteUint32(mob.GetID())

	p.WriteUint32(0) // ?

	p.WriteInt16(mob.GetX())
	p.WriteInt16(mob.GetY())
	p.WriteByte(0x02)               // which direction it faces?
	p.WriteInt16(mob.GetFoothold()) // foothold to oscillate around
	p.WriteInt16(mob.GetFoothold()) // spawn foothold

	if isNewSpawn {
		p.WriteByte(0xFE)
	} else {
		p.WriteByte(0xFF)
	}
	p.WriteUint32(0)

	return p
}

func controlAckPacket(mobID uint32, moveID uint16, useSkill bool, skill byte, level byte, mp uint16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_CONTROL_MOB_ACK)
	p.WriteUint32(mobID)
	p.WriteUint16(moveID)
	p.WriteBool(useSkill)
	p.WriteByte(0)
	p.WriteUint16(mp)
	p.WriteByte(skill)
	p.WriteByte(level)

	return p
}

func moveMobPacket(mobID uint32, skillUsed bool, skill byte, unknown uint32, buf []byte) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNE_MOVE_MOB)
	p.WriteUint32(mobID)
	p.WriteBool(skillUsed)
	p.WriteByte(skill)
	p.WriteUint32(unknown)
	p.WriteBytes(buf)

	return p

}

func endMobControlPacket(mobID uint32) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_CONTROL_MOB)
	p.WriteByte(0)
	p.WriteUint32(mobID)

	return p
}

func removeMobPacket(mobID uint32, deathType byte) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_REMOVE_MOB)
	p.WriteUint32(mobID)
	p.WriteByte(deathType)

	return p
}
