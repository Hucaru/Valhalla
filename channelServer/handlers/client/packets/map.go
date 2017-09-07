package packet

import (
	"github.com/Hucaru/Valhalla/common/constants"
	"github.com/Hucaru/gopacket"
)

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
