package nx

import (
	"encoding/binary"
	"os"
)

type header struct {
	Magic                   [4]byte
	NodeCount               uint32
	NodeBlockOffset         int64
	StringCount             uint32
	StringOffsetTableOffset int64
	BitmapCount             uint32
	BitmapOffsetTableOffset int64
	AudioCount              uint32
	AudioOffsetTableOffset  int64
}

type node struct {
	ID         uint32
	NameID     uint32
	ChildID    uint32
	ChildCount uint32
	Type       uint16
	Data       uint64
}

type text struct {
	ID   uint32
	Text string
}

type equip struct{}
type useable struct{}
type setup struct{}
type etc struct{}
type cash struct{}
type skill struct{}
type npc struct{}
type field struct{} // map is langauge keyword

func Parse(fname string) {
	f, err := os.Open(fname)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	head := header{}
	binary.Read(f, binary.LittleEndian, &head)

	if head.Magic != [4]byte{0x50, 0x4B, 0x47, 0x34} {
		panic("Not valid nx magic number")
	}

	_, err = f.Seek(head.NodeBlockOffset, 0)

	if err != nil {
		panic(err)
	}

	var stringOffsets []int64

	nodeMemoryMap := make([]byte, 20*head.NodeCount)
	_, err = f.Read(nodeMemoryMap)
	if err != nil {
		panic(err)
	}

	_, err = f.Seek(head.StringOffsetTableOffset, 0)
	if err != nil {
		panic(err)
	}

	for i := uint32(0); i < head.StringCount; i++ {
		tmp := make([]byte, 8)
		_, err = f.Read(tmp)
		if err != nil {
			panic(err)
		}
		offset := int64(tmp[0]) | int64(tmp[1])<<8 | int64(tmp[2])<<16 | int64(tmp[3])<<24 | int64(tmp[4])<<32 | int64(tmp[5])<<40 | int64(tmp[6])<<48 | int64(tmp[7])<<56
		stringOffsets = append(stringOffsets, offset)
	}

	for _, v := range stringOffsets {
		_, err = f.Seek(v, 0)

		if err != nil {
			panic(err)
		}
		length := make([]byte, 2)
		f.Read(length)
		str := make([]byte, uint16(length[0])|uint16(length[1])<<8)
		_, err = f.Read(str)
		if err != nil {
			panic(err)
		}
	}
}
