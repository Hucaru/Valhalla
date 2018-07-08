package character

import (
	"math"

	"github.com/Hucaru/Valhalla/maplepacket"
)

func WriteDisplayCharacter(char Character, p *maplepacket.Packet) {
	p.WriteByte(char.GetGender()) // gender
	p.WriteByte(char.GetSkin())   // skin
	p.WriteInt32(char.GetFace())  // face
	p.WriteByte(0x00)             // ?
	p.WriteInt32(char.GetHair())  // hair
	cashWeapon := int32(0)

	for _, b := range char.GetItems() {
		if b.GetSlotNumber() < 0 && b.GetSlotNumber() > -20 {
			p.WriteByte(byte(math.Abs(float64(b.GetSlotNumber()))))
			p.WriteInt32(b.GetItemID())
		}
	}

	for _, b := range char.GetItems() {
		if b.GetSlotNumber() < -100 {
			if b.GetSlotNumber() == -111 {
				cashWeapon = b.GetItemID()
			} else {
				p.WriteByte(byte(math.Abs(float64(b.GetSlotNumber() + 100))))
				p.WriteInt32(b.GetItemID())
			}
		}
	}

	p.WriteByte(0xFF)
	// What items go here?
	p.WriteByte(0xFF)
	p.WriteInt32(cashWeapon)
}
