package channelhandlers

import (
	"github.com/Hucaru/Valhalla/maplepacket"
	"github.com/Hucaru/Valhalla/types"
)

// values from WvsGlobal
var movementType = struct {
	normalMovement  byte
	jump            byte
	jumpKb          byte
	immediate       byte
	teleport        byte
	normalMovement2 byte
	flashJump       byte
	assaulter       byte
	falling         byte
}{
	normalMovement:  0,
	jump:            1,
	jumpKb:          2,
	immediate:       3, // GM F1 teleport
	teleport:        4,
	normalMovement2: 5,
	flashJump:       6,
	assaulter:       7,
	falling:         8,
}

func parseMovement(reader maplepacket.Reader) (types.MovementData, types.MovementFrag) {
	// http://mapleref.wikia.com/wiki/Movement

	mData := types.MovementData{}

	mData.OrigX = reader.ReadInt16()
	mData.OrigY = reader.ReadInt16()

	nFrags := reader.ReadByte()

	mData.Frags = make([]types.MovementFrag, nFrags)

	final := types.MovementFrag{}

	for i := byte(0); i < nFrags; i++ {
		frag := types.MovementFrag{}

		frag.MType = reader.ReadByte()

		switch frag.MType {
		case movementType.normalMovement:
			fallthrough
		case movementType.normalMovement2:
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
			frag.Vx = reader.ReadInt16()
			frag.Vy = reader.ReadInt16()
			frag.Stance = reader.ReadByte()
			frag.Duration = reader.ReadInt16()

		case movementType.immediate:
			fallthrough
		case movementType.teleport:
			fallthrough
		case movementType.assaulter:
			frag.X = reader.ReadInt16()
			frag.Y = reader.ReadInt16()
			frag.Foothold = reader.ReadInt16()
			frag.Stance = reader.ReadByte()
			frag.Duration = reader.ReadInt16()

		case movementType.falling:
			frag.Stance = reader.ReadByte()

		default:
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

func generateMovementBytes(moveData types.MovementData) maplepacket.Packet {
	p := maplepacket.NewPacket()

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
