package packets

import (
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/Valhalla/common/nx"
	"github.com/Hucaru/gopacket"
)

func SpawnNPC(index uint32, npc nx.Life) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(0x97)
	p.WriteUint32(index)
	p.WriteUint32(npc.ID)
	p.WriteInt16(npc.X)
	p.WriteInt16(npc.Y)

	if npc.F == 0 {
		p.WriteByte(1)
	} else {
		p.WriteByte(0)
	}

	p.WriteInt16(npc.Fh)
	p.WriteInt16(npc.Rx0)
	p.WriteInt16(npc.Rx1)

	p.WriteByte(0x9B)
	p.WriteByte(0x1)
	p.WriteUint32(npc.ID)
	p.WriteUint32(npc.ID)
	p.WriteInt16(npc.X)
	p.WriteInt16(npc.Y)

	if npc.F == 0 {
		p.WriteByte(1)
	} else {
		p.WriteByte(0)
	}
	if npc.F == 0 {
		p.WriteByte(1)
	} else {
		p.WriteByte(0)
	}
	p.WriteInt16(npc.Fh)
	p.WriteInt16(npc.Rx0)
	p.WriteInt16(npc.Rx1)

	return p
}

func sendLevelUpAnimation() gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_LEVEL_UP_ANIMATION)
	p.WriteByte(0) // animation to use

	return p
}

func spawnDoor(x int16, y int16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_SPAWN_DOOR)
	p.WriteByte(0)  // ?
	p.WriteInt32(0) // ?
	p.WriteInt16(x) // x pos
	p.WriteInt16(y) // y pos

	return p
}

func removeDoor(x int16, y int16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_SPAWN_DOOR)
	p.WriteByte(0)  // ?
	p.WriteInt32(0) // ?
	p.WriteInt16(x) // x pos
	p.WriteInt16(y) // y pos

	return p
}

func quizQuestionAndAnswer(isQuestion bool, questionSet byte, questionNumber int16) gopacket.Packet {
	p := gopacket.NewPacket()
	p.WriteByte(constants.SEND_CHANNEL_QUIZ_Q_AND_A)
	if isQuestion {
		p.WriteByte(0x01)
	} else {
		p.WriteByte(0x00)
	}
	p.WriteByte(questionSet)
	p.WriteInt16(questionNumber)

	return p
}
