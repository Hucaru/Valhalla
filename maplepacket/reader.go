package maplepacket

// Reader -
type Reader struct {
	pos    int
	packet *Packet
}

// NewReader -
func NewReader(p *Packet) Reader {
	return Reader{pos: 0, packet: p}
}

func (r Reader) String() string {
	return r.packet.String()
}

// GetBuffer -
func (r *Reader) GetBuffer() []byte {
	return *r.packet
}

func (r *Reader) GetRestAsBytes() []byte {
	return (*r.packet)[r.pos:]
}

// ReadByte -
func (r *Reader) ReadByte() byte {
	return r.packet.readByte(&r.pos)
}

// ReadBytes -
func (r *Reader) ReadBytes(size int) []byte {
	return r.packet.readBytes(&r.pos, size)
}

// ReadInt16 -
func (r *Reader) ReadInt16() int16 {
	return r.packet.readInt16(&r.pos)
}

// ReadInt32 -
func (r *Reader) ReadInt32() int32 {
	return r.packet.readInt32(&r.pos)
}

// ReadInt64 -
func (r *Reader) ReadInt64() int64 {
	return r.packet.readInt64(&r.pos)
}

// ReadUint16 -
func (r *Reader) ReadUint16() uint16 {
	return r.packet.readUint16(&r.pos)
}

// ReadUint32 -
func (r *Reader) ReadUint32() uint32 {
	return r.packet.readUint32(&r.pos)
}

// ReadUint64 -
func (r *Reader) ReadUint64() uint64 {
	return r.packet.readUint64(&r.pos)
}

// ReadString -
func (r *Reader) ReadString(size int) string {
	return r.packet.readString(&r.pos, size)
}
