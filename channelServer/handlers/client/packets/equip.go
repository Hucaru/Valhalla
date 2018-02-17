package packets

import (
	"math"

	"github.com/Hucaru/Valhalla/common/character"
	"github.com/Hucaru/gopacket"
)

func addEquip(item character.Equip) gopacket.Packet {
	p := gopacket.NewPacket()

	if item.SlotID < -100 {
		p.WriteByte(byte(math.Abs(float64(item.SlotID + 100))))
	} else {
		p.WriteByte(byte(math.Abs(float64(item.SlotID))))
	}
	p.WriteByte(byte(item.ItemID / 1000000))
	p.WriteUint32(item.ItemID)

	if item.SlotID < -100 { // TODO:need to replce this with a isCashCheck once nx can be read
		p.WriteByte(1)                     // is cash item
		p.WriteUint64(uint64(item.ItemID)) // ? some form of id this seems to work
	} else {
		p.WriteByte(0) // not cash item
	}
	p.WriteUint64(item.ExpireTime)
	p.WriteByte(item.UpgradeSlots)
	p.WriteByte(item.Level)
	p.WriteUint16(item.Str)
	p.WriteUint16(item.Dex)
	p.WriteUint16(item.Intt)
	p.WriteUint16(item.Luk)
	p.WriteUint16(item.HP)
	p.WriteUint16(item.MP)
	p.WriteUint16(item.Watk)
	p.WriteUint16(item.Matk)
	p.WriteUint16(item.Wdef)
	p.WriteUint16(item.Mdef)
	p.WriteUint16(item.Accuracy)
	p.WriteUint16(item.Avoid)
	p.WriteUint16(item.Hands)
	p.WriteUint16(item.Speed)
	p.WriteUint16(item.Jump)
	p.WriteString(item.OwnerName) // Name of creator
	p.WriteInt16(2)               // lock, show, spikes, cape, cold protection etc ?
	return p
}
