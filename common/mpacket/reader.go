package mpacket

// Reader -
type Reader struct {
	pos    int
	packet *Packet
	Time   int64
}

// NewReader -
func NewReader(p *Packet, time int64) Reader {
	return Reader{pos: 0, packet: p, Time: time}
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

func (r *Reader) Skip(ammount int) {
	if len(*r.packet)-(r.pos+ammount) >= 0 {
		r.pos += ammount
	}
}

// ReadByte -
func (r *Reader) ReadByte() byte {
	if len(*r.packet)-r.pos > 0 {
		return r.packet.readByte(&r.pos)
	}

	return 0
}

// ReadInt8 -
func (r *Reader) ReadInt8() int8 {
	if len(*r.packet)-r.pos > 0 {
		return r.packet.readInt8(&r.pos)
	}

	return 0
}

// ReadBool -
func (r *Reader) ReadBool() bool {
	if len(*r.packet)-r.pos > 0 {
		return r.packet.readBool(&r.pos)
	}

	return false
}

// ReadBytes -
func (r *Reader) ReadBytes(size int) []byte {
	if len(*r.packet)-r.pos >= size {
		return r.packet.readBytes(&r.pos, size)
	}

	return []byte{0}
}

// ReadInt16 -
func (r *Reader) ReadInt16() int16 {
	if len(*r.packet)-r.pos > 1 {
		return r.packet.readInt16(&r.pos)
	}

	return 0
}

// ReadInt32 -
func (r *Reader) ReadInt32() int32 {
	if len(*r.packet)-r.pos > 3 {
		return r.packet.readInt32(&r.pos)
	}

	return 0
}

// ReadInt64 -
func (r *Reader) ReadInt64() int64 {
	if len(*r.packet)-r.pos > 7 {
		return r.packet.readInt64(&r.pos)
	}

	return 0
}

// ReadUint16 -
func (r *Reader) ReadUint16() uint16 {
	if len(*r.packet)-r.pos > 1 {
		return r.packet.readUint16(&r.pos)
	}

	return 0
}

// ReadUint32 -
func (r *Reader) ReadUint32() uint32 {
	if len(*r.packet)-r.pos > 3 {
		return r.packet.readUint32(&r.pos)
	}

	return 0
}

// ReadUint64 -
func (r *Reader) ReadUint64() uint64 {
	if len(*r.packet)-r.pos > 7 {
		return r.packet.readUint64(&r.pos)
	}

	return 0
}

// ReadFloat32 -
func (r *Reader) ReadFloat32() float32 {
	if len(*r.packet)-r.pos > 3 {
		return r.packet.readFloat32(&r.pos)
	}

	return 0
}

// ReadString -
func (r *Reader) ReadString(size int16) string {
	if len(*r.packet)-r.pos >= int(size) {
		return r.packet.readString(&r.pos, int(size))
	}

	return ""
}
