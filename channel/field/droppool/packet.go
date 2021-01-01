package droppool

import (
	"github.com/Hucaru/Valhalla/common/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
)

func packetShowDrop(spawnType byte, drop drop) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelDrobEnterMap)
	p.WriteByte(spawnType) // 0 = disappears on land, 1 = normal drop, 2 = show drop, 3 = fade at top of drop
	p.WriteInt32(drop.ID)

	if drop.mesos > 0 {
		p.WriteByte(1)
		p.WriteInt32(drop.mesos)
	} else {
		p.WriteByte(0)
		p.WriteInt32(drop.item.ID())
	}

	p.WriteInt32(drop.ownerID)
	p.WriteByte(drop.dropType) // drop type 0 = timeout for non owner, 1 = timeout for non-owner party, 2 = free for all, 3 = explosive free for all
	p.WriteInt16(drop.finalPos.X())
	p.WriteInt16(drop.finalPos.Y())

	if drop.dropType == DropTimeoutNonOwner {
		p.WriteInt32(drop.ownerID)
	} else {
		p.WriteInt32(0)
	}

	if spawnType != SpawnShow {
		p.WriteInt16(drop.originPos.X())        // drop from x
		p.WriteInt16(drop.originPos.Y())        // drop from y
		p.WriteInt16(drop.originPos.Foothold()) // foothold
	}

	if drop.mesos == 0 {
		p.WriteByte(0)    // ?
		p.WriteByte(0x80) // constants to indicate it's for item
		p.WriteByte(0x05)

		if drop.neverExpire {
			p.WriteInt32(400967355)
			p.WriteByte(2)
		} else {
			p.WriteInt32(int32((drop.expireTime - 946681229830) / 1000 / 60)) // TODO: figure out what time this is for
			p.WriteByte(1)
		}
	}

	p.WriteByte(0) // Did player drop it, used by pet with equip?

	return p
}

// PacketRemoveDrop on field
func packetRemoveDrop(instant bool, dropID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelDropExitMap)
	p.WriteBool(instant) // 0 - fade away, 1 - instant, 2,3,5 - player id? , 4 - int16
	p.WriteInt32(dropID)

	return p
}
