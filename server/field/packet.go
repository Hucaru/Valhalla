package field

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/field/mob"
	"github.com/Hucaru/Valhalla/server/field/npc"
)

func packetMapPlayerEnter(plr player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCharacterEnterField)
	p.WriteInt32(plr.ID())
	p.WriteString(plr.Name())

	if true {
		p.WriteString("[Admins]")
		p.WriteInt16(1030) // logo background
		p.WriteByte(3)     // logo bg colour
		p.WriteInt16(4017) // logo
		p.WriteByte(2)     // logo colour
		p.WriteInt32(0)
		p.WriteInt32(0)
	} else {
		p.WriteInt32(0)
		p.WriteInt32(0)
		p.WriteInt32(0)
		p.WriteInt32(0)
	}

	p.WriteBytes(plr.DisplayBytes())

	p.WriteInt32(0)             // ?
	p.WriteInt32(0)             // ?
	p.WriteInt32(0)             // ?
	p.WriteInt32(plr.ChairID()) // 0 means no chair in use, stance needs to be changed to match

	p.WriteInt16(plr.Pos().X())
	p.WriteInt16(plr.Pos().Y())
	p.WriteByte(plr.Stance())
	p.WriteInt16(plr.Pos().Foothold())
	p.WriteInt32(0) // ?

	return p
}

func packetMapPlayerLeft(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCharacterLeaveField)
	p.WriteInt32(charID)

	return p
}

func packetPlayerMove(charID int32, bytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerMovement)
	p.WriteInt32(charID)
	p.WriteBytes(bytes)

	return p
}

func packetNpcShow(npc npc.Data) mpacket.Packet {
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

func packetMobShow(mob mob.Data) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelShowMob)
	p.Append(mob.DisplayBytes())

	return p
}

func packetMobRemove(spawnID int32, deathType byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRemoveMob)
	p.WriteInt32(spawnID)
	p.WriteByte(deathType)

	return p
}

func packetMobShowBossHP(mobID, hp, maxHP int32, colourFg, colourBg byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMapEffect)
	p.WriteByte(5)
	p.WriteInt32(mobID)
	p.WriteInt32(hp)
	p.WriteInt32(maxHP)
	p.WriteByte(colourFg)
	p.WriteByte(colourBg)

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

func packetMapShowGameBox(displayBytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoomBox)
	p.WriteBytes(displayBytes)

	return p
}

func packetMapRemoveGameBox(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRoomBox)
	p.WriteInt32(charID)
	p.WriteInt32(0)

	return p
}
