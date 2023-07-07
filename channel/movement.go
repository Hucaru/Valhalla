package channel

import (
	"fmt"

	"github.com/Hucaru/Valhalla/mpacket"
)

type movementFrag struct {
	x, y, vx, vy, foothold, duration int16
	stance, mType                    byte
	posSet                           bool
}

type movement struct {
	origX, origY int16
	frags        []movementFrag
}

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

func parseMovement(reader mpacket.Reader) (movement, movementFrag) {
	// http://mapleref.wikia.com/wiki/Movement

	mData := movement{}

	mData.origX = reader.ReadInt16()
	mData.origY = reader.ReadInt16()

	nFrags := reader.ReadByte()

	mData.frags = make([]movementFrag, nFrags)

	final := movementFrag{}

	for i := byte(0); i < nFrags; i++ {
		frag := movementFrag{posSet: false}

		frag.mType = reader.ReadByte()

		switch frag.mType {
		case movementType.normalMovement:
			fallthrough
		case movementType.normalMovement2:
			fallthrough
		case movementType.normalMovement3:
			frag.x = reader.ReadInt16()
			frag.y = reader.ReadInt16()
			frag.vx = reader.ReadInt16()
			frag.vy = reader.ReadInt16()
			frag.foothold = reader.ReadInt16()
			frag.stance = reader.ReadByte()
			frag.duration = reader.ReadInt16()

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
			frag.vx = reader.ReadInt16()
			frag.vy = reader.ReadInt16()
			frag.stance = reader.ReadByte()
			frag.duration = reader.ReadInt16()

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
			frag.x = reader.ReadInt16()
			frag.y = reader.ReadInt16()
			frag.foothold = reader.ReadInt16()
			frag.stance = reader.ReadByte()
			frag.duration = reader.ReadInt16()

		case movementType.falling:
			reader.ReadByte() // what is this

		default:
			fmt.Println("Unknown movement fragment type: ", frag.mType)
			frag.stance = reader.ReadByte()
			frag.duration = reader.ReadInt16()
		}

		final.x = frag.x
		final.y = frag.y
		final.foothold = frag.foothold
		final.stance = frag.stance

		mData.frags[i] = frag
	}

	return mData, final
}

func generateMovementBytes(moveData movement) mpacket.Packet {
	p := mpacket.NewPacket()

	p.WriteInt16(moveData.origX)
	p.WriteInt16(moveData.origY)

	p.WriteByte(byte(len(moveData.frags)))

	for _, frag := range moveData.frags {
		p.WriteByte(frag.mType)

		switch frag.mType {
		case movementType.normalMovement:
			fallthrough
		case movementType.normalMovement2:
			p.WriteInt16(frag.x)
			p.WriteInt16(frag.y)
			p.WriteInt16(frag.vx)
			p.WriteInt16(frag.vy)
			p.WriteInt16(frag.foothold)
			p.WriteByte(frag.stance)
			p.WriteInt16(frag.duration)

		case movementType.jump:
			fallthrough
		case movementType.jumpKb:
			fallthrough
		case movementType.flashJump:
			p.WriteInt16(frag.vx)
			p.WriteInt16(frag.vy)
			p.WriteByte(frag.stance)
			p.WriteInt16(frag.duration)

		case movementType.immediate:
			fallthrough
		case movementType.teleport:
			fallthrough
		case movementType.assaulter:
			p.WriteInt16(frag.x)
			p.WriteInt16(frag.y)
			p.WriteInt16(frag.foothold)
			p.WriteByte(frag.stance)
			p.WriteInt16(frag.duration)

		case movementType.falling:
			p.WriteByte(frag.stance)

		default:
			p.WriteByte(frag.stance)
			p.WriteInt16(frag.duration)
		}
	}

	return p
}

func (data movement) validateChar(player *player) bool {
	// run through the movement data and make sure characters are not moving too fast (going to have to take into account gear and buffs "-_- )

	return true
}

type mob interface {
}

// ValidateMob movement
func (data movement) validateMob(mob mob) bool {
	// run through the movement data and make sure monsters are not moving too fast

	return true
}
