package field

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
	mob "github.com/Hucaru/Valhalla/server/field/mob"
	npc "github.com/Hucaru/Valhalla/server/field/npc"
)

func packetMapPlayerEnter(plr player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCharacterEnterField)
	p.WriteInt32(plr.ID())
	p.WriteString(plr.Name())

	if true {
		p.WriteString("test")
		p.WriteInt16(0)
		p.WriteByte(0)
		p.WriteByte(0)
		p.WriteInt16(0)
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
	p.WriteInt16(plr.Foothold())
	p.WriteInt32(0) // ?

	return p
}

func packetMapPlayerLeft(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCharacterLeaveField)
	p.WriteInt32(charID)

	return p
}

func packetNpcShow(npc npc.Data) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelNpcShow)
	p.WriteInt32(npc.SpawnID())
	p.WriteInt32(npc.ID())
	p.WriteInt16(npc.Pos().X())
	p.WriteInt16(npc.Pos().Y())

	p.WriteBool(!npc.FaceLeft())

	p.WriteInt16(npc.Foothold())
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
