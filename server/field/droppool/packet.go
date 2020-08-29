package droppool

import (
	"github.com/Hucaru/Valhalla/constant/opcode"
	"github.com/Hucaru/Valhalla/mpacket"
	"github.com/Hucaru/Valhalla/server/pos"
)

func test(spawnType byte, finalPos pos.Data, dropFrom pos.Data, neverExpire bool, expireTimestamp int64) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelDrobEnterMap)
	p.WriteByte(spawnType) // 0 = disappears on land, 1 = normal drop, 2 = show drop, 3 = fade at top of drop
	p.WriteInt32(1)        // drop id

	// if data.mesos {
	// 	p.WriteByte(1)
	// } else {
	// 	p.WriteByte(0)
	// }
	p.WriteByte(0)

	p.WriteInt32(1332020)      // mesos amount, itemID
	p.WriteInt32(2)            // owner id - player
	p.WriteByte(2)             // drop type 0 = timeout for non owner, 1 = timeout for non-owner party, 2 = free for all, 3 = explosive free for all
	p.WriteInt16(finalPos.X()) // drop to x
	p.WriteInt16(finalPos.Y()) // drop to y
	p.WriteInt32(0)            // if drop type == 0 place owner id, otherwise 0

	if spawnType != 2 {
		p.WriteInt16(dropFrom.X())        // drop from x
		p.WriteInt16(dropFrom.Y())        // drop from y
		p.WriteInt16(dropFrom.Foothold()) // foothold
	}

	// if !data.mesos {
	if 1 == 1 {
		p.WriteByte(0)    // ?
		p.WriteByte(0x80) // constants to indicate it's for item
		p.WriteByte(0x05)

		if neverExpire {
			p.WriteInt32(400967355)
			p.WriteByte(2)
		} else {
			p.WriteInt32(int32(expireTimestamp-946681229830) / 1000 / 60)
			p.WriteByte(0)
		}
	}

	p.WriteByte(0) // pet pickup?

	return p
}

// PacketShowDrop to field
func PacketShowDrop(spawnType byte, drop drop) mpacket.Packet {
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

	if drop.dropType == 0 {
		p.WriteInt32(drop.ownerID)
	} else {
		p.WriteInt32(0)
	}

	if spawnType != 2 {
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
			p.WriteInt32(int32(drop.expireTime-946681229830) / 1000 / 60)
			p.WriteByte(0)
		}
	}

	p.WriteByte(0) // pet pickup?

	return p
}

// PacketRemoveDrop on field
func PacketRemoveDrop(instant bool, dropID int32) mpacket.Packet {
	p := mpacket.CreateWithOpcode(opcode.SendChannelDropExitMap)
	p.WriteBool(instant)
	p.WriteInt32(dropID)

	return p
}
