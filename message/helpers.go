package message

import (
	"github.com/Hucaru/Valhalla/interfaces"
	"github.com/Hucaru/Valhalla/maplepacket"
)

var charsPtr interfaces.Characters

// RegisterCharactersObj -
func RegisterCharactersObj(chars interfaces.Characters) {
	charsPtr = chars
}

// CreateNoticePacket -
func CreateNoticePacket(msg string) maplepacket.Packet {
	return noticePacket(msg)
}

func CreateDialogeMessage(msg string) maplepacket.Packet {
	return dialogueBoxPacket(msg)
}
