package player

import (
	"log"

	"github.com/Hucaru/Valhalla/channelServer/maps"
	"github.com/Hucaru/Valhalla/channelServer/playerConn"
	"github.com/Hucaru/gopacket"
)

// TODO: Add cheat detection - use an audit thread maybe? to not block current socket
func HandlePlayerMovement(reader gopacket.Reader, conn *playerConn.Conn) {
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
	reader.ReadBytes(5) // used in movement validation
	char := conn.GetCharacter()

	nFragaments := reader.ReadByte()

	for i := byte(0); i < nFragaments; i++ {
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

			reader.ReadUint16()

			state := reader.ReadByte()
			duration := reader.ReadUint16()

			// Do I need to apply kinematics equations here?
			char.SetX(posX + velX*int16(duration))
			char.SetY(posY + velY*int16(duration))
			char.SetState(state)

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
			reader.ReadUint16() // duration

			char.SetState(state)

		// Instant movement
		case 0x03:
			fallthrough
		case 0x04:
			fallthrough
		case 0x07:
			fallthrough
		case 0x08:
			fallthrough
		case 0x09:
			fallthrough
		case 0x014:
			posX := reader.ReadInt16()
			posY := reader.ReadInt16()
			reader.ReadInt16() // velX
			reader.ReadInt16() // velY

			state := reader.ReadByte()

			char.SetX(posX)
			char.SetY(posY)
			char.SetState(state)

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

			foothold := reader.ReadUint16()
			duration := reader.ReadUint16()

			char.SetX(posX + velX*int16(duration))
			char.SetY(posY + velY*int16(duration))
			char.SetFh(foothold)
		default:
			log.Println("Unkown movement type received", movementType, reader.GetRestAsBytes())

		}
	}

	reader.GetRestAsBytes() // used in movement validation

	maps.PlayerMove(conn, reader.GetBuffer()[2:])
}
