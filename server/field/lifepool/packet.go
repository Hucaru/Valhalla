package lifepool

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/field/lifepool/mob"
	"github.com/Hucaru/Valhalla/server/field/lifepool/npc"
)

func packetNpcShow(npc *npc.Data) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcShow)
	p.WriteInt32(npc.SpawnID())
	p.WriteInt32(npc.ID())
	p.WriteInt16(npc.Pos().X())
	p.WriteInt16(npc.Pos().Y())

	p.WriteBool(!npc.FaceLeft())

	p.WriteInt16(npc.Pos().Foothold())
	p.WriteInt16(npc.Rx0())
	p.WriteInt16(npc.Rx1())

	return p
}

func packetNpcRemove(npcID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcRemove)
	p.WriteInt32(npcID)

	return p
}

func packetMobShow(mob *mob.Data) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelShowMob)
	p.Append(mob.DisplayBytes())

	return p
}

func packetMobMove(mobID int32, allowedToUseSkill bool, action byte, skillData uint32, moveBytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMoveMob)
	p.WriteInt32(mobID)
	p.WriteBool(allowedToUseSkill)
	p.WriteByte(action)
	p.WriteUint32(skillData)
	p.WriteBytes(moveBytes)

	return p

}

func packetMobRemove(spawnID int32, deathType byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRemoveMob)
	p.WriteInt32(spawnID)
	p.WriteByte(deathType)

	return p
}

func packetMobShowBossHP(mobID, hp, maxHP int32, colourFg, colourBg byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMapEffect) // field effect
	p.WriteByte(5)                                             // 1, tremble effect, 3 - mapEffect (string), 4 - mapSound (string), arbitary - environemnt change int32 followed by string
	p.WriteInt32(mobID)
	p.WriteInt32(hp)
	p.WriteInt32(maxHP)
	p.WriteByte(colourFg)
	p.WriteByte(colourBg)

	return p
}
