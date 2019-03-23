package channel

import (
	"fmt"

	"github.com/Hucaru/Valhalla/game"
	"github.com/Hucaru/Valhalla/game/def"
	"github.com/Hucaru/Valhalla/mpacket"
)

// values from WvsGlobal
var movementType = struct {
	normalMovement   byte
	jump             byte
	jumpKb           byte
	immediate        byte
	teleport         byte
	normalMovement2  byte
	flashJump        byte
	assaulter        byte
	falling          byte
	equipMovement    byte
	jumpdownMovement byte
	normalMovement3  byte
}{
	normalMovement:   0,
	jump:             1,
	jumpKb:           2,
	immediate:        3, // GM F1 teleport
	teleport:         4,
	normalMovement2:  5,
	flashJump:        6,
	assaulter:        7,
	falling:          8,
	equipMovement:    10,
	jumpdownMovement: 11,
	normalMovement3:  17,
}

func parseMovement(reader mpacket.Reader) (def.MovementData, def.MovementFrag) {
	// http://mapleref.wikia.com/wiki/Movement

	mData := def.MovementData{}

	mData.OrigX = reader.ReadInt16()
	mData.OrigY = reader.ReadInt16()

	nFrags := reader.ReadByte()

	mData.Frags = make([]def.MovementFrag, nFrags)

	final := def.MovementFrag{}

	for i := byte(0); i < nFrags; i++ {
		frag := def.MovementFrag{PosSet: false}

		frag.MType = reader.ReadByte()

		switch frag.MType {
		case movementType.normalMovement:
			fallthrough
		case movementType.normalMovement2:
			fallthrough
		case movementType.normalMovement3:
			frag.X = reader.ReadInt16()
			frag.Y = reader.ReadInt16()
			frag.Vx = reader.ReadInt16()
			frag.Vy = reader.ReadInt16()
			frag.Foothold = reader.ReadInt16()
			frag.Stance = reader.ReadByte()
			frag.Duration = reader.ReadInt16()

		case movementType.jump:
			fallthrough
		case movementType.jumpKb:
			fallthrough
		case movementType.flashJump:
			fallthrough
		case 12:
			fallthrough
		case 13:
			fallthrough
		case 16:
			frag.Vx = reader.ReadInt16()
			frag.Vy = reader.ReadInt16()
			frag.Stance = reader.ReadByte()
			frag.Duration = reader.ReadInt16()

		case movementType.immediate:
			fallthrough
		case movementType.teleport:
			fallthrough
		case movementType.assaulter:
			fallthrough
		// case movementType.falling:
		// 	fallthrough
		case 9:
			fallthrough
		case 14:
			frag.X = reader.ReadInt16()
			frag.Y = reader.ReadInt16()
			frag.Foothold = reader.ReadInt16()
			frag.Stance = reader.ReadByte()
			frag.Duration = reader.ReadInt16()

		case movementType.falling:
			reader.ReadByte() // what is this

		default:
			fmt.Println("unkown movement fragment type: ", frag.MType)
			frag.Stance = reader.ReadByte()
			frag.Duration = reader.ReadInt16()
		}

		final.X = frag.X
		final.Y = frag.Y
		final.Foothold = frag.Foothold
		final.Stance = frag.Stance

		mData.Frags[i] = frag
	}

	return mData, final
}

func generateMovementBytes(moveData def.MovementData) mpacket.Packet {
	p := mpacket.NewPacket()

	p.WriteInt16(moveData.OrigX)
	p.WriteInt16(moveData.OrigY)

	p.WriteByte(byte(len(moveData.Frags)))

	for _, frag := range moveData.Frags {
		p.WriteByte(frag.MType)

		switch frag.MType {
		case movementType.normalMovement:
			fallthrough
		case movementType.normalMovement2:
			p.WriteInt16(frag.X)
			p.WriteInt16(frag.Y)
			p.WriteInt16(frag.Vx)
			p.WriteInt16(frag.Vy)
			p.WriteInt16(frag.Foothold)
			p.WriteByte(frag.Stance)
			p.WriteInt16(frag.Duration)

		case movementType.jump:
			fallthrough
		case movementType.jumpKb:
			fallthrough
		case movementType.flashJump:
			p.WriteInt16(frag.Vx)
			p.WriteInt16(frag.Vy)
			p.WriteByte(frag.Stance)
			p.WriteInt16(frag.Duration)

		case movementType.immediate:
			fallthrough
		case movementType.teleport:
			fallthrough
		case movementType.assaulter:
			p.WriteInt16(frag.X)
			p.WriteInt16(frag.Y)
			p.WriteInt16(frag.Foothold)
			p.WriteByte(frag.Stance)
			p.WriteInt16(frag.Duration)

		case movementType.falling:
			p.WriteByte(frag.Stance)

		default:
			p.WriteByte(frag.Stance)
			p.WriteInt16(frag.Duration)
		}
	}

	return p
}

func validateCharMovement(char def.Character, moveData def.MovementData) bool {
	// run through the movement data and make sure characters are not moving too fast (going to have to take into account gear and buffs "-_- )

	return true
}

func validateMobMovement(mob game.Mob, moveData def.MovementData) bool {
	// run through the movement data and make sure monsters are not moving too fast

	return true
}
