package field

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/pos"
)

func packetMapPlayerEnter(plr player) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCharacterEnterField)
	p.WriteInt32(plr.ID())
	p.WriteString(plr.Name())

	if true {
		p.WriteString("[Admins]")
		p.WriteInt16(1030) // logo background
		p.WriteByte(3)     // logo bg colour
		p.WriteInt16(4017) // logo
		p.WriteByte(2)     // logo colour
		p.WriteInt32(0)
		p.WriteInt32(0)
	} else {
		p.WriteInt32(0)
		p.WriteInt32(0)
		p.WriteInt32(0)
		p.WriteInt32(0)
	}

	p.WriteBytes(plr.DisplayBytes())

	p.WriteInt32(0)             // ?
	p.WriteInt32(0)             // ?
	p.WriteInt32(0)             // ?
	p.WriteInt32(plr.ChairID()) // 0 means no chair in use, stance needs to be changed to match

	p.WriteInt16(plr.Pos().X())
	p.WriteInt16(plr.Pos().Y())
	p.WriteByte(plr.Stance())
	p.WriteInt16(plr.Pos().Foothold())
	p.WriteInt32(0) // ?

	return p
}

func packetMapPlayerLeft(charID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelCharacterLeaveField)
	p.WriteInt32(charID)

	return p
}

func packetPlayerMove(charID int32, bytes []byte) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelPlayerMovement)
	p.WriteInt32(charID)
	p.WriteBytes(bytes)

	return p
}

func packetMapSpawnMysticDoor(spawnID int32, pos pos.Data, instant bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelSpawnDoor)
	p.WriteBool(instant)
	p.WriteInt32(spawnID)
	p.WriteInt16(pos.X())
	p.WriteInt16(pos.Y())

	return p
}

func packetMapSpawnTownMysticDoor(dstMap int32, destPos pos.Data) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelTownPortal)
	p.WriteInt32(dstMap)
	p.WriteInt32(dstMap)
	p.WriteInt16(destPos.X())
	p.WriteInt16(destPos.Y())

	return p
}

func packetMapRemoveMysticDoor(spawnID int32, instant bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelRemoveDoor)
	p.WriteBool(instant)
	p.WriteInt32(spawnID)

	return p
}

func packetMapBoat(show bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelBoat)
	if show {
		p.WriteInt16(0x01)
	} else {
		p.WriteInt16(0x02)
	}

	return p
}

func packetShowMovingObject(docked bool) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMovingObj)

	p.WriteByte(0x0a)

	if docked {
		p.WriteByte(4)
	} else {
		p.WriteByte(5)
	}

	return p
}

func packetShowEffect(path string) mpacket.Packet {
	return packetEnvironmentChange(3, path)
}

func packetPlaySound(path string) mpacket.Packet {
	return packetEnvironmentChange(4, path)
}

func packetBgmChange(path string) mpacket.Packet {
	return packetEnvironmentChange(6, path)
}

func packetEnvironmentChange(setting int32, value string) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelMapEffect)
	p.WriteInt32(setting)
	p.WriteString(value)
	return p
}

// func packetMapPortal(srcMap, dstmap int32, pos pos.Data) mpacket.Packet {
// 	p := mpacket.CreateWithOpcode(0x2d)
// 	p.WriteByte(26)
// 	p.WriteByte(0) // ?
// 	p.WriteInt32(srcMap)
// 	p.WriteInt32(dstmap)
// 	p.WriteInt16(pos.X())
// 	p.WriteInt16(pos.Y())

// 	return p
// }
