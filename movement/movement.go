package movement

import (
	"fmt"
	"log"

	"github.com/Hucaru/Valhalla/maplepacket"
)

type fragObj interface {
	SetX(int16)
	SetY(int16)
	SetState(byte)
	SetFoothold(int16)
}

func ParseFragments(nFrags byte, life fragObj, reader maplepacket.Reader) {
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
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Panic in handling movement packet:", reader)
		}
	}()

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

			reader.ReadInt16() // velx
			reader.ReadInt16() // vely

			foothold := reader.ReadInt16()

			state := reader.ReadByte()
			reader.ReadUint16() //duration

			life.SetX(posX)
			life.SetY(posY)
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
			reader.ReadInt16() // velX
			reader.ReadInt16() // velY

			reader.ReadUint16()

			foothold := reader.ReadInt16()
			reader.ReadUint16() //duration

			life.SetX(posX)
			life.SetY(posY)
			life.SetFoothold(foothold)
		case 0x08:
			reader.ReadByte()
		default:
			log.Println("Unkown movement type received", movementType, reader.GetRestAsBytes())
		}
	}
}
