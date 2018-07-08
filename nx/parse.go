package nx

import (
	"encoding/binary"
	"os"
	"strings"
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
	NameID     uint32
	ChildID    uint32
	ChildCount uint16
	Type       uint16
	Data       [8]byte
}

var strLookup []string
var nodes []node

func Parse(fname string) {
	f, err := os.Open(fname)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	head := header{}
	err = binary.Read(f, binary.LittleEndian, &head)

	if err != nil {
		panic(err)
	}

	if head.Magic != [4]byte{0x50, 0x4B, 0x47, 0x34} {
		panic("Not valid nx magic number")
	}

	readStrings(f, head)
	readNodes(f, head)

	getMapInfo()
	getEquipInfo()
	getItemInfo()
	getMobInfo()
}

func searchNode(search string, fnc func(*node)) bool {
	cursor := &nodes[0]

	path := strings.Split(search, "/")

	if strings.Compare(path[0], "/") == 0 {
		path = path[1:]
	}

	for j, p := range path {
		for i := uint32(0); i < uint32(cursor.ChildCount); i++ {

			if cursor.ChildID+i > uint32(len(nodes))-1 {
				return false
			}

			if strings.Compare(strLookup[nodes[cursor.ChildID+i].NameID], p) == 0 {
				cursor = &nodes[cursor.ChildID+i]

				if j == len(path)-1 {
					fnc(cursor)
					return true
				}

				break
			}
		}
	}

	return false
}

func readNodes(f *os.File, head header) {
	_, err := f.Seek(head.NodeBlockOffset, 0)

	if err != nil {
		panic(err)
	}

	nodes = make([]node, head.NodeCount)
	err = binary.Read(f, binary.LittleEndian, &nodes)

	if err != nil {
		panic(err)
	}
}

func readStrings(f *os.File, head header) {
	_, err := f.Seek(head.StringOffsetTableOffset, 0)

	if err != nil {
		panic(err)
	}

	stringOffsets := make([]int64, head.StringCount)
	err = binary.Read(f, binary.LittleEndian, &stringOffsets)

	if err != nil {
		panic(err)
	}

	strLookup = make([]string, head.StringCount)

	for i, v := range stringOffsets {
		_, err = f.Seek(v, 0)

		if err != nil {
			panic(err)
		}

		var length uint16
		err = binary.Read(f, binary.LittleEndian, &length)

		if err != nil {
			panic(err)
		}

		str := make([]byte, length)
		_, err = f.Read(str)

		if err != nil {
			panic(err)
		}

		strLookup[i] = string(str)
	}
}

func dataToInt64(data [8]byte) int64 {
	return int64(data[0]) |
		int64(data[1])<<8 |
		int64(data[2])<<16 |
		int64(data[3])<<24 |
		int64(data[4])<<32 |
		int64(data[5])<<40 |
		int64(data[6])<<48 |
		int64(data[7])<<56
}

func dataToUint64(data [8]byte) uint64 {
	return uint64(data[0]) |
		uint64(data[1])<<8 |
		uint64(data[2])<<16 |
		uint64(data[3])<<24 |
		uint64(data[4])<<32 |
		uint64(data[5])<<40 |
		uint64(data[6])<<48 |
		uint64(data[7])<<56
}

func dataToUint32(data [8]byte) uint32 {
	return uint32(data[0]) |
		uint32(data[1])<<8 |
		uint32(data[2])<<16 |
		uint32(data[3])<<24
}

func dataToInt32(data [8]byte) int32 {
	return int32(data[0]) |
		int32(data[1])<<8 |
		int32(data[2])<<16 |
		int32(data[3])<<24
}

func dataToInt16(data [8]byte) int16 {
	return int16(data[0]) |
		int16(data[1])<<8
}

func dataToUint16(data [8]byte) uint16 {
	return uint16(data[0]) |
		uint16(data[1])<<8
}
