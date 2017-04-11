package common

type Packet []byte

func NewPacket() *Packet {
	return &Packet{}
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

func (p *Packet) WriteByte() {

}

func (p *Packet) WriteShort() {

}

func (p *Packet) WriteInt() {

}

func (p *Packet) WriteLong() {

}

func (p *Packet) WriteString() {

}

func (p *Packet) ReadByte() {

}

func (p *Packet) ReadShort() {

}

func (p *Packet) ReadInt() {

}

func (p *Packet) ReadLong() {

}

func (p *Packet) ReadString() {

}
