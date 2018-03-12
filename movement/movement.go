package movement

import (
	"log"

	"github.com/Hucaru/gopacket"

	"github.com/Hucaru/Valhalla/interfaces"
)

func ParseFragments(nFrags byte, life interfaces.FragObj, reader gopacket.Reader) {
	// http://mapleref.wikia.com/wiki/Movement
	/*
		State enum:
			left / right: Action
			3 / 2: Walk
			5 / 4: Standing
			7 / 6: Jumping & Falling
			9 / 8: Normal attack
			11 / 10: Prone
			13 / 12: Rope
			15 / 14: Ladder
	*/
	for frag := byte(0); frag < nFrags; frag++ {
		movementType := reader.ReadByte()

		switch movementType { // Movement type
		// Absolute movement
		case 0x00: // normal move
			fallthrough
		case 0x05: // normal move
			fallthrough
		case 0x17:
			posX := reader.ReadInt16()
			posY := reader.ReadInt16()

			velX := reader.ReadInt16()
			velY := reader.ReadInt16()

			foothold := reader.ReadInt16()

			state := reader.ReadByte()
			duration := reader.ReadUint16()

			life.SetX(posX + velX*int16(duration))
			life.SetY(posY + velY*int16(duration))
			life.SetFoothold(foothold)
			life.SetState(state)

		// Relative movement
		case 0x01: // jump
			fallthrough
		case 0x02:
			fallthrough
		case 0x06:
			fallthrough
		case 0x12:
			fallthrough
		case 0x13:
			fallthrough
		case 0x16:
			reader.ReadInt16() // velX
			reader.ReadInt16() // velY

			state := reader.ReadByte()
			foothold := reader.ReadInt16()

			life.SetState(state)
			life.SetFoothold(foothold)

		// Instant movement
		case 0x03:
			fallthrough
		case 0x04: // teleport
			fallthrough
		case 0x07: // assaulter
			fallthrough

		case 0x09:
			fallthrough
		case 0x014:
			posX := reader.ReadInt16()
			posY := reader.ReadInt16()
			reader.ReadInt16() // velX
			reader.ReadInt16() // velY

			state := reader.ReadByte()

			life.SetX(posX)
			life.SetY(posY)
			life.SetState(state)

		// Equip movement
		case 0x10:
			reader.ReadByte() // ?

		// Jump down movement
		case 0x11:
			posX := reader.ReadInt16()
			posY := reader.ReadInt16()
			velX := reader.ReadInt16()
			velY := reader.ReadInt16()

			reader.ReadUint16()

			foothold := reader.ReadInt16()
			duration := reader.ReadUint16()

			life.SetX(posX + velX*int16(duration))
			life.SetY(posY + velY*int16(duration))
			life.SetFoothold(foothold)
		case 0x08:
			reader.ReadByte()
		default:
			log.Println("Unkown movement type received", movementType, reader.GetRestAsBytes())
		}
	}
}
