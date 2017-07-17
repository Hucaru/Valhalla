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

// WriteByte -
func (p *Packet) WriteByte(data byte) {
	*p = append(*p, data)
}

// WriteUint16 -
func (p *Packet) WriteUint16(data uint16) {
	*p = append(*p, byte(data), byte(data>>8))
}

// WriteUin16 -
func (p *Packet) WriteUint32(data uint32) {
	*p = append(*p, byte(data), byte(data>>8), byte(data>>16), byte(data>>24))
}

// WriteUint64 -
func (p *Packet) WriteUint64(data uint64) {
	*p = append(*p, byte(data), byte(data>>8), byte(data>>16), byte(data>>24),
		byte(data>>32), byte(data>>40), byte(data>>48), byte(data>>56))
}

// WriteBytes -
func (p *Packet) WriteBytes(data []byte) {
	p.Append(data)
}

// WriteString -
func (p *Packet) WriteString(str string) {
	p.WriteUint16(uint16(len(str)))
	p.WriteBytes([]byte(str))
}

// WriteInt16 -
func (p *Packet) WriteInt16(data int16) { p.WriteUint16(uint16(data)) }

// WriteInt32 -
func (p *Packet) WriteInt32(data int32) { p.WriteUint32(uint32(data)) }

// WriteInt64 -
func (p *Packet) WriteInt64(data int64) { p.WriteUint64(uint64(data)) }

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

// ReadInt16 -
func (p *Packet) ReadInt16(pos *int) int16 {
	r := p.ReadByte(pos) | (p.ReadByte(pos) << 8)
	return int16(r)
}

// ReadInt32 -
func (p *Packet) ReadInt32(pos *int) int32 {
	r := p.ReadByte(pos) | (p.ReadByte(pos) << 8) | (p.ReadByte(pos) << 16) | (p.ReadByte(pos) << 24)
	return int32(r)
}

// ReadLong -
func (p *Packet) ReadInt64(pos *int) int64 {
	return 0
}

// ReadString -
func (p *Packet) ReadString(pos *int, length int) string {
	return string(p.ReadBytes(pos, length))
}
