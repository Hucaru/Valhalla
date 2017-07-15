package packet

import "fmt"

// Packet -
type Packet []byte

// NewPacket -
func NewPacket() Packet {
	return make(Packet, 0)
}

// Append -
func (p *Packet) Append(data []byte) {
	*p = append(*p, data...)
}

// Size -
func (p *Packet) Size() int {
	return int(len(*p))
}

// String -
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

// WriteByte -
func (p *Packet) WriteByte(data byte) {
	*p = append(*p, data)
}

// WriteShort -
func (p *Packet) WriteShort(data uint16) {
	*p = append(*p, byte(data), byte(data>>8))
}

// WriteInt -
func (p *Packet) WriteInt(data uint32) {
	*p = append(*p, byte(data), byte(data>>8), byte(data>>16), byte(data>>24))
}

// WriteLong -
func (p *Packet) WriteLong(data uint64) {
	*p = append(*p, byte(data), byte(data>>8), byte(data>>16), byte(data>>24),
		byte(data>>32), byte(data>>40), byte(data>>48), byte(data>>56))
}

// WriteBuffer -
func (p *Packet) WriteBuffer(data []byte) {
	p.WriteShort(uint16(len(data)))
	p.Append(data)
}

// WriteString -
func (p *Packet) WriteString(str string) {
	p.WriteBuffer([]byte(str))
}

// WriteShortS -
func (p *Packet) WriteShortS(data int16) { p.WriteShort(uint16(data)) }

// WriteIntS -
func (p *Packet) WriteIntS(data int32) { p.WriteInt(uint32(data)) }

// WriteLongS -
func (p *Packet) WriteLongS(data int64) { p.WriteLong(uint64(data)) }

// ReadByte -
func (p *Packet) ReadByte(pos *int) byte {
	r := byte((*p)[*pos])
	*pos++
	return r
}

// ReadBytes -
func (p *Packet) ReadBytes(pos *int, length int) []byte {
	r := []byte((*p)[*pos : *pos+length])
	*pos += length
	return r
}

// ReadShort -
func (p *Packet) ReadShort(pos *int) int {
	r := p.ReadByte(pos) | (p.ReadByte(pos) << 8)
	return int(r)
}

// ReadInt -
func (p *Packet) ReadInt(pos *int) int {
	r := p.ReadByte(pos) | (p.ReadByte(pos) << 8) | (p.ReadByte(pos) << 16) | (p.ReadByte(pos) << 24)
	return int(r)
}

// ReadLong -
func (p *Packet) ReadLong(pos *int) int64 {
	return 0
}

// ReadString -
func (p *Packet) ReadString(pos *int, length int) string {
	return string(p.ReadBytes(pos, length))
}
