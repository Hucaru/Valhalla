package character

import (
	"math"

	"github.com/Hucaru/gopacket"
)

func WriteDisplayCharacter(char *Character, p *gopacket.Packet) {
	p.WriteByte(char.GetGender()) // gender
	p.WriteByte(char.GetSkin())   // skin
	p.WriteUint32(char.GetFace()) // face
	p.WriteByte(0x00)             // ?
	p.WriteUint32(char.GetHair()) // hair
	cashWeapon := uint32(0)

	for _, b := range char.GetEquips() {
		if b.GetSlotID() < 0 && b.GetSlotID() > -20 {
			p.WriteByte(byte(math.Abs(float64(b.GetSlotID()))))
			p.WriteUint32(b.GetItemID())
		}
	}

	for _, b := range char.GetEquips() {
		if b.GetSlotID() < -100 {
			if b.GetSlotID() == -111 {
				cashWeapon = b.GetItemID()
			} else {
				p.WriteByte(byte(math.Abs(float64(b.GetSlotID() + 100))))
				p.WriteUint32(b.GetItemID())
			}
		}
	}

	p.WriteByte(0xFF)
	// What items go here?
	p.WriteByte(0xFF)
	p.WriteUint32(cashWeapon)
}
