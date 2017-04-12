package packet

import "fmt"

type Packet []byte

func NewPacket(size int) Packet {
	return make(Packet, size)
}

func (p *Packet) Append(data []byte) {
	*p = append(*p, data...)
}

func (p *Packet) AddSize() {
	size := NewPacket(0)
	size.WriteShort(uint16(len(*p)))
	size.Append(*p)
	*p = size
}

func (p *Packet) Size() int {
	return int(len(*p))
}

func (p Packet) String() string {
	return fmt.Sprintf("[Packet] (%d) : % X", len(p), string(p))
}

//////////////////////////////////////////////////////////
// Maplestory only uses the following types in its packets
//////////////////////////////////////////////////////////
/*
Byte - 1 byte.
Short - 2 bytes.
Int (Integer) - 4 bytes.
Long - 8 bytes.
String - 2 bytes (denoting the length of the string) + length of the string in bytes
*/

func (p *Packet) WriteByte(data byte) {
	*p = append(*p, data)
}

func (p *Packet) WriteShort(data uint16) {
	*p = append(*p, byte(data), byte(data>>8))
}

func (p *Packet) WriteInt(data uint32) {
	*p = append(*p, byte(data), byte(data>>8), byte(data>>16), byte(data>>24))
}

func (p *Packet) WriteLong(data uint64) {
	*p = append(*p, byte(data), byte(data>>8), byte(data>>16), byte(data>>24),
		byte(data>>32), byte(data>>40), byte(data>>48), byte(data>>56))
}

func (p *Packet) WriteBuffer(data []byte) {
	p.WriteShort(uint16(len(data)))
	p.Append(data)
}

func (p *Packet) WriteString(str string) {
	p.WriteBuffer([]byte(str))
}

// Signed wrappers
func (p *Packet) WriteShortS(data int16) { p.WriteShort(uint16(data)) }
func (p *Packet) WriteIntS(data int32)   { p.WriteInt(uint32(data)) }
func (p *Packet) WriteLongS(data int64)  { p.WriteLong(uint64(data)) }

func (p *Packet) ReadByte() {

}

func (p *Packet) ReadShort() uint16 {
	return 0
}

func (p *Packet) ReadInt() {

}

func (p *Packet) ReadLong() {

}

func (p *Packet) ReadString() {

}

type PacketIterator struct {
}
